package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	_ "github.com/lib/pq"
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
		fmt.Fprint(os.Stderr, "$DATABASE_URL is not provided\n")
		os.Exit(1)
	}

	if err := start(flag.Arg(0), flag.Arg(1), databaseURL); err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
}

func start(root, pwdfile, databaseURL string) error {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return err
	}

	m, err := mgr.New(root, pwdfile, db)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/users", makeUsersHandler(m))

	srv := &http.Server{Addr: addrFlag, Handler: mux}
	closeOnSignal(srv)

	if certFileFlag != "" && keyFileFlag != "" {
		err = srv.ListenAndServeTLS(certFileFlag, keyFileFlag)
	} else {
		err = srv.ListenAndServe()
	}

	// ignore closed server errors
	if err.Error() == "http: Server closed" {
		err = nil
	}
	return err
}

// closeOnSignal closes srv when an interrupt signal received
func closeOnSignal(srv *http.Server) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		srv.Close()
	}()
}

// GET /health
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok\n"))
	w.WriteHeader(http.StatusOK)
}

// GET    /users
// POST   /users {"name": "", "password": "", "structure": [{""}]}
// DELETE /users {"name": ""}
func makeUsersHandler(m *mgr.Mgr) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
