#!/bin/bash

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./scripts/remote/remote_sim/simulation simulation.go simtests.go client.go
