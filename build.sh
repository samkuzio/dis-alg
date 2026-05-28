#!/bin/bash
CGO_ENABLED=0 go build -ldflags="-w -s" -o dis-alg ./cmd/dis-alg/
