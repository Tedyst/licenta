package buildid

import (
	"net/http"

	"runtime/debug"
)

func BuildIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, ok := debug.ReadBuildInfo()
		if ok {
			for _, setting := range b.Settings {
				switch setting.Key {
				case "vcs.revision":
					w.Header().Set("X-Revision", setting.Value)
				case "vcs.branch":
					w.Header().Set("X-Branch", setting.Value)
				case "vcs.modified":
					w.Header().Set("X-Revision-Modified", setting.Value)
				case "vcs.time":
					w.Header().Set("X-Revision-Time", setting.Value)
				default:
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
