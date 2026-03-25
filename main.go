package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

const SEPARATOR = "\r\n"

type Request struct {
	RequestLine RequestLine
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

	buffer := make([]byte, 1024)

	n, err := conn.Read(buffer)
	if err != nil {
		log.Println(err)
		return
	}

	request := string(buffer[:n])
	reqLine, restOfMsg, err := parseRequestLine(request)
	if err != nil {
		log.Println(err)
		return
	}

	// test request line parsing
	fmt.Printf("Parsed:\nMethod: %s, Target: %s, Version: %s\n", reqLine.HttpMethod, reqLine.RequestTarget, reqLine.HttpVersion)
	fmt.Printf("Rest of the request:\n%s\n", restOfMsg)

	body := fmt.Sprintf("You sent:\n%s", buffer[:n])
	statusLine := "HTTP/1.1 200 OK"
	lenHeader := fmt.Sprintf("Content-Length: %d", len(body))
	typeHeader := "Content-Type: text/plain"

	res := fmt.Sprintf("%s\r\n%s\r\n%s\r\n\r\n%s", statusLine, lenHeader, typeHeader, body)
	_, err = conn.Write([]byte(res))
	if err != nil {
		log.Println(err)
		return
	}
}

func parseRequestLine(request string) (*RequestLine, string, error) {
	idx := strings.Index(request, SEPARATOR)
	if idx == -1 {
		return nil, request, fmt.Errorf("parsing error: no CRLF found")
	}

	line := request[:idx]
	restOfMsg := request[idx+len(SEPARATOR):]

	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, restOfMsg, fmt.Errorf("parsing error: split by single space failed")
	}

	return &RequestLine{
		HttpMethod:    parts[0],
		RequestTarget: parts[1],
		HttpVersion:   parts[2],
	}, restOfMsg, nil
}
