package httphelp

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net/http"
)

// ListenAndServe starts listening on srv.Addr, TLS encryption
// is enabled when certFile and keyFile are passed.
// caFile enables mutual TLS authentication, requires TLS encryption.
func ListenAndServe(srv *http.Server, certFile, keyFile, caFile string) error {
	if caFile != "" {
		if err := enableTLSMutualAuth(srv, caFile); err != nil {
			return err
		}
	}

	if certFile != "" || keyFile != "" {
		return srv.ListenAndServeTLS(certFile, keyFile)
	}
	return srv.ListenAndServe()
}

func enableTLSMutualAuth(srv *http.Server, caFile string) error {
	b, err := ioutil.ReadFile(caFile)
	if err != nil {
		return err
	}

	clientCAs := x509.NewCertPool()
	if ok := clientCAs.AppendCertsFromPEM(b); !ok {
		return errors.New("unable to append ca cert")
	}

	tlsConfig := &tls.Config{
		ClientCAs:  clientCAs,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	tlsConfig.BuildNameToCertificate()
	srv.TLSConfig = tlsConfig
	return nil
}
