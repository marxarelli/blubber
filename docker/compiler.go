package docker

import (
	"bytes"
	"strings"
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
	Writeln(buffer, "WORKDIR /srv")
	CompileToCommands(buffer, vcfg.Apt)
	CompileToCommands(buffer, vcfg.Runs)

	if vcfg.Runs.As != "" {
		Writeln(buffer, "USER ", vcfg.Runs.As)
	}

	if vcfg.Runs.In != "" {
		Writeln(buffer, "WORKDIR ", vcfg.Runs.In)
	}

	if vcfg.SharedVolume {
		Writeln(buffer, "VOLUME [\"", vcfg.Runs.In, "\"]")
  } else {
		Writeln(buffer, "COPY . \"", vcfg.Runs.In, "\"")
	}

	CompileToCommands(buffer, vcfg.Npm)

	// Artifact copying
	for _, artifact := range vcfg.Artifacts {
		Write(buffer, "COPY ")

		if artifact.From != "" {
			Write(buffer, "--from=", artifact.From, " ")
		}

		Writeln(buffer, artifact.Source, " ", artifact.Destination)
	}

	if len(vcfg.EntryPoint) > 0 {
		Writeln(buffer, "ENTRYPOINT [\"", strings.Join(vcfg.EntryPoint, "\", \""), "\"]")
	}
}

func CompileToCommands(buffer *bytes.Buffer, compileable config.CommandCompileable) {
	for _, command := range compileable.Commands() {
		Writeln(buffer, "RUN ", command)
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
