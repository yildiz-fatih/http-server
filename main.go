package main

import (
	"bufio"
	"bytes"
	"errors"
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

const errorTemplate = `
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

const dirListingTemplate = `
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

func main() {
	port := 8080

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
	defer conn.Close()

	reader := bufio.NewReader(conn)

	reqLine, err := parseRequestLine(reader)
	if err != nil {
		log.Println(err)
		return
	}

	headers, err := parseHeaders(reader)
	if err != nil {
		log.Println(err)
		return
	}

	body, err := parseBody(reader, headers)
	if err != nil {
		log.Println(err)
		return
	}

	req := Request{RequestLine: *reqLine, Headers: headers, Body: body}

	switch req.RequestLine.RequestTarget {
	case "/ping":
		err := handlePing(conn, &req)
		if err != nil {
			log.Println(err)
			return
		}
	case "/echo":
		err := handleEcho(conn, &req)
		if err != nil {
			log.Println(err)
			return
		}
	default:
		err := handleFile(conn, &req)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func parseRequestLine(reader *bufio.Reader) (*RequestLine, error) {
	rawLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	trimmedLine, _, found := strings.Cut(rawLine, SEPARATOR)
	if !found {
		return nil, fmt.Errorf("parsing error: no separator found")
	}

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

	out += string(res.Body)

	_, err := conn.Write([]byte(out))
	return err
}

func handlePing(conn net.Conn, req *Request) error {
	if req.RequestLine.HttpMethod != "GET" {
		res := Response{
			StatusCode: "405 Method Not Allowed",
			Headers:    map[string]string{},
			Body:       []byte{},
		}

		err := writeResponse(conn, &res)
		if err != nil {
			return err
		}

		return nil
	}

	res := Response{
		StatusCode: "200 OK",
		Headers:    map[string]string{"Content-Type": "text/plain"},
		Body:       []byte("pong"),
	}

	return writeResponse(conn, &res)
}

func handleEcho(conn net.Conn, req *Request) error {
	resBody := ""
	resBody += fmt.Sprintf("%s %s %s\r\n", req.RequestLine.HttpMethod, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)
	for name, value := range req.Headers {
		resBody += fmt.Sprintf("%s: %s\r\n", name, value)
	}
	resBody += "\r\n"
	resBody += string(req.Body)

	res := Response{
		StatusCode: "200 OK",
		Headers:    map[string]string{"Content-Type": "text/plain"},
		Body:       []byte(resBody),
	}

	return writeResponse(conn, &res)
}

func handleFile(conn net.Conn, req *Request) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	targetFilename := filepath.Join(root, filepath.Clean(req.RequestLine.RequestTarget))

	targetFileInfo, err := os.Stat(targetFilename)
	if err != nil {
		// return 404 page
		type ErrorPage struct {
			Code    string
			Message string
		}

		tmpl, err := template.New("error").Parse(errorTemplate)
		if err != nil {
			return err
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, ErrorPage{
			Code:    "404 Not Found",
			Message: "The requested resource was not found on this server.",
		})
		if err != nil {
			return err
		}

		res := Response{
			StatusCode: "404 Not Found",
			Headers:    map[string]string{"Content-Type": "text/html"},
			Body:       buf.Bytes(),
		}
		return writeResponse(conn, &res)
	}

	if targetFileInfo.IsDir() {
		// if you want a directory, put a slash at the end of the url
		if req.RequestLine.RequestTarget[len(req.RequestLine.RequestTarget)-1] != '/' { // if you don't
			// i'll redirect you to the url with the slash at the end
			res := Response{
				StatusCode: "301 Moved Permanently",
				Headers:    map[string]string{"Location": req.RequestLine.RequestTarget + "/"},
				Body:       []byte{},
			}
			return writeResponse(conn, &res)
		}

		// check for index.html in the directory
		_, err := os.Stat(filepath.Join(targetFilename, "index.html"))
		if !errors.Is(err, os.ErrNotExist) {
			content, err := os.ReadFile(filepath.Join(targetFilename, "index.html"))
			if err != nil {
				return err
			}

			res := Response{
				StatusCode: "200 OK",
				Headers:    map[string]string{"Content-Type": "text/html"},
				Body:       content,
			}
			return writeResponse(conn, &res)
		}

		// return directory listing
		type DirListingPage struct {
			Path  string
			Files []string
		}

		tmpl, err := template.New("listing").Parse(dirListingTemplate)
		if err != nil {
			return err
		}

		entries, err := os.ReadDir(targetFilename)
		if err != nil {
			return err
		}

		var files []string
		for _, entry := range entries {
			files = append(files, entry.Name())
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, DirListingPage{
			Path:  req.RequestLine.RequestTarget,
			Files: files,
		})
		if err != nil {
			return err
		}

		res := Response{
			StatusCode: "200 OK",
			Headers:    map[string]string{"Content-Type": "text/html"},
			Body:       buf.Bytes(),
		}
		return writeResponse(conn, &res)
	} else {
		contentType := mime.TypeByExtension(filepath.Ext(targetFilename))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		resBody, err := os.ReadFile(targetFilename)
		if err != nil {
			return err
		}

		res := Response{
			StatusCode: "200 OK",
			Headers:    map[string]string{"Content-Type": contentType},
			Body:       resBody,
		}

		return writeResponse(conn, &res)
	}
}
