package events

type Phase string

const (
	PhaseBeginDuration     Phase = "B"
	PhaseEndDuration       Phase = "E"
	PhaseComplete          Phase = "X"
	PhaseInstant           Phase = "I"
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

type Event interface {
	Phase() Phase
	Core() *EventCore
}

type StackFrame struct {
	Category string
	Name     string
	Parent   string
}

type StackTrace struct {
	Trace []*StackFrame
}

type EventCore struct {
	Name            string
	Categories      []string
	Timestamp       int64
	ThreadTimestamp *int64
	ProcessID       *int64
	ThreadID        *int64
}

func (ec *EventCore) Core() *EventCore {
	return ec
}

type EventWithArgs struct {
	EventCore
	Args map[string]interface{}
}

type BeginDuration struct {
	EventWithArgs
	StackTrace *StackTrace
}

func (BeginDuration) Phase() Phase { return PhaseBeginDuration }

type EndDuration struct {
	EventWithArgs
	StackTrace *StackTrace
}

func (EndDuration) Phase() Phase { return PhaseEndDuration }

type Complete struct {
	EventWithArgs
	StackTrace    *StackTrace
	EndStackTrace *StackTrace
}

func (Complete) Phase() Phase { return PhaseComplete }

type InstantScope string

const (
	InstantScopeThread  InstantScope = "t"
	InstantScopeProcess InstantScope = "p"
	InstantScopeGlobal  InstantScope = "g"
)

type Instant struct {
	EventCore
	Scope      InstantScope
	StackTrace *StackTrace
}

func (Instant) Phase() Phase { return PhaseInstant }

type Counter struct {
	EventCore
	Values map[string]float64
}

func (Counter) Phase() Phase { return PhaseCounter }

type AsyncBegin struct {
	EventWithArgs
	Id    string
	Scope string
}

func (AsyncBegin) Phase() Phase { return PhaseAsyncBegin }

type AsyncEnd struct {
	EventWithArgs
	Id    string
	Scope string
}

func (AsyncEnd) Phase() Phase { return PhaseAsyncEnd }

type AsyncInstant struct {
	EventWithArgs
	Id    string
	Scope string
}

func (AsyncInstant) Phase() Phase { return PhaseAsyncInstant }

type FlowStart struct {
	EventWithArgs
}

func (FlowStart) Phase() Phase { return PhaseFlowStart }

type FlowInstant struct {
	EventWithArgs
}

func (FlowInstant) Phase() Phase { return PhaseFlowInstant }

type BindingPoint int

const (
	BindingPointEnclosing BindingPoint = iota
	BindingPointNext
)

type FlowFinish struct {
	EventWithArgs
	BindingPoint BindingPoint
}

func (FlowFinish) Phase() Phase { return PhaseFlowFinish }

type ObjectCreated struct {
	EventCore
	Id string
}

func (ObjectCreated) Phase() Phase { return PhaseObjectCreated }

type ObjectSnapshot struct {
	EventWithArgs
	Id string
}

func (ObjectSnapshot) Phase() Phase { return PhaseObjectSnapshot }

type ObjectDeleted struct {
	EventCore
	Id string
}

func (ObjectDeleted) Phase() Phase { return PhaseObjectDeleted }

type MetadataKind string

const (
	MetadataKindProcessName      MetadataKind = "process_name"
	MetadataKindProcessLabels    MetadataKind = "process_labels"
	MetadataKindProcessSortIndex MetadataKind = "process_sort_index"
	MetadataKindThreadName       MetadataKind = "thread_name"
	MetadataKindThreadSortIndex  MetadataKind = "thread_sort_index"
)

type MetadataProcessName struct {
	EventCore
	ProcessName string
}

func (MetadataProcessName) Phase() Phase { return PhaseMetadata }

type MetadataThreadName struct {
	EventCore
	ThreadName string
}

func (MetadataThreadName) Phase() Phase { return PhaseMetadata }

type MetadataProcessLabels struct {
	EventCore
	Labels string
}

func (MetadataProcessLabels) Phase() Phase { return PhaseMetadata }

type MetadataProcessSortIndex struct {
	EventCore
	SortIndex int64
}

func (MetadataProcessSortIndex) Phase() Phase { return PhaseMetadata }

type MetadataThreadSortIndex struct {
	EventCore
	SortIndex int64
}

func (MetadataThreadSortIndex) Phase() Phase { return PhaseMetadata }

type MetadataMisc struct {
	EventWithArgs
}

func (MetadataMisc) Phase() Phase { return PhaseMetadata }

type GlobalMemoryDump struct {
	EventWithArgs
}

func (GlobalMemoryDump) Phase() Phase { return PhaseGlobalMemoryDump }

type ProcessMemoryDump struct {
	EventWithArgs
}

func (ProcessMemoryDump) Phase() Phase { return PhaseProcessMemoryDump }

type Mark struct {
	EventWithArgs
}

func (Mark) Phase() Phase { return PhaseMark }

type ClockSync struct {
	EventWithArgs
	SyncId  string
	IssueTs *int64
}

func (ClockSync) Phase() Phase { return PhaseClockSync }

type ContextEnter struct {
	EventWithArgs
	Id string
}

func (ContextEnter) Phase() Phase { return PhaseContextEnter }

type ContextExit struct {
	EventWithArgs
	Id string
}

func (ContextExit) Phase() Phase { return PhaseContextExit }

type LinkIds struct {
	EventWithArgs
	Id       string
	LinkedId string
}

func (LinkIds) Phase() Phase { return PhaseLinkIds }
