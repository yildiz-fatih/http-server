package main

import (
	"bufio"
	"strings"
	"testing"
)

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
