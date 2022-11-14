package http_server

import (
	"fmt"
	"net/http"
	"time"
)

func New(port string, host string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf("%s:%s", host, port),
		Handler:           h,
		ReadTimeout:       time.Second * 15,
		ReadHeaderTimeout: time.Second * 15,
		WriteTimeout:      time.Second * 15,
		IdleTimeout:       time.Second * 15,
		MaxHeaderBytes:    1 << 20,
	}
}
