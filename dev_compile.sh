#!/bin/sh
NOW=`date '+%Y-%m-%d_%H:%M:%S'`
MSG="Alpha"
export GOPATH="`pwd`"
ls -1 src/giss/exec | while read row ; do
  GOOS=linux GOARCH=amd64 go install -ldflags "-s -w -X giss/values.DevStr=${MSG}_${NOW}" giss/exec/$row
  GOOS=windows GOARCH=amd64 go install -ldflags "-s -w -X giss/values.DevStr=${MSG}_${NOW}" giss/exec/$row
  GOOS=darwin GOARCH=amd64 go install -ldflags "-s -w -X giss/values.DevStr=${MSG}_${NOW}" giss/exec/$row
done
