package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Tower sets PORT. Read it rather than hardcoding a number: the port is a setting on
	// the function and the default can change.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "Hello from Tower.",
			"path":    r.URL.Path,
			"method":  r.Method,
		})
	})

	// 0.0.0.0, not localhost: a server bound to the loopback address is reachable only
	// from inside its own instance, so every request from outside fails.
	srv := &http.Server{Addr: "0.0.0.0:" + port, Handler: mux}

	// Tower sends SIGTERM before stopping an instance. Draining here is optional, but if
	// you do drain you have to wait for it: Shutdown makes ListenAndServe return
	// ErrServerClosed straight away, so without the drained channel below main() would
	// return while requests were still being served, and they would be cut off.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	drained := make(chan struct{})
	go func() {
		defer close(drained)
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		// Shutdown closes idle connections itself and waits for active ones. The timeout
		// bounds the wait so a slow request cannot hold the instance open indefinitely.
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}()

	log.Printf("listening on 0.0.0.0:%s", port)
	// ListenAndServe blocks. Returning from main() is what makes a function build
	// successfully and then fail to serve.
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	<-drained
}
