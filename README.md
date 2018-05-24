### How to use

First install the dependencies:

```
dep ensure
```

Then run the application:

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
query SayHello{
  hello(msg: "Hello world!") {
    id
    msg
  }
}
```

### Dependencies

* [Dep](https://github.com/golang/dep)
