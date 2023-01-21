# algoexplore

[![GoBuild](https://github.com/joekir/algoexplore/workflows/GoBuild/badge.svg?branch=main)](https://github.com/joekir/algoexplore/actions/workflows/go_build.yml)
[![codecov](https://codecov.io/gh/joekir/algoexplore/branch/main/graph/badge.svg?token=ZOOIRH3QET)](https://codecov.io/gh/joekir/algoexplore)
[![Go Report Card](https://goreportcard.com/badge/github.com/joekir/algoexplore)](https://goreportcard.com/report/github.com/joekir/algoexplore)

[![CodeQL](https://github.com/joekir/algoexplore/workflows/CodeQL/badge.svg?event=push)](https://github.com/joekir/algoexplore/actions/workflows/codeql-analysis.yml)
[![GoSec](https://github.com/joekir/algoexplore/workflows/GoSec/badge.svg?event=workflow_dispatch)](https://github.com/joekir/algoexplore/actions/workflows/main.yml)

A web framework for visualizing bit-level algorithms.    
The intent to help either your understanding or for teaching of algorithms to others

## Leveraging the framework

You need to implement an "algo" in Golang that implements the interfaces in Algo.go     
See internal/algos/ctph as an example implementation

<`TODO` frontend instructions>

## Examples of usage

- [https://algoexplore.ca](https://algoexplore.ca)

## Running

```
$ COOKIE_SESSION_KEY=0x`openssl rand -hex 8` go run cmd/web_server/main.go
```

## Running with debug logging

_via [glog](https://pkg.go.dev/github.com/golang/glog)_
```
$ COOKIE_SESSION_KEY=0x`openssl rand -hex 8` go run cmd/web_server/main.go --logtostderr=1
```

## Deploying to fly.io

```
1. flyctl secrets set COOKIE_SESSION_KEY=0x`openssl rand -hex 8`
2. fly launch
```

## Running Tests 

```
$ go test -race ./...
```
