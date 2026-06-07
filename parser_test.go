package main

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

func TestParseRequest(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        *Request
		expectError bool
	}{
		{
			name:  "valid GET request",
			input: "GET /index.html HTTP/1.1\r\nHost: example.com\r\n\r\n",
			want: &Request{
				RequestLine: RequestLine{
					HttpMethod:    "GET",
					RequestTarget: "/index.html",
					HttpVersion:   "HTTP/1.1",
				},
				Headers: map[string]string{"host": "example.com"},
				Body:    nil,
			},
		},
		{
			name:  "valid POST request",
			input: "POST /api/cats HTTP/1.1\r\nHost: example.com\r\nContent-Length: 11\r\n\r\nhello world",
			want: &Request{
				RequestLine: RequestLine{
					HttpMethod:    "POST",
					RequestTarget: "/api/cats",
					HttpVersion:   "HTTP/1.1",
				},
				Headers: map[string]string{
					"host":           "example.com",
					"content-length": "11",
				},
				Body: []byte("hello world"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))

			got, err := parseRequest(reader)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected an error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if !reflect.DeepEqual(*tt.want, *got) {
					t.Errorf("want %+v, got %+v", *tt.want, *got)
				}
			}
		})
	}
}

func TestParseRequestLine(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        RequestLine
		expectError bool
	}{
		{
			name:  "valid GET request",
			input: "GET /index.html HTTP/1.1\r\n",
			want:  RequestLine{HttpMethod: "GET", RequestTarget: "/index.html", HttpVersion: "HTTP/1.1"},
		},
		{
			name:  "valid POST request",
			input: "POST /api/v2/cats?lives_remaining=7 HTTP/1.1\r\n",
			want:  RequestLine{HttpMethod: "POST", RequestTarget: "/api/v2/cats?lives_remaining=7", HttpVersion: "HTTP/1.1"},
		},
		{
			name:        "error on missing carriage return",
			input:       "GET /index.html HTTP/1.1\n",
			expectError: true,
		},
		{
			name:        "error on missing newline",
			input:       "GET /index.html HTTP/1.1",
			expectError: true,
		},
		{
			name:        "error on empty input",
			input:       "",
			expectError: true,
		},
		{
			name:        "error on too few parts",
			input:       "GET /index.html\r\n",
			expectError: true,
		},
		{
			name:        "error on too many parts",
			input:       "GET /index.html HTTP/1.1 extra\r\n",
			expectError: true,
		},
		{
			name:        "error on extra whitespace",
			input:       "GET   /index.html   HTTP/1.1\r\n",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))

			got, err := parseRequestLine(reader)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected an error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if tt.want != *got {
					t.Errorf("want %+v, got %+v", tt.want, *got)
				}
			}
		})
	}
}

func TestParseHeaders(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        map[string]string
		expectError bool
	}{
		{
			name:  "single header",
			input: "Host: example.com\r\n\r\n",
			want:  map[string]string{"host": "example.com"},
		},
		{
			name:  "no headers",
			input: "\r\n",
			want:  map[string]string{},
		},
		{
			name:  "duplicate header names get combined",
			input: "Accept: text/html\r\nAccept: application/json\r\n\r\n",
			want:  map[string]string{"accept": "text/html, application/json"},
		},
		{
			name:  "header names get lowercased",
			input: "HOST: example.com\r\n\r\n",
			want:  map[string]string{"host": "example.com"},
		},
		{
			name:  "extra whitespace gets trimmed",
			input: "Host:   example.com   \r\n\r\n",
			want:  map[string]string{"host": "example.com"},
		},
		{
			name:        "error on missing carriage return",
			input:       "Host: example.com\n",
			expectError: true,
		},
		{
			name:        "error on missing newline",
			input:       "Host: example.com",
			expectError: true,
		},
		{
			name:        "error on malformed header (no colon)",
			input:       "Host example.com\r\n",
			expectError: true,
		},
		{
			name:        "error on empty input",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))

			got, err := parseHeaders(reader)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected an error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if !reflect.DeepEqual(tt.want, got) {
					t.Errorf("want %+v, got %+v", tt.want, got)
				}
			}
		})
	}
}

func TestParseBody(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		headers     map[string]string
		want        []byte
		expectError bool
	}{
		{
			name:    "no content-length header returns nil body",
			input:   "",
			headers: map[string]string{},
			want:    nil,
		},
		{
			name:    "content-length header with valid body returns body bytes",
			input:   "hello world",
			headers: map[string]string{"content-length": "11"},
			want:    []byte("hello world"),
		},
		{
			name:        "non-numeric content-length header returns error",
			input:       "",
			headers:     map[string]string{"content-length": "abc"},
			expectError: true,
		},
		{
			name:        "body shorter than content-length returns error",
			input:       "hello world",
			headers:     map[string]string{"content-length": "100"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))

			got, err := parseBody(reader, tt.headers)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected an error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if !reflect.DeepEqual(tt.want, got) {
					t.Errorf("want %+v, got %+v", tt.want, got)
				}
			}
		})
	}
}
