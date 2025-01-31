package winrm

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"sync"
)

type commandWriter struct {
	*Command
	mutex sync.Mutex
	eof   bool
}

type commandReader struct {
	*Command
	write  *io.PipeWriter
	read   *io.PipeReader
	stream string
}

// Command represents a given command running on a Shell. This structure allows to get access
// to the various stdout, stderr and stdin pipes.
type Command struct {
	client   *Client
	shell    *Shell
	id       string
	exitCode int
	err      error

	Stdin  *commandWriter
	Stdout *commandReader
	Stderr *commandReader

	done   chan struct{}
	cancel chan struct{}
	ctx    context.Context
}

func newCommand(ctx context.Context, shell *Shell, ids string) *Command {
	command := &Command{
		shell:    shell,
		client:   shell.client,
		id:       ids,
		exitCode: 0,
		err:      nil,
		done:     make(chan struct{}),
		cancel:   make(chan struct{}),

		ctx: ctx,
	}

	command.Stdout = newCommandReader("stdout", command)
	command.Stdin = &commandWriter{
		Command: command,
		eof:     false,
	}
	command.Stderr = newCommandReader("stderr", command)

	return command
}

func newCommandReader(stream string, command *Command) *commandReader {
	read, write := io.Pipe()
	return &commandReader{
		Command: command,
		stream:  stream,
		write:   write,
		read:    read,
	}
}

func fetchOutput(ctx context.Context, command *Command) {
	ctxDone := ctx.Done()
	for {
		select {
		case <-command.cancel:
			_, _ = command.slurpAllOutput()
			err := errors.New("canceled")
			command.Stderr.write.CloseWithError(err)
			command.Stdout.write.CloseWithError(err)
			close(command.done)
			return
		case <-ctxDone:
			command.err = ctx.Err()
			ctxDone = nil
			command.Close()
		default:
			finished, err := command.slurpAllOutput()
			if finished {
				command.err = err
				close(command.done)
				return
			}
		}
	}
}

func (c *Command) check() error {
	if c.id == "" {
		return errors.New("Command has already been closed")
	}
	if c.shell == nil {
		return errors.New("Command has no associated shell")
	}
	if c.client == nil {
		return errors.New("Command has no associated client")
	}
	return nil
}

// Close will terminate the running command
func (c *Command) Close() error {
	if err := c.check(); err != nil {
		return err
	}

	select { // close cancel channel if it's still open
	case <-c.cancel:
	default:
		close(c.cancel)
	}

	request := NewSignalRequest(c.client.url, c.shell.id, c.id, &c.client.Parameters)
	defer request.Free()

	_, err := c.client.sendRequest(request)
	return err
}

func (c *Command) slurpAllOutput() (bool, error) {
	if err := c.check(); err != nil {
		c.Stderr.write.CloseWithError(err)
		c.Stdout.write.CloseWithError(err)
		return true, err
	}

	request := NewGetOutputRequest(c.client.url, c.shell.id, c.id, "stdout stderr", &c.client.Parameters)
	defer request.Free()

	response, err := c.client.sendRequest(request)
	if err != nil {
		if strings.Contains(err.Error(), "OperationTimeout") ||
			strings.Contains(err.Error(), "timeout awaiting response headers") {
			// Operation timeout because there was no command output
			return false, err
		}
		if strings.Contains(err.Error(), "EOF") {
			c.exitCode = 16001
		}

		c.Stderr.write.CloseWithError(err)
		c.Stdout.write.CloseWithError(err)
		return true, err
	}

	var exitCode int
	var stdout, stderr bytes.Buffer
	finished, exitCode, err := ParseSlurpOutputErrResponse(response, &stdout, &stderr)
	if err != nil {
		c.Stderr.write.CloseWithError(err)
		c.Stdout.write.CloseWithError(err)
		return true, err
	}
	if stdout.Len() > 0 {
		_, _ = c.Stdout.write.Write(stdout.Bytes())
	}
	if stderr.Len() > 0 {
		_, _ = c.Stderr.write.Write(stderr.Bytes())
	}
	if finished {
		c.exitCode = exitCode
		_ = c.Stderr.write.Close()
		_ = c.Stdout.write.Close()
	}

	return finished, nil
}

func (c *Command) sendInput(data []byte, eof bool) error {
	if err := c.check(); err != nil {
		return err
	}

	request := NewSendInputRequest(c.client.url, c.shell.id, c.id, data, eof, &c.client.Parameters)
	defer request.Free()

	_, err := c.client.sendRequest(request)
	return err
}

// ExitCode returns command exit code when it is finished. Before that the result is always 0.
func (c *Command) ExitCode() int {
	return c.exitCode
}

// Error returns command error.
func (c *Command) Error() error {
	return c.err
}

func (c *Command) Result() (stdout []byte, stderr []byte, exitCode int, err error) {
	go fetchOutput(c.ctx, c)

	var outWriter, errWriter bytes.Buffer
	wg := sync.WaitGroup{}
	wg.Add(2)

	_ = c.Stdin.Close()
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&outWriter, c.Stdout)
	}()
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&errWriter, c.Stderr)
	}()

	wg.Wait()

	err = c.Wait()
	if err != nil {
		return
	}

	return outWriter.Bytes(), CleanupStderr(errWriter.Bytes()), c.exitCode, c.err
}

// Wait function will block the current goroutine until the remote command terminates.
func (c *Command) Wait() error {
	// block until finished
	<-c.done

	return c.err
}

// Write data to this Pipe
// commandWriter implements io.Writer and io.Closer interface
func (w *commandWriter) Write(data []byte) (int, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.eof {
		return 0, io.ErrClosedPipe
	}

	var (
		written int
		err     error
	)
	origLen := len(data)
	for len(data) > 0 {
		// never send more data than our EnvelopeSize.
		n := min(w.client.Parameters.EnvelopeSize-1000, len(data))
		if err := w.sendInput(data[:n], false); err != nil {
			break
		}
		data = data[n:]
		written += n
	}

	// signal that we couldn't write all data
	if err == nil && written < origLen {
		err = io.ErrShortWrite
	}

	return written, err
}

// Write data to this Pipe and mark EOF
func (w *commandWriter) WriteClose(data []byte) (int, error) {
	w.eof = true
	return w.Write(data)
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

// Close method wrapper
// commandWriter implements io.Closer interface
func (w *commandWriter) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.eof {
		return io.ErrClosedPipe
	}
	w.eof = true
	return w.sendInput(nil, w.eof)
}

// Read data from this Pipe
func (r *commandReader) Read(buf []byte) (int, error) {
	n, err := r.read.Read(buf)
	if err != nil && errors.Is(err, io.EOF) {
		return 0, err
	}
	return n, err
}
