# go-fizzbuzz

A golang fizzbuzz implementation

## Objective 

The objective is to implement an http endpoint that returns the fizzbuzz sequence based on query parameters.

This is the original briefing, verbatim:

> The original fizz-buzz consists in writing all numbers from 1 to 100, and just replacing all multiples of 3 by "fizz", all multiples of 5 by "buzz", and all multiples of 15 by "fizzbuzz".
> The output would look like this: "1,2,fizz,4,buzz,fizz,7,8,fizz,buzz,11,fizz,13,14,fizzbuzz,16,...".
> 
> Your goal is to implement a web server that will expose a REST API endpoint that:
> - Accepts five parameters: three integers int1, int2 and limit, and two strings str1 and str2.
> - Returns a list of strings with numbers from 1 to limit, where: all multiples of int1 are replaced by str1, all multiples of int2 are replaced by str2, all multiples of int1 and int2 are replaced by str1str2.
> 
> The server needs to be:
> - Ready for production
> - Easy to maintain by other developers
> 
> Bonus: add a statistics endpoint allowing users to know what the most frequent request has been.
> 
> This endpoint should:
> - Accept no parameter
> - Return the parameters corresponding to the most used request, as well as the number of hits for this request

## Quick Start

To start the server we can execute `make run`:

```console
$ make run
go run ./cmd/server/main.go
2026/09/28 11:50:53 INFO starting server, will listen to :8080
```

We can use curl to perform a quick http request:

```console
$ curl -s "http://127.0.0.1:8080/fizzbuzz"
["1","2","fizz","4","buzz","fizz","7","8","fizz","buzz","11","fizz","13","14","fizzbuzz","16","17","fizz","19","buzz","fizz","22","23","fizz","buzz","26","fizz","28","29","fizzbuzz","31","32","fizz","34","buzz","fizz","37","38","fizz","buzz","41","fizz","43","44","fizzbuzz","46","47","fizz","49","buzz","fizz","52","53","fizz","buzz","56","fizz","58","59","fizzbuzz","61","62","fizz","64","buzz","fizz","67","68","fizz","buzz","71","fizz","73","74","fizzbuzz","76","77","fizz","79","buzz","fizz","82","83","fizz","buzz","86","fizz","88","89","fizzbuzz","91","92","fizz","94","buzz","fizz","97","98","fizz","buzz"]
```

To create the docker image you can do:

```console
$ docker build . 
```

## API

### the /fizzbuzz endpoint

This endpoint responds to HTTP verb GET only, it expect up to 5 parameters ( `int1`, `int2`, `limit`, `str1`, `str2` ) and will return a json response with the fizzbuzz sequence if the parameters are considered valid.

### the /healthz endpoint

This endpoint is used both readiness and liveness probes for kubernetes. Returns a simple HTTP 200 OK response.

## the /metrics endpoint

TBD

## Design Decisions

In order to be easier to maintain by other developers and be ready to production I make the following decisions:

- This project must be coded by myself. But I may use AI to perform code reviews.
- The golang standard library should be enough to implement the core features.
- No premature optimizations. Correctness and bounded resources first.
- Output will be a JSON array of strings: this format helps to escape the input strings and avoid any ambiguity.
- Omitted parameters take the fizzbuzz classic defaults. Invalid values will be rejected.
- Inputs are bounded to certain limits in order to keep the response size under control and avoid a denial of service risk.

## Limitations

TBD

## Development 

TBD

## How I worked.

TBD
