#!/bin/sh

go build -ldflags "-s -w" -o license.dll --buildmode=c-shared .