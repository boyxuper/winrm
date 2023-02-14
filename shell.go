package winrm

import "context"

// Shell is the local view of a WinRM Shell of a given Client
type Shell struct {
	client *Client
	id     string
}

// Execute command on the given Shell, returning either an error or a Command
//
// Deprecated: user ExecuteWithContext
func (s *Shell) Execute(command string, arguments ...string) (*Command, error) {
	return s.ExecuteWithContext(context.Background(), command, arguments...)
}
func (s *Shell) Client() *Client {
	return s.client
}

// ExecuteWithContext command on the given Shell, returning either an error or a Command
func (s *Shell) ExecuteWithContext(ctx context.Context, command string, arguments ...string) (*Command, error) {
	request := NewExecuteCommandRequest(s.client.url, s.id, command, arguments, &s.client.Parameters)
	defer request.Free()

	response, err := s.client.sendRequest(request)
	if err != nil {
		return nil, err
	}

	commandID, err := ParseExecuteCommandResponse(response)
	if err != nil {
		return nil, err
	}

	cmd := newCommand(ctx, s, commandID)

	return cmd, nil
}

// ExecutePSWithContext powershell command on the given Shell, returning either an error or a Command
func (s *Shell) ExecutePSWithContext(ctx context.Context, command string) (*Command, error) {
	command = Powershell(command)
	request := NewExecuteCommandRequest(s.client.url, s.id, command, nil, &s.client.Parameters)
	defer request.Free()

	response, err := s.client.sendRequest(request)
	if err != nil {
		return nil, err
	}

	commandID, err := ParseExecuteCommandResponse(response)
	if err != nil {
		return nil, err
	}

	cmd := newCommand(ctx, s, commandID)

	return cmd, nil
}

// Close will terminate this shell. No commands can be issued once the shell is closed.
func (s *Shell) Close() error {
	request := NewDeleteShellRequest(s.client.url, s.id, &s.client.Parameters)
	defer request.Free()

	_, err := s.client.sendRequest(request)
	return err
}
