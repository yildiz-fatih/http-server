package main

import (
	"bytes"
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

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
