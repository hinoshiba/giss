#!/bin/sh
export GOPATH
GOPATH="`pwd`"
cd src/giss
dep ensure
dep status
