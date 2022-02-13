#!/usr/bin/env bash

set -x

find . -name '*.out' -delete &&
go fmt ./... &&
go vet ./... &&
golint ./... | grep --color=never -v -e 'exported type .* should have comment or be unexported' -e 'exported function .* should have comment or be unexported'

