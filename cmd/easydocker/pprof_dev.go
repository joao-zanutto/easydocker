//go:build dev

package main

import (
	"log"
	"net/http"
	"time"

	_ "net/http/pprof"
)

func init() {
	go func() {
		srv := &http.Server{
			Addr:              "localhost:6060",
			ReadHeaderTimeout: 5 * time.Second,
		}
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("pprof server: %v", err)
		}
	}()
}
