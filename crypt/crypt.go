package crypt

import (
	"fmt"
	"sync"
	"unsafe"
	"math/rand"
	"time"
)

/*
#cgo LDFLAGS: -lcrypt

#define _GNU_SOURCE
#define _XOPEN_SOURCE

#include <stdlib.h>
#include <unistd.h>
*/
import "C"

var mu sync.Mutex

// TODO: implement without using cgo
// Crypt is a wrapper of C crypt.
// See `man 3 crypt` for more information.
func Crypt(pass, salt string) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	cPass := C.CString(pass)
	defer C.free(unsafe.Pointer(cPass))

	cSalt := C.CString(salt)
	defer C.free(unsafe.Pointer(cSalt))

	enc, err := C.crypt(cPass, cSalt)
	if enc == nil {
		return "", err
	}
	// no need to free enc
	// crypt uses the same memory address every time

	return C.GoString(enc), nil
}

// Salt dictionary [a-zA-Z0-9./].
var salt = [...]byte{
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
	'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'L', 'L', 'M',
	'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.', '/'}

// Salt random number generator.
var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

// MD5 hashes the provided password with a random salt.
// Equivalent of C's `crypt(pass, "$1$salt$")`.
func MD5(pass string) (string, error) {
	b := make([]byte, 8)
	for i := 0; i < 8; i ++ {
		b[i] = salt[rnd.Intn(len(salt))]
	}
	return Crypt(pass, fmt.Sprintf("$1$%s$", string(b)))
}
