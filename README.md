# GraphQL subscription example
[![Build Status](https://travis-ci.org/matiasanaya/go-graphql-subscription-example.svg?branch=master)](https://travis-ci.org/matiasanaya/go-graphql-subscription-example)

This application uses:

* [github.com/graph-gophers/graphql-go](https://github.com/graph-gophers/graphql-go)
* [github.com/graph-gophers/graphql-transport-ws](https://github.com/graph-gophers/graphql-transport-ws)

## Pre-requisites

**Requires Go 1.11.x** or above, which support Go modules. Read more about them [here](https://github.com/golang/go/wiki/Modules).

Remember to set ```GO111MODULE=on```

## How to use

Run the application:

```
go run main.go
```

Navigate to [localhost:8080](http://localhost:8080) and use GraphiQL to subscribe using the following example:

```
subscription onHelloSaid {
  helloSaid {
    id
    msg
  }
}
```

On a separate tab run:

```
mutation SayHello{
  sayHello(msg: "Hello world!") {
    id
    msg
  }
}
```
