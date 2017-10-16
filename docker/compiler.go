package docker

import (
	"bytes"
	"strings"

	"phabricator.wikimedia.org/source/blubber/build"
	"phabricator.wikimedia.org/source/blubber/config"
	"phabricator.wikimedia.org/source/blubber/meta"
)

// Compile blubber yaml file into Dockerfile
func Compile(cfg *config.Config, variant string) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)

	vcfg, err := config.ExpandVariant(cfg, variant)

	if err != nil {
		return nil, err
	}

	// omit the main stage name unless multi-stage is required below
	mainStage := ""

	// write multi-stage sections for each variant dependency
	for _, stage := range vcfg.VariantDependencies() {
		dependency, err := config.ExpandVariant(cfg, stage)

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
	baseAndStage := vcfg.Base

	if stage != "" {
		baseAndStage += " AS " + stage
	}

	writeln(buffer, "FROM ", baseAndStage)

	writeln(buffer, "USER root")

	compilePhase(buffer, vcfg, build.PhasePrivileged)

	if vcfg.Runs.As != "" {
		writeln(buffer, "USER ", vcfg.Runs.As)
	}

	compilePhase(buffer, vcfg, build.PhasePrivilegeDropped)

	if vcfg.Runs.In != "" {
		writeln(buffer, "WORKDIR ", vcfg.Runs.In)
	}

	compilePhase(buffer, vcfg, build.PhasePreInstall)

	compilePhase(buffer, vcfg, build.PhaseInstall)

	compilePhase(buffer, vcfg, build.PhasePostInstall)

	if len(vcfg.EntryPoint) > 0 {
		writeln(buffer, "ENTRYPOINT [\"", strings.Join(vcfg.EntryPoint, "\", \""), "\"]")
	}
}

func compileInstructions(buffer *bytes.Buffer, instructions ...build.Instruction) {
	for _, instruction := range instructions {
		dockerInstruction, _ := NewDockerInstruction(instruction)
		write(buffer, dockerInstruction.Compile())
	}
}

func compilePhase(buffer *bytes.Buffer, vcfg *config.VariantConfig, phase build.Phase) {
	compileInstructions(buffer, vcfg.InstructionsForPhase(phase)...)
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
