package io

import (
	"encoding/json"
	"fmt"
	"github.com/omaskery/teffy/pkg/events"
	"io"
	"strconv"
	"strings"
)

func WriteJsonObject(w io.Writer, data TefData) error {
	jsonFile := jsonObjectFile{
		TraceEvents:            make([]json.RawMessage, 0, len(data.Events())),
		DisplayTimeUnit:        string(data.DisplayTimeUnit()),
		StackFrames:            make(map[string]*stackFrame),
		SystemTraceEvents:      data.SystemTraceEvents(),
		PowerTraceAsString:     data.PowerTraceAsString(),
		ControllerTraceDataKey: data.ControllerTraceDataKey(),
		Metadata:               data.Metadata(),
	}

	for id, frame := range data.StackFrames() {
		jsonFile.StackFrames[id] = &stackFrame{
			Category: frame.Category,
			Name:     frame.Name,
			Parent:   frame.Parent,
		}
	}

	for _, event := range data.Events() {
		e := &jsonEvent{
			Name:            event.Core().Name,
			Phase:           string(event.Phase()),
			Categories:      strings.Join(event.Core().Categories, ","),
			Timestamp:       event.Core().Timestamp,
			ThreadTimestamp: event.Core().ThreadTimestamp,
			ProcessID:       event.Core().ProcessID,
			ThreadID:        event.Core().ThreadID,
			Args:            nil,
			Extra:           nil,
		}

		err := writeJsonEvent(event, e)
		if err != nil {
			return fmt.Errorf("failed while augmenting json event: %w", err)
		}

		msg, err := json.Marshal(e)
		if err != nil {
			return fmt.Errorf("failed to serialise json event: %w", err)
		}

		jsonFile.TraceEvents = append(jsonFile.TraceEvents, msg)
	}

	encoder := json.NewEncoder(w)
	err := encoder.Encode(&jsonFile)
	if err != nil {
		return fmt.Errorf("failed to write JSON object file: %w", err)
	}

	return nil
}

func writeJsonEvent(event events.Event, out *jsonEvent) error {
	switch e := event.(type) {
	case *events.BeginDuration:
		copyArgs(e.Args, out)
		writeStack(e.StackTrace, out, "stack")
	case *events.EndDuration:
		copyArgs(e.Args, out)
		writeStack(e.StackTrace, out, "stack")

	case *events.Complete:
		copyArgs(e.Args, out)
		writeStack(e.StackTrace, out, "stack")
		writeStack(e.EndStackTrace, out, "estack")

	case *events.Instant:
		copyArgs(e.Args, out)
		out.Extra = setEntry(out.Extra, "s", string(e.Scope))
		writeStack(e.StackTrace, out, "stack")

	case *events.Counter:
		if len(e.Values) > 0 {
			out.Args = ensureDict(out.Args)
			for key, value := range e.Values {
				out.Args[key] = strconv.FormatInt(value, 10)
			}
		}

	case *events.AsyncBegin:
		copyArgs(e.Args, out)
	case *events.AsyncInstant:
		copyArgs(e.Args, out)
	case *events.AsyncEnd:
		copyArgs(e.Args, out)

	case *events.FlowStart:
		copyArgs(e.Args, out)
	case *events.FlowInstant:
		copyArgs(e.Args, out)
	case *events.FlowFinish:
		copyArgs(e.Args, out)
		if e.BindingPoint != events.BindingPointNext {
			out.Extra = setEntry(out.Extra, "bp", "e")
		}
	case *events.ObjectCreated:
	case *events.ObjectSnapshot:
		copyArgs(e.Args, out)
	case *events.ObjectDeleted:

	case *events.MetadataProcessName:
		out.Args = setEntry(out.Args, "name", e.ProcessName)
	case *events.MetadataProcessLabels:
		out.Args = setEntry(out.Args, "labels", e.Labels)
	case *events.MetadataProcessSortIndex:
		out.Args = setEntry(out.Args, "sort_index", strconv.FormatInt(e.SortIndex, 10))
	case *events.MetadataThreadName:
		out.Args = setEntry(out.Args, "name", e.ThreadName)
	case *events.MetadataThreadSortIndex:
		out.Args = setEntry(out.Args, "sort_index", strconv.FormatInt(e.SortIndex, 10))
	case *events.MetadataMisc:
		copyArgs(e.Args, out)

	case *events.GlobalMemoryDump:
		copyArgs(e.Args, out)
	case *events.ProcessMemoryDump:
		copyArgs(e.Args, out)

	case *events.Mark:
		copyArgs(e.Args, out)

	case *events.ClockSync:
		copyArgs(e.Args, out)
		if e.SyncId != "" {
			out.Args = setEntry(out.Args, "sync_id", e.SyncId)
		}
		if e.IssueTs != nil {
			out.Args = setEntry(out.Args, "issue_ts", *e.IssueTs)
		}

	case *events.ContextEnter:
		copyArgs(e.Args, out)
	case *events.ContextExit:
		copyArgs(e.Args, out)

	case *events.LinkIds:
		copyArgs(e.Args, out)
		e.Args = setEntry(e.Args, "linked_id", e.LinkedId)
	}
	return nil
}

func writeStack(trace *events.StackTrace, out *jsonEvent, key string) {
	entries := make([]string, 0, len(trace.Trace))
	for _, frame := range trace.Trace {
		entries = append(entries, frame.Name)
	}
	out.Extra = setEntry(out.Extra, key, entries)
}

func copyArgs(args map[string]interface{}, out *jsonEvent) {
	if args == nil {
		return
	}
	if len(args) < 1 {
		return
	}

	out.Args = ensureDict(out.Args)

	for k, v := range args {
		out.Args[k] = v
	}
}

func setEntry(current map[string]interface{}, k string, v interface{}) map[string]interface{} {
	current = ensureDict(current)
	current[k] = v
	return current
}

func ensureDict(current map[string]interface{}) map[string]interface{} {
	if current != nil {
		return current
	}
	return map[string]interface{}{}
}
