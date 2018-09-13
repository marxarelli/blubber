// Package main provides the blubberoid server.
//
package main

import (
	"fmt"
	"io/ioutil"
	"log"
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

	cfg, err := config.ReadConfig(body)

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
