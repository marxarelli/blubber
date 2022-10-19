// Package main provides the blubber-buildkit-frontend server.
//
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/moby/buildkit/frontend/gateway/grpcclient"
	"github.com/moby/buildkit/util/appcontext"
	"github.com/pborman/getopt/v2"

	"gitlab.wikimedia.org/repos/releng/blubber/buildkit"
	"gitlab.wikimedia.org/repos/releng/blubber/meta"
)

var (
	showHelp    = getopt.BoolLong("help", 'h', "show help/usage")
	showVersion = getopt.BoolLong("version", 'V', "show version")
)

func main() {
	getopt.Parse()

	if *showHelp {
		getopt.Usage()
		os.Exit(0)
	}

	if *showVersion {
		fmt.Println(meta.FullVersion())
		os.Exit(0)
	}

	err := grpcclient.RunFromEnvironment(appcontext.Context(), buildkit.Build)

	if err != nil {
		log.Panicf("fatal error:\n%v", err)
	}
}
