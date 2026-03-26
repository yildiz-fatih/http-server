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

	resBody := "Parsed!\n"
	resBody += fmt.Sprintf("Method: %s\nTarget: %s\nVersion: %s\n", req.RequestLine.HttpMethod, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)
	resBody += fmt.Sprintf("Headers found: %d\n", len(req.Headers))
	for name, value := range req.Headers {
		resBody += fmt.Sprintf("%s: %s\n", name, value)
	}
	resBody += fmt.Sprintf("Body:\n%s", req.Body)

	statusLine := "HTTP/1.1 200 OK"
	lenHeader := fmt.Sprintf("Content-Length: %d", len(resBody))
	typeHeader := "Content-Type: text/plain"

	res := fmt.Sprintf("%s\r\n%s\r\n%s\r\n\r\n%s", statusLine, lenHeader, typeHeader, resBody)
	_, err = conn.Write([]byte(res))
	if err != nil {
		log.Println(err)
		return
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

		headers[fieldName] = fieldValue
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
