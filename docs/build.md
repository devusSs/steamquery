### Building the app yourself

To build the program yourself you will need a few tools:
- the [Go](https://go.dev) programming language
- [GNU Make](https://www.gnu.org/software/make/) by the Free Software Foundation
- [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) for debugging / checking purposes
- [golangci-lint](https://github.com/golangci/golangci-lint) for debugging / checking purposes
- [gocritic](https://github.com/go-critic/go-critic) for debugging / checking purposes
- [golines](https://github.com/segmentio/golines) for debugging / checking purposes

### Using make to build and run

There are a few options available:
- `make build` to build the app in the `.release` directory
- `make dev` to build the app and run it with debugging parameters
- `make check` to check the code for potential errors
- `make clean` to check the code for potential errors and cleanup build artifacts