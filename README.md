# http-server

An HTTP/1.1 server written in Go, built over TCP. Point it at a directory and it serves the files over HTTP.

This is a learning project. The goal is to understand how HTTP works over TCP.

## Usage

```bash
# Default: Serve the current directory on port 8080
go run .

# Custom: Serve the ./public directory on port 9090
go run . -port 9090 -root ./public
```

## Options

- `-port`: port to listen on (default 8080)
- `-root`: root directory to serve files from (default current directory)

## Testing
```bash
make test         # run all tests
make test-v       # run all tests (verbose)
make cover        # print coverage summary
make cover-html   # open coverage report in the browser
```

## References

- [RFC 9110: HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110)
- [RFC 9112: HTTP/1.1](https://www.rfc-editor.org/rfc/rfc9112)
