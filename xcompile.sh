#!/bin/sh
export GOPATH="`pwd`"
ls -1 src/giss/exec | while read row ; do
  GOOS=linux GOARCH=amd64 go install -ldflags '-s -w' giss/exec/$row
  GOOS=windows GOARCH=amd64 go install -ldflags '-s -w' giss/exec/$row
  GOOS=darwin GOARCH=amd64 go install -ldflags '-s -w' giss/exec/$row
done
