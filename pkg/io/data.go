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

func (td *TefData) SetDisplayTimeUnit(d DisplayTimeUnit) {
	td.displayTimeUnit = d
}

func (td *TefData) SetSystemTraceEvents(s string) {
	td.systemTraceEvents = s
}

func (td *TefData) SetPowerTraceString(s string) {
	td.powerTraceAsString = s
}

func (td *TefData) SetControllerTraceDataKey(s string) {
	td.controllerTraceDataKey = s
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

type jsonObjectFile struct {
	TraceEvents            []json.RawMessage      `json:"traceEvents"`
	DisplayTimeUnit        string                 `json:"displayTimeUnit,omitempty"`
	StackFrames            map[string]*stackFrame `json:"stackFrames,omitempty"`
	SystemTraceEvents      string                 `json:"systemTraceEvents,omitempty"`
	PowerTraceAsString     string                 `json:"powerTraceAsString,omitempty"`
	ControllerTraceDataKey string                 `json:"controllerTraceDataKey,omitempty"`
	Metadata               map[string]interface{} `json:"-"`
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

type jsonId2 struct {
	Local  string `json:"local,omitempty"`
	Global string `json:"global,omitempty"`
}

type jsonId struct {
	Id  string  `json:"id,omitempty"`
	Id2 jsonId2 `json:"id2,omitempty"`
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
