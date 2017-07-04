package httputil

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// ListenAndServe creates a http server and listens on addr.
// It uses TLS encryption when certFile and keyFile are not empty.
// If an interrupt signal received the server is shut down.
// If multiple interrupt signals receiver current process exits
// immediately, e.g. if you press Ctrl^C twice.
func ListenAndServe(addr string, handler http.Handler, certFile, keyFile string) error {
	srv := http.Server{Addr: addr, Handler: handler}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		var exit bool
		for range sigCh {
			if exit {
				os.Exit(1)
			}
			exit = true

			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "shutdown error: %v\n", err)
				os.Exit(1)
			}
		}
	}()

	var err error
	if certFile != "" && keyFile != "" {
		err = srv.ListenAndServeTLS(certFile, keyFile)
	} else {
		err = srv.ListenAndServe()
	}

	// ignore closed server errors
	if err.Error() == "http: Server closed" {
		err = nil
	}

	return err
}
