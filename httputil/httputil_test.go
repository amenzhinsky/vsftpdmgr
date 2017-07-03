package httputil

import (
	"os"
	"testing"
	"time"
)

func TestListenAndServe(t *testing.T) {
	t.Parallel()

	go func() {
		time.Sleep(100 * time.Millisecond)

		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			t.Fatal(err)
		}
		p.Signal(os.Interrupt)
	}()

	if err := ListenAndServe(":0", nil, "", ""); err != nil {
		t.Fatal(err)
	}
}
