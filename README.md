# Funicular [![Build Status](https://travis-ci.com/defgenx/funicular.svg?branch=master)](https://travis-ci.com/Defgenx/funicular)
###### 01000110 01010101 01001110 01001001 01000011 01010101 01001100 01000001 01010010

## Run commands

```bash
$ export GO111MODULE=on
$ go get -v -t ./...
$ cp .env-example .env
$ cd cmd/<cmd>
$ go build ./<cmd>
```

## Run tests

```bash
$ export GO111MODULE=on
$ go get ./...
$ go get github.com/onsi/gomega
$ go install github.com/onsi/ginkgo/ginkgo
$ ginkgo -r --randomizeAllSpecs --randomizeSuites --race --trace
```