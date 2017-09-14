package main

import (
	"fmt"
	"log"
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
		log.Printf("Error reading config: %v\n", err)
		os.Exit(2)
	}

	dockerFile, err := docker.Compile(cfg, os.Args[2])

	if err != nil {
		log.Printf("Error compiling config: %v\n", err)
		os.Exit(3)
	}

	dockerFile.WriteTo(os.Stdout)
}
