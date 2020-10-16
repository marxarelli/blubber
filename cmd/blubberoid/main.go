// Package main provides the blubberoid server.
//
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/pborman/getopt/v2"

	"gerrit.wikimedia.org/r/blubber/config"
	"gerrit.wikimedia.org/r/blubber/docker"
	"gerrit.wikimedia.org/r/blubber/meta"
)

var (
	showHelp    = getopt.BoolLong("help", 'h', "show help/usage")
	address     = getopt.StringLong("address", 'a', ":8748", "socket address/port to listen on (default ':8748')", "address:port")
	endpoint    = getopt.StringLong("endpoint", 'e', "/", "server endpoint (default '/')", "path")
	policyURI   = getopt.StringLong("policy", 'p', "", "policy file URI", "uri")
	policy      *config.Policy
	openAPISpec []byte
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

	setup()

	log.Printf("listening on %s for requests to %sv1/[variant]\n", *address, *endpoint)

	http.HandleFunc(*endpoint, blubberoid)
	log.Fatal(http.ListenAndServe(*address, nil))
}

func blubberoid(res http.ResponseWriter, req *http.Request) {
	if len(req.URL.Path) <= len(*endpoint) {
		if req.URL.RawQuery == "spec" {
			res.Header().Set("Content-Type", "text/plain")
			res.Write(openAPISpec)
			return
		}

		res.WriteHeader(http.StatusNotFound)
		res.Write(responseBody("request a variant at %sv1/[variant]", *endpoint))
		return
	}

	requestPath := req.URL.Path[len(*endpoint):]
	pathSegments := strings.Split(requestPath, "/")

	// Request should have been to v1/[variant]
	if len(pathSegments) != 2 || pathSegments[0] != "v1" {
		res.WriteHeader(http.StatusNotFound)
		res.Write(responseBody("request a variant at %sv1/[variant]", *endpoint))
		return
	}

	variant, err := url.PathUnescape(pathSegments[1])

	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed to unescape variant name '%s': %s\n", pathSegments[1], err)
		return
	}

	body, err := ioutil.ReadAll(req.Body)

	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed to read request body: %s\n", err)
		return
	}

	var cfg *config.Config
	mediaType, _, _ := mime.ParseMediaType(req.Header.Get("content-type"))

	// Default to application/json
	if mediaType == "" {
		mediaType = "application/json"
	}

	switch mediaType {
	case "application/json":
		cfg, err = config.ReadConfig(body)
	case "application/yaml", "application/x-yaml":
		cfg, err = config.ReadYAMLConfig(body)
	default:
		res.WriteHeader(http.StatusUnsupportedMediaType)
		res.Write(responseBody("'%s' media type is not supported", mediaType))
		return
	}

	if err != nil {
		if config.IsValidationError(err) {
			res.WriteHeader(http.StatusUnprocessableEntity)
			res.Write(responseBody(config.HumanizeValidationError(err)))
			return
		}

		res.WriteHeader(http.StatusBadRequest)
		res.Write(responseBody(
			"Failed to read '%s' config from request body. Error: %s",
			mediaType,
			err.Error(),
		))
		return
	}

	err = config.ExpandIncludesAndCopies(cfg, variant)
	if err != nil {
		res.WriteHeader(http.StatusUnprocessableEntity)
		res.Write(responseBody(
			"Failed to process the config for '%v': Error: %s",
			variant, err,
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

func readOpenAPISpec() []byte {
	var buffer bytes.Buffer
	tmpl, _ := template.New("spec").Parse(openAPISpecTemplate)

	tmpl.Execute(&buffer, struct {
		Version string
	}{
		Version: meta.FullVersion(),
	})

	return buffer.Bytes()
}

func setup() {
	// Ensure endpoint is always an absolute path starting and ending with "/"
	*endpoint = path.Clean("/" + *endpoint)

	if *endpoint != "/" {
		*endpoint += "/"
	}

	// Evaluate OpenAPI spec template and store results for ?spec requests
	openAPISpec = readOpenAPISpec()
}
