package crypt

import "testing"

func TestCrypt(t *testing.T) {
	got, err := Crypt("test", "$1$Bb6jzHiC$")
	if err != nil {
		t.Fatal(err)
	}

	want := "$1$Bb6jzHiC$Yt25IchKE4VSFK5Vg7qFp/"
	if got != want {
		t.Errorf("s, _ := Crypt(%q, %q); s = %q, want %q", "test", "$1$Bb6jzHiC", got, want)
	}
}

func TestMD5(t *testing.T) {
	p, err := MD5("test")
	if err != nil {
		t.Fatal(err)
	}

	// $ + 1 + $ + salt(8) + $ + hash(22) = 34
	if len(p) != 34 {
		t.Errorf("s, _, MD5(%q); len(s) = %d, want %d", "test", len(p), 34)
	}
}
