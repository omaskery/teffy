package trace

import (
	"fmt"
	"github.com/omaskery/teffy/pkg/events"
	tio "github.com/omaskery/teffy/pkg/io"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/go-logr/logr"
)

type TracerOption = func(t *Tracer)

type ErrorHandler = func(err error)

type TimestampFn = func() int64

func WithLogger(logger logr.Logger) TracerOption {
	return func(t *Tracer) {
		t.logger = logger
	}
}

func WithErrorHandler(handler ErrorHandler) TracerOption {
	return func(t *Tracer) {
		t.errHandler = handler
	}
}

func WithTimestampFn(f TimestampFn) TracerOption {
	return func(t *Tracer) {
		t.timestampFn = f
	}
}

type Tracer struct {
	stream      tio.EventWriter
	logger      logr.Logger
	errHandler  ErrorHandler
	timestampFn TimestampFn
}

func NewTracer(stream tio.EventWriter, options ...TracerOption) *Tracer {
	t := &Tracer{
		stream:      stream,
		timestampFn: MillisecondTimestampFn,
	}
	for _, opt := range options {
		opt(t)
	}
	return t
}

func TracerToWriter(w io.WriteCloser, options ...TracerOption) *Tracer {
	return NewTracer(tio.NewStreamingWriter(w), options...)
}

func TraceToFile(path string, options ...TracerOption) (*Tracer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return TracerToWriter(f, options...), nil
}

func (t *Tracer) Close() error {
	if err := t.stream.Close(); err != nil {
		return fmt.Errorf("error closing stream writer: %w", err)
	}
	return nil
}

type EventOption = func(e events.Event)

func WithCategories(categories ...string) EventOption {
	return func(e events.Event) {
		e.Core().Categories = categories
	}
}

func WithArgs(args map[string]interface{}) EventOption {
	return func(e events.Event) {
		switch event := e.(type) {
		case events.ArgSetter:
			event.SetArgs(args)
		default:
			panic(fmt.Sprintf("cannot set arguments on this event type: %v", e))
		}
	}
}

func WithStackTrace() EventOption {
	return func(e events.Event) {
		switch event := e.(type) {
		case events.StackTraceSetter:
			event.SetStackTrace(buildStackTrace())
		default:
			panic(fmt.Sprintf("cannot set stack traces on this event type: %v", e))
		}
	}
}

func WithEndStackTrace() EventOption {
	return func(e events.Event) {
		switch event := e.(type) {
		case events.EndStackTraceSetter:
			event.SetEndStackTrace(buildStackTrace())
		default:
			panic(fmt.Sprintf("cannot set end stack traces on this event type: %v", e))
		}
	}
}

func buildStackTrace() *events.StackTrace {
	s := &events.StackTrace{
		Trace: nil,
	}

	// TODO: this probably shouldn't skip a hard coded number of stack levels ¯\_(ツ)_/¯
	stackLevelsToSkip := 5

	pc := make([]uintptr, 10)
	n := runtime.Callers(stackLevelsToSkip, pc)
	if n == 0 {
		return s
	}
	pc = pc[:n]

	frames := runtime.CallersFrames(pc)
	for {
		frame, more := frames.Next()

		s.Trace = append(s.Trace, &events.StackFrame{
			Category: frame.File,
			Name:     fmt.Sprintf("%s:%v", frame.Function, frame.Line),
		})

		if !more {
			break
		}
	}

	return s
}

type Duration struct {
	name string
	pid  int64
	t    *Tracer
}

func (t *Tracer) BeginDuration(name string, options ...EventOption) Duration {
	duration := Duration{
		name: name,
		pid:  getPid(),
		t:    t,
	}

	event := &events.BeginDuration{
		EventWithArgs: events.EventWithArgs{
			EventCore: events.EventCore{
				Name:      name,
				Timestamp: t.getTimestamp(),
				ProcessID: &duration.pid,
			},
		},
	}

	t.writeEvent(event, options...)

	return duration
}

func (d *Duration) End(options ...EventOption) {
	event := &events.EndDuration{
		EventWithArgs: events.EventWithArgs{
			EventCore: events.EventCore{
				Name:      d.name,
				Timestamp: d.t.getTimestamp(),
				ProcessID: &d.pid,
			},
		},
	}

	d.t.writeEvent(event, options...)
}

func (t *Tracer) Instant(name string, options ...EventOption) {
	t.ScopedInstant(name, events.InstantScopeThread, options...)
}

func (t *Tracer) ScopedInstant(name string, scope events.InstantScope, options ...EventOption) {
	pid := getPid()

	event := &events.Instant{
		EventCore: events.EventCore{
			Name:      name,
			Timestamp: t.getTimestamp(),
			ProcessID: &pid,
		},
		Scope: scope,
	}

	t.writeEvent(event, options...)
}

func (t *Tracer) writeEvent(e events.Event, options ...EventOption) {
	for _, opt := range options {
		opt(e)
	}

	err := t.stream.Write(e)
	if err != nil {
		t.handleError("failed to write begin duration event", err)
	}
}

func (t *Tracer) getTimestamp() int64 {
	return (t.timestampFn)()
}

func (t *Tracer) handleError(context string, err error) {
	if t.logger != nil {
		t.logger.Error(err, context)
	}
	err = fmt.Errorf("%s: %w", context, err)
	if t.errHandler != nil {
		(t.errHandler)(err)
	}
}

func MillisecondTimestampFn() int64 {
	nanoToMs := int64(1e6)
	return NanosecondTimestampFn() / nanoToMs
}

func NanosecondTimestampFn() int64 {
	return time.Now().UTC().UnixNano()
}

func getPid() int64 {
	return int64(os.Getpid())
}
