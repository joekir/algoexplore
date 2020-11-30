# algoexplore

A web framework for visualizing bit-level algorithms.    
The intent to help either your understanding or for teaching of algorithms to others

## Examples of usage

- [ssdeepviz.herokuapp.com](https://ssdeepviz.herokuapp.com)

## Running

```
$ cd cmd/web_server
$ COOKIE_SESSION_KEY=0x`openssl rand -hex 8` go run .
```


## Running Tests 

```
$ COOKIE_SESSION_KEY=0x`openssl rand -hex 8` go test ./...
```
