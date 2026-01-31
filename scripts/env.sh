#!/bin/bash
# Source this file to set up the development environment
# Usage: source scripts/env.sh

export PATH="/c/Program Files (x86)/GnuWin32/bin:/c/Users/rober/sdk/go1.25.6/bin:$PATH"
export GOROOT="/c/Users/rober/sdk/go1.25.6"

echo "Environment configured:"
echo "  Go: $(go version)"
echo "  Make: $(make --version | head -1)"
