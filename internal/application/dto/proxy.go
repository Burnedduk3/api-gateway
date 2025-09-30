package dto

import "net/http"

type ProxyRequest struct {
	URL     string
	Method  string
	Headers map[string][]string
	Body    []byte
}

type ProxyResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}
