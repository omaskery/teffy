package io_test

import (
	"encoding/json"
	"fmt"
	"github.com/omaskery/teffy/pkg/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"strings"

	teffyio "github.com/omaskery/teffy/pkg/io"
)

var _ = Describe("WriteJsonObject", func() {
	var writer strings.Builder
	var data teffyio.TefData
	var err error
	var output string

	BeforeEach(func() {
		writer = strings.Builder{}
		data = teffyio.TefData{}
		output = ""
		err = nil
	})

	JustBeforeEach(func() {
		err = teffyio.WriteJsonObject(&writer, data)
		output = writer.String()
	})

	When("using empty trace data", func() {
		It("generates valid output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile()))
		})
	})

	When("defaults are overriden", func() {
		BeforeEach(func() {
			data.SetDisplayTimeUnit(teffyio.DisplayTimeNs)
			data.SetSystemTraceEvents("hello")
			data.SetPowerTraceString("bye")
			data.SetControllerTraceDataKey("wow-key")
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(mustJson(map[string]interface{}{
				"traceEvents":            []interface{}{},
				"displayTimeUnit":        "ns",
				"systemTraceEvents":      "hello",
				"powerTraceAsString":     "bye",
				"controllerTraceDataKey": "wow-key",
			})))
		})
	})

	When("stack frames are stored", func() {
		BeforeEach(func() {
			data.SetStackFrame("one", &events.StackFrame{
				Category: "cat1",
				Name:     "name1",
				Parent:   "parent1",
			})
			data.SetStackFrame("two", &events.StackFrame{
				Category: "cat2",
				Name:     "name2",
				Parent:   "parent2",
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(mustJson(map[string]interface{}{
				"traceEvents": []interface{}{},
				"stackFrames": map[string]interface{}{
					"one": map[string]interface{}{
						"category": "cat1",
						"name":     "name1",
						"parent":   "parent1",
					},
					"two": map[string]interface{}{
						"category": "cat2",
						"name":     "name2",
						"parent":   "parent2",
					},
				},
			})))
		})
	})

	When("a single event is written", func() {
		Context("with minimal fields", func() {
			BeforeEach(func() {
				data.Write(&events.BeginDuration{
					EventWithArgs: events.EventWithArgs{
						EventCore: events.EventCore{
							Name:      "wow-an-event",
							Timestamp: 5023,
						},
					},
				})
			})

			It("generates expected output", func() {
				Expect(err).To(Succeed())
				Expect(output).To(MatchJSON(testJsonObjFile(
					mustJson(map[string]interface{}{
						"name": "wow-an-event",
						"ph":   "B",
						"ts":   5023,
					}),
				)))
			})
		})

		Context("with all fields", func() {
			BeforeEach(func() {
				tts := int64(1)
				pid := int64(2)
				tid := int64(3)
				data.Write(&events.BeginDuration{
					EventWithArgs: events.EventWithArgs{
						EventCore: events.EventCore{
							Name:            "wow-an-event",
							Categories:      []string{"one", "two"},
							Timestamp:       5023,
							ThreadTimestamp: &tts,
							ProcessID:       &pid,
							ThreadID:        &tid,
						},
						Args: map[string]interface{}{
							"cute": "kittens",
						},
					},
					EventStackTrace: events.EventStackTrace{
						StackTrace: &events.StackTrace{
							Trace: []*events.StackFrame{
								{
									Name: "some stack frame",
								},
							},
						},
					},
				})
			})

			It("generates expected output", func() {
				Expect(err).To(Succeed())
				Expect(output).To(MatchJSON(mustJson(map[string]interface{}{
					"traceEvents": []interface{}{
						map[string]interface{}{
							"name":  "wow-an-event",
							"cat":   "one,two",
							"ph":    "B",
							"ts":    5023,
							"tts":   1,
							"pid":   2,
							"tid":   3,
							"stack": []interface{}{"some stack frame"},
							"args":  minimalArgs(),
						},
					},
				})))
			})
		})
	})

	When("a BeginDuration event is written", func() {
		BeforeEach(func() {
			data.Write(&events.BeginDuration{
				EventWithArgs: minimalEventWithArgs(minimalArgs()),
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseBeginDuration, minimalArgs(), nil),
			)))
		})
	})

	When("an EndDuration event is written", func() {
		BeforeEach(func() {
			data.Write(&events.EndDuration{
				EventWithArgs: minimalEventWithArgs(minimalArgs()),
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseEndDuration, minimalArgs(), nil),
			)))
		})
	})

	When("a Complete event is written", func() {
		BeforeEach(func() {
			data.Write(&events.Complete{
				EventWithArgs: minimalEventWithArgs(minimalArgs()),
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseComplete, minimalArgs(), nil),
			)))
		})
	})

	When("an Instant event is written", func() {
		Context("with no scope specified", func() {
			BeforeEach(func() {
				data.Write(&events.Instant{
					EventCore: minimalEventCore(),
				})
			})

			It("generates expected output", func() {
				Expect(err).To(Succeed())
				Expect(output).To(MatchJSON(testJsonObjFile(
					eventJson(events.PhaseInstant, nil, nil),
				)))
			})
		})

		Context("with a scope specified", func() {
			BeforeEach(func() {
				data.Write(&events.Instant{
					EventCore: minimalEventCore(),
					Scope:     events.InstantScopeProcess,
				})
			})

			It("generates expected output", func() {
				Expect(err).To(Succeed())
				Expect(output).To(MatchJSON(testJsonObjFile(
					eventJson(events.PhaseInstant, nil, map[string]interface{}{
						"s": string(events.InstantScopeProcess),
					}),
				)))
			})
		})
	})

	When("a Counter event is written", func() {
		BeforeEach(func() {
			data.Write(&events.Counter{
				EventCore: minimalEventCore(),
				Values: map[string]float64{
					"hello": 24,
					"meow":  10,
				},
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseCounter, map[string]interface{}{
					"hello": 24,
					"meow":  10,
				}, nil),
			)))
		})
	})

	When("a AsyncBegin event is written", func() {
		BeforeEach(func() {
			data.Write(&events.AsyncBegin{
				EventWithArgs: minimalEventWithArgs(minimalArgs()),
				Id:            "some-id",
				Scope:         "some-scope",
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseAsyncBegin, minimalArgs(), minimalId(true)),
			)))
		})
	})

	When("a AsyncInstant event is written", func() {
		BeforeEach(func() {
			data.Write(&events.AsyncInstant{
				EventWithArgs: minimalEventWithArgs(minimalArgs()),
				Id:            "some-id",
				Scope:         "some-scope",
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseAsyncInstant, minimalArgs(), minimalId(true)),
			)))
		})
	})

	When("a AsyncEnd event is written", func() {
		BeforeEach(func() {
			data.Write(&events.AsyncEnd{
				EventWithArgs: minimalEventWithArgs(minimalArgs()),
				Id:            "some-id",
				Scope:         "some-scope",
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseAsyncEnd, minimalArgs(), minimalId(true)),
			)))
		})
	})

	When("a ObjectCreated event is written", func() {
		BeforeEach(func() {
			data.Write(&events.ObjectCreated{
				EventCore: minimalEventCore(),
				Id:        "some-id",
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseObjectCreated, nil, minimalId(false)),
			)))
		})
	})

	When("a ObjectSnapshot event is written", func() {
		BeforeEach(func() {
			data.Write(&events.ObjectSnapshot{
				EventWithArgs: minimalEventWithArgs(minimalArgs()),
				Id:            "some-id",
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseObjectSnapshot, minimalArgs(), minimalId(false)),
			)))
		})
	})

	When("a ObjectDeleted event is written", func() {
		BeforeEach(func() {
			data.Write(&events.ObjectDeleted{
				EventCore: minimalEventCore(),
				Id:        "some-id",
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseObjectDeleted, nil, minimalId(false)),
			)))
		})
	})

	When("a Metadata (Process Name) event is written", func() {
		BeforeEach(func() {
			data.Write(&events.MetadataProcessName{
				EventCore:   minimalEventCore(),
				ProcessName: "some-process-name",
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseMetadata, map[string]interface{}{
					"name": "some-process-name",
				}, withEventName(string(events.MetadataKindProcessName))),
			)))
		})
	})

	When("a Metadata (Process Labels) event is written", func() {
		BeforeEach(func() {
			data.Write(&events.MetadataProcessLabels{
				EventCore: minimalEventCore(),
				Labels:    "some-label",
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseMetadata, map[string]interface{}{
					"labels": "some-label",
				}, withEventName(string(events.MetadataKindProcessLabels))),
			)))
		})
	})

	When("a Metadata (Process Sort Index) event is written", func() {
		BeforeEach(func() {
			data.Write(&events.MetadataProcessSortIndex{
				EventCore: minimalEventCore(),
				SortIndex: 3,
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseMetadata, map[string]interface{}{
					"sort_index": 3,
				}, withEventName(string(events.MetadataKindProcessSortIndex))),
			)))
		})
	})

	When("a Metadata (Thread Name) event is written", func() {
		BeforeEach(func() {
			data.Write(&events.MetadataThreadName{
				EventCore:  minimalEventCore(),
				ThreadName: "some-thread-name",
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseMetadata, map[string]interface{}{
					"name": "some-thread-name",
				}, withEventName(string(events.MetadataKindThreadName))),
			)))
		})
	})

	When("a Metadata (Thread Sort Index) event is written", func() {
		BeforeEach(func() {
			data.Write(&events.MetadataThreadSortIndex{
				EventCore: minimalEventCore(),
				SortIndex: 3,
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseMetadata, map[string]interface{}{
					"sort_index": 3,
				}, withEventName(string(events.MetadataKindThreadSortIndex))),
			)))
		})
	})

	When("a Metadata (Misc) event is written", func() {
		BeforeEach(func() {
			data.Write(&events.MetadataMisc{
				EventWithArgs: minimalEventWithArgs(minimalArgs()),
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseMetadata, minimalArgs(), nil),
			)))
		})
	})

	When("a Global Memory Dump event is written", func() {
		BeforeEach(func() {
			data.Write(&events.GlobalMemoryDump{
				EventWithArgs: minimalEventWithArgs(minimalArgs()),
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseGlobalMemoryDump, minimalArgs(), nil),
			)))
		})
	})

	When("a Process Memory Dump event is written", func() {
		BeforeEach(func() {
			data.Write(&events.ProcessMemoryDump{
				EventWithArgs: minimalEventWithArgs(minimalArgs()),
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseProcessMemoryDump, minimalArgs(), nil),
			)))
		})
	})

	When("a Mark event is written", func() {
		BeforeEach(func() {
			data.Write(&events.Mark{
				EventWithArgs: minimalEventWithArgs(minimalArgs()),
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseMark, minimalArgs(), nil),
			)))
		})
	})

	When("a ClockSync event is written", func() {
		BeforeEach(func() {
			issueTs := int64(1)
			data.Write(&events.ClockSync{
				EventWithArgs: minimalEventWithArgs(minimalArgs()),
				SyncId:        "hello",
				IssueTs:       &issueTs,
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseClockSync, map[string]interface{}{
					"cute":     "kittens",
					"sync_id":  "hello",
					"issue_ts": int64(1),
				}, nil),
			)))
		})
	})

	When("a ContextEnter event is written", func() {
		BeforeEach(func() {
			data.Write(&events.ContextEnter{
				EventWithArgs: minimalEventWithArgs(minimalArgs()),
				Id:            "some-id",
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseContextEnter, minimalArgs(), minimalId(false)),
			)))
		})
	})

	When("a ContextExit event is written", func() {
		BeforeEach(func() {
			data.Write(&events.ContextExit{
				EventWithArgs: minimalEventWithArgs(minimalArgs()),
				Id:            "some-id",
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseContextExit, minimalArgs(), minimalId(false)),
			)))
		})
	})

	When("a LinkedIds event is written", func() {
		BeforeEach(func() {
			data.Write(&events.LinkIds{
				EventWithArgs: minimalEventWithArgs(minimalArgs()),
				Id:            "some-id",
				LinkedId:      "some-other-id",
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonObjFile(
				eventJson(events.PhaseLinkIds, map[string]interface{}{
					"cute":      "kittens",
					"linked_id": "some-other-id",
				}, minimalId(false)),
			)))
		})
	})
})

var _ = Describe("WriteJsonArray", func() {
	var writer strings.Builder
	var data []events.Event
	var err error
	var output string

	BeforeEach(func() {
		writer = strings.Builder{}
		data = make([]events.Event, 0)
		output = ""
		err = nil
	})

	JustBeforeEach(func() {
		err = teffyio.WriteJsonArray(&writer, data)
		output = writer.String()
	})

	When("writing an empty array", func() {
		It("produces expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON("[]"))
		})
	})

	When("writing a single event", func() {
		BeforeEach(func() {
			data = append(data, &events.BeginDuration{
				EventWithArgs: minimalEventWithArgs(minimalArgs()),
			})
		})

		It("produces expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(testJsonArrFile(
				eventJson(events.PhaseBeginDuration, minimalArgs(), nil),
			)))
		})
	})
})

var _ = Describe("StreamingWriter", func() {
	var writer strings.Builder
	var stream teffyio.EventWriter
	var output string

	BeforeEach(func() {
		writer = strings.Builder{}
		stream = teffyio.NewStreamingWriter(writerNoopCloser(&writer))
		output = ""
	})

	Context("when the stream is not closed properly", func() {
		JustBeforeEach(func() {
			output = writer.String()
		})

		When("writing no entries", func() {
			It("produces no output", func() {
				Expect(output).To(Equal(""))
			})
		})

		When("writing one entry", func() {
			BeforeEach(func() {
				Expect(stream.Write(&events.BeginDuration{
					EventWithArgs: minimalEventWithArgs(minimalArgs()),
				})).To(Succeed())
			})

			It("produces the valid start of an array", func() {
				Expect(output + "]").To(MatchJSON(testJsonArrFile(
					eventJson(events.PhaseBeginDuration, minimalArgs(), nil),
				)))
			})
		})

		When("writing two events", func() {
			BeforeEach(func() {
				Expect(stream.Write(&events.BeginDuration{
					EventWithArgs: minimalEventWithArgs(minimalArgs()),
				})).To(Succeed())
				Expect(stream.Write(&events.EndDuration{
					EventWithArgs: minimalEventWithArgs(minimalArgs()),
				})).To(Succeed())
			})

			It("produces the valid start of an array", func() {
				Expect(output + "]").To(MatchJSON(testJsonArrFile(
					eventJson(events.PhaseBeginDuration, minimalArgs(), nil),
					eventJson(events.PhaseEndDuration, minimalArgs(), nil),
				)))
			})
		})
	})

	Context("when the stream is closed on completion", func() {
		JustBeforeEach(func() {
			Expect(stream.Close()).To(Succeed())
			output = writer.String()
		})

		When("writing no entries", func() {
			It("produces empty array output", func() {
				Expect(output).To(MatchJSON("[]"))
			})
		})

		When("writing a single event", func() {
			BeforeEach(func() {
				Expect(stream.Write(&events.BeginDuration{
					EventWithArgs: minimalEventWithArgs(minimalArgs()),
				})).To(Succeed())
			})

			It("produces array with single expected element", func() {
				Expect(output).To(MatchJSON(testJsonArrFile(
					eventJson(events.PhaseBeginDuration, minimalArgs(), nil),
				)))
			})
		})

		When("writing two events", func() {
			BeforeEach(func() {
				Expect(stream.Write(&events.BeginDuration{
					EventWithArgs: minimalEventWithArgs(minimalArgs()),
				})).To(Succeed())
				Expect(stream.Write(&events.EndDuration{
					EventWithArgs: minimalEventWithArgs(minimalArgs()),
				})).To(Succeed())
			})

			It("produces array with two expected elements", func() {
				Expect(output).To(MatchJSON(testJsonArrFile(
					eventJson(events.PhaseBeginDuration, minimalArgs(), nil),
					eventJson(events.PhaseEndDuration, minimalArgs(), nil),
				)))
			})
		})
	})
})

type wrapper struct {
	io.Writer
}

func (w *wrapper) Close() error {
	return nil
}

func writerNoopCloser(w io.Writer) io.WriteCloser {
	return &wrapper{w}
}

func testJsonObjFile(events ...string) string {
	return fmt.Sprintf(`{
		"traceEvents": [
			%s
		]
	}`, strings.Join(events, ","))
}

func testJsonArrFile(events ...string) string {
	return fmt.Sprintf("[%s]", strings.Join(events, ","))
}

func mustJson(j map[string]interface{}) string {
	result, err := json.Marshal(j)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal test data to JSON: %v", err))
	}
	return string(result)
}

func eventJson(phase events.Phase, args map[string]interface{}, extra map[string]interface{}) string {
	j := map[string]interface{}{
		"name": "event-name",
		"ph":   string(phase),
		"ts":   1,
	}
	if args != nil {
		j["args"] = args
	}
	if extra != nil {
		for k, v := range extra {
			j[k] = v
		}
	}
	return mustJson(j)
}

func minimalEventCore() events.EventCore {
	return events.EventCore{
		Name:      "event-name",
		Timestamp: 1,
	}
}

func minimalEventWithArgs(args map[string]interface{}) events.EventWithArgs {
	return events.EventWithArgs{
		EventCore: minimalEventCore(),
		Args:      args,
	}
}

func minimalArgs() map[string]interface{} {
	return map[string]interface{}{
		"cute": "kittens",
	}
}

func withEventName(name string) map[string]interface{} {
	return map[string]interface{}{
		"name": name,
	}
}

func minimalId(scoped bool) map[string]interface{} {
	result := map[string]interface{}{
		"id": "some-id",
	}
	if scoped {
		result["scope"] = "some-scope"
	}
	return result
}
