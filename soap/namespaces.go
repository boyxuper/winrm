package soap

import (
	"github.com/ChrisTrenkamp/goxpath"
	"github.com/masterzen/simplexml/dom"
)

// Namespaces
//nolint:stylecheck // we keep the ALL_CAPS names
const (
	NS_SOAP_ENV    = "http://www.w3.org/2003/05/soap-envelope"
	NS_ADDRESSING  = "http://schemas.xmlsoap.org/ws/2004/08/addressing"
	NS_CIMBINDING  = "http://schemas.dmtf.org/wbem/wsman/1/cimbinding.xsd"
	NS_ENUM        = "http://schemas.xmlsoap.org/ws/2004/09/enumeration"
	NS_TRANSFER    = "http://schemas.xmlsoap.org/ws/2004/09/transfer"
	NS_WSMAN_DMTF  = "http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd"
	NS_WSMAN_MSFT  = "http://schemas.microsoft.com/wbem/wsman/1/wsman.xsd"
	NS_SCHEMA_INST = "http://www.w3.org/2001/XMLSchema-instance"
	NS_WIN_SHELL   = "http://schemas.microsoft.com/wbem/wsman/1/windows/shell"
	NS_WSMAN_FAULT = "http://schemas.microsoft.com/wbem/wsman/1/wsmanfault"
	NS_POWERSHELL  = "http://schemas.microsoft.com/powershell/2004/04"
)

// Namespace Prefixes
//nolint:stylecheck // we keep the ALL_CAPS names
const (
	NSP_SOAP_ENV    = "env"
	NSP_ADDRESSING  = "a"
	NSP_CIMBINDING  = "b"
	NSP_ENUM        = "n"
	NSP_TRANSFER    = "x"
	NSP_WSMAN_DMTF  = "w"
	NSP_WSMAN_MSFT  = "p"
	NSP_SCHEMA_INST = "xsi"
	NSP_WIN_SHELL   = "rsp"
	NSP_WSMAN_FAULT = "f"
	NSP_POWERSHELL  = "ps"
)

// DOM Namespaces
//nolint:stylecheck
var (
	DOM_NS_SOAP_ENV    = dom.Namespace{Prefix: "env", Uri: "http://www.w3.org/2003/05/soap-envelope"}
	DOM_NS_ADDRESSING  = dom.Namespace{Prefix: "a", Uri: "http://schemas.xmlsoap.org/ws/2004/08/addressing"}
	DOM_NS_CIMBINDING  = dom.Namespace{Prefix: "b", Uri: "http://schemas.dmtf.org/wbem/wsman/1/cimbinding.xsd"}
	DOM_NS_ENUM        = dom.Namespace{Prefix: "n", Uri: "http://schemas.xmlsoap.org/ws/2004/09/enumeration"}
	DOM_NS_TRANSFER    = dom.Namespace{Prefix: "x", Uri: "http://schemas.xmlsoap.org/ws/2004/09/transfer"}
	DOM_NS_WSMAN_DMTF  = dom.Namespace{Prefix: "w", Uri: "http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd"}
	DOM_NS_WSMAN_MSFT  = dom.Namespace{Prefix: "p", Uri: "http://schemas.microsoft.com/wbem/wsman/1/wsman.xsd"}
	DOM_NS_SCHEMA_INST = dom.Namespace{Prefix: "xsi", Uri: "http://www.w3.org/2001/XMLSchema-instance"}
	DOM_NS_WIN_SHELL   = dom.Namespace{Prefix: "rsp", Uri: "http://schemas.microsoft.com/wbem/wsman/1/windows/shell"}
	DOM_NS_WSMAN_FAULT = dom.Namespace{Prefix: "f", Uri: "http://schemas.microsoft.com/wbem/wsman/1/wsmanfault"}
)

var MostUsed = [...]dom.Namespace{
	DOM_NS_SOAP_ENV,
	DOM_NS_ADDRESSING,
	DOM_NS_WIN_SHELL,
	DOM_NS_WSMAN_DMTF,
	DOM_NS_WSMAN_MSFT,
}

func AddUsualNamespaces(node *dom.Element) {
	for _, ns := range MostUsed {
		node.DeclareNamespace(ns)
	}
}

func GetAllXPathNamespaces() func(o *goxpath.Opts) {
	ns := map[string]string{
		NSP_SOAP_ENV:    NS_SOAP_ENV,
		NSP_ADDRESSING:  NS_ADDRESSING,
		NSP_CIMBINDING:  NS_CIMBINDING,
		NSP_ENUM:        NS_ENUM,
		NSP_TRANSFER:    NS_TRANSFER,
		NSP_WSMAN_DMTF:  NS_WSMAN_DMTF,
		NSP_WSMAN_MSFT:  NS_WSMAN_MSFT,
		NSP_SCHEMA_INST: NS_SCHEMA_INST,
		NSP_WIN_SHELL:   NS_WIN_SHELL,
		NSP_WSMAN_FAULT: NS_WSMAN_FAULT,
		NSP_POWERSHELL:  NS_POWERSHELL,
	}

	return func(o *goxpath.Opts) {
		o.NS = ns
	}
}
