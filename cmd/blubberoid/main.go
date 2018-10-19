// Package main provides the blubberoid server.
//
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path"

	"github.com/pborman/getopt/v2"

	"gerrit.wikimedia.org/r/blubber/config"
	"gerrit.wikimedia.org/r/blubber/docker"
)

var (
	showHelp  = getopt.BoolLong("help", 'h', "show help/usage")
	address   = getopt.StringLong("address", 'a', ":8748", "socket address/port to listen on (default ':8748')", "address:port")
	endpoint  = getopt.StringLong("endpoint", 'e', "/", "server endpoint (default '/')", "path")
	policyURI = getopt.StringLong("policy", 'p', "", "policy file URI", "uri")
	policy    *config.Policy
)

func main() {
	getopt.Parse()

	if *showHelp {
		getopt.Usage()
		os.Exit(1)
	}

	if *policyURI != "" {
		var err error

		policy, err = config.ReadPolicyFromURI(*policyURI)

		if err != nil {
			log.Fatalf("Error loading policy from %s: %v\n", *policyURI, err)
		}
	}

	// Ensure endpoint is always an absolute path starting and ending with "/"
	*endpoint = path.Clean("/" + *endpoint)

	if *endpoint != "/" {
		*endpoint += "/"
	}

	log.Printf("listening on %s for requests to %s[variant]\n", *address, *endpoint)

	http.HandleFunc(*endpoint, blubberoid)
	log.Fatal(http.ListenAndServe(*address, nil))
}

func blubberoid(res http.ResponseWriter, req *http.Request) {
	if len(req.URL.Path) <= len(*endpoint) {
		res.WriteHeader(http.StatusNotFound)
		res.Write(responseBody("request a variant at %s[variant]", *endpoint))
		return
	}

	variant := req.URL.Path[len(*endpoint):]
	body, err := ioutil.ReadAll(req.Body)

	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed to read request body: %s\n", err)
		return
	}

	switch mt, _, _ := mime.ParseMediaType(req.Header.Get("content-type")); mt {
	case "application/json":
		// Enforce strict JSON syntax if specified, even though the config parser
		// would technically handle anything that's at least valid YAML
		if !json.Valid(body) {
			res.WriteHeader(http.StatusBadRequest)
			res.Write(responseBody("'%s' media type given but request contains invalid JSON", mt))
			return
		}
	case "application/yaml", "application/x-yaml":
		// Let the config parser validate YAML syntax
	default:
		res.WriteHeader(http.StatusUnsupportedMediaType)
		res.Write(responseBody("'%s' media type is not supported", mt))
		return
	}

	cfg, err := config.ReadYAMLConfig(body)

	if err != nil {
		if config.IsValidationError(err) {
			res.WriteHeader(http.StatusUnprocessableEntity)
			res.Write(responseBody(config.HumanizeValidationError(err)))
			return
		}

		res.WriteHeader(http.StatusBadRequest)
		res.Write(responseBody(
			"Failed to read config YAML from request body. "+
				"Was it formatted correctly and encoded as binary data?\nerror: %s",
			err.Error(),
		))
		return
	}

	if policy != nil {
		err = policy.Validate(*cfg)

		if err != nil {
			res.WriteHeader(http.StatusUnprocessableEntity)
			res.Write(responseBody(
				"Configuration fails policy check against:\npolicy: %s\nviolation: %v",
				*policyURI, err,
			))
			return
		}
	}

	dockerFile, err := docker.Compile(cfg, variant)

	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		res.Write(responseBody(err.Error()))
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.Write(dockerFile.Bytes())
}

func responseBody(msg string, a ...interface{}) []byte {
	return []byte(fmt.Sprintf(msg+"\n", a...))
}
