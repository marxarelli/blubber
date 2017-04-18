package main

import (
	"fmt"
	"os"
	"github.com/davecgh/go-spew/spew"
	"github.com/marxarelli/blubber/config"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: blubber config.json")
		os.Exit(1)
	}

	config, err := config.ReadConfigFile(os.Args[1])

	if err != nil {
		fmt.Println("Error reading config:\n", err)
		os.Exit(2)
	} 

	spew.Dump(config)
}
