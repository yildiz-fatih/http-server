package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const SEPARATOR = "\r\n"

func parseRequest(reader *bufio.Reader) (*Request, error) {
	reqLine, err := parseRequestLine(reader)
	if err != nil {
		return nil, err
	}

	headers, err := parseHeaders(reader)
	if err != nil {
		return nil, err
	}

	body, err := parseBody(reader, headers)
	if err != nil {
		return nil, err
	}

	return &Request{RequestLine: *reqLine, Headers: headers, Body: body}, nil
}

func parseRequestLine(reader *bufio.Reader) (*RequestLine, error) {
	rawLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(rawLine, SEPARATOR) {
		return nil, fmt.Errorf("parsing error: request line missing CRLF")
	}

	trimmedLine := strings.TrimSuffix(rawLine, SEPARATOR)

	parts := strings.Split(trimmedLine, " ")

	if len(parts) != 3 {
		return nil, fmt.Errorf("parsing error: malformed request line")
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

		if !strings.HasSuffix(rawLine, SEPARATOR) {
			return nil, fmt.Errorf("parsing error: header is missing CRLF")
		}

		trimmedLine := strings.TrimSuffix(rawLine, SEPARATOR)

		if len(trimmedLine) == 0 {
			break // empty line indicates end of headers
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
