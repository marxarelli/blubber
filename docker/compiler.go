// Package docker implements a compiler for turning Blubber configuration into
// a valid single- or multi-stage Dockerfile.
package docker

import (
	"bytes"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/config"
	"gitlab.wikimedia.org/repos/releng/blubber/meta"
)

// Compile takes a parsed config.Config and a configured variant name and
// returns the bytes of a resulting Dockerfile. In the case where artifacts
// are defined or the shorthand "copies" configured is set, a multi-stage
// Dockerfile will be returned.
func Compile(cfg *config.Config, variant string) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)

	vcfg, err := config.GetVariant(cfg, variant)

	if err != nil {
		return nil, err
	}

	// omit the main stage name unless multi-stage is required below
	mainStage := ""

	// write multi-stage sections for each variant dependency
	copiesDeps, err := cfg.CopiesDepGraph.GetDeps(variant)

	if err != nil {
		return nil, err
	}

	for _, stage := range copiesDeps {
		dependency, err := config.GetVariant(cfg, stage)

		if err != nil {
			return nil, err
		}

		compileStage(buffer, stage, dependency)
		mainStage = variant
	}

	compileStage(buffer, mainStage, vcfg)

	// add meta-data labels to the final stage
	compileInstructions(buffer, build.Label{map[string]string{
		"blubber.variant": variant,
		"blubber.version": meta.FullVersion(),
	}})

	return buffer, nil
}

func compileStage(buffer *bytes.Buffer, stage string, vcfg *config.VariantConfig) {
	for _, phase := range build.Phases() {
		compileInstructions(buffer, vcfg.InstructionsForPhase(phase)...)
	}

	// Add a blank line at the end of the stage for easier reading
	writeln(buffer)
}

func compileInstructions(buffer *bytes.Buffer, instructions ...build.Instruction) {
	for _, instruction := range instructions {
		dockerInstruction, _ := NewInstruction(instruction)
		write(buffer, dockerInstruction.Compile())
	}
}

func write(buffer *bytes.Buffer, strings ...string) {
	for _, str := range strings {
		buffer.WriteString(str)
	}
}

func writeln(buffer *bytes.Buffer, strings ...string) {
	write(buffer, strings...)
	buffer.WriteString("\n")
}
