package main

import (
	"fmt"
	"os"
	"phabricator.wikimedia.org/source/blubber/config"
	"phabricator.wikimedia.org/source/blubber/docker"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: blubber config.yaml variant")
		os.Exit(1)
	}

	cfg, err := config.ReadConfigFile(os.Args[1])

	if err != nil {
		fmt.Println("Error reading config:\n", err)
		os.Exit(2)
	}

	docker.Compile(cfg, os.Args[2]).WriteTo(os.Stdout)
}
