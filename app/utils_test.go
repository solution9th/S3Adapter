package app

import (
	"fmt"
	"testing"
)

func TestGenRandomString(t *testing.T) {

	for i := 0; i < 10; i++ {
		a := genAccessKey()
		b := genSecretKey()
		c := GenRandomString(-1)
		fmt.Println(a, b, c, len(a) == 20, len(b) == 40, len(c) == 20)
	}
}

func TestGetRequestID(t *testing.T) {

	getRequestID()
}
