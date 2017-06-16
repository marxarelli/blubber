package build

type InstructionType int

const (
	Run InstructionType = iota
	Copy
	Env
)

type Instruction struct {
	Type InstructionType
	Arguments []string
}
