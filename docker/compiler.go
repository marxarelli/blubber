package docker

import (
	"bytes"
	"fmt"
	"github.com/marxarelli/blubber/config"
)

func Compile(cfg config.ConfigType, variant string) []byte {
	buffer := new(bytes.Buffer)

	buffer.WriteString("FROM ")
}

func Compile(base config.Base) []byte {

}

func Quote(arg string) {
}

func Write()
