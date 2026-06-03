package main

func routeRequest(req *Request) (*Response, error) {
	var res *Response
	var err error

	switch req.RequestLine.RequestTarget {
	case "/ping":
		res, err = handlePing(req)
	case "/echo":
		res, err = handleEcho(req)
	default:
		res, err = handleFile(req)
	}

	if err != nil {
		return nil, err
	}

	return res, nil
}
