package build

// Phase enum type
type Phase int

// Distinct build phases that each compiler implementation should pass to
// PhaseCompileable configuration (in the order they are defined here) to
// allow for dependency injection during compilation.
//
const (
	PhasePrivileged       Phase = iota // first, copies/execution done as root
	PhasePrivilegeDropped              // second, copies/execution done as unprivileged user from here on
	PhasePreInstall                    // third, before application files and artifacts are copied
	PhaseInstall                       // fourth, application files and artifacts are copied
	PhasePostInstall                   // fifth, after application files and artifacts are copied
)

// PhaseCompileable defines and interface that all configuration types should
// implement if they want to inject build instructions into any of the defined
// build phases.
//
type PhaseCompileable interface {
	InstructionsForPhase(phase Phase) []Instruction
}
