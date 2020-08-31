package io

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/omaskery/teffy/pkg/events"
	"io"
	"strconv"
	"strings"
)

var (
	ErrInvalidDisplayTimeUnit = errors.New("invalid display time unit")
	ErrRawStackNotStrArray    = errors.New("raw stack trace is expected to be a string array")
	ErrInvalidStackId         = errors.New("stack frame ids must be a string or integer")
	ErrStackIdNotFound        = errors.New("stack frame id not found in known stack frames")
	ErrInvalidDataType        = errors.New("data found in file does not match expected type")
	ErrSyntaxError            = errors.New("file format contained a syntax error")
)

func ParseJsonArray(r io.Reader) (*TefData, error) {
	decoder := json.NewDecoder(r)

	t, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to parse first token: %w", err)
	}
	if t != json.Delim('[') {
		return nil, fmt.Errorf("expected '[' at start of json array format: %w", ErrSyntaxError)
	}

	result := &TefData{
		displayTimeUnit:        DisplayTimeMs,
		metadata:               map[string]interface{}{},
		stackFrames:            map[string]*events.StackFrame{},
		controllerTraceDataKey: "traceEvents",
	}

	postprocessing := make([]postProcessStep, 0)

	for decoder.More() {
		var e json.RawMessage
		err = decoder.Decode(&e)
		if err != nil && errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error parsing JSON: %w", err)
		}

		event, err := parseJsonEvent(e, func(step postProcessStep) {
			postprocessing = append(postprocessing, step)
		})
		if err != nil {
			return nil, fmt.Errorf("error parsing event: %w", err)
		}

		result.traceEvents = append(result.traceEvents, event)
	}

	for _, step := range postprocessing {
		err = step.Process(result)
		if err != nil {
			return nil, fmt.Errorf("error performing postprocessing step: %w", err)
		}
	}

	return result, nil
}

func ParseJsonObj(r io.Reader) (*TefData, error) {
	var jsonFile jsonObjectFile
	decoder := json.NewDecoder(r)
	err := decoder.Decode(&jsonFile)
	if err != nil {
		return nil, fmt.Errorf("JSON decode error while parsing: %w", err)
	}

	result := &TefData{
		displayTimeUnit:        DisplayTimeMs,
		metadata:               map[string]interface{}{},
		stackFrames:            map[string]*events.StackFrame{},
		controllerTraceDataKey: "traceEvents",
	}

	switch jsonFile.DisplayTimeUnit {
	case "":
		fallthrough
	case string(DisplayTimeMs):
		result.displayTimeUnit = DisplayTimeMs
	case string(DisplayTimeNs):
		result.displayTimeUnit = DisplayTimeNs
	default:
		return nil, ErrInvalidDisplayTimeUnit
	}

	result.powerTraceAsString = jsonFile.PowerTraceAsString
	result.systemTraceEvents = jsonFile.SystemTraceEvents
	if jsonFile.ControllerTraceDataKey != "" {
		result.controllerTraceDataKey = jsonFile.ControllerTraceDataKey
	}

	for id, f := range jsonFile.StackFrames {
		frame := &events.StackFrame{
			Category: f.Category,
			Name:     f.Name,
			Parent:   f.Parent,
		}
		result.stackFrames[id] = frame
	}

	postprocessing := make([]postProcessStep, 0)

	for _, e := range jsonFile.TraceEvents {
		event, err := parseJsonEvent(e, func(step postProcessStep) {
			postprocessing = append(postprocessing, step)
		})
		if err != nil {
			return nil, fmt.Errorf("error parsing event: %w", err)
		}
		result.traceEvents = append(result.traceEvents, event)
	}

	for _, step := range postprocessing {
		err = step.Process(result)
		if err != nil {
			return nil, fmt.Errorf("error performing postprocessing step: %w", err)
		}
	}

	return result, nil
}

func parseJsonEvent(rawEvent json.RawMessage, postProcess func(step postProcessStep)) (events.Event, error) {

	var e jsonEvent
	err := json.Unmarshal(rawEvent, &e)
	if err != nil {
		return nil, fmt.Errorf("error parsing event JSON: %w", err)
	}

	err = json.Unmarshal(rawEvent, &e.Extra)
	if err != nil {
		return nil, fmt.Errorf("error while parsing extra event JSON: %w", err)
	}

	categories := make([]string, 0)
	if e.Categories != "" {
		categories = strings.Split(e.Categories, ",")
	}
	base := events.EventCore{
		Name:            e.Name,
		Categories:      categories,
		Timestamp:       e.Timestamp,
		ThreadTimestamp: e.ThreadTimestamp,
		ProcessID:       e.ProcessID,
		ThreadID:        e.ThreadID,
	}

	var event events.Event
	switch events.Phase(e.Phase) {
	case events.PhaseBeginDuration:
		bd := &events.BeginDuration{
			EventWithArgs: withArgs(base, e),
			StackTrace:    nil,
		}
		bd.StackTrace, err = parseRawStackTrace(e, "stack")
		if err != nil {
			return nil, fmt.Errorf("error while parsing raw stack trace: %w", err)
		}
		if stackRef, ok := e.Extra["sf"]; ok {
			target := &events.StackTrace{}
			postProcess(&buildStackTrace{
				reference: stackRef,
				target:    target,
			})
			bd.StackTrace = target
		}
		event = bd

	case events.PhaseComplete:
		c := &events.Complete{
			EventWithArgs: withArgs(base, e),
			StackTrace:    nil,
			EndStackTrace: nil,
		}
		c.StackTrace, err = parseRawStackTrace(e, "stack")
		if err != nil {
			return nil, fmt.Errorf("error while parsing raw stack trace: %w", err)
		}
		if stackRef, ok := e.Extra["sf"]; ok {
			target := &events.StackTrace{}
			postProcess(&buildStackTrace{
				reference: stackRef,
				target:    target,
			})
			c.StackTrace = target
		}
		c.EndStackTrace, err = parseRawStackTrace(e, "estack")
		if err != nil {
			return nil, fmt.Errorf("error while parsing raw stack trace: %w", err)
		}
		if stackRef, ok := e.Extra["esf"]; ok {
			target := &events.StackTrace{}
			postProcess(&buildStackTrace{
				reference: stackRef,
				target:    target,
			})
			c.EndStackTrace = target
		}

	case events.PhaseInstant:
		scope := events.InstantScopeThread
		if scopeVal, ok := e.Extra["s"]; ok {
			s, err := expectStr(scopeVal)
			if err != nil {
				return nil, fmt.Errorf("error parsing instant event scope: %w", err)
			}
			scope = events.InstantScope(s)
		}
		i := &events.Instant{
			EventWithArgs: withArgs(base, e),
			Scope:         scope,
			StackTrace:    nil,
		}
		i.StackTrace, err = parseRawStackTrace(e, "stack")
		if err != nil {
			return nil, fmt.Errorf("error while parsing raw stack trace: %w", err)
		}
		if stackRef, ok := e.Extra["sf"]; ok {
			target := &events.StackTrace{}
			postProcess(&buildStackTrace{
				reference: stackRef,
				target:    target,
			})
			i.StackTrace = target
		}

	case "S": // deprecated async start
		event = &events.AsyncBegin{
			EventWithArgs: withArgs(base, e),
		}

	case "T": // deprecated async step into
		event = &events.AsyncInstant{
			EventWithArgs: withArgs(base, e),
		}

	case "p": // deprecated async step past
		event = &events.AsyncInstant{
			EventWithArgs: withArgs(base, e),
		}

	case "F": // deprecated async finish
		event = &events.AsyncEnd{
			EventWithArgs: withArgs(base, e),
		}

	case events.PhaseAsyncBegin:
		event = &events.AsyncBegin{
			EventWithArgs: withArgs(base, e),
		}

	case events.PhaseAsyncInstant:
		event = &events.AsyncInstant{
			EventWithArgs: withArgs(base, e),
		}

	case events.PhaseAsyncEnd:
		event = &events.AsyncEnd{
			EventWithArgs: withArgs(base, e),
		}

	case events.PhaseObjectCreated:
		event = &events.ObjectCreated{
			EventCore: base,
		}

	case events.PhaseObjectSnapshot:
		event = &events.ObjectSnapshot{
			EventWithArgs: withArgs(base, e),
		}

	case events.PhaseObjectDeleted:
		event = &events.ObjectDeleted{
			EventCore: base,
		}

	case events.PhaseMetadata:
		switch events.MetadataKind(e.Name) {
		case events.MetadataKindProcessName:
			name, err := expectStr(e.Args["name"])
			if err != nil {
				return nil, fmt.Errorf("failed to get process name metadata: %w", err)
			}
			event = &events.MetadataProcessName{
				EventCore:   base,
				ProcessName: name,
			}
		case events.MetadataKindProcessLabels:
			labels, err := expectStr(e.Args["labels"])
			if err != nil {
				return nil, fmt.Errorf("failed to get process labels metadata: %w", err)
			}
			event = &events.MetadataProcessLabels{
				EventCore: base,
				Labels:    labels,
			}
		case events.MetadataKindProcessSortIndex:
			sortIndex, err := expectInt(e.Args["sort_index"])
			if err != nil {
				return nil, fmt.Errorf("failed to get process sort index metadata: %w", err)
			}
			event = &events.MetadataProcessSortIndex{
				EventCore: base,
				SortIndex: sortIndex,
			}
		case events.MetadataKindThreadName:
			name, err := expectStr(e.Args["name"])
			if err != nil {
				return nil, fmt.Errorf("failed to get thread name metadata: %w", err)
			}
			event = &events.MetadataThreadName{
				EventCore:  base,
				ThreadName: name,
			}
		case events.MetadataKindThreadSortIndex:
			sortIndex, err := expectInt(e.Args["sort_index"])
			if err != nil {
				return nil, fmt.Errorf("failed to get thread sort index metadata: %w", err)
			}
			event = &events.MetadataThreadSortIndex{
				EventCore: base,
				SortIndex: sortIndex,
			}
		default:
			event = &events.MetadataMisc{
				EventWithArgs: withArgs(base, e),
			}
		}

	case events.PhaseMark:
		event = &events.Mark{
			EventWithArgs: withArgs(base, e),
		}

	case events.PhaseContextEnter:
		event = &events.ContextEnter{
			EventWithArgs: withArgs(base, e),
		}

	case events.PhaseContextExit:
		event = &events.ContextExit{
			EventWithArgs: withArgs(base, e),
		}

	default:
		return nil, fmt.Errorf("unknown phase encountered: '%v'", e.Phase)
	}

	return event, nil
}

func withArgs(base events.EventCore, e jsonEvent) events.EventWithArgs {
	return events.EventWithArgs{
		EventCore: base,
		Args:      e.Args,
	}
}

func expectInt(v interface{}) (int64, error) {
	if s, ok := v.(float64); ok {
		return int64(s), nil
	}
	return 0, fmt.Errorf("expected number, got '%v': %w", v, ErrInvalidDataType)
}

func expectStr(v interface{}) (string, error) {
	if s, ok := v.(string); ok {
		return s, nil
	}
	return "", fmt.Errorf("expected string, got '%v': %w", v, ErrInvalidDataType)
}

type postProcessStep interface {
	Process(data *TefData) error
}

type buildStackTrace struct {
	reference interface{}
	target    *events.StackTrace
}

func (s buildStackTrace) Process(data *TefData) error {
	var stackRef string
	switch r := s.reference.(type) {
	case string:
		stackRef = r
	case int:
		stackRef = strconv.Itoa(r)
	default:
		return fmt.Errorf("invalid stack ref: %w", ErrInvalidStackId)
	}

	for {
		frame, ok := data.stackFrames[stackRef]
		if !ok {
			return fmt.Errorf("invalid stack ref '%s': %w", stackRef, ErrStackIdNotFound)
		}

		s.target.Trace = append([]*events.StackFrame{frame}, s.target.Trace...)
		if frame.Parent == "" {
			break
		}

		stackRef = frame.Parent
	}

	return nil
}

func parseRawStackTrace(event jsonEvent, key string) (*events.StackTrace, error) {
	stack, ok := event.Extra[key]
	if !ok {
		return nil, nil
	}

	stackEntries, ok := stack.([]interface{})
	if !ok {
		return nil, ErrRawStackNotStrArray
	}

	trace := &events.StackTrace{}
	for _, entry := range stackEntries {
		entryStr, ok := entry.(string)
		if !ok {
			return nil, ErrRawStackNotStrArray
		}
		trace.Trace = append(trace.Trace, &events.StackFrame{
			Name: entryStr,
		})
	}

	return trace, nil
}
