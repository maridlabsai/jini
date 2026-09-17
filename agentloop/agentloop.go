// Package agentloop is the PUBLIC extension seam for Jini's native agentic
// loop. The paid decision-tree recorder and git checkpointer (commercial repo,
// for the Pro backtrack capability) implement these interfaces and register
// them; the public loop consults them per step. A pure public build registers
// nothing, so the loop runs unrecorded and uncheckpointed — behavior unchanged.
//
// See specs/native-agentic-loop-design.md and
// specs/decision-tree-backtrack-design.md.
package agentloop

// Step is one decision point in a loop run.
type Step struct {
	Index       int
	Tool        string
	Args        map[string]string
	Observation string
}

// DecisionRecorder records each step so the tree can be inspected/branched.
type DecisionRecorder interface {
	RecordStep(step Step)
}

// Checkpointer snapshots workspace state before a write step so a branch can
// rewind to it (git in the commercial implementation). Returning an error does
// not stop the loop; it is best-effort.
type Checkpointer interface {
	Checkpoint(label string) error
}

var (
	recorder     DecisionRecorder
	checkpointer Checkpointer
)

// RegisterRecorder installs the process-wide recorder (nil clears).
func RegisterRecorder(r DecisionRecorder) { recorder = r }

// RegisterCheckpointer installs the process-wide checkpointer (nil clears).
func RegisterCheckpointer(c Checkpointer) { checkpointer = c }

// RegisteredRecorder returns the installed recorder, if any.
func RegisteredRecorder() (DecisionRecorder, bool) { return recorder, recorder != nil }

// RegisteredCheckpointer returns the installed checkpointer, if any.
func RegisteredCheckpointer() (Checkpointer, bool) { return checkpointer, checkpointer != nil }
