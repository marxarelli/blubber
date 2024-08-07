// Package main provides the command line interface.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/pborman/getopt/v2"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/buildkit"
	"gitlab.wikimedia.org/repos/releng/blubber/config"
	"gitlab.wikimedia.org/repos/releng/blubber/meta"
)

const parameters = "config.yaml variant"

var (
	showHelp    = getopt.BoolLong("help", 'h', "show help/usage")
	policyURI   = getopt.StringLong("policy", 'p', "", "policy file URI", "uri")
	showVersion = getopt.BoolLong("version", 'v', "show version information")
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

	err = config.ExpandIncludesAndCopies(cfg, variant)
	if err != nil {
		if config.IsValidationError(err) {
			log.Printf("%s is invalid:\n%v", cfgPath, config.HumanizeValidationError(err))
			os.Exit(4)
		} else {
			log.Printf("Error: Failed to process config for '%s': %s\n", variant, err)
			os.Exit(3)
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

	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	defer func() {
		signal.Stop(ch)
		cancel()
	}()

	go func() {
		select {
		case <-ch:
			cancel()
		case <-ctx.Done():
		}
	}()

	opts := buildkit.BuildOptions{
		Options: build.NewOptions(),
	}
	opts.Variant = variant

	target, err := buildkit.Compile(ctx, &opts, cfg, nil)

	if err != nil {
		log.Printf("Error compiling config: %v\n", err)
		os.Exit(3)
	}

	err = target.WriteTo(ctx, os.Stdout)

	if err != nil {
		log.Printf("Error marshaling target: %v\n", err)
		os.Exit(3)
	}
}
