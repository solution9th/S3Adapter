package app

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestVersion(t *testing.T) {

	convey.Convey("Version", t, func() {

		r := httptest.NewRequest("GET", "/version", nil)
		w := httptest.NewRecorder()

		Version(w, r)

		convey.So(w.Header().Get("X-Build-Name"), convey.ShouldEqual, BuildAppName)
		convey.So(w.Header().Get("X-Build-Version"), convey.ShouldEqual, BuildVersion)
		convey.So(w.Body.String(), convey.ShouldEqual, fmt.Sprintf("%s ok", BuildAppName))
	})
}
