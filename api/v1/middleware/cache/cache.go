package cache

import "net/http"

func CacheControlHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private, max-age=31536000")
		next.ServeHTTP(w, r)
	})
}
