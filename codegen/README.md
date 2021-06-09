# Directory Contents

* `functions`: Constains the "Functions SDK". Declares the `firebase.com/functions` module.
* `acme`: Contains the user code written using the "Functions SDK".
* `deployer`: Contains the CLI/deployment tools. Takes the user code, and generates an HTTP server
  to serve the functions.
* `generated`: A directory to dump generated code.

# How To

* From the `deployer` directory:

```
$ go run decorator_discovery.go > ../generated/main.go
```

* From the `generated` directory:

```
$ go run main.go
```

* Send sample requests to the triggers:

```
$ curl -v localhost:8080/MyFunction
$ curl -v -X POST -d '{"uid": "alice"}' localhost:8080/MyOtherFunction
```
