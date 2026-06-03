package main

type Request struct {
	RequestLine RequestLine
	Headers     map[string]string
	Body        []byte
}

type RequestLine struct {
	HttpMethod    string
	RequestTarget string
	HttpVersion   string
}

type Response struct {
	StatusCode string
	Headers    map[string]string
	Body       []byte
}
