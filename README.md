# http-server

HTTP/1.1 web server written in Go, built over TCP

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

## References

- [RFC 9110](https://www.rfc-editor.org/rfc/rfc9110)
- [RFC 9112](https://www.rfc-editor.org/rfc/rfc9112)
