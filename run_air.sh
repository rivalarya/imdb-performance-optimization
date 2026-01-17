#!/usr/bin/env bash

# Detect OS
OS="$(uname -s)"

# Set binary output name based on OS
if [[ "$OS" == "MINGW"* || "$OS" == "CYGWIN"* || "$OS" == "MSYS_NT"* ]]; then
    BINARY="./tmp/main.exe"
else
    BINARY="./tmp/main"
fi

# Build air config file dynamically (optional)
cat <<EOF > .air.toml
root = "."
tmp_dir = "tmp"

[build]
  bin = "$BINARY"
  cmd = "go build -o $BINARY ."
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_regex = ["_test.go"]
  include_ext = ["go", "tpl", "tmpl", "html"]
  log = "build-errors.log"

[color]
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
EOF

# Run air
echo "Running air with binary: $BINARY"
air
