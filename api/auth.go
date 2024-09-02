package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/user"
)

func ProtectRouteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

		claims, err := jwt.Verify(r.Context(), &jwt.VerifyParams{
			Token: sessionToken,
		})
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"access": "unauthorized"}`))
			return
		}

		usr, err := user.Get(r.Context(), claims.Subject)
		if err != nil {
			slog.Error("Error getting user from Clerk")
			return
		}
		fmt.Fprintf(w, `{"user_id": "%s", "user_banned": "%t"}`, usr.ID, usr.Banned)
		next.ServeHTTP(w, r)
	})
}
