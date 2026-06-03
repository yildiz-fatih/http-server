package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

var (
	root string
	port int
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
