# go-acme-proxy
Tiny Let's Encrypt enabled reverse proxy

## Usage

To run the proxy just run `go-acme-proxy --port 8080` to add https to the service running at http://localhost:8080.
Use `--author` to add the mail used to register this service on let's encrypt.

## Getting Started

If you want to use it for debugging, development. 
Just get it using go `go get github.com/tawalaya/go-acme-proxy` and run it with `$GOPATH/bin/go-acme-proxy`. You can also download binary's form the [release tab](https://github.com/tawalaya/go-acme-proxy/releases).

### Prerequisites

- Go Lang 1.10+


## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/tawalaya/go-acme-proxy/tags). 

## Authors

* **Sebastian Werner** - *Initial work* - [tawalaya](https://github.com/tawalaya)

See also the list of [contributors](https://github.com/tawalaya/go-acme-proxy/contributors) who participated in this project.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE.md](LICENSE.md) file for details
