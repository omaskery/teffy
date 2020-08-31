package io

import "github.com/omaskery/teffy/pkg/events"

type DisplayTimeUnit string

const (
	DisplayTimeNs DisplayTimeUnit = "ns"
	DisplayTimeMs DisplayTimeUnit = "ms"
)

type TefData struct {
	traceEvents            []events.Event
	displayTimeUnit        DisplayTimeUnit
	systemTraceEvents      string
	powerTraceAsString     string
	stackFrames            map[string]*events.StackFrame
	controllerTraceDataKey string
	metadata               map[string]interface{}
}

func (td TefData) Events() []events.Event {
	return td.traceEvents
}

func (td TefData) DisplayTimeUnit() DisplayTimeUnit {
	return td.displayTimeUnit
}

func (td TefData) SystemTraceEvents() string {
	return td.systemTraceEvents
}

func (td TefData) PowerTraceAsString() string {
	return td.powerTraceAsString
}

func (td TefData) StackFrames() map[string]*events.StackFrame {
	return td.stackFrames
}

func (td TefData) ControllerTraceDataKey() string {
	return td.controllerTraceDataKey
}

func (td TefData) Metadata() map[string]interface{} {
	return td.metadata
}
