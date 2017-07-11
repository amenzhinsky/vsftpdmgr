package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/amenzhinsky/vsftpdmgr/httphelp"
	"github.com/amenzhinsky/vsftpdmgr/mgr"
)

var (
	addrFlag     = ":8080"
	certFileFlag = ""
	keyFileFlag  = ""
	caFileFlag   = ""
	gzipFlag     = false
	syncFlag     = false
	traceFlag    = false
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s ROOT PWDFILE\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.StringVar(&addrFlag, "addr", addrFlag, "address to listen to")
	flag.StringVar(&certFileFlag, "cert-file", certFileFlag, "path to TLS certificate file")
	flag.StringVar(&keyFileFlag, "key-file", keyFileFlag, "path to TLS key file")
	flag.StringVar(&caFileFlag, "ca-file", caFileFlag, "path to TLS CA file, enables TLS mutual authentication")
	flag.BoolVar(&gzipFlag, "gzip", gzipFlag, "enables gzip compression of HTTP traffic")
	flag.BoolVar(&syncFlag, "sync", syncFlag, "sync pwdfile with database and exit immediately")
	flag.BoolVar(&traceFlag, "trace", traceFlag, "enable http requests tracing")
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	if err := start(flag.Arg(0), flag.Arg(1)); err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
}

// we use separate function here to make sure that all
// defer callback are executed before the process exits.
func start(root, pwdfile string) error {
	// DATABASE_URL has to passed to the program
	// via environment variable for security reasons
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fmt.Fprint(os.Stderr, "DATABASE_URL is not provided\n")
		os.Exit(1)
	}

	m, err := mgr.New(root, pwdfile, databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()

	if syncFlag {
		return m.Sync(context.Background())
	}

	log.Printf("Listening on %s", addrFlag)

	srv := &http.Server{Addr: addrFlag, Handler: handler(m, traceFlag, gzipFlag)}

	// stop server on SIGINT
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

	return httphelp.ListenAndServe(srv, certFileFlag, keyFileFlag, caFileFlag)
}

// handler is needed for integrated testing.
func handler(m *mgr.Mgr, trace, gzip bool) http.Handler {
	mk := func(f httphelp.HandlerFunc) http.Handler {
		if trace {
			f = httphelp.Trace(f)
		}
		if gzip {
			f = httphelp.Gzip(f)
		}
		f = httphelp.Log(f)

		return httphelp.HandlerFunc(f)
	}

	mux := http.NewServeMux()
	mux.Handle("/health/", mk(healthHandler))
	mux.Handle("/users/", mk(makeUsersHandler(m)))
	return mux
}

// GET /health
func healthHandler(w http.ResponseWriter, _ *http.Request) error {
	return httphelp.Text(w, http.StatusOK, "ok\n")
}

// GET    /users
// POST   /users {"username": "...", "password": "..."}
// DELETE /users {"username": "..."}
func makeUsersHandler(m *mgr.Mgr) httphelp.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		switch r.Method {
		case http.MethodGet:
			users, err := m.List(r.Context())
			if err != nil {
				return err
			}

			return httphelp.JSON(w, users)
		case http.MethodPost:
			var u mgr.User
			if err := httphelp.Bind(r, &u); err != nil {
				return err
			}

			if err := m.Save(r.Context(), &u); err != nil {
				if err == mgr.ErrInvalidUser {
					err = &httphelp.HTTPError{
						Code: http.StatusUnprocessableEntity,
						Err:  err,
					}
				}
				return err
			}

			return httphelp.Empty(w, http.StatusOK)
		case http.MethodDelete:
			var u mgr.User
			if err := httphelp.Bind(r, &u); err != nil {
				return err
			}

			if err := m.Delete(r.Context(), &u); err != nil {
				return err
			}
			return httphelp.Empty(w, http.StatusOK)
		default:
			return httphelp.ErrMethodNotAllowed
		}
	}
}
