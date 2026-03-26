package main

import (
	"bufio"
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

	reader := bufio.NewReader(conn)

	reqLine, err := parseRequestLine(reader)
	if err != nil {
		log.Println(err)
		return
	}

	body := fmt.Sprintf("Parsed!\nMethod: %s\nTarget: %s\nVersion: %s\n", reqLine.HttpMethod, reqLine.RequestTarget, reqLine.HttpVersion)
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
