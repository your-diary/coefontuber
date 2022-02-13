#!/usr/bin/env bash

set -x

go fmt ./... &&
go vet ./... &&
golint ./... | grep --color=never -v -e 'exported type .* should have comment or be unexported' -e 'exported function .* should have comment or be unexported'

