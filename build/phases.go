package build

type Phase int

const (
	PhasePrivileged Phase = iota
	PhasePrivilegeDropped
	PhasePreInstall
	PhaseInstall
	PhasePostInstall
)

type PhaseCompileable interface {
	InstructionsForPhase(phase Phase) []Instruction
}
