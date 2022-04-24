package util

import (
	"net"
	"net/http"
)

func ListenAndServeNotify(srv *http.Server, ready chan int) error {
	// No shutdown logic
	addr := srv.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	if ready != nil {
		ready <- 1
	}
	return srv.Serve(ln)
}
