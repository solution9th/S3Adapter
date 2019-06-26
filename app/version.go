package app

import (
	"fmt"
	"net/http"
)

var (
	// BuildTime build time
	BuildTime = ""
	// BuildVersion build version
	BuildVersion = "devlen"
	// BuildAppName app name
	BuildAppName = "haozibi"
)

// Version get version
func Version(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("X-Build-Name", BuildAppName)
	w.Header().Set("X-Build-Time", BuildTime)
	w.Header().Set("X-Build-Version", BuildVersion)

	fmt.Fprintf(w, "%s ok", BuildAppName)
	return
}
