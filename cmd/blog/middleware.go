package main

import "net/http"

func basicAuth(username string, password string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || username != u || password != p {
			w.Header().Set("WWW-Authenticate", `Basic realm="Admin"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
