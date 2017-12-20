# go-snowboy

The Go bindings for snowboy audio detection (https://github.com/Kitt-AI/snowboy) are generated using swig which
 creates a lot of extra types and uses calls with variable arguments. This makes writing integrations in golang difficult
 because the types aren't explicit. go-snowboy is intended to be a wrapper around the swig-generated Go code which will
 provide Go-style usage.

## Docs
See https://godoc.org/github.com/brentnd/go-snowboy

## Dependencies
* SWIG (v 3.0.12 recommended)

### Go Packages
* github.com/Kitt-AI/snowboy/swig/Go

## Example

Example hotword detection usage in `example/detect.go`.
Example API hotword training usage in `example/api.go`.

### Building
```
go build -o build/snowboy-detect example/detect.go
go build -o build/snowboy-api example/api.go
go build -o build/snowboy-listen example/listen.go
```

### Running
```
usage: ./build/snowboy-detect <resource> <keyword.umdl> <audio file>
usage: ./build/snowboy-listen <resource> <keyword.umdl>
```

### See Also
`Makefile` has some standard targets to do `go build` steps
