package docker

import (
	"bytes"
	"strings"
	"github.com/marxarelli/blubber/build"
	"github.com/marxarelli/blubber/config"
)

func Compile(cfg *config.Config, variant string) *bytes.Buffer {
	buffer := new(bytes.Buffer)

	vcfg, err := config.ExpandVariant(cfg, variant)

	if err == nil {
		// write multi-stage sections for each artifact dependency
		for _, artifact := range vcfg.Artifacts {
			if artifact.From != "" {
				dependency, err := config.ExpandVariant(cfg, artifact.From)

				if err == nil {
					CompileStage(buffer, artifact.From, dependency)
				}
			}
		}

		CompileStage(buffer, variant, vcfg)
	}

	return buffer
}

func CompileStage(buffer *bytes.Buffer, stage string, vcfg *config.VariantConfig) {
	Writeln(buffer, "FROM ", vcfg.Base, " AS ", stage)

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

	if vcfg.SharedVolume {
		Writeln(buffer, "VOLUME [\"", vcfg.Runs.In, "\"]")
  } else {
		Writeln(buffer, "COPY . \"", vcfg.Runs.In, "\"")
	}

	// Artifact copying
	for _, artifact := range vcfg.Artifacts {
		Write(buffer, "COPY ")

		if artifact.From != "" {
			Write(buffer, "--from=", artifact.From, " ")
		}

		Writeln(buffer, artifact.Source, " ", artifact.Destination)
	}

	CompilePhase(buffer, vcfg, build.PhasePostInstall)

	if len(vcfg.EntryPoint) > 0 {
		Writeln(buffer, "ENTRYPOINT [\"", strings.Join(vcfg.EntryPoint, "\", \""), "\"]")
	}
}

func CompilePhase(buffer *bytes.Buffer, vcfg *config.VariantConfig, phase build.Phase) {
	for _, instruction := range vcfg.InstructionsForPhase(phase) {
		CompileInstruction(buffer, instruction)
	}
}

func CompileInstruction(buffer *bytes.Buffer, instruction build.Instruction) {
	switch instruction.Type {
	case build.Run:
		Writeln(buffer, append([]string{"RUN "}, instruction.Arguments...)...)
	case build.Copy:
		Writeln(buffer, "COPY \"", instruction.Arguments[0], "\" \"", instruction.Arguments[1], "\"")
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
