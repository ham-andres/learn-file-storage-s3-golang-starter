package main

import "net/http"

/*
	no - cache: Doesn't mean "dont cache this" 
	it means "cache this,
	but revalidate it before serving it again".
*/

func noCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
