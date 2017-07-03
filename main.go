package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/amenzhinsky/vsftpdmgr/httputil"
	"github.com/amenzhinsky/vsftpdmgr/mgr"
)

var (
	addrFlag     = ":8080"
	certFileFlag = ""
	keyFileFlag  = ""
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s ROOT PWDFILE\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.StringVar(&addrFlag, "addr", addrFlag, "address to listen to")
	flag.StringVar(&certFileFlag, "cert-file", certFileFlag, "path to TLS certificate file")
	flag.StringVar(&keyFileFlag, "key-file", keyFileFlag, "path to TLS key file")
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	// DATABASE_URL has to passed to the program
	// via environment variable for security reasons
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fmt.Fprint(os.Stderr, "DATABASE_URL is not provided\n")
		os.Exit(1)
	}

	if err := start(addrFlag, certFileFlag, keyFileFlag, flag.Arg(0), flag.Arg(1), databaseURL); err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
}

// we use separate function here to make sure that all
// defer callback are executed before the process exits.
func start(addr, certFile, keyFile, root, pwdfile, databaseURL string) error {
	m, err := mgr.New(root, pwdfile, databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()

	log.Printf("Listening on %s", addr)
	return httputil.ListenAndServe(addr, handler(m), certFile, keyFile)
}

// handler is needed for integrated testing.
func handler(m *mgr.Mgr) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/users", makeUsersHandler(m))
	return mux
}

// GET /health
func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("ok\n"))
	w.WriteHeader(http.StatusOK)
}

// GET    /users
// POST   /users {"name": "", "password": "", "structure": [{""}]}
// DELETE /users {"name": ""}
func makeUsersHandler(m *mgr.Mgr) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			users, err := m.List(r.Context())
			if err != nil {
				httpError(w, err)
				return
			}

			b, err := json.Marshal(users)
			if err != nil {
				httpError(w, err)
				return
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(b)
			if err != nil {
				fmt.Fprintf(os.Stderr, "write error: %v", err)
			}
		case http.MethodPost:
			u, err := userFromRequest(r)
			if err != nil {
				httpError(w, err)
				return
			}

			if err := m.Save(r.Context(), u); err != nil {
				httpError(w, err)
				return
			}

			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			u, err := userFromRequest(r)
			if err != nil {
				httpError(w, err)
				return
			}

			if err := m.Delete(r.Context(), u); err != nil {
				httpError(w, err)
				return
			}

			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

// userFromRequest parses json request into User structure.
func userFromRequest(r *http.Request) (*mgr.User, error) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var u mgr.User
	return &u, json.Unmarshal(b, &u)
}

func httpError(w http.ResponseWriter, err error) {
	fmt.Printf("http error: %v\n", err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
