#!/bin/bash
protoc --go_out=plugins=grpc:. -I $GOPATH/src:. -I/usr/local/include *.proto
