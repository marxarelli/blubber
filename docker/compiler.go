package docker

import (
	"bytes"
	"log"
	"strings"

	"phabricator.wikimedia.org/source/blubber/build"
	"phabricator.wikimedia.org/source/blubber/config"
)

func Compile(cfg *config.Config, variant string) *bytes.Buffer {
	buffer := new(bytes.Buffer)

	vcfg, err := config.ExpandVariant(cfg, variant)

	if err != nil {
		log.Fatal(err)
	}

	// omit the main stage name unless multi-stage is required below
	mainStage := ""

	// write multi-stage sections for each variant dependency
	for _, stage := range vcfg.VariantDependencies() {
		dependency, err := config.ExpandVariant(cfg, stage)

		if err == nil {
			CompileStage(buffer, stage, dependency)
			mainStage = variant
		}
	}

	CompileStage(buffer, mainStage, vcfg)

	return buffer
}

func CompileStage(buffer *bytes.Buffer, stage string, vcfg *config.VariantConfig) {
	baseAndStage := vcfg.Base

	if stage != "" {
		baseAndStage += " AS " + stage
	}

	Writeln(buffer, "FROM ", baseAndStage)

	Writeln(buffer, "USER root")

	CompilePhase(buffer, vcfg, build.PhasePrivileged)

	if vcfg.Runs.As != "" {
		Writeln(buffer, "USER ", vcfg.Runs.As)
	}

	CompilePhase(buffer, vcfg, build.PhasePrivilegeDropped)

	if vcfg.Runs.In != "" {
		Writeln(buffer, "WORKDIR ", vcfg.Runs.In)
	}

	CompilePhase(buffer, vcfg, build.PhasePreInstall)

	CompilePhase(buffer, vcfg, build.PhaseInstall)

	CompilePhase(buffer, vcfg, build.PhasePostInstall)

	if len(vcfg.EntryPoint) > 0 {
		Writeln(buffer, "ENTRYPOINT [\"", strings.Join(vcfg.EntryPoint, "\", \""), "\"]")
	}
}

func CompilePhase(buffer *bytes.Buffer, vcfg *config.VariantConfig, phase build.Phase) {
	for _, instruction := range vcfg.InstructionsForPhase(phase) {
		dockerInstruction, _ := NewDockerInstruction(instruction)
		Write(buffer, dockerInstruction.Compile())
	}
}

func Write(buffer *bytes.Buffer, strings ...string) {
	for _, str := range strings {
		buffer.WriteString(str)
	}
}

func Writeln(buffer *bytes.Buffer, strings ...string) {
	Write(buffer, strings...)
	buffer.WriteString("\n")
}
