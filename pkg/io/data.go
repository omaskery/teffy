package io

import (
	"encoding/json"
	"strconv"

	"github.com/omaskery/teffy/pkg/events"
)

// DisplayTimeUnit indicates whether time should be displayed in nano or milliseconds
type DisplayTimeUnit string

const (
	DisplayTimeNs DisplayTimeUnit = "ns"
	DisplayTimeMs DisplayTimeUnit = "ms"
)

// TefData is an in-memory representation of a JSON Object Format variant of Trace Event Format file
type TefData struct {
	traceEvents            []events.Event
	displayTimeUnit        DisplayTimeUnit
	systemTraceEvents      string
	powerTraceAsString     string
	stackFrames            map[string]*events.StackFrame
	controllerTraceDataKey string
	metadata               map[string]interface{}
}

// Write records the given trace event
func (td *TefData) Write(e events.Event) {
	td.traceEvents = append(td.traceEvents, e)
}

// SetDisplayTimeUnit sets what units timestamps should be displayed in
func (td *TefData) SetDisplayTimeUnit(d DisplayTimeUnit) {
	td.displayTimeUnit = d
}

// SetSystemTraceEvents stores the provided system trace text
func (td *TefData) SetSystemTraceEvents(s string) {
	td.systemTraceEvents = s
}

// SetSystemTraceString stores the provided power trace string
func (td *TefData) SetPowerTraceString(s string) {
	td.powerTraceAsString = s
}

// SetControllerTraceDataKey records which key this tracing agent stores traces in
func (td *TefData) SetControllerTraceDataKey(s string) {
	td.controllerTraceDataKey = s
}

// SetStackFrame internally associates the given stack frame with the given id
func (td *TefData) SetStackFrame(id string, frame *events.StackFrame) {
	if td.stackFrames == nil {
		td.stackFrames = map[string]*events.StackFrame{}
	}
	td.stackFrames[id] = frame
}

// Events retrieves the events stored in the file
func (td TefData) Events() []events.Event {
	return td.traceEvents
}

// DisplayTimeUnit gets the desired units to display timestamps from this file
func (td TefData) DisplayTimeUnit() DisplayTimeUnit {
	return td.displayTimeUnit
}

// SystemTraceEvents retrieves the system trace string
func (td TefData) SystemTraceEvents() string {
	return td.systemTraceEvents
}

// PowerTraceAsString retrieves the power trace string
func (td TefData) PowerTraceAsString() string {
	return td.powerTraceAsString
}

// StackFrames retrieves the stack frames recorded in this file
func (td TefData) StackFrames() map[string]*events.StackFrame {
	return td.stackFrames
}

// ControllerTraceDataKey retrieves the key that trace events are stored under for this trace file
func (td TefData) ControllerTraceDataKey() string {
	return td.controllerTraceDataKey
}

// Metadata retrieves additional, non standard key values stored at the top level of this file
func (td TefData) Metadata() map[string]interface{} {
	return td.metadata
}

type stackFrame struct {
	Category string `json:"category"`
	Name     string `json:"name"`
	Parent   string `json:"parent,omitempty"`
}

type jsonObjectFile struct {
	TraceEvents            []json.RawMessage      `json:"traceEvents"`
	DisplayTimeUnit        string                 `json:"displayTimeUnit,omitempty"`
	StackFrames            map[string]*stackFrame `json:"stackFrames,omitempty"`
	SystemTraceEvents      string                 `json:"systemTraceEvents,omitempty"`
	PowerTraceAsString     string                 `json:"powerTraceAsString,omitempty"`
	ControllerTraceDataKey string                 `json:"controllerTraceDataKey,omitempty"`
	Metadata               map[string]interface{} `json:"otherData,omitempty"`
}

type jsonEventPhase struct {
	Phase string `json:"ph"`
}

type jsonEventCore struct {
	jsonEventPhase
	Name            string `json:"name"`
	Categories      string `json:"cat,omitempty"`
	Timestamp       int64  `json:"ts"`
	ThreadTimestamp *int64 `json:"tts,omitempty"`
	ProcessID       *int64 `json:"pid,omitempty"`
	ThreadID        *int64 `json:"tid,omitempty"`
}

type jsonEventWithArgs struct {
	jsonEventCore
	Args map[string]interface{} `json:"args,omitempty"`
}

type jsonStackInfo struct {
	Stack      []string `json:"stack,omitempty"`
	StackFrame string   `json:"sf,omitempty"`
}

type jsonDurationEvent struct {
	jsonEventWithArgs
	jsonStackInfo
}

type jsonCompleteEvent struct {
	jsonEventWithArgs
	jsonStackInfo
	Duration      int64    `json:"dur,omitempty"`
	EndStack      []string `json:"estack,omitempty"`
	EndStackFrame string   `json:"esf,omitempty"`
}

type jsonInstantEvent struct {
	jsonEventCore
	jsonStackInfo
	Scope string `json:"s,omitempty"`
}

type jsonCounterEvent struct {
	jsonEventCore
	Values map[string]float64 `json:"args,omitempty"`
}

type numberOrString struct {
	number float64
	str    string
}

func (nos *numberOrString) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &nos.str); err != nil {
		return json.Unmarshal(data, &nos.number)
	}

	return nil
}

type tempJsonCounterEvent struct {
	jsonEventCore
	Values map[string]numberOrString `json:"args,omitempty"`
}

func (ce *jsonCounterEvent) UnmarshalJSON(data []byte) error {
	t := &tempJsonCounterEvent{}
	if err := json.Unmarshal(data, t); err != nil {
		return err
	}
	ce.jsonEventCore = t.jsonEventCore
	ce.Values = make(map[string]float64)
	for k, numberOrStr := range t.Values {
		value := numberOrStr.number

		if numberOrStr.str != "" {
			f, err := strconv.ParseFloat(numberOrStr.str, 64)
			if err != nil {
				return err
			}
			value = f
		}

		ce.Values[k] = value
	}
	return nil
}

type jsonId2 struct {
	Local  string `json:"local,omitempty"`
	Global string `json:"global,omitempty"`
}

type jsonId struct {
	Id  string   `json:"id,omitempty"`
	Id2 *jsonId2 `json:"id2,omitempty"`
}

type jsonScopedId struct {
	jsonId
	Scope string `json:"scope,omitempty"`
}

type jsonAsyncEvent struct {
	jsonEventWithArgs
	jsonScopedId
}

type jsonObjectEvent struct {
	jsonEventWithArgs
	jsonScopedId
}

type jsonMetadataEvent struct {
	jsonEventWithArgs
}

type jsonMemoryDumpEvent struct {
	jsonEventWithArgs
}

type jsonMarkEvent struct {
	jsonEventWithArgs
}

type jsonClockSyncEvent struct {
	jsonEventWithArgs
}

type jsonContextEvent struct {
	jsonEventWithArgs
	jsonId
}

type jsonLinkedIdEvent struct {
	jsonEventWithArgs
	jsonId
}
