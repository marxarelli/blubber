// Package main provides the command line interface.
//
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pborman/getopt/v2"

	"gerrit.wikimedia.org/r/blubber/config"
	"gerrit.wikimedia.org/r/blubber/docker"
	"gerrit.wikimedia.org/r/blubber/meta"
)

const parameters = "config.yaml variant"

var (
	showHelp    = getopt.BoolLong("help", 'h', "show help/usage")
	showVersion = getopt.BoolLong("version", 'v', "show version information")
	policyURI   = getopt.StringLong("policy", 'p', "", "policy file URI", "uri")
)

func main() {
	getopt.SetParameters(parameters)
	getopt.Parse()

	if *showHelp {
		getopt.Usage()
		os.Exit(1)
	}

	if *showVersion {
		fmt.Println(meta.FullVersion())
		os.Exit(0)
	}

	args := getopt.Args()

	if len(args) < 2 {
		getopt.Usage()
		os.Exit(1)
	}

	cfgPath, variant := args[0], args[1]

	cfg, err := config.ReadConfigFile(cfgPath)

	if err != nil {
		if config.IsValidationError(err) {
			log.Printf("%s is invalid:\n%v", cfgPath, config.HumanizeValidationError(err))
			os.Exit(4)
		} else {
			log.Printf("Error reading %s: %v\n", cfgPath, err)
			os.Exit(2)
		}
	}

	if *policyURI != "" {
		policy, err := config.ReadPolicyFromURI(*policyURI)

		if err != nil {
			log.Printf("Error loading policy from %s: %v\n", *policyURI, err)
			os.Exit(5)
		}

		err = policy.Validate(*cfg)

		if err != nil {
			log.Printf("Configuration fails policy check against:\npolicy: %s\nviolation: %v\n", *policyURI, err)
			os.Exit(6)
		}
	}

	dockerFile, err := docker.Compile(cfg, variant)

	if err != nil {
		log.Printf("Error compiling config: %v\n", err)
		os.Exit(3)
	}

	dockerFile.WriteTo(os.Stdout)
}
