#!/bin/bash
go get -u github.com/golang/lint/golint || exit 1
go get golang.org/x/tools/cmd/goimports || exit 1
go get github.com/jgautheron/goconst/cmd/goconst || exit 1
go get github.com/jgautheron/usedexports || exit 1
go get github.com/kisielk/errcheck || exit 1
echo 'OK'
