package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/amenzhinsky/vsftpdmgr/mgr"
)

var (
	addrFlag     = ":8080"
	certFileFlag = ""
	keyFileFlag  = ""
	syncFlag     = false
)

func main() {
	log.SetFlags(0)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: %s ROOT PWDFILE

Options:
`, filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.StringVar(&addrFlag, "addr", addrFlag, "`address` to listen to")
	flag.StringVar(&certFileFlag, "cert-file", certFileFlag, "`path` to TLS certificate file")
	flag.StringVar(&keyFileFlag, "key-file", keyFileFlag, "`path` to TLS key file")
	flag.BoolVar(&syncFlag, "sync", syncFlag, "sync pwdfile with database and exit immediately")
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}
	if err := run(flag.Arg(0), flag.Arg(1)); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func run(root, pwdfile string) error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL not provided")
	}

	m, err := mgr.New(root, pwdfile, databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()

	if syncFlag {
		return m.Sync(context.Background())
	}

	lis, err := net.Listen("tcp", addrFlag)
	if err != nil {
		return err
	}
	defer lis.Close()
	log.Printf("listening to %s", addrFlag)

	srv := &http.Server{Handler: handler(m)}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		signal.Reset()
		log.Print("shutting down...")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("shutdown error: %s", err)
		}
	}()

	if certFileFlag != "" && keyFileFlag != "" {
		err = srv.ServeTLS(lis, certFileFlag, keyFileFlag)
	} else {
		err = srv.Serve(lis)
	}
	if err != nil {
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
	return nil
}

func handler(m *mgr.Mgr) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/health", handlerFunc(healthHandler))
	mux.Handle("/users", usersHandler(m))
	mux.Handle("/users/", usersHandler(m))
	mux.Handle("/", handlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return nil
	}))
	return mux
}

// GET /health
func healthHandler(w http.ResponseWriter, _ *http.Request) error {
	_, err := w.Write([]byte("ok\n"))
	return err
}

// GET    /users
// POST   /users {"username": "...", "password": "..."}
// DELETE /users {"username": "..."}
func usersHandler(m *mgr.Mgr) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		switch r.Method {
		case http.MethodGet:
			users, err := m.List(r.Context())
			if err != nil {
				return err
			}
			b, err := json.Marshal(users)
			if err != nil {
				return err
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(b)
			return err
		case http.MethodPost:
			var u mgr.User
			if err := bind(r, &u); err != nil {
				return err
			}
			if err := m.Save(r.Context(), &u); err != nil {
				if err == mgr.ErrInvalidUser {
					http.Error(w, err.Error(), http.StatusUnprocessableEntity)
					return nil
				}
				return err
			}
			w.WriteHeader(http.StatusOK)
			return nil
		case http.MethodDelete:
			var u mgr.User
			if err := bind(r, &u); err != nil {
				return err
			}
			if err := m.Delete(r.Context(), &u); err != nil {
				return err
			}
			w.WriteHeader(http.StatusOK)
			return nil
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return nil
		}
	}
}

func bind(r *http.Request, v interface{}) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

type handlerFunc func(w http.ResponseWriter, r *http.Request) error

func (f handlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n := time.Now()
	rw := &responseWriter{http.StatusOK, w}
	if err := f(rw, r); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
	log.Printf("%s %s %d %s", r.Method, r.URL.Path, rw.code, time.Since(n))
}

type responseWriter struct {
	code int
	http.ResponseWriter
}

func (w *responseWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}
