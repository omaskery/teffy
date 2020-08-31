package io

import (
	"encoding/json"
	"github.com/omaskery/teffy/pkg/events"
)

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

func (td *TefData) Write(e events.Event) {
	td.traceEvents = append(td.traceEvents, e)
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

type stackFrame struct {
	Category string `json:"category"`
	Name     string `json:"name"`
	Parent   string `json:"parent,omitempty"`
}

type jsonEvent struct {
	Name            string                 `json:"name"`
	Phase           string                 `json:"ph"`
	Categories      string                 `json:"cat,omitempty"`
	Timestamp       int64                  `json:"ts"`
	ThreadTimestamp *int64                 `json:"tts,omitempty"`
	ProcessID       *int64                 `json:"pid,omitempty"`
	ThreadID        *int64                 `json:"tid,omitempty"`
	Args            map[string]interface{} `json:"args,omitempty"`
	Extra           map[string]interface{} `json:"-"`
}

type jsonObjectFile struct {
	TraceEvents            []json.RawMessage      `json:"traceEvents"`
	DisplayTimeUnit        string                 `json:"displayTimeUnit,omitempty"`
	StackFrames            map[string]*stackFrame `json:"stackFrames,omitempty"`
	SystemTraceEvents      string                 `json:"systemTraceEvents,omitempty"`
	PowerTraceAsString     string                 `json:"powerTraceAsString,omitempty"`
	ControllerTraceDataKey string                 `json:"controllerTraceDataKey,omitempty"`
	Metadata               map[string]interface{} `json:"-"`
}
