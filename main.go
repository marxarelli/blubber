// Package main provides the command line interface.
//
package main

import (
	"fmt"
	"log"
	"os"

	"phabricator.wikimedia.org/source/blubber/config"
	"phabricator.wikimedia.org/source/blubber/docker"
	"phabricator.wikimedia.org/source/blubber/meta"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println(meta.FullVersion())
		os.Exit(0)
	}

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
