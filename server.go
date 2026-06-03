package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

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
		conn.SetReadDeadline(time.Now().Add(15 * time.Second)) // timeout for idle connections
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
