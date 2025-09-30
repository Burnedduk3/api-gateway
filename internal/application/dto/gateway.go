package dto

import "net/url"

type GatewayRequest struct {
	Path        string
	Method      string
	Headers     map[string][]string
	Body        []byte
	QueryParams url.Values
}

type GatewayResponse struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
}
