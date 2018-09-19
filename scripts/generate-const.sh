#!/bin/bash
#
# Generates a Go const from the contents of a static file, and appends it at
# the line following the go:generate directive that calls this script.
#
set -euo pipefail

sed -i.sed.bak -e "$((GOLINE+1)),\$d" "$GOFILE"
rm "$GOFILE.sed.bak"

(echo -n "const $1 = \`"; cat "$2"; echo "\`") >> "$GOFILE"
