package main

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlubberoidOpenAPISpecTemplateMatchesFile(t *testing.T) {
	specFile, err := ioutil.ReadFile("../../api/openapi-spec/blubberoid.yaml")

	if assert.NoError(t, err) {
		assert.Equal(t, string(specFile), openAPISpecTemplate)
	}
}
