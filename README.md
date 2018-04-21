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
    msg
  }
}
```

### Dependencies

* [Dep](https://github.com/golang/dep)
