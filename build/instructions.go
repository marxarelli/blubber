package build

type InstructionType int

const (
	Run InstructionType = iota
	Copy
)

type Instruction struct {
	Type InstructionType
	Arguments []string
}
