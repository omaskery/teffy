package trace_test

import (
	"github.com/omaskery/teffy/pkg/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"

	"github.com/omaskery/teffy/pkg/util/trace"
)

type mockEventWriter struct {
	events []events.Event
}

func (m *mockEventWriter) Write(e events.Event) error {
	m.events = append(m.events, e)
	return nil
}

func (m *mockEventWriter) Close() error {
	return nil
}

func (m *mockEventWriter) lastEvent() events.Event {
	l := len(m.events)
	if l < 1 {
		return nil
	}
	return m.events[l-1]
}

type mockTimestamp struct {
	time int64
}

func (m *mockTimestamp) getTimestamp() int64 {
	return m.time
}

var _ = Describe("Tracer", func() {
	var mockTime mockTimestamp
	var tracer *trace.Tracer
	var options []trace.TracerOption
	var eventWriter mockEventWriter
	pid := int64(os.Getpid())

	JustBeforeEach(func() {
		mockTime = mockTimestamp{}
		eventWriter = mockEventWriter{}
		baseOptions := []trace.TracerOption{
			trace.WithTimestampFn(mockTime.getTimestamp),
		}
		allOptions := append(baseOptions, options...)
		tracer = trace.NewTracer(&eventWriter, allOptions...)
	})

	When("no events are written", func() {
		It("is an empty list of events", func() {
			Expect(tracer.Close()).To(Succeed())
			Expect(eventWriter.events).To(BeEmpty())
		})
	})

	When("a duration is started", func() {
		var d trace.Duration

		When("there are no options", func() {
			JustBeforeEach(func() {
				d = tracer.BeginDuration("such-duration")
			})

			It("emits a single BeginDuration event", func() {
				Expect(eventWriter.events).To(HaveLen(1))
				Expect(eventWriter.lastEvent()).To(Equal(&events.BeginDuration{
					EventWithArgs: events.EventWithArgs{
						EventCore: events.EventCore{
							Name:      "such-duration",
							Timestamp: 0,
							ProcessID: &pid,
						},
					},
				}))
			})

			When("the duration is ended", func() {
				var endOptions []trace.EventOption

				JustBeforeEach(func() {
					mockTime.time = 10
					d.End(endOptions...)
				})

				It("emits an EndDuration event", func() {
					Expect(eventWriter.events).To(HaveLen(2))
					Expect(eventWriter.lastEvent()).To(Equal(&events.EndDuration{
						EventWithArgs: events.EventWithArgs{
							EventCore: events.EventCore{
								Name:      "such-duration",
								Timestamp: 10,
								ProcessID: &pid,
							},
						},
					}))
				})
			})
		})

		When("there are options", func() {
			JustBeforeEach(func() {
				d = tracer.BeginDuration("such-duration",
					trace.WithCategories("one", "two"),
					trace.WithArgs(map[string]interface{}{"a": 5}))
			})

			It("includes the options in the output", func() {
				Expect(eventWriter.events).To(HaveLen(1))
				Expect(eventWriter.lastEvent()).To(Equal(&events.BeginDuration{
					EventWithArgs: events.EventWithArgs{
						EventCore: events.EventCore{
							Name:       "such-duration",
							Timestamp:  0,
							ProcessID:  &pid,
							Categories: []string{"one", "two"},
						},
						Args: map[string]interface{}{
							"a": 5,
						},
					},
				}))
			})
		})
	})

	When("an instant is emitted", func() {
		Context("without extra options", func() {
			JustBeforeEach(func() {
				tracer.Instant("such-instant")
			})

			It("emits a sensible event", func() {
				Expect(eventWriter.events).To(HaveLen(1))
				Expect(eventWriter.lastEvent()).To(Equal(&events.Instant{
					EventCore: events.EventCore{
						Name:      "such-instant",
						Timestamp: 0,
						ProcessID: &pid,
					},
					Scope: events.InstantScopeThread,
				}))
			})
		})

		Context("with stack traces", func() {
			JustBeforeEach(func() {
				tracer.Instant("such-instant", trace.WithStackTrace())
			})

			It("emits a sensible event", func() {
				Expect(eventWriter.events).To(HaveLen(1))
				e, ok := eventWriter.lastEvent().(*events.Instant)
				Expect(ok).To(BeTrue())
				Expect(e.StackTrace.Trace).ToNot(BeEmpty())
			})
		})
	})
})
