package app

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const (
	alphabetnum = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func getRequestID() string {
	return fmt.Sprintf("%X", time.Now().UnixNano())
}

func genAccessKey() string {
	return GenRandomString(20)
}

func genSecretKey() string {
	return GenRandomString(40)
}

// GenRandomString generate random string
func GenRandomString(size int) string {
	if size <= 0 {
		size = 20
	}
	bts := genRandomBytes(size, alphabetnum)
	if bts[0] == '0' {
		bts[0] = 'x'
	}
	return string(bts)
}

func genRandomBytes(size int, base string) []byte {
	bts := make([]byte, size)
	n := len(base)
	// ignore error
	rand.Read(bts)
	for i, b := range bts {
		bts[i] = base[b%byte(n)]
	}
	return bts
}

// todo, 生成 request id 等跟踪信息
func newContext(w http.ResponseWriter, r *http.Request, api string) context.Context {

	vars := mux.Vars(r)
	bucket, object, prefix := vars["bucket"], vars["object"], vars["prefix"]

	_, _, _ = bucket, object, prefix
	// ReqInfo
	return context.Background()
}
