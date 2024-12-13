package utils

import (
	"net/http"
)

type ResponseInterceptor struct {
	http.ResponseWriter
	Status    int
	Body      []byte
	HeaderMap http.Header
}

func NewResponseInterceptor(w http.ResponseWriter) *ResponseInterceptor {
	return &ResponseInterceptor{
		ResponseWriter: w,
		Status:         200,
		HeaderMap:      make(http.Header),
	}
}

func (ri *ResponseInterceptor) Header() http.Header {
	return ri.HeaderMap
}

func (ri *ResponseInterceptor) WriteHeader(code int) {
	ri.Status = code
}

func (ri *ResponseInterceptor) Write(b []byte) (int, error) {
	ri.Body = make([]byte, len(b))
	copy(ri.Body, b)
	return len(b), nil
}

// SSRResponse is only used when worker pool is disabled
type SSRResponse struct {
	HTML  string `json:"html"`
	Error string `json:"error,omitempty"`
	Code  int    `json:"code,omitempty"`
}
