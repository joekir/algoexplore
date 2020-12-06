# algoexplore

[![Build Status](https://travis-ci.org/joekir/algoexplore.svg?branch=main)](https://travis-ci.org/joekir/algoexplore)
[![codecov](https://codecov.io/gh/joekir/algoexplore/branch/main/graph/badge.svg?token=ZOOIRH3QET)](https://codecov.io/gh/joekir/algoexplore)

A web framework for visualizing bit-level algorithms.    
The intent to help either your understanding or for teaching of algorithms to others

## Leveraging the framework

You need to implement an "algo" in Golang that implements the interfaces in Algo.go     
See internal/algos/ctph as an example implementation

<TODO frontend intstructions>

## Examples of usage

- [ssdeepviz.herokuapp.com](https://ssdeepviz.herokuapp.com)

## Running

```
$ cd cmd/web_server
$ COOKIE_SESSION_KEY=0x`openssl rand -hex 8` go run .
```

## Running Tests 

```
$ go test -race ./...
```