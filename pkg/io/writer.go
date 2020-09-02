package io

import (
	"encoding/json"
	"fmt"
	"github.com/omaskery/teffy/pkg/events"
	"io"
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
		jsonEvent, err := writeJsonEvent(event)
		if err != nil {
			return fmt.Errorf("failed while preparing json event: %w", err)
		}

		msg, err := json.Marshal(jsonEvent)
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

func writeJsonEvent(event events.Event) (interface{}, error) {
	switch e := event.(type) {
	case *events.BeginDuration:
		return jsonDurationEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCore(event),
				Args:          e.Args,
			},
			jsonStackInfo: writeStackInfo(e.StackTrace),
		}, nil
	case *events.EndDuration:
		return jsonDurationEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCore(event),
				Args:          e.Args,
			},
			jsonStackInfo: writeStackInfo(e.StackTrace),
		}, nil

	case *events.Complete:
		return jsonCompleteEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCore(event),
				Args:          e.Args,
			},
			jsonStackInfo: writeStackInfo(e.StackTrace),
			EndStack:      writeStackInfo(e.EndStackTrace).Stack,
		}, nil

	case *events.Instant:
		return jsonInstantEvent{
			jsonEventCore: writeJsonEventCore(event),
			jsonStackInfo: writeStackInfo(e.StackTrace),
			Scope:         string(e.Scope),
		}, nil

	case *events.Counter:
		return jsonCounterEvent{
			jsonEventCore: writeJsonEventCore(event),
			Values:        e.Values,
		}, nil

	case *events.AsyncBegin:
		return jsonAsyncEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCore(event),
				Args:          e.Args,
			},
			jsonScopedId: jsonScopedId{
				jsonId: jsonId{
					Id: e.Id,
				},
				Scope: e.Scope,
			},
		}, nil
	case *events.AsyncInstant:
		return jsonAsyncEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCore(event),
				Args:          e.Args,
			},
			jsonScopedId: jsonScopedId{
				jsonId: jsonId{
					Id: e.Id,
				},
				Scope: e.Scope,
			},
		}, nil
	case *events.AsyncEnd:
		return jsonAsyncEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCore(event),
				Args:          e.Args,
			},
			jsonScopedId: jsonScopedId{
				jsonId: jsonId{
					Id: e.Id,
				},
				Scope: e.Scope,
			},
		}, nil

	case *events.ObjectCreated:
		return jsonObjectEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCore(event),
			},
			jsonScopedId: jsonScopedId{
				jsonId: jsonId{
					Id: e.Id,
				},
			},
		}, nil
	case *events.ObjectSnapshot:
		return jsonObjectEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCore(event),
				Args:          e.Args,
			},
			jsonScopedId: jsonScopedId{
				jsonId: jsonId{
					Id: e.Id,
				},
			},
		}, nil
	case *events.ObjectDeleted:
		return jsonObjectEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCore(event),
			},
			jsonScopedId: jsonScopedId{
				jsonId: jsonId{
					Id: e.Id,
				},
			},
		}, nil

	case *events.MetadataProcessName:
		return jsonMetadataEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCoreWithName(event, string(events.MetadataKindProcessName)),
				Args: map[string]interface{}{
					"name": e.ProcessName,
				},
			},
		}, nil
	case *events.MetadataProcessLabels:
		return jsonMetadataEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCoreWithName(event, string(events.MetadataKindProcessLabels)),
				Args: map[string]interface{}{
					"labels": e.Labels,
				},
			},
		}, nil
	case *events.MetadataProcessSortIndex:
		return jsonMetadataEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCoreWithName(event, string(events.MetadataKindProcessSortIndex)),
				Args: map[string]interface{}{
					"sort_index": e.SortIndex,
				},
			},
		}, nil
	case *events.MetadataThreadName:
		return jsonMetadataEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCoreWithName(event, string(events.MetadataKindThreadName)),
				Args: map[string]interface{}{
					"name": e.ThreadName,
				},
			},
		}, nil
	case *events.MetadataThreadSortIndex:
		return jsonMetadataEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCoreWithName(event, string(events.MetadataKindThreadSortIndex)),
				Args: map[string]interface{}{
					"sort_index": e.SortIndex,
				},
			},
		}, nil
	case *events.MetadataMisc:
		return jsonMetadataEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCore(event),
				Args:          e.Args,
			},
		}, nil

	case *events.GlobalMemoryDump:
		return jsonMemoryDumpEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCore(event),
				Args:          e.Args,
			},
		}, nil
	case *events.ProcessMemoryDump:
		return jsonMemoryDumpEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCore(event),
				Args:          e.Args,
			},
		}, nil

	case *events.Mark:
		return jsonMarkEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCore(event),
				Args:          e.Args,
			},
		}, nil

	case *events.ClockSync:
		return jsonClockSyncEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCore(event),
				Args: mergeDicts(e.Args, map[string]interface{}{
					"sync_id":  e.SyncId,
					"issue_ts": e.IssueTs,
				}),
			},
		}, nil

	case *events.ContextEnter:
		return jsonContextEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCore(event),
				Args:          e.Args,
			},
			jsonId: jsonId{
				Id: e.Id,
			},
		}, nil
	case *events.ContextExit:
		return jsonContextEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCore(event),
				Args:          e.Args,
			},
			jsonId: jsonId{
				Id: e.Id,
			},
		}, nil

	case *events.LinkIds:
		return jsonLinkedIdEvent{
			jsonEventWithArgs: jsonEventWithArgs{
				jsonEventCore: writeJsonEventCore(event),
				Args: mergeDicts(e.Args, map[string]interface{}{
					"linked_id": e.LinkedId,
				}),
			},
			jsonId: jsonId{
				Id: e.Id,
			},
		}, nil
	}

	return nil, fmt.Errorf("unknown phase encountered: '%v'", event.Phase())
}

func mergeDicts(a, b map[string]interface{}) map[string]interface{} {
	r := map[string]interface{}{}
	for k, v := range a {
		if v != nil {
			r[k] = v
		}
	}
	for k, v := range b {
		if v != nil {
			r[k] = v
		}
	}
	return r
}

func writeStackInfo(trace *events.StackTrace) jsonStackInfo {
	var stack []string

	if trace != nil {
		stack = make([]string, 0, len(trace.Trace))
		for _, frame := range trace.Trace {
			stack = append(stack, frame.Name)
		}
	}

	return jsonStackInfo{
		Stack: stack,
	}
}

func writeJsonEventCoreWithName(e events.Event, name string) jsonEventCore {
	core := writeJsonEventCore(e)
	core.Name = name
	return core
}

func writeJsonEventCore(e events.Event) jsonEventCore {
	core := e.Core()
	return jsonEventCore{
		jsonEventPhase: jsonEventPhase{
			Phase: string(e.Phase()),
		},
		Name:            core.Name,
		Categories:      strings.Join(core.Categories, ","),
		Timestamp:       core.Timestamp,
		ThreadTimestamp: core.ThreadTimestamp,
		ProcessID:       core.ProcessID,
		ThreadID:        core.ThreadID,
	}
}
