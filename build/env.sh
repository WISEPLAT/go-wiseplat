#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
wshdir="$workspace/src/github.com/wiseplat"
if [ ! -L "$wshdir/go-wiseplat" ]; then
    mkdir -p "$wshdir"
    cd "$wshdir"
    ln -s ../../../../../. go-wiseplat
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

# Run the command inside the workspace.
cd "$wshdir/go-wiseplat"
PWD="$wshdir/go-wiseplat"

# Launch the arguments with the configured environment.
exec "$@"
