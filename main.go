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

	buffer := make([]byte, 8)
	var request string
	var reqLine *RequestLine

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Println(err)
			return
		}
		// append the new chunk
		request += string(buffer[:n])

		// try to parse
		parsedLine, restOfMsg, err := parseRequestLine(request)
		if err != nil {
			log.Println(err)
			return
		}

		if parsedLine != nil {
			reqLine = parsedLine
			request = restOfMsg
			break
		}
	}
	/*
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
	*/
	body := fmt.Sprintf("Parsed!\nMethod: %s\nTarget: %s\nVersion: %s\n", reqLine.HttpMethod, reqLine.RequestTarget, reqLine.HttpVersion)
	statusLine := "HTTP/1.1 200 OK"
	lenHeader := fmt.Sprintf("Content-Length: %d", len(body))
	typeHeader := "Content-Type: text/plain"

	res := fmt.Sprintf("%s\r\n%s\r\n%s\r\n\r\n%s", statusLine, lenHeader, typeHeader, body)
	_, err := conn.Write([]byte(res))
	if err != nil {
		log.Println(err)
		return
	}
}

func parseRequestLine(request string) (*RequestLine, string, error) {
	idx := strings.Index(request, SEPARATOR)
	if idx == -1 {
		return nil, request, nil
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
