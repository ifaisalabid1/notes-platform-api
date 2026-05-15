package internalmw

import (
	"crypto/subtle"
	"net/http"

	"github.com/ifaisalabid1/notes-platform-api/internal/http/response"
)

func RequireWorkerSecret(expectedSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			providedSecret := r.Header.Get("X-Worker-Secret")

			if expectedSecret == "" || providedSecret == "" {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "Unauthorized.")
				return
			}

			if subtle.ConstantTimeCompare([]byte(providedSecret), []byte(expectedSecret)) != 1 {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "Unauthorized.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
