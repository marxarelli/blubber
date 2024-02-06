# Blubber configuration (version v4)

## `.version` _string_ (required)

Blubber configuration version.

## `.apt` _object_


### `.apt.packages` _array<string>_

Packages to install from APT sources of base image.

#### `.apt.packages[]` _string_

### `.apt.packages` _object_

Key-Value pairs of target release and packages to install from APT sources.

#### `.apt.packages.*` _array<string>_

The packages to install using the target release.

#### `.apt.packages.*[]` _string_

### `.apt.proxies` _array<object|string>_

HTTP/HTTPS proxies to use during package installation.


#### `.apt.proxies[]` _string_

Shorthand configuration of a proxy that applies to all sources of its protocol.

#### `.apt.proxies[]` _object_

Proxy for either all sources of a given protocol or a specific source.

#### `.apt.proxies[].source` _string_

APT source to which this proxy applies.

#### `.apt.proxies[].url` _string_ (required)

HTTP/HTTPS proxy URL.

### `.apt.sources` _array<object>_

Additional APT sources to configure prior to package installation.

#### `.apt.sources[]` _object_

APT source URL, distribution/release name, and components.

#### `.apt.sources[].components` _array<string>_

List of distribution components (e.g. main, contrib).

#### `.apt.sources[].components[]` _string_

#### `.apt.sources[].distribution` _string_

Debian distribution/release name (e.g. buster).

#### `.apt.sources[].url` _string_ (required)

APT source URL.

## `.base` _null|string_

Base image reference.

## `.builder` _object_

### `.builder.command` _array<string>_

Command and arguments of an arbitrary build command.

#### `.builder.command[]` _string_

### `.builder.requirements` _array<object|string>_


#### `.builder.requirements[]` _string_

Path of files/directories to copy from the local build context.

#### `.builder.requirements[]` _object_

#### `.builder.requirements[].destination` _string_

Destination path. Defaults to source path.

#### `.builder.requirements[].from` _null|string_

Variant from which to copy files.

#### `.builder.requirements[].source` _string_

Path of files/directories to copy.

## `.builders` _array<object>_

Multiple builders to be executed in an explicit order.


### `.builders[]` _object_

#### `.builders[].custom` _object_

#### `.builders[].custom.command` _array<string>_

Command and arguments of an arbitrary build command.

#### `.builders[].custom.command[]` _string_

#### `.builders[].custom.requirements` _array<object|string>_


#### `.builders[].custom.requirements[]` _string_

Path of files/directories to copy from the local build context.

#### `.builders[].custom.requirements[]` _object_

#### `.builders[].custom.requirements[].destination` _string_

Destination path. Defaults to source path.

#### `.builders[].custom.requirements[].from` _null|string_

Variant from which to copy files.

#### `.builders[].custom.requirements[].source` _string_

Path of files/directories to copy.

### `.builders[]` _object_

#### `.builders[].node` _object_

#### `.builders[].node.allow-dedupe-failure` _boolean_

Whether to allow npm dedupe to fail; can be used to temporarily unblock CI while conflicts are resolved.

#### `.builders[].node.env` _string_

Node environment (e.g. production, etc.).

#### `.builders[].node.requirements` _array<object|string>_


#### `.builders[].node.requirements[]` _string_

Path of files/directories to copy from the local build context.

#### `.builders[].node.requirements[]` _object_

#### `.builders[].node.requirements[].destination` _string_

Destination path. Defaults to source path.

#### `.builders[].node.requirements[].from` _null|string_

Variant from which to copy files.

#### `.builders[].node.requirements[].source` _string_

Path of files/directories to copy.

#### `.builders[].node.use-npm-ci` _boolean_

Whether to run npm ci instead of npm install.

### `.builders[]` _object_

#### `.builders[].php` _object_

#### `.builders[].php.production` _boolean_

Whether to inject the --no-dev flag into the install command.

#### `.builders[].php.requirements` _array<object|string>_


#### `.builders[].php.requirements[]` _string_

Path of files/directories to copy from the local build context.

#### `.builders[].php.requirements[]` _object_

#### `.builders[].php.requirements[].destination` _string_

Destination path. Defaults to source path.

#### `.builders[].php.requirements[].from` _null|string_

Variant from which to copy files.

#### `.builders[].php.requirements[].source` _string_

Path of files/directories to copy.

### `.builders[]` _object_

#### `.builders[].python` _object_

#### `.builders[].python.no-deps` _boolean_

Inject --no-deps into the pip install command. All transitive requirements thus must be explicitly listed in the requirements file. pip check will be run to verify all dependencies are fulfilled.

#### `.builders[].python.poetry` _object_

#### `.builders[].python.poetry.devel` _boolean_

Whether to install development dependencies or not when using Poetry.

#### `.builders[].python.poetry.version` _string_

Version constraint for installing Poetry package.

#### `.builders[].python.requirements` _array<object|string>_


#### `.builders[].python.requirements[]` _string_

Path of files/directories to copy from the local build context.

#### `.builders[].python.requirements[]` _object_

#### `.builders[].python.requirements[].destination` _string_

Destination path. Defaults to source path.

#### `.builders[].python.requirements[].from` _null|string_

Variant from which to copy files.

#### `.builders[].python.requirements[].source` _string_

Path of files/directories to copy.

#### `.builders[].python.use-system-flag` _boolean_

Whether to inject the --system flag into the install command.

#### `.builders[].python.version` _string_

Python binary present in the system (e.g. python3).

## `.entrypoint` _array<string>_

Runtime entry point command and arguments.

### `.entrypoint[]` _string_

## `.lives` _object_

### `.lives.as` _string_

Owner (name) of application files within the container.

### `.lives.gid` _integer_

Group owner (GID) of application files within the container.

### `.lives.in` _string_

Application working directory within the container.

### `.lives.uid` _integer_

Owner (UID) of application files within the container.

## `.node` _object_

### `.node.allow-dedupe-failure` _boolean_

Whether to allow npm dedupe to fail; can be used to temporarily unblock CI while conflicts are resolved.

### `.node.env` _string_

Node environment (e.g. production, etc.).

### `.node.requirements` _array<object|string>_


#### `.node.requirements[]` _string_

Path of files/directories to copy from the local build context.

#### `.node.requirements[]` _object_

#### `.node.requirements[].destination` _string_

Destination path. Defaults to source path.

#### `.node.requirements[].from` _null|string_

Variant from which to copy files.

#### `.node.requirements[].source` _string_

Path of files/directories to copy.

### `.node.use-npm-ci` _boolean_

Whether to run npm ci instead of npm install.

## `.php` _object_

### `.php.production` _boolean_

Whether to inject the --no-dev flag into the install command.

### `.php.requirements` _array<object|string>_


#### `.php.requirements[]` _string_

Path of files/directories to copy from the local build context.

#### `.php.requirements[]` _object_

#### `.php.requirements[].destination` _string_

Destination path. Defaults to source path.

#### `.php.requirements[].from` _null|string_

Variant from which to copy files.

#### `.php.requirements[].source` _string_

Path of files/directories to copy.

## `.python` _object_

### `.python.no-deps` _boolean_

Inject --no-deps into the pip install command. All transitive requirements thus must be explicitly listed in the requirements file. pip check will be run to verify all dependencies are fulfilled.

### `.python.poetry` _object_

#### `.python.poetry.devel` _boolean_

Whether to install development dependencies or not when using Poetry.

#### `.python.poetry.version` _string_

Version constraint for installing Poetry package.

### `.python.requirements` _array<object|string>_


#### `.python.requirements[]` _string_

Path of files/directories to copy from the local build context.

#### `.python.requirements[]` _object_

#### `.python.requirements[].destination` _string_

Destination path. Defaults to source path.

#### `.python.requirements[].from` _null|string_

Variant from which to copy files.

#### `.python.requirements[].source` _string_

Path of files/directories to copy.

### `.python.use-system-flag` _boolean_

Whether to inject the --system flag into the install command.

### `.python.version` _string_

Python binary present in the system (e.g. python3).

## `.runs` _object_

### `.runs.as` _string_

Runtime process owner (name) of application entrypoint.

### `.runs.environment` _object_

Environment variables and values to be set before entrypoint execution.

### `.runs.gid` _integer_

Runtime process group (GID) of application entrypoint.

### `.runs.insecurely` _boolean_

Skip dropping of priviledge to the runtime process owner before entrypoint execution.

### `.runs.uid` _integer_

Runtime process owner (UID) of application entrypoint.

## `.variants` _object_

Configuration variants (e.g. development, test, production).

### `.variants.*` _object_

#### `.variants.*.apt` _object_


#### `.variants.*.apt.packages` _array<string>_

Packages to install from APT sources of base image.

#### `.variants.*.apt.packages[]` _string_

#### `.variants.*.apt.packages` _object_

Key-Value pairs of target release and packages to install from APT sources.

#### `.variants.*.apt.packages.*` _array<string>_

The packages to install using the target release.

#### `.variants.*.apt.packages.*[]` _string_

#### `.variants.*.apt.proxies` _array<object|string>_

HTTP/HTTPS proxies to use during package installation.


#### `.variants.*.apt.proxies[]` _string_

Shorthand configuration of a proxy that applies to all sources of its protocol.

#### `.variants.*.apt.proxies[]` _object_

Proxy for either all sources of a given protocol or a specific source.

#### `.variants.*.apt.proxies[].source` _string_

APT source to which this proxy applies.

#### `.variants.*.apt.proxies[].url` _string_ (required)

HTTP/HTTPS proxy URL.

#### `.variants.*.apt.sources` _array<object>_

Additional APT sources to configure prior to package installation.

#### `.variants.*.apt.sources[]` _object_

APT source URL, distribution/release name, and components.

#### `.variants.*.apt.sources[].components` _array<string>_

List of distribution components (e.g. main, contrib).

#### `.variants.*.apt.sources[].components[]` _string_

#### `.variants.*.apt.sources[].distribution` _string_

Debian distribution/release name (e.g. buster).

#### `.variants.*.apt.sources[].url` _string_ (required)

APT source URL.

#### `.variants.*.base` _null|string_

Base image reference.

#### `.variants.*.builder` _object_

#### `.variants.*.builder.command` _array<string>_

Command and arguments of an arbitrary build command.

#### `.variants.*.builder.command[]` _string_

#### `.variants.*.builder.requirements` _array<object|string>_


#### `.variants.*.builder.requirements[]` _string_

Path of files/directories to copy from the local build context.

#### `.variants.*.builder.requirements[]` _object_

#### `.variants.*.builder.requirements[].destination` _string_

Destination path. Defaults to source path.

#### `.variants.*.builder.requirements[].from` _null|string_

Variant from which to copy files.

#### `.variants.*.builder.requirements[].source` _string_

Path of files/directories to copy.

#### `.variants.*.builders` _array<object>_

Multiple builders to be executed in an explicit order.


#### `.variants.*.builders[]` _object_

#### `.variants.*.builders[].custom` _object_

#### `.variants.*.builders[].custom.command` _array<string>_

Command and arguments of an arbitrary build command.

#### `.variants.*.builders[].custom.command[]` _string_

#### `.variants.*.builders[].custom.requirements` _array<object|string>_


#### `.variants.*.builders[].custom.requirements[]` _string_

Path of files/directories to copy from the local build context.

#### `.variants.*.builders[].custom.requirements[]` _object_

#### `.variants.*.builders[].custom.requirements[].destination` _string_

Destination path. Defaults to source path.

#### `.variants.*.builders[].custom.requirements[].from` _null|string_

Variant from which to copy files.

#### `.variants.*.builders[].custom.requirements[].source` _string_

Path of files/directories to copy.

#### `.variants.*.builders[]` _object_

#### `.variants.*.builders[].node` _object_

#### `.variants.*.builders[].node.allow-dedupe-failure` _boolean_

Whether to allow npm dedupe to fail; can be used to temporarily unblock CI while conflicts are resolved.

#### `.variants.*.builders[].node.env` _string_

Node environment (e.g. production, etc.).

#### `.variants.*.builders[].node.requirements` _array<object|string>_


#### `.variants.*.builders[].node.requirements[]` _string_

Path of files/directories to copy from the local build context.

#### `.variants.*.builders[].node.requirements[]` _object_

#### `.variants.*.builders[].node.requirements[].destination` _string_

Destination path. Defaults to source path.

#### `.variants.*.builders[].node.requirements[].from` _null|string_

Variant from which to copy files.

#### `.variants.*.builders[].node.requirements[].source` _string_

Path of files/directories to copy.

#### `.variants.*.builders[].node.use-npm-ci` _boolean_

Whether to run npm ci instead of npm install.

#### `.variants.*.builders[]` _object_

#### `.variants.*.builders[].php` _object_

#### `.variants.*.builders[].php.production` _boolean_

Whether to inject the --no-dev flag into the install command.

#### `.variants.*.builders[].php.requirements` _array<object|string>_


#### `.variants.*.builders[].php.requirements[]` _string_

Path of files/directories to copy from the local build context.

#### `.variants.*.builders[].php.requirements[]` _object_

#### `.variants.*.builders[].php.requirements[].destination` _string_

Destination path. Defaults to source path.

#### `.variants.*.builders[].php.requirements[].from` _null|string_

Variant from which to copy files.

#### `.variants.*.builders[].php.requirements[].source` _string_

Path of files/directories to copy.

#### `.variants.*.builders[]` _object_

#### `.variants.*.builders[].python` _object_

#### `.variants.*.builders[].python.no-deps` _boolean_

Inject --no-deps into the pip install command. All transitive requirements thus must be explicitly listed in the requirements file. pip check will be run to verify all dependencies are fulfilled.

#### `.variants.*.builders[].python.poetry` _object_

#### `.variants.*.builders[].python.poetry.devel` _boolean_

Whether to install development dependencies or not when using Poetry.

#### `.variants.*.builders[].python.poetry.version` _string_

Version constraint for installing Poetry package.

#### `.variants.*.builders[].python.requirements` _array<object|string>_


#### `.variants.*.builders[].python.requirements[]` _string_

Path of files/directories to copy from the local build context.

#### `.variants.*.builders[].python.requirements[]` _object_

#### `.variants.*.builders[].python.requirements[].destination` _string_

Destination path. Defaults to source path.

#### `.variants.*.builders[].python.requirements[].from` _null|string_

Variant from which to copy files.

#### `.variants.*.builders[].python.requirements[].source` _string_

Path of files/directories to copy.

#### `.variants.*.builders[].python.use-system-flag` _boolean_

Whether to inject the --system flag into the install command.

#### `.variants.*.builders[].python.version` _string_

Python binary present in the system (e.g. python3).

#### `.variants.*.copies` _array<object|string>_


#### `.variants.*.copies[]` _string_

Variant from which to copy application and library files.

#### `.variants.*.copies[]` _object_

#### `.variants.*.copies[].destination` _string_

Destination path. Defaults to source path.

#### `.variants.*.copies[].from` _null|string_

Variant from which to copy files.

#### `.variants.*.copies[].source` _string_

Path of files/directories to copy.

#### `.variants.*.entrypoint` _array<string>_

Runtime entry point command and arguments.

#### `.variants.*.entrypoint[]` _string_

#### `.variants.*.includes` _array<string>_

Names of other variants to inherit configuration from.

#### `.variants.*.includes[]` _string_

Variant name.

#### `.variants.*.lives` _object_

#### `.variants.*.lives.as` _string_

Owner (name) of application files within the container.

#### `.variants.*.lives.gid` _integer_

Group owner (GID) of application files within the container.

#### `.variants.*.lives.in` _string_

Application working directory within the container.

#### `.variants.*.lives.uid` _integer_

Owner (UID) of application files within the container.

#### `.variants.*.node` _object_

#### `.variants.*.node.allow-dedupe-failure` _boolean_

Whether to allow npm dedupe to fail; can be used to temporarily unblock CI while conflicts are resolved.

#### `.variants.*.node.env` _string_

Node environment (e.g. production, etc.).

#### `.variants.*.node.requirements` _array<object|string>_


#### `.variants.*.node.requirements[]` _string_

Path of files/directories to copy from the local build context.

#### `.variants.*.node.requirements[]` _object_

#### `.variants.*.node.requirements[].destination` _string_

Destination path. Defaults to source path.

#### `.variants.*.node.requirements[].from` _null|string_

Variant from which to copy files.

#### `.variants.*.node.requirements[].source` _string_

Path of files/directories to copy.

#### `.variants.*.node.use-npm-ci` _boolean_

Whether to run npm ci instead of npm install.

#### `.variants.*.php` _object_

#### `.variants.*.php.production` _boolean_

Whether to inject the --no-dev flag into the install command.

#### `.variants.*.php.requirements` _array<object|string>_


#### `.variants.*.php.requirements[]` _string_

Path of files/directories to copy from the local build context.

#### `.variants.*.php.requirements[]` _object_

#### `.variants.*.php.requirements[].destination` _string_

Destination path. Defaults to source path.

#### `.variants.*.php.requirements[].from` _null|string_

Variant from which to copy files.

#### `.variants.*.php.requirements[].source` _string_

Path of files/directories to copy.

#### `.variants.*.python` _object_

#### `.variants.*.python.no-deps` _boolean_

Inject --no-deps into the pip install command. All transitive requirements thus must be explicitly listed in the requirements file. pip check will be run to verify all dependencies are fulfilled.

#### `.variants.*.python.poetry` _object_

#### `.variants.*.python.poetry.devel` _boolean_

Whether to install development dependencies or not when using Poetry.

#### `.variants.*.python.poetry.version` _string_

Version constraint for installing Poetry package.

#### `.variants.*.python.requirements` _array<object|string>_


#### `.variants.*.python.requirements[]` _string_

Path of files/directories to copy from the local build context.

#### `.variants.*.python.requirements[]` _object_

#### `.variants.*.python.requirements[].destination` _string_

Destination path. Defaults to source path.

#### `.variants.*.python.requirements[].from` _null|string_

Variant from which to copy files.

#### `.variants.*.python.requirements[].source` _string_

Path of files/directories to copy.

#### `.variants.*.python.use-system-flag` _boolean_

Whether to inject the --system flag into the install command.

#### `.variants.*.python.version` _string_

Python binary present in the system (e.g. python3).

#### `.variants.*.runs` _object_

#### `.variants.*.runs.as` _string_

Runtime process owner (name) of application entrypoint.

#### `.variants.*.runs.environment` _object_

Environment variables and values to be set before entrypoint execution.

#### `.variants.*.runs.gid` _integer_

Runtime process group (GID) of application entrypoint.

#### `.variants.*.runs.insecurely` _boolean_

Skip dropping of priviledge to the runtime process owner before entrypoint execution.

#### `.variants.*.runs.uid` _integer_

Runtime process owner (UID) of application entrypoint.
