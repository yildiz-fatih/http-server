package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const errorTemplateHTML = `
<!DOCTYPE HTML>
<html lang="en">
   <head>
      <meta charset="utf-8">
      <style type="text/css">
         :root {
         color-scheme: light dark;
         }
      </style>
      <title>Error response</title>
   </head>
   <body>
      <h1>Error response</h1>
      <p>Error code: {{ .Code }}</p>
      <p>Message: {{ .Message }}</p>
   </body>
</html>
`

const dirListingTemplateHTML = `
<!DOCTYPE HTML>
<html lang="en">
   <head>
      <meta charset="utf-8">
      <style type="text/css">
         :root {
         color-scheme: light dark;
         }
      </style>
      <title>Directory listing for {{ .Path }}</title>
   </head>
   <body>
      <h1>Directory listing for {{ .Path }}</h1>
      <hr>
      <ul>
        {{ range .Files }}
         <li><a href="{{ . }}">{{ . }}</a></li>
        {{ end }}
      </ul>
      <hr>
   </body>
</html>
`

var (
	root string
	port int
)

const SEPARATOR = "\r\n"

type Request struct {
	RequestLine RequestLine
	Headers     map[string]string
	Body        []byte
}

type RequestLine struct {
	HttpMethod    string
	RequestTarget string
	HttpVersion   string
}

type Response struct {
	StatusCode string
	Headers    map[string]string
	Body       []byte
}

// parse templates at startup
var (
	errorTemplate      = template.Must(template.New("error").Parse(errorTemplateHTML))
	dirListingTemplate = template.Must(template.New("dirListing").Parse(dirListingTemplateHTML))
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&root, "root", cwd, "root directory to serve files from")
	flag.IntVar(&port, "port", 8080, "port to listen on")
	flag.Parse()

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Server running on port %d...\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	// defer conn.Close()
	defer func() {
		log.Printf("Closing connection from %s", conn.RemoteAddr().String())
		conn.Close()
	}()

	log.Printf("New connection from %s\n", conn.RemoteAddr().String())

	// prevent over-reading into the next request,
	// by re-using the same reader for the all requests on the same connection
	reader := bufio.NewReader(conn)

	for {
		log.Printf("New request from %s\n", conn.RemoteAddr().String())

		req, err := parseRequest(reader)
		if err != nil {
			log.Println(err)
			return
		}

		shouldClose := req.Headers["connection"] == "close"

		res, err := routeRequest(req)
		if err != nil {
			log.Println(err)
			return
		}

		if shouldClose {
			res.Headers["connection"] = "close"
		}

		err = writeResponse(conn, res)
		if err != nil {
			log.Println(err)
			return
		}

		if shouldClose {
			break
		}
	}
}

func routeRequest(req *Request) (*Response, error) {
	var res *Response
	var err error

	switch req.RequestLine.RequestTarget {
	case "/ping":
		res, err = handlePing(req)
	case "/echo":
		res, err = handleEcho(req)
	default:
		res, err = handleFile(req)
	}

	if err != nil {
		return nil, err
	}

	return res, nil
}

func parseRequest(reader *bufio.Reader) (*Request, error) {
	reqLine, err := parseRequestLine(reader)
	if err != nil {
		return nil, err
	}

	headers, err := parseHeaders(reader)
	if err != nil {
		return nil, err
	}

	body, err := parseBody(reader, headers)
	if err != nil {
		return nil, err
	}

	return &Request{RequestLine: *reqLine, Headers: headers, Body: body}, nil
}

func parseRequestLine(reader *bufio.Reader) (*RequestLine, error) {
	rawLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	trimmedLine := strings.TrimSuffix(rawLine, SEPARATOR)

	parts := strings.Split(trimmedLine, " ")

	if len(parts) != 3 {
		return nil, fmt.Errorf("parsing error: split by single space failed")
	}

	return &RequestLine{
		HttpMethod:    parts[0],
		RequestTarget: parts[1],
		HttpVersion:   parts[2],
	}, nil
}

func parseHeaders(reader *bufio.Reader) (map[string]string, error) {
	headers := make(map[string]string)

	for {
		rawLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		trimmedLine, _, found := strings.Cut(rawLine, SEPARATOR)
		if !found {
			return nil, fmt.Errorf("parsing error: no separator found")
		}

		if len(trimmedLine) == 0 {
			break
		}

		fieldName, fieldValue, found := strings.Cut(trimmedLine, ":")
		if !found {
			return nil, fmt.Errorf("parsing error: no colon found")
		}

		fieldName = strings.ToLower(fieldName)
		fieldValue = strings.TrimSpace(fieldValue)

		// handle multiple headers with the same field name
		existingValue, found := headers[fieldName] // check if header already exists
		if found {                                 // header already exists, concatenate values
			headers[fieldName] = existingValue + ", " + fieldValue
		} else { // header does not exist, add it to the map
			headers[fieldName] = fieldValue
		}
	}

	return headers, nil
}

func parseBody(reader *bufio.Reader, headers map[string]string) ([]byte, error) {
	contentLenStr, found := headers["content-length"]
	if !found {
		return nil, nil
	}

	contentLen, err := strconv.Atoi(contentLenStr)
	if err != nil {
		return nil, fmt.Errorf("parsing error: invalid content-length")
	}

	body := make([]byte, contentLen)

	_, err = io.ReadFull(reader, body)
	if err != nil {
		return nil, fmt.Errorf("parsing error: failed to read body")
	}

	return body, nil
}

func writeResponse(conn net.Conn, res *Response) error {
	out := ""

	httpVersion := "HTTP/1.1"
	statusLine := fmt.Sprintf("%s %s", httpVersion, res.StatusCode)

	out += statusLine + "\r\n"

	out += fmt.Sprintf("Content-Length: %d\r\n", len(res.Body))

	for name, value := range res.Headers {
		out += fmt.Sprintf("%s: %s\r\n", name, value)
	}
	out += "\r\n"

	_, err := conn.Write([]byte(out))
	if err != nil {
		return err
	}

	_, err = conn.Write(res.Body)
	return err
}

func handlePing(req *Request) (*Response, error) {
	if req.RequestLine.HttpMethod != "GET" {
		return &Response{
			StatusCode: "405 Method Not Allowed",
			Headers:    map[string]string{},
			Body:       []byte{},
		}, nil
	}

	return &Response{
		StatusCode: "200 OK",
		Headers:    map[string]string{"Content-Type": "text/plain"},
		Body:       []byte("pong"),
	}, nil
}

func handleEcho(req *Request) (*Response, error) {
	resBody := ""
	resBody += fmt.Sprintf("%s %s %s\r\n", req.RequestLine.HttpMethod, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)
	for name, value := range req.Headers {
		resBody += fmt.Sprintf("%s: %s\r\n", name, value)
	}
	resBody += "\r\n"
	resBody += string(req.Body)

	return &Response{
		StatusCode: "200 OK",
		Headers:    map[string]string{"Content-Type": "text/plain"},
		Body:       []byte(resBody),
	}, nil
}

func handleFile(req *Request) (*Response, error) {
	targetFilename := filepath.Join(root, filepath.Clean(req.RequestLine.RequestTarget))

	targetFileInfo, err := os.Stat(targetFilename)
	if err != nil {
		// return 404 page
		var buf bytes.Buffer
		err = errorTemplate.Execute(&buf, struct{ Code, Message string }{
			Code:    "404 Not Found",
			Message: "The requested resource was not found on this server.",
		})
		if err != nil {
			return nil, err
		}

		return &Response{
			StatusCode: "404 Not Found",
			Headers:    map[string]string{"Content-Type": "text/html"},
			Body:       buf.Bytes(),
		}, nil
	}

	if targetFileInfo.IsDir() {
		// if you want a directory, put a slash at the end of the url
		if !strings.HasSuffix(req.RequestLine.RequestTarget, "/") { // if you don't
			// i'll redirect you to the url with the slash at the end
			return &Response{
				StatusCode: "301 Moved Permanently",
				Headers:    map[string]string{"Location": req.RequestLine.RequestTarget + "/"},
				Body:       []byte{},
			}, nil
		}

		// check for index.html in the directory
		_, err := os.Stat(filepath.Join(targetFilename, "index.html"))
		if !errors.Is(err, os.ErrNotExist) {
			content, err := os.ReadFile(filepath.Join(targetFilename, "index.html"))
			if err != nil {
				return nil, err
			}

			return &Response{
				StatusCode: "200 OK",
				Headers:    map[string]string{"Content-Type": "text/html"},
				Body:       content,
			}, nil
		}

		// return directory listing
		entries, err := os.ReadDir(targetFilename)
		if err != nil {
			return nil, err
		}

		var files []string
		for _, entry := range entries {
			files = append(files, entry.Name())
		}

		var buf bytes.Buffer
		err = dirListingTemplate.Execute(&buf, struct {
			Path  string
			Files []string
		}{
			Path:  req.RequestLine.RequestTarget,
			Files: files,
		})
		if err != nil {
			return nil, err
		}

		return &Response{
			StatusCode: "200 OK",
			Headers:    map[string]string{"Content-Type": "text/html"},
			Body:       buf.Bytes(),
		}, nil
	} else {
		contentType := mime.TypeByExtension(filepath.Ext(targetFilename))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		resBody, err := os.ReadFile(targetFilename)
		if err != nil {
			return nil, err
		}

		return &Response{
			StatusCode: "200 OK",
			Headers:    map[string]string{"Content-Type": contentType},
			Body:       resBody,
		}, nil
	}
}
