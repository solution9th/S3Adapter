package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/haozibi/zlog"
)

const (
	responseRequestIDKey = "x-amz-request-id"
	responseAMZIDKey     = "x-amz-id-2"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)

		t2 := time.Since(t1)

		vars := mux.Vars(r)

		zlog.ZDebug().Str("Method", r.Method).Str("Host", r.Host).Str("URL", r.RequestURI).Str("From", r.RemoteAddr).Str("UA", r.UserAgent()).Str("Bucket", vars["bucket"]).Str("Object", vars["object"]).Str("Time", fmt.Sprintf("%v", t2)).Msg("[http]")
	})
}
