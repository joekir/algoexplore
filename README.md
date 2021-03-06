# algoexplore

[![Build Status](https://travis-ci.org/joekir/algoexplore.svg?branch=main)](https://travis-ci.org/joekir/algoexplore)
[![codecov](https://codecov.io/gh/joekir/algoexplore/branch/main/graph/badge.svg?token=ZOOIRH3QET)](https://codecov.io/gh/joekir/algoexplore)
[![Go Report Card](https://goreportcard.com/badge/github.com/joekir/algoexplore)](https://goreportcard.com/report/github.com/joekir/algoexplore)
![GoSec](https://github.com/joekir/algoexplore/workflows/GoSec/badge.svg)

A web framework for visualizing bit-level algorithms.    
The intent to help either your understanding or for teaching of algorithms to others

## Leveraging the framework

You need to implement an "algo" in Golang that implements the interfaces in Algo.go     
See internal/algos/ctph as an example implementation

<TODO frontend intstructions>

## Examples of usage

- [https://algoexplore.ca](https://algoexplore.ca)

## Running

```
$ COOKIE_SESSION_KEY=0x`openssl rand -hex 8` go run cmd/web_server/main.go
```

## Running Tests 

```
$ go test -race ./...
```
