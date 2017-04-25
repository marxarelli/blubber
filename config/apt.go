package config

import (
	"bytes"
	"strings"
)

type AptConfig struct {
	Packages []string `yaml:"packages"`
}

func (apt *AptConfig) Merge(apt2 AptConfig) {
	apt.Packages = append(apt.Packages, apt2.Packages...)
}

func (apt AptConfig) Commands() []string {
	if len(apt.Packages) < 1 {
		return []string{}
	}

	buffer := new(bytes.Buffer)

	buffer.WriteString("apt-get update && apt-get install -y ")
	buffer.WriteString(strings.Join(apt.Packages, " "))
	buffer.WriteString(" && rm -rf /var/lib/apt/lists/*")

	return []string{buffer.String()}
}
