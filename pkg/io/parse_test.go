package io_test

import (
	"fmt"
	"github.com/omaskery/teffy/pkg/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strings"

	"github.com/omaskery/teffy/pkg/io"
)

var _ = Describe("ParseJsonFile", func() {
	var testFileContents string
	var data *io.TefData
	var err error

	JustBeforeEach(func() {
		r := strings.NewReader(testFileContents)
		data, err = io.ParseJsonObj(r)
	})

	When("there are no traces", func() {
		BeforeEach(func() {
			testFileContents = `
				{
					"traceEvents": []
				}
			`
		})

		It("correctly parses with reasonable defaults", func() {
			Expect(err).To(Succeed())
			Expect(data.DisplayTimeUnit()).To(Equal(io.DisplayTimeMs))
			Expect(data.PowerTraceAsString()).To(Equal(""))
			Expect(data.SystemTraceEvents()).To(Equal(""))
			Expect(data.Events()).To(BeEmpty())
			Expect(data.StackFrames()).To(BeEmpty())
			Expect(data.ControllerTraceDataKey()).To(Equal("traceEvents"))
		})

		When("it has additional config data", func() {
			BeforeEach(func() {
				testFileContents = `
					{
						"traceEvents": [],
						"displayTimeUnit": "ns",
						"powerTraceAsString": "hello",
						"systemTraceEvents": "hi",
						"stackFrames": {
							"id1": {
								"category": "MyCategory1",
								"name": "MyName1",
								"parent": "id2"
							},
							"id2": {
								"category": "MyCategory2",
								"name": "MyName2"
							}
						},
						"controllerTraceDataKey": "kittens"
					}
				`
			})

			It("correctly stores the additional config data", func() {
				Expect(err).To(Succeed())
				Expect(data.DisplayTimeUnit()).To(Equal(io.DisplayTimeNs))
				Expect(data.PowerTraceAsString()).To(Equal("hello"))
				Expect(data.SystemTraceEvents()).To(Equal("hi"))
				Expect(data.Events()).To(BeEmpty())
				Expect(data.ControllerTraceDataKey()).To(Equal("kittens"))
				Expect(data.StackFrames()).To(HaveLen(2))
			})
		})
	})
})

var _ = Describe("ParseJsonArray", func() {
	var testFileContents string
	var data *io.TefData
	var err error

	JustBeforeEach(func() {
		r := strings.NewReader(testFileContents)
		data, err = io.ParseJsonArray(r)
	})

	When("when there is a well formed but empty array", func() {
		BeforeEach(func() {
			testFileContents = `
				[]
			`
		})

		It("correctly parses with reasonable defaults", func() {
			Expect(err).To(Succeed())
			Expect(data.DisplayTimeUnit()).To(Equal(io.DisplayTimeMs))
			Expect(data.PowerTraceAsString()).To(Equal(""))
			Expect(data.SystemTraceEvents()).To(Equal(""))
			Expect(data.Events()).To(BeEmpty())
			Expect(data.StackFrames()).To(BeEmpty())
			Expect(data.ControllerTraceDataKey()).To(Equal("traceEvents"))
		})
	})

	When("when there is a well formed array with 1 entry", func() {
		BeforeEach(func() {
			testFileContents = `
				[{
					"name": "namesies",
					"ph": "B"
				}]
			`
		})

		It("successfully parses an entry", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(1))
			Expect(data.Events()[0].Phase()).To(Equal(events.PhaseBeginDuration))
			Expect(data.Events()[0].Core().Name).To(Equal("namesies"))
		})
	})

	When("when there is a well formed array with 2 entries", func() {
		BeforeEach(func() {
			testFileContents = `
				[{
					"name": "namesies1",
					"ph": "B",
					"ts": 0
				},{
					"name": "namesies2",
					"ph": "B",
					"ts": 10
				}]
			`
		})

		It("successfully parses two entries", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(2))
			Expect(data.Events()[0].Phase()).To(Equal(events.PhaseBeginDuration))
			Expect(data.Events()[0].Core().Name).To(Equal("namesies1"))
			Expect(data.Events()[0].Core().Timestamp).To(BeNumerically("==", 0))
			Expect(data.Events()[1].Phase()).To(Equal(events.PhaseBeginDuration))
			Expect(data.Events()[1].Core().Name).To(Equal("namesies2"))
			Expect(data.Events()[1].Core().Timestamp).To(BeNumerically("==", 10))
		})
	})

	When("when there is an incomplete array with 2 entries and a trailing comma", func() {
		BeforeEach(func() {
			testFileContents = `
				[{
					"name": "namesies1",
					"ph": "B",
					"ts": 0
				},{
					"name": "namesies2",
					"ph": "B",
					"ts": 10
				},
			`
		})

		It("successfully parses two entries", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(2))
			Expect(data.Events()[0].Phase()).To(Equal(events.PhaseBeginDuration))
			Expect(data.Events()[0].Core().Name).To(Equal("namesies1"))
			Expect(data.Events()[0].Core().Timestamp).To(BeNumerically("==", 0))
			Expect(data.Events()[1].Phase()).To(Equal(events.PhaseBeginDuration))
			Expect(data.Events()[1].Core().Name).To(Equal("namesies2"))
			Expect(data.Events()[1].Core().Timestamp).To(BeNumerically("==", 10))
		})
	})

	When("when there is an incomplete array with 2 entries and no trailing comma", func() {
		BeforeEach(func() {
			testFileContents = `
				[{
					"name": "namesies1",
					"ph": "B",
					"ts": 0
				},{
					"name": "namesies2",
					"ph": "B",
					"ts": 10
				}
			`
		})

		It("successfully parses two entries", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(2))
			Expect(data.Events()[0].Phase()).To(Equal(events.PhaseBeginDuration))
			Expect(data.Events()[0].Core().Name).To(Equal("namesies1"))
			Expect(data.Events()[0].Core().Timestamp).To(BeNumerically("==", 0))
			Expect(data.Events()[1].Phase()).To(Equal(events.PhaseBeginDuration))
			Expect(data.Events()[1].Core().Name).To(Equal("namesies2"))
			Expect(data.Events()[1].Core().Timestamp).To(BeNumerically("==", 10))
		})
	})
})

var _ = Describe("Parsing EventCore", func() {
	var testFileContents string
	var data *io.TefData
	var err error

	JustBeforeEach(func() {
		r := strings.NewReader(testFileContents)
		data, err = io.ParseJsonArray(r)
	})

	When("when all mandatory fields are present", func() {
		BeforeEach(func() {
			testFileContents = `[{
				"name": "A",
				"ph": "B",
				"ts": 0
			}]`
		})

		It("correctly parses with reasonable defaults", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(1))
			event := data.Events()[0]
			Expect(event.Core().Name).To(Equal("A"))
			Expect(event.Core().Timestamp).To(Equal(int64(0)))
			Expect(event.Core().ThreadTimestamp).To(BeNil())
			Expect(event.Core().ProcessID).To(BeNil())
			Expect(event.Core().ThreadID).To(BeNil())
			Expect(event.Core().Categories).To(BeEmpty())
		})
	})

	When("when all fields are present", func() {
		BeforeEach(func() {
			testFileContents = `[{
				"name": "A",
				"cat": "one,two",
				"ph": "B",
				"ts": 0,
				"tts": 10,
				"pid": 1,
				"tid": 2
			}]`
		})

		It("correctly parses with reasonable defaults", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(1))
			event := data.Events()[0]
			Expect(event.Core().Name).To(Equal("A"))
			Expect(event.Core().Timestamp).To(Equal(int64(0)))
			Expect(event.Core().ThreadTimestamp).ToNot(BeNil())
			Expect(*event.Core().ThreadTimestamp).To(BeNumerically("==", int64(10)))
			Expect(event.Core().ProcessID).ToNot(BeNil())
			Expect(*event.Core().ProcessID).To(BeNumerically("==", int64(1)))
			Expect(event.Core().ThreadID).ToNot(BeNil())
			Expect(*event.Core().ThreadID).To(BeNumerically("==", int64(2)))
			Expect(event.Core().Categories).To(HaveLen(2))
			Expect(event.Core().Categories[0]).To(Equal("one"))
			Expect(event.Core().Categories[1]).To(Equal("two"))
		})
	})
})

var _ = Describe("Parsing Begin Duration", func() {
	var testFileContents string
	var data *io.TefData
	var err error

	JustBeforeEach(func() {
		r := strings.NewReader(testFileContents)
		data, err = io.ParseJsonArray(r)
	})

	When("when only essentials are present", func() {
		BeforeEach(func() {
			testFileContents = `[{
				"name": "A",
				"ph": "B",
				"ts": 0
			}]`
		})

		It("correctly defaults values", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(1))
			event, ok := data.Events()[0].(*events.BeginDuration)
			Expect(ok).To(BeTrue())
			Expect(event.StackTrace).To(BeNil())
			Expect(event.Args).To(BeEmpty())
		})
	})

	When("when stacktrace is present", func() {
		BeforeEach(func() {
			testFileContents = `[{
				"name": "A",
				"ph": "B",
				"ts": 0,
				"stack": [
					"one", "two"
				]
			}]`
		})

		It("correctly parses the arguments", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(1))
			event, ok := data.Events()[0].(*events.BeginDuration)
			Expect(ok).To(BeTrue())
			Expect(event.Args).To(BeEmpty())
			Expect(event.StackTrace).ToNot(BeNil())
			Expect(event.StackTrace.Trace).To(HaveLen(2))
			Expect(event.StackTrace.Trace[0].Parent).To(Equal(""))
			Expect(event.StackTrace.Trace[0].Category).To(Equal(""))
			Expect(event.StackTrace.Trace[0].Name).To(Equal("one"))
			Expect(event.StackTrace.Trace[1].Parent).To(Equal(""))
			Expect(event.StackTrace.Trace[1].Category).To(Equal(""))
			Expect(event.StackTrace.Trace[1].Name).To(Equal("two"))
		})
	})

	When("when arguments are present", func() {
		BeforeEach(func() {
			testFileContents = `[{
				"name": "A",
				"ph": "B",
				"ts": 0,
				"args": {
					"one": 1,
					"two": "too"
				}
			}]`
		})

		It("correctly parses the arguments", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(1))
			event, ok := data.Events()[0].(*events.BeginDuration)
			Expect(ok).To(BeTrue())
			Expect(event.StackTrace).To(BeNil())
			Expect(event.Args).To(HaveLen(2))
			Expect(event.Args).To(HaveKeyWithValue("one", float64(1)))
			Expect(event.Args).To(HaveKeyWithValue("two", "too"))
		})
	})
})

var _ = Describe("Parsing Async Start", func() {
	var testFileContents string
	var data *io.TefData
	var err error

	JustBeforeEach(func() {
		r := strings.NewReader(testFileContents)
		data, err = io.ParseJsonArray(r)
	})

	When("parsing its deprecated form", func() {
		BeforeEach(func() {
			testFileContents = makeTrivialEventWithPhase("S")
		})

		It("generates the correct type", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(1))
			_, ok := data.Events()[0].(*events.AsyncBegin)
			Expect(ok).To(BeTrue())
		})
	})

	When("parsing its current form", func() {
		BeforeEach(func() {
			testFileContents = makeTrivialEventWithPhase(events.PhaseAsyncBegin)
		})

		It("generates the correct type", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(1))
			_, ok := data.Events()[0].(*events.AsyncBegin)
			Expect(ok).To(BeTrue())
		})
	})
})

var _ = Describe("Parsing Async Instant", func() {
	var testFileContents string
	var data *io.TefData
	var err error

	JustBeforeEach(func() {
		r := strings.NewReader(testFileContents)
		data, err = io.ParseJsonArray(r)
	})

	When("parsing its deprecated 'step into' form", func() {
		BeforeEach(func() {
			testFileContents = makeTrivialEventWithPhase("T")
		})

		It("generates the correct type", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(1))
			_, ok := data.Events()[0].(*events.AsyncInstant)
			Expect(ok).To(BeTrue())
		})
	})

	When("parsing its deprecated 'step past' form", func() {
		BeforeEach(func() {
			testFileContents = makeTrivialEventWithPhase("p")
		})

		It("generates the correct type", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(1))
			_, ok := data.Events()[0].(*events.AsyncInstant)
			Expect(ok).To(BeTrue())
		})
	})

	When("parsing its current form", func() {
		BeforeEach(func() {
			testFileContents = makeTrivialEventWithPhase(events.PhaseAsyncInstant)
		})

		It("generates the correct type", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(1))
			_, ok := data.Events()[0].(*events.AsyncInstant)
			Expect(ok).To(BeTrue())
		})
	})
})

var _ = Describe("Parsing Async End", func() {
	var testFileContents string
	var data *io.TefData
	var err error

	JustBeforeEach(func() {
		r := strings.NewReader(testFileContents)
		data, err = io.ParseJsonArray(r)
	})

	When("parsing its deprecated form", func() {
		BeforeEach(func() {
			testFileContents = makeTrivialEventWithPhase("F")
		})

		It("generates the correct type", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(1))
			_, ok := data.Events()[0].(*events.AsyncEnd)
			Expect(ok).To(BeTrue())
		})
	})

	When("parsing its current form", func() {
		BeforeEach(func() {
			testFileContents = makeTrivialEventWithPhase(events.PhaseAsyncEnd)
		})

		It("generates the correct type", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(1))
			_, ok := data.Events()[0].(*events.AsyncEnd)
			Expect(ok).To(BeTrue())
		})
	})
})

var _ = Describe("Parsing Object Created", func() {
	var testFileContents string
	var data *io.TefData
	var err error

	JustBeforeEach(func() {
		r := strings.NewReader(testFileContents)
		data, err = io.ParseJsonArray(r)
	})

	When("parsing", func() {
		BeforeEach(func() {
			testFileContents = makeTrivialEventWithPhase(events.PhaseObjectCreated)
		})

		It("generates the correct type", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(1))
			_, ok := data.Events()[0].(*events.ObjectCreated)
			Expect(ok).To(BeTrue())
		})
	})
})

var _ = Describe("Parsing Object Snapshot", func() {
	var testFileContents string
	var data *io.TefData
	var err error

	JustBeforeEach(func() {
		r := strings.NewReader(testFileContents)
		data, err = io.ParseJsonArray(r)
	})

	When("parsing", func() {
		BeforeEach(func() {
			testFileContents = makeTrivialEventWithPhase(events.PhaseObjectSnapshot)
		})

		It("generates the correct type", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(1))
			_, ok := data.Events()[0].(*events.ObjectSnapshot)
			Expect(ok).To(BeTrue())
		})
	})
})

var _ = Describe("Parsing Object Deleted", func() {
	var testFileContents string
	var data *io.TefData
	var err error

	JustBeforeEach(func() {
		r := strings.NewReader(testFileContents)
		data, err = io.ParseJsonArray(r)
	})

	When("parsing", func() {
		BeforeEach(func() {
			testFileContents = makeTrivialEventWithPhase(events.PhaseObjectDeleted)
		})

		It("generates the correct type", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(1))
			_, ok := data.Events()[0].(*events.ObjectDeleted)
			Expect(ok).To(BeTrue())
		})
	})
})

var _ = Describe("Parsing Mark", func() {
	var testFileContents string
	var data *io.TefData
	var err error

	JustBeforeEach(func() {
		r := strings.NewReader(testFileContents)
		data, err = io.ParseJsonArray(r)
	})

	When("parsing", func() {
		BeforeEach(func() {
			testFileContents = makeTrivialEventWithPhase(events.PhaseMark)
		})

		It("generates the correct type", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(1))
			_, ok := data.Events()[0].(*events.Mark)
			Expect(ok).To(BeTrue())
		})
	})
})

var _ = Describe("Parsing Context Enter", func() {
	var testFileContents string
	var data *io.TefData
	var err error

	JustBeforeEach(func() {
		r := strings.NewReader(testFileContents)
		data, err = io.ParseJsonArray(r)
	})

	When("parsing", func() {
		BeforeEach(func() {
			testFileContents = makeTrivialEventWithPhase(events.PhaseContextEnter)
		})

		It("generates the correct type", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(1))
			_, ok := data.Events()[0].(*events.ContextEnter)
			Expect(ok).To(BeTrue())
		})
	})
})

var _ = Describe("Parsing Context Exit", func() {
	var testFileContents string
	var data *io.TefData
	var err error

	JustBeforeEach(func() {
		r := strings.NewReader(testFileContents)
		data, err = io.ParseJsonArray(r)
	})

	When("parsing", func() {
		BeforeEach(func() {
			testFileContents = makeTrivialEventWithPhase(events.PhaseContextExit)
		})

		It("generates the correct type", func() {
			Expect(err).To(Succeed())
			Expect(data.Events()).To(HaveLen(1))
			_, ok := data.Events()[0].(*events.ContextExit)
			Expect(ok).To(BeTrue())
		})
	})
})

func makeTrivialEventWithPhase(phase events.Phase) string {
	return fmt.Sprintf(`[{
		"name": "event-name",
		"ph": "%s",
		"ts": 0
	}]`, string(phase))
}
