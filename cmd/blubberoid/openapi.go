// Code generated by ../../scripts/generate-const.sh DO NOT EDIT.
package main

//go:generate ../../scripts/generate-const.sh openAPISpecTemplate ../../api/openapi-spec/blubberoid.yaml
const openAPISpecTemplate = `---
openapi: '3.0.0'
info:
  title: Blubberoid
  description: >
    Blubber is a highly opinionated abstraction for container build
    configurations.
  version: {{ .Version }}
paths:
  /v1/{variant}:
    parameters:
      - name: variant
        description: Name of the variant to generate
        in: path
        required: true
        schema:
          type: string
    post:
      summary: >
        Generates a valid Dockerfile based on Blubber YAML configuration
        provided in the request body and the given variant name.
      requestBody:
        description: A valid Blubber configuration.
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/v4.Config'
          application/yaml:
            schema:
              type: string
          application/x-yaml:
            schema:
              type: string
      responses:
        '200':
          description: OK. Response body should be a valid Dockerfile.
          content:
            text/plain:
              schema:
                type: string
        '400':
          description: Bad request. The YAML request body failed to parse.
        '404':
          description: No variant name was provided in the request path.
        '422':
          description: Provided Blubber config parsed correctly but failed validation.
        '5XX':
          description: An unexpected service-side error.

      x-amples:
        - title: Mathoid test variant
          request:
            params:
              variant: test
            headers:
              Content-Type: application/json
            body: {
              "version": "v4",
              "base": "docker-registry.wikimedia.org/buster-nodejs10-slim",
              "apt": { "packages": ["librsvg2-2"] },
              "lives": { "in": "/srv/service" },
              "variants": {
                "build": {
                  "base": "docker-registry.wikimedia.org/buster-nodejs10-devel",
                  "apt": {
                    "packages": ["librsvg2-dev", "git", "pkg-config", "build-essential"]
                  },
                  "node": { "requirements": ["package.json"] },
                  "runs": { "environment": { "LINK": "g++" } }
                },
                "test": { "includes": ["build"], "entrypoint": ["npm", "test"] }
              }
            }
          response:
            status: 200
            headers:
              content-type: text/plain
            body: /^FROM docker-registry.wikimedia.org\/buster-nodejs10-devel/

components:
  schemas:
    v4.Config:
      title: Top-level blubber configuration (version v4)
      allOf:
        - $ref: '#/components/schemas/v4.CommonConfig'
        - type: object
          required: [version, variants]
          properties:
            version:
              type: string
              description: Blubber configuration version
            variants:
              type: object
              description: Configuration variants (e.g. development, test, production)
              additionalProperties:
                $ref: '#/components/schemas/v4.VariantConfig'

    v4.CommonConfig:
      type: object
      properties:
        base:
          type: string
          description: Base image reference
        apt:
          type: object
          properties:
            packages:
              type: array
              description: Packages to install from APT sources of base image
              items:
                type: string
        node:
          type: object
          properties:
            env:
              type: string
              description: Node environment (e.g. production, etc.)
            requirements:
              type: array
              description: Files needed for Node package installation (e.g. package.json, package-lock.json)
              items:
                type: string
            use-npm-ci:
              type: boolean
              description: Whether to run npm ci instead of npm install
        php:
          type: object
          properties:
            requirements:
              type: array
              description: Files needed for PHP package installation (e.g. composer.json)
              items:
                type: string
            production:
              type: boolean
              description: Whether to inject the --no-dev flag into the install command
        python:
          type: object
          properties:
            version:
              type: string
              description: Python binary present in the system (e.g. python3)
            requirements:
              type: array
              description: Files needed for Python package installation (e.g. requirements.txt, etc.)
              items:
                type: string
            use-system-flag:
              type: boolean
              description: Whether to inject the --system flag into the install command
            poetry:
              type: object
              properties:
                version:
                  type: string
                  description: Version constraint for installing Poetry package
                devel:
                  type: boolean
                  description: Whether to install development dependencies or not when using Poetry
        builder:
          type: object
          properties:
            command:
              type: array
              description: Command and arguments of an arbitrary build command
              items:
                type: string
            requirements:
              type: array
              description: Files needed by the build command (e.g. Makefile, ./src/, etc.)
              items:
                type: string
        lives:
          type: object
          properties:
            as:
              type: string
              description: Owner (name) of application files within the container
            uid:
              type: integer
              description: Owner (UID) of application files within the container
            gid:
              type: integer
              description: Group owner (GID) of application files within the container
            in:
              type: string
              description: Application working directory within the container
        runs:
          type: object
          properties:
            as:
              type: string
              description: Runtime process owner (name) of application entrypoint
            uid:
              type: integer
              description: Runtime process owner (UID) of application entrypoint
            gid:
              type: integer
              description: Runtime process group (GID) of application entrypoint
            environment:
              type: object
              description: Environment variables and values to be set before entrypoint execution
              additionalProperties: true
            insecurely:
              type: boolean
              description: Skip dropping of priviledge to the runtime process owner before entrypoint execution
        entrypoint:
          type: array
          description: Runtime entry point command and arguments
          items:
            type: string
    v4.VariantConfig:
      allOf:
        - $ref: '#/components/schemas/v4.CommonConfig'
        - type: object
          properties:
            includes:
              type: array
              description: Names of other variants to inherit configuration from
              items:
                description: Variant name
                type: string
            copies:
              type: string
              description: Name of variant from which to copy application files, resulting in a multi-stage build
            artifacts:
              type: array
              items:
                type: object
                description: Artifacts to copy from another variant, resulting in a multi-stage build
                required: [from, source, destination]
                properties:
                  from:
                    type: string
                    description: Variant name
                  source:
                    type: string
                    description: Path of files/directories to copy
                  destination:
                    type: string
                    description: Destination path
`
