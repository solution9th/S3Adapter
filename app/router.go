package app

import (
	"github.com/gorilla/mux"
)

// NewAPIRouter router list
func NewAPIRouter(r *mux.Router, api *API) {

	r.Use(loggingMiddleware)

	apiRouter := r.PathPrefix("/").Subrouter()

	// version, notice: router order
	apiRouter.Methods("GET").Path("/version").HandlerFunc(Version)

	routers := make([]*mux.Router, 0)
	routers = append(routers, apiRouter.Host("{bucket:.+}."+EndPointDomain).Subrouter())
	routers = append(routers, apiRouter.Host("{bucket:.+}."+EndPointDomain+":{port:.*}").Subrouter())
	routers = append(routers, apiRouter.PathPrefix("/{bucket}").Subrouter())

	for _, bucket := range routers {

		// Object
		// HeadObject
		bucket.Methods("HEAD").Path("/{object:.+}").HandlerFunc(api.HeadObject)
		// GetObject
		bucket.Methods("GET").Path("/{object:.+}").HandlerFunc(api.GetObject)
		// CopyObject
		bucket.Methods("PUT").Path("/{object:.+}").HeadersRegexp("X-Amz-Copy-Source", ".*?(\\/|%2F).*?").HandlerFunc(api.CopyObject)
		// PutObject
		bucket.Methods("PUT").Path("/{object:.+}").HandlerFunc(api.PutObject)
		// DeleteObject
		bucket.Methods("DELETE").Path("/{object:.+}").HandlerFunc(api.DeleteObject)

		// Bucket
		// Head Bucket
		bucket.Methods("HEAD").HandlerFunc(api.HeadBucket)
		// GET Bucket (List Objects) Version 2
		bucket.Methods("GET").HandlerFunc(api.GetBucketV2).Queries("list-type", "2")
		// GET Bucket (List Objects) Version 1
		bucket.Methods("GET").HandlerFunc(api.GetBucketV1)
		// Put Bucket
		bucket.Methods("PUT").HandlerFunc(api.PutBucket)
		// DELETE Bucket
		bucket.Methods("DELETE").HandlerFunc(api.DeleteBucket)

	}

	// ListBuckets
	apiRouter.Methods("GET").Path("/").HandlerFunc(api.ListBuckets)

	// PutApplication create application, change key
	apiRouter.Methods("PUT").Path("/").HandlerFunc(api.PutApplication)

	// DeleteApplication delete application
	apiRouter.Methods("DELETE").Path("/").HandlerFunc(api.DeleteApplication)

}
