package middleware

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

type hijackableResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (h *hijackableResponseWriter) WriteHeader(statusCode int) {
	h.statusCode = statusCode
	h.ResponseWriter.WriteHeader(statusCode)
}

func (h *hijackableResponseWriter) Write(b []byte) (int, error) {
	size, err := h.ResponseWriter.Write(b)
	h.size += size
	return size, err
}

func (h *hijackableResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := h.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("response writer does not implement http.Hijacker")
}
