package io

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/omaskery/teffy/pkg/events"
	"io"
	"strings"
)

var (
	ErrInvalidDisplayTimeUnit = errors.New("invalid display time unit")
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

	for decoder.More() {
		var e json.RawMessage
		err = decoder.Decode(&e)
		if err != nil && errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error parsing JSON: %w", err)
		}

		event, err := parseJsonEvent(e)
		if err != nil {
			return nil, fmt.Errorf("error parsing event: %w", err)
		}

		result.traceEvents = append(result.traceEvents, event)
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

	for _, e := range jsonFile.TraceEvents {
		event, err := parseJsonEvent(e)
		if err != nil {
			return nil, fmt.Errorf("error parsing event: %w", err)
		}
		result.traceEvents = append(result.traceEvents, event)
	}

	return result, nil
}

func parseJsonEvent(rawEvent json.RawMessage) (events.Event, error) {
	phase, err := decodeEventPhase(rawEvent)
	if err != nil {
		return nil, fmt.Errorf("error decoding json event: %w", err)
	}

	var event events.Event
	switch phase {
	case events.PhaseBeginDuration:
		var j jsonDurationEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode begin duration event: %w", err)
		}
		event = &events.BeginDuration{
			EventWithArgs: events.EventWithArgs{
				EventCore: decodeEventCore(j.jsonEventCore),
				Args:      j.Args,
			},
			EventStackTrace: events.EventStackTrace{
				StackTrace: decodeRawStackTrace(j.Stack),
			},
		}
	case events.PhaseEndDuration:
		var j jsonDurationEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode end duration event: %w", err)
		}
		event = &events.EndDuration{
			EventWithArgs: events.EventWithArgs{
				EventCore: decodeEventCore(j.jsonEventCore),
				Args:      j.Args,
			},
			EventStackTrace: events.EventStackTrace{
				StackTrace: decodeRawStackTrace(j.Stack),
			},
		}

	case events.PhaseComplete:
		var j jsonCompleteEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode complete event: %w", err)
		}
		event = &events.Complete{
			EventWithArgs: events.EventWithArgs{
				EventCore: decodeEventCore(j.jsonEventCore),
				Args:      j.Args,
			},
			EventStackTrace: events.EventStackTrace{
				StackTrace:    decodeRawStackTrace(j.Stack),
			},
			EventEndStackTrace: events.EventEndStackTrace{
				EndStackTrace: decodeRawStackTrace(j.EndStack),
			},
		}

	case events.PhaseInstant:
		var j jsonInstantEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode instant event: %w", err)
		}
		scope := events.InstantScope(j.Scope)
		if scope == "" {
			scope = events.InstantScopeGlobal
		}
		event = &events.Instant{
			EventCore:  decodeEventCore(j.jsonEventCore),
			EventStackTrace: events.EventStackTrace{
				StackTrace: decodeRawStackTrace(j.Stack),
			},
			Scope:      scope,
		}

	case events.PhaseCounter:
		var j jsonCounterEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode counter event: %w", err)
		}
		event = &events.Counter{
			EventCore: decodeEventCore(j.jsonEventCore),
			Values:    j.Values,
		}

	case "S": // deprecated async start
		var j jsonAsyncEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode (deprecated) async start event: %w", err)
		}
		event = &events.AsyncBegin{
			EventWithArgs: events.EventWithArgs{
				EventCore: decodeEventCore(j.jsonEventCore),
				Args:      j.Args,
			},
		}
	case "T": // deprecated async step into
		var j jsonAsyncEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode (deprecated) async step into event: %w", err)
		}
		event = &events.AsyncInstant{
			EventWithArgs: events.EventWithArgs{
				EventCore: decodeEventCore(j.jsonEventCore),
				Args:      j.Args,
			},
		}
	case "p": // deprecated async step past
		var j jsonAsyncEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode (deprecated) async step past event: %w", err)
		}
		event = &events.AsyncInstant{
			EventWithArgs: events.EventWithArgs{
				EventCore: decodeEventCore(j.jsonEventCore),
				Args:      j.Args,
			},
		}
	case "F": // deprecated async finish
		var j jsonAsyncEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode (deprecated) async finish event: %w", err)
		}
		event = &events.AsyncEnd{
			EventWithArgs: events.EventWithArgs{
				EventCore: decodeEventCore(j.jsonEventCore),
				Args:      j.Args,
			},
		}

	case events.PhaseAsyncBegin:
		var j jsonAsyncEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode async begin event: %w", err)
		}
		event = &events.AsyncBegin{
			EventWithArgs: events.EventWithArgs{
				EventCore: decodeEventCore(j.jsonEventCore),
				Args:      j.Args,
			},
		}
	case events.PhaseAsyncInstant:
		var j jsonAsyncEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode async instant event: %w", err)
		}
		event = &events.AsyncInstant{
			EventWithArgs: events.EventWithArgs{
				EventCore: decodeEventCore(j.jsonEventCore),
				Args:      j.Args,
			},
		}
	case events.PhaseAsyncEnd:
		var j jsonAsyncEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode async end event: %w", err)
		}
		event = &events.AsyncEnd{
			EventWithArgs: events.EventWithArgs{
				EventCore: decodeEventCore(j.jsonEventCore),
				Args:      j.Args,
			},
		}

	case events.PhaseObjectCreated:
		var j jsonObjectEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode object created event: %w", err)
		}
		event = &events.ObjectCreated{
			EventCore: decodeEventCore(j.jsonEventCore),
		}
	case events.PhaseObjectSnapshot:
		var j jsonObjectEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode object snapshot event: %w", err)
		}
		event = &events.ObjectSnapshot{
			EventWithArgs: events.EventWithArgs{
				EventCore: decodeEventCore(j.jsonEventCore),
				Args:      j.Args,
			},
		}
	case events.PhaseObjectDeleted:
		var j jsonObjectEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode object deleted event: %w", err)
		}
		event = &events.ObjectDeleted{
			EventCore: decodeEventCore(j.jsonEventCore),
		}

	case events.PhaseMetadata:
		var j jsonMetadataEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode metadata event: %w", err)
		}
		switch events.MetadataKind(j.Name) {
		case events.MetadataKindProcessName:
			name, err := requireStrEntry(j.Args, "name")
			if err != nil {
				return nil, fmt.Errorf("failed to get process name metadata: %w", err)
			}
			event = &events.MetadataProcessName{
				EventCore:   decodeEventCore(j.jsonEventCore),
				ProcessName: name,
			}
		case events.MetadataKindProcessLabels:
			labels, err := requireStrEntry(j.Args, "labels")
			if err != nil {
				return nil, fmt.Errorf("failed to get process labels metadata: %w", err)
			}
			event = &events.MetadataProcessLabels{
				EventCore: decodeEventCore(j.jsonEventCore),
				Labels:    labels,
			}
		case events.MetadataKindProcessSortIndex:
			sortIndex, err := requireIntEntry(j.Args, "sort_index")
			if err != nil {
				return nil, fmt.Errorf("failed to get process sort index metadata: %w", err)
			}
			event = &events.MetadataProcessSortIndex{
				EventCore: decodeEventCore(j.jsonEventCore),
				SortIndex: sortIndex,
			}
		case events.MetadataKindThreadName:
			name, err := requireStrEntry(j.Args, "name")
			if err != nil {
				return nil, fmt.Errorf("failed to get thread name metadata: %w", err)
			}
			event = &events.MetadataThreadName{
				EventCore:  decodeEventCore(j.jsonEventCore),
				ThreadName: name,
			}
		case events.MetadataKindThreadSortIndex:
			sortIndex, err := requireIntEntry(j.Args, "sort_index")
			if err != nil {
				return nil, fmt.Errorf("failed to get thread sort index metadata: %w", err)
			}
			event = &events.MetadataThreadSortIndex{
				EventCore: decodeEventCore(j.jsonEventCore),
				SortIndex: sortIndex,
			}
		default:
			event = &events.MetadataMisc{
				EventWithArgs: events.EventWithArgs{
					EventCore: decodeEventCore(j.jsonEventCore),
					Args:      j.Args,
				},
			}
		}

	case events.PhaseGlobalMemoryDump:
		var j jsonMemoryDumpEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode global memory dump event: %w", err)
		}
		event = &events.GlobalMemoryDump{
			EventWithArgs: events.EventWithArgs{
				EventCore: decodeEventCore(j.jsonEventCore),
				Args:      j.Args,
			},
		}
	case events.PhaseProcessMemoryDump:
		var j jsonMemoryDumpEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode process memory dump event: %w", err)
		}
		event = &events.ProcessMemoryDump{
			EventWithArgs: events.EventWithArgs{
				EventCore: decodeEventCore(j.jsonEventCore),
				Args:      j.Args,
			},
		}

	case events.PhaseMark:
		var j jsonMarkEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode mark event: %w", err)
		}
		event = &events.Mark{
			EventWithArgs: events.EventWithArgs{
				EventCore: decodeEventCore(j.jsonEventCore),
				Args:      j.Args,
			},
		}

	case events.PhaseClockSync:
		var j jsonClockSyncEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode clock sync event: %w", err)
		}
		issueTs, err := getIntEntry(j.Args, "issue_ts")
		if err != nil {
			return nil, fmt.Errorf("failed to extract issue timestamp: %w", err)
		}
		syncId, err := requireStrEntry(j.Args, "sync_id")
		if err != nil {
			return nil, fmt.Errorf("failed to extract sync ID: %w", err)
		}
		event = &events.ClockSync{
			EventWithArgs: events.EventWithArgs{
				EventCore: decodeEventCore(j.jsonEventCore),
				Args:      j.Args,
			},
			IssueTs: issueTs,
			SyncId:  syncId,
		}

	case events.PhaseContextEnter:
		var j jsonContextEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode context enter event: %w", err)
		}
		event = &events.ContextEnter{
			EventWithArgs: events.EventWithArgs{
				EventCore: decodeEventCore(j.jsonEventCore),
				Args:      j.Args,
			},
		}
	case events.PhaseContextExit:
		var j jsonContextEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode context exit event: %w", err)
		}
		event = &events.ContextExit{
			EventWithArgs: events.EventWithArgs{
				EventCore: decodeEventCore(j.jsonEventCore),
				Args:      j.Args,
			},
		}

	case events.PhaseLinkIds:
		var j jsonLinkedIdEvent
		if err := json.Unmarshal(rawEvent, &j); err != nil {
			return nil, fmt.Errorf("unable to decode linked id event: %w", err)
		}
		linkedId, err := requireStrEntry(j.Args, "linked_id")
		if err != nil {
			return nil, fmt.Errorf("failed to extract linked ID: %w", err)
		}
		event = &events.LinkIds{
			EventWithArgs: events.EventWithArgs{
				EventCore: decodeEventCore(j.jsonEventCore),
				Args:      j.Args,
			},
			LinkedId: linkedId,
		}

	default:
		return nil, fmt.Errorf("unknown phase encountered: '%v'", phase)
	}

	return event, nil
}

func requireIntEntry(args map[string]interface{}, key string) (int64, error) {
	v, err := getIntEntry(args, key)
	if err != nil {
		return 0, err
	}
	if v == nil {
		return 0, fmt.Errorf("integer '%s' expected but was not found", key)
	}
	return *v, nil
}

func getIntEntry(args map[string]interface{}, key string) (*int64, error) {
	v, ok := args[key]
	if !ok {
		return nil, nil
	}

	if f, ok := v.(float64); ok {
		i := int64(f)
		return &i, nil
	}

	return nil, fmt.Errorf("expected number, got '%v': %w", v, ErrInvalidDataType)
}

func requireStrEntry(args map[string]interface{}, key string) (string, error) {
	v, err := getStrEntry(args, key)
	if err != nil {
		return "", err
	}
	if v == nil {
		return "", fmt.Errorf("string '%s' expected but was not found", key)
	}
	return *v, nil
}

func getStrEntry(args map[string]interface{}, key string) (*string, error) {
	v, ok := args[key]
	if !ok {
		return nil, nil
	}

	if s, ok := v.(string); ok {
		return &s, nil
	}

	return nil, fmt.Errorf("expected string, got '%v': %w", v, ErrInvalidDataType)
}

func decodeRawStackTrace(trace []string) *events.StackTrace {
	if len(trace) < 1 {
		return nil
	}

	t := events.StackTrace{}
	for _, entry := range trace {
		t.Trace = append(t.Trace, &events.StackFrame{
			Name: entry,
		})
	}
	return &t
}

func decodeEventPhase(j json.RawMessage) (events.Phase, error) {
	var jsonPhase jsonEventPhase
	err := json.Unmarshal(j, &jsonPhase)
	if err != nil {
		return "", fmt.Errorf("unable to decode phase from JSON event: %w", err)
	}
	return events.Phase(jsonPhase.Phase), nil
}

func decodeEventCore(jsonCore jsonEventCore) events.EventCore {
	categories := make([]string, 0)
	if jsonCore.Categories != "" {
		categories = strings.Split(jsonCore.Categories, ",")
	}

	core := events.EventCore{
		Name:            jsonCore.Name,
		Categories:      categories,
		Timestamp:       jsonCore.Timestamp,
		ThreadTimestamp: jsonCore.ThreadTimestamp,
		ProcessID:       jsonCore.ProcessID,
		ThreadID:        jsonCore.ThreadID,
	}

	return core
}
