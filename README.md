# Funicular
[![GoDoc](https://godoc.org/github.com/defgenx/funicular?status.svg)](http://godoc.org/github.com/defgenx/funicular) [![Build Status](https://travis-ci.com/defgenx/funicular.svg?branch=master)](https://travis-ci.com/defgenx/funicular) [![codecov](https://codecov.io/gh/defgenx/funicular/branch/master/graph/badge.svg)](https://codecov.io/gh/defgenx/funicular) [![license](https://img.shields.io/github/license/defgenx/funicular.svg?maxAge=2592000)](https://github.com/defgenx/funicular/LICENSE)

###### 01000110 01010101 01001110 01001001 01000011 01010101 01001100 01000001 01010010

A WIP simple clients wrapper to create commands.

## Run commands

```bash
$ export GO111MODULE=on
$ go get ./...
$ cp .env-example .env
$ cd cmd/<cmd>
$ go build ./<cmd>
```

## Run tests locally

```bash
$ export GO111MODULE=on
$ go get ./...
$ go get github.com/onsi/gomega
$ go install github.com/onsi/ginkgo/ginkgo
$ go install github.com/joho/godotenv/cmd/godotenv
$ godotenv -f <env_file> ginkgo -r --randomizeAllSpecs --randomizeSuites --race --trace
```