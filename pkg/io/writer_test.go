package io_test

import (
	"github.com/omaskery/teffy/pkg/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strings"

	"github.com/omaskery/teffy/pkg/io"
)

var _ = Describe("WriteJsonObject", func() {
	var writer strings.Builder
	var data io.TefData
	var err error
	var output string

	BeforeEach(func() {
		writer = strings.Builder{}
		data = io.TefData{}
		output = ""
		err = nil
	})

	JustBeforeEach(func() {
		err = io.WriteJsonObject(&writer, data)
		output = writer.String()
	})

	When("using empty trace data", func() {
		It("generates valid output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(`{
				"traceEvents": []
			}`))
		})
	})

	When("defaults are overriden", func() {
		BeforeEach(func() {
			data.SetDisplayTimeUnit(io.DisplayTimeNs)
			data.SetSystemTraceEvents("hello")
			data.SetPowerTraceString("bye")
			data.SetControllerTraceDataKey("wow-key")
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(`{
				"traceEvents": [],
				"displayTimeUnit": "ns",
				"systemTraceEvents": "hello",
				"powerTraceAsString": "bye",
				"controllerTraceDataKey": "wow-key"
			}`))
		})
	})

	When("a single event is written", func() {
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
			Expect(output).To(MatchJSON(`{
					"traceEvents": [
						{
							"name": "wow-an-event",
							"ph": "B",
							"ts": 5023
						}
					]
			}`))
		})
	})

	When("a BeginDuration event is written", func() {
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
				StackTrace: &events.StackTrace{
					Trace: []*events.StackFrame{
						{
							Name: "some stack frame",
						},
					},
				},
			})
		})

		It("generates expected output", func() {
			Expect(err).To(Succeed())
			Expect(output).To(MatchJSON(`{
				"traceEvents": [
					{
						"name": "wow-an-event",
						"cat": "one,two",
						"ph": "B",
						"ts": 5023,
						"tts": 1,
						"pid": 2,
						"tid": 3,
						"stack": ["some stack frame"],
                        "args": {
							"cute": "kittens"
						}
					}
				]
			}`))
		})
	})
})
