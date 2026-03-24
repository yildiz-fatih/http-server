package main

import (
	"fmt"
	"log"
	"net"
)

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
