package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
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
		// do echo
	default:
		// serve files from "root"
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
