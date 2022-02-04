// events provides logical representations for trace events
package events

// Phase is the discriminator for identifying the type of an event in a Trace Event Format file
type Phase string

const (
	PhaseBeginDuration     Phase = "B"
	PhaseEndDuration       Phase = "E"
	PhaseComplete          Phase = "X"
	PhaseInstant           Phase = "I"
	PhaseInstantLegacy     Phase = "i"
	PhaseCounter           Phase = "C"
	PhaseAsyncBegin        Phase = "b"
	PhaseAsyncEnd          Phase = "e"
	PhaseAsyncInstant      Phase = "n"
	PhaseFlowStart         Phase = "s"
	PhaseFlowInstant       Phase = "t"
	PhaseFlowFinish        Phase = "f"
	PhaseObjectCreated     Phase = "N"
	PhaseObjectSnapshot    Phase = "O"
	PhaseObjectDeleted     Phase = "D"
	PhaseMetadata          Phase = "M"
	PhaseGlobalMemoryDump  Phase = "V"
	PhaseProcessMemoryDump Phase = "v"
	PhaseMark              Phase = "R"
	PhaseClockSync         Phase = "c"
	PhaseContextEnter      Phase = "("
	PhaseContextExit       Phase = ")"
	PhaseLinkIds           Phase = "="
)

// Event represents common information to all events
type Event interface {
	// Phase indicates the discriminator for identifying what kind of event this is, primarily for (un)marshaling
	Phase() Phase
	// Core provides mutable access to common event fields
	Core() *EventCore
}

// StackFrame represents a single entry in a stack trace, for inline traces only the Name field is populated,
// whereas formats that populate the StackFrames map these form a graph of stack frames
type StackFrame struct {
	// Category seems, from the examples, to often represent the filename that the symbol resides in
	Category string
	// Name seems, from examples, to often represent the current function of this stack frame
	Name string
	// Parent is an optional ID that, when present, indicates the outer (calling) stack frame in the StackFrames map
	Parent string
}

// StackTrace represents a full stack trace
type StackTrace struct {
	// Trace represents the individual frames of the stack trace starting from least recent to most recently called
	Trace []*StackFrame
}

// EventCore represents fields that are common to all events
type EventCore struct {
	// Name to associate with this event, often used with (Begin/End)Duration events to convey the current function name
	Name string
	// Categories is an optional collection of tags to help categorise events for filtering in viewers
	Categories []string
	// Timestamp is the event time in microseconds
	Timestamp int64
	// ThreadTimestamp is an optional timestamp to order events within a single thread
	ThreadTimestamp *int64
	// ProcessID is an optional identifier for the ID of the process that output this event
	ProcessID *int64
	// ThreadID is an optional identifier for the ID of the thread that output this event
	ThreadID *int64
}

// ArgSetter allows setting the arguments of events that allow it
type ArgSetter interface {
	// SetArgs sets the event arguments
	SetArgs(args map[string]interface{})
}

// StackTraceSetter allows setting the stack trace of events that allow it
type StackTraceSetter interface {
	// SetStackTrace sets the event stack trace
	SetStackTrace(trace *StackTrace)
}

// StackTraceSetter allows setting the ending stack trace of events that allow it
type EndStackTraceSetter interface {
	// SetEndStackTrace sets the event's end stack trace
	SetEndStackTrace(trace *StackTrace)
}

// Core provides mutable access to the common fields of events
func (ec *EventCore) Core() *EventCore {
	return ec
}

// EventWithArgs represents events that allow for a map of arbitrary arguments
type EventWithArgs struct {
	EventCore
	// Args are arbitrary values associated with the event by the traced software, though sometimes events
	// store their additional fields in the args
	Args map[string]interface{}
}

// SetArgs allows for events with arguments to have those arguments updated
func (e *EventWithArgs) SetArgs(args map[string]interface{}) {
	e.Args = args
}

// EventStackTrace represents the fields included in events that have a stack trace
type EventStackTrace struct {
	StackTrace *StackTrace
}

// SetStackTrace allows events with stack traces to have those stack traces updated
func (e *EventStackTrace) SetStackTrace(trace *StackTrace) {
	e.StackTrace = trace
}

// EventEndStackTrace represents the fields included in events that have an 'ending' stack trace
type EventEndStackTrace struct {
	EndStackTrace *StackTrace
}

// SetEndStackTrace allows events with ending stack traces to have those stack traces updated
func (e *EventEndStackTrace) SetEndStackTrace(trace *StackTrace) {
	e.EndStackTrace = trace
}

// BeginDuration represents the start of work on a given thread
type BeginDuration struct {
	EventWithArgs
	EventStackTrace
}

func (BeginDuration) Phase() Phase { return PhaseBeginDuration }

// EndDuration represents the end of work on a given thread
type EndDuration struct {
	EventWithArgs
	EventStackTrace
}

func (EndDuration) Phase() Phase { return PhaseEndDuration }

// Complete represents the start and end of work on a given thread
// primarily used to reduce the size of stored traces, given traces
// are made primarily of BeginDuration & EndDuration events
type Complete struct {
	EventWithArgs
	EventStackTrace
	EventEndStackTrace
	// Duration of the event in microseconds
	Duration int64
	// ThreadDuration is an optional duration of the event according to the thread clock
	ThreadDuration *int64
}

func (Complete) Phase() Phase { return PhaseComplete }

// InstantScope represents how widely an instantaneous event is relevant within a trace file
type InstantScope string

const (
	// InstantScopeThread means this instant event is only relevant to one thread of a single process
	InstantScopeThread InstantScope = "t"
	// InstantScopeThread means this instant event is relevant to one process, but across all threads in that process
	InstantScopeProcess InstantScope = "p"
	// InstantScopeThread means this instant event is relevant to the entire trace across all processes
	InstantScopeGlobal InstantScope = "g"
)

// Instant corresponds to something that happens but has no duration associated with it
type Instant struct {
	EventCore
	EventStackTrace
	// Scope indicates how widely this event is relevant, within the thread, process, or globally
	Scope InstantScope
}

func (Instant) Phase() Phase { return PhaseInstant }

// Counter is used to track one or more values as they change over time
type Counter struct {
	EventCore
	// Values records a snapshot of named values for tracking over time
	Values map[string]float64
}

func (Counter) Phase() Phase { return PhaseCounter }

// AsyncBegin represents the start of an asynchronous operation
type AsyncBegin struct {
	EventWithArgs
	// Id is a unique identifier to correlate the chain of causally related asynchronous events
	Id string
	// Scope is an optional extra component to the identifier to help prevent name collisions for common Id values
	Scope string
}

func (AsyncBegin) Phase() Phase { return PhaseAsyncBegin }

// AsyncEnd represents the end of an asynchronous operation
type AsyncEnd struct {
	EventWithArgs
	// Id is a unique identifier to correlate the chain of causally related asynchronous events
	Id string
	// Scope is an optional extra component to the identifier to help prevent name collisions for common Id values
	Scope string
}

func (AsyncEnd) Phase() Phase { return PhaseAsyncEnd }

// AsyncInstant represents an event with no duration that occurs as part of a chain of causally related async events
type AsyncInstant struct {
	EventWithArgs
	// Id is a unique identifier to correlate the chain of causally related asynchronous events
	Id string
	// Scope is an optional extra component to the identifier to help prevent name collisions for common Id values
	Scope string
}

func (AsyncInstant) Phase() Phase { return PhaseAsyncInstant }

// FlowStart is like an AsyncBegin but are used to represent links between Begin/End Duration events
type FlowStart struct {
	EventWithArgs
}

func (FlowStart) Phase() Phase { return PhaseFlowStart }

// FlowInstant is like an AsyncInstant but ... the documentation isn't particularly clear on what that means ^_^;
type FlowInstant struct {
	EventWithArgs
}

func (FlowInstant) Phase() Phase { return PhaseFlowInstant }

// BindingPoint indicates whether a FlowFinish event binds to the enclosing slice or the next slice
type BindingPoint int

const (
	// BindingPointEnclosing means the FlowFinish event will bind to the current slice enclosing this event
	BindingPointEnclosing BindingPoint = iota
	// BindingPointNext means the FlowFinish event will bind to the next slice after this event's timestamp
	BindingPointNext
)

// FlowFinish is like an AsyncEnd but is used to represent the links between Begin/End Duration events
type FlowFinish struct {
	EventWithArgs
	// BindingPoint indicates whether the event binds to the enclosing slice or next slice after this event
	// but defaults to the enclosing slice
	BindingPoint BindingPoint
}

func (FlowFinish) Phase() Phase { return PhaseFlowFinish }

// ObjectCreated allow for tracking the creation of complex data structures in trace
type ObjectCreated struct {
	EventCore
	// Id uniquely identifies the created object
	Id string
}

func (ObjectCreated) Phase() Phase { return PhaseObjectCreated }

// ObjectSnapshot allows for tracking the current state of a complex data structure in a trace
type ObjectSnapshot struct {
	EventWithArgs
	// Id uniquely identifies the object for which this event records the state
	Id string
}

func (ObjectSnapshot) Phase() Phase { return PhaseObjectSnapshot }

// ObjectDeleted allows for tracking the deletion of complex datastructures in the trace
type ObjectDeleted struct {
	EventCore
	// Id uniquely identifies the deleted object
	Id string
}

func (ObjectDeleted) Phase() Phase { return PhaseObjectDeleted }

// MetadataKind helps identify common well-known metadata values included in traces
type MetadataKind string

const (
	MetadataKindProcessName      MetadataKind = "process_name"
	MetadataKindProcessLabels    MetadataKind = "process_labels"
	MetadataKindProcessSortIndex MetadataKind = "process_sort_index"
	MetadataKindThreadName       MetadataKind = "thread_name"
	MetadataKindThreadSortIndex  MetadataKind = "thread_sort_index"
)

// MetadataProcessName is a metadata event conveying the name of the process the trace is from
type MetadataProcessName struct {
	EventCore
	ProcessName string
}

func (MetadataProcessName) Phase() Phase { return PhaseMetadata }

// MetadataThreadName is a metadata event conveying the name of the thread the trace is from
type MetadataThreadName struct {
	EventCore
	ThreadName string
}

func (MetadataThreadName) Phase() Phase { return PhaseMetadata }

// MetadataProcessLabels is a metadata event tagging a process with a convenient label
type MetadataProcessLabels struct {
	EventCore
	Labels string
}

func (MetadataProcessLabels) Phase() Phase { return PhaseMetadata }

// MetadataProcessSortIndex is a metadata event helping control the order that processes are drawn in a Trace Viewer
type MetadataProcessSortIndex struct {
	EventCore
	// SortIndex informs the order of processes drawn in the Trace Viewer, lower numbers are higher on the screen
	SortIndex int64
}

func (MetadataProcessSortIndex) Phase() Phase { return PhaseMetadata }

// MetadataThreadSortIndex is a metadata event helping control the order that threads are drawn in a Trace Viewer
type MetadataThreadSortIndex struct {
	EventCore
	// SortIndex informs the order of threads drawn in the Trace Viewer, lower numbers are higher on the screen
	SortIndex int64
}

func (MetadataThreadSortIndex) Phase() Phase { return PhaseMetadata }

// MetadataMisc is metadata that is not well known and so no attempt to decode its values has been performed
type MetadataMisc struct {
	EventWithArgs
}

func (MetadataMisc) Phase() Phase { return PhaseMetadata }

// GlobalMemoryDump events convey system memory information such as the size of RAM
type GlobalMemoryDump struct {
	EventWithArgs
}

func (GlobalMemoryDump) Phase() Phase { return PhaseGlobalMemoryDump }

// ProcessMemoryDump events convey information about a single processes memory usage
type ProcessMemoryDump struct {
	EventWithArgs
}

func (ProcessMemoryDump) Phase() Phase { return PhaseProcessMemoryDump }

// Mark events are for Chrome's "navigation timing API"
type Mark struct {
	EventWithArgs
}

func (Mark) Phase() Phase { return PhaseMark }

// ClockSync events are used to try and synchronise clock domains of multiple trace logs from different tracing agents
type ClockSync struct {
	EventWithArgs
	// SyncId is an identifier used to identify the ClockSync event in the issuing and receiving tracing agents' events
	SyncId string
	// IssueTs is a measurement of the time the receiver spent recording the ClockSync event for improving accuracy
	IssueTs *int64
}

func (ClockSync) Phase() Phase { return PhaseClockSync }

// ContextEnter denotes following events as belonging to a given context until a matching ContextExit event
type ContextEnter struct {
	EventWithArgs
	// Id uniquely identifies the context that is being entered
	Id string
}

func (ContextEnter) Phase() Phase { return PhaseContextEnter }

// ContextExit causes events to stop being associated with a context entered by the corresponding ContextEnter event
type ContextExit struct {
	EventWithArgs
	// Id uniquely identifying the context that has been exited
	Id string
}

func (ContextExit) Phase() Phase { return PhaseContextExit }

// LinkIds is used to indicate that two Ids are identical
type LinkIds struct {
	EventWithArgs
	// Id is one of the Ids that is being specified as equivalent
	Id string
	// LinkedId is the second of the Ids that is being marked as equivalent
	LinkedId string
}

func (LinkIds) Phase() Phase { return PhaseLinkIds }
