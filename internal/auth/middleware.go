package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/alvii147/flagger-api/pkg/api"
	"github.com/alvii147/flagger-api/pkg/errutils"
	"github.com/alvii147/flagger-api/pkg/httputils"
)

// JWTAuthMiddleware parses and validates JWT from authorization header.
// If authentication fails, it returns 401.
// If authentication is successful, it sets User UUID in context.
func JWTAuthMiddleware(next httputils.HandlerFunc) httputils.HandlerFunc {
	return httputils.HandlerFunc(func(w *httputils.ResponseWriter, r *http.Request) {
		token, ok := httputils.GetAuthorizationHeader(r.Header, "Bearer")
		if !ok {
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeMissingCredentials,
					Detail: api.ErrDetailMissingCredentials,
				},
				http.StatusUnauthorized,
			)
			return
		}

		claims, ok := validateAuthJWT(token, JWTTypeAccess)
		if !ok {
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeInvalidCredentials,
					Detail: api.ErrDetailInvalidToken,
				},
				http.StatusUnauthorized,
			)
			return
		}

		next.ServeHTTP(w, r.Clone(context.WithValue(r.Context(), AuthContextKeyUserUUID, claims.Subject)))
	})
}

// APIKeyAuthMiddleware authenticates user using provided API Key.
// If authentication fails, it returns 401.
// If authentication is successful, it sets User UUID in context.
func APIKeyAuthMiddleware(next httputils.HandlerFunc, svc Service) httputils.HandlerFunc {
	return httputils.HandlerFunc(func(w *httputils.ResponseWriter, r *http.Request) {
		rawKey, ok := httputils.GetAuthorizationHeader(r.Header, "X-API-Key")
		if !ok {
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeMissingCredentials,
					Detail: api.ErrDetailMissingCredentials,
				},
				http.StatusUnauthorized,
			)
			return
		}

		apiKey, err := svc.FindAPIKey(r.Context(), rawKey)
		if err != nil {
			switch {
			case errors.Is(err, errutils.ErrAPIKeyNotFound):
				w.WriteJSON(
					api.ErrorResponse{
						Code:   api.ErrCodeInvalidCredentials,
						Detail: api.ErrDetailInvalidToken,
					},
					http.StatusUnauthorized,
				)
			default:
				w.WriteJSON(
					api.ErrorResponse{
						Code:   api.ErrCodeInvalidCredentials,
						Detail: api.ErrDetailInvalidToken,
					},
					http.StatusUnauthorized,
				)
			}
			return
		}

		next.ServeHTTP(w, r.Clone(context.WithValue(r.Context(), AuthContextKeyUserUUID, apiKey.UserUUID)))
	})
}
