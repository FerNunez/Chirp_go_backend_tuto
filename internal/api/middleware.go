package api

import "net/http"

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {

	// CHECK AND LEARN THIS??
	// create wrapping handler Func
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Increment 1
		cfg.FileserverHits.Add(1)
		// call wrapped handler
		next.ServeHTTP(w, r)
	})

}
