package httphelp

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestListenAndServe_HTTP(t *testing.T) {
	s := listenAndServe(t, "", "", "")
	defer stopServer(t, s)

	testRequest(t, s, http.DefaultTransport.(*http.Transport))
}

func TestListenAndServe_HTTPS(t *testing.T) {
	s := listenAndServe(t, "testdata/server.crt", "testdata/server.key", "")
	defer stopServer(t, s)

	testRequest(t, s, &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})
}

func TestListenAndServe_TLSMutualAuth(t *testing.T) {
	s := listenAndServe(t, "testdata/server.crt", "testdata/server.key", "testdata/server.crt")
	defer stopServer(t, s)

	b, err := ioutil.ReadFile("testdata/server.crt")
	if err != nil {
		t.Fatal(err)
	}

	c, err := tls.LoadX509KeyPair("testdata/client.crt", "testdata/client.key")
	if err != nil {
		t.Fatal(err)
	}

	p := x509.NewCertPool()
	if ok := p.AppendCertsFromPEM(b); !ok {
		t.Fatal("unable to add cert to pool")
	}

	testRequest(t, s, &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      p,
			Certificates: []tls.Certificate{c},
		},
	})
}

func listenAndServe(t *testing.T, certFile, keyFile, caFile string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// here we use hardcoded port instead of ephemeral :0
	// I couldn't figure out how to determine real port
	// after the server starts.
	s := &http.Server{Addr: ":9999", Handler: mux}
	c := make(chan error)
	go func() {
		c <- ListenAndServe(s, certFile, keyFile, caFile)
	}()

	// wait for errors 250ms
	select {
	case err := <-c:
		t.Fatal(err)
	case <-time.After(250 * time.Millisecond):
	}
	return s
}

func stopServer(t *testing.T, s *http.Server) {
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func testRequest(t *testing.T, s *http.Server, transport *http.Transport) {
	proto := "http"
	if transport.TLSClientConfig != nil {
		proto = "https"
	}

	req, err := http.NewRequest(http.MethodGet, proto+"://localhost"+s.Addr+"/", nil)
	if err != nil {
		t.Fatal(err)
	}

	c := http.Client{Transport: transport}
	res, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", res.StatusCode, http.StatusOK)
	}
}
