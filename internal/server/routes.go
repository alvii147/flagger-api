package server

import (
	"net/http"

	"github.com/alvii147/flagger-api/internal/auth"
	"github.com/alvii147/flagger-api/pkg/httputils"
	"github.com/alvii147/flagger-api/pkg/logging"
)

// Route sets up routes for the controller and returns a router.
func (ctrl *controller) Route() *http.ServeMux {
	mux := http.NewServeMux()
	loggerMiddleware := func(next httputils.HandlerFunc) httputils.HandlerFunc {
		return logging.LoggerMiddleware(next, ctrl.logger)
	}
	apiKeyMiddleware := func(next httputils.HandlerFunc) httputils.HandlerFunc {
		return auth.APIKeyAuthMiddleware(next, ctrl.authService)
	}

	httputils.HandleWithMiddlewaresPOST(mux, "/auth/users", ctrl.HandleCreateUser, loggerMiddleware)
	httputils.HandleWithMiddlewaresGET(mux, "/auth/users/me", ctrl.HandleGetUserMe, auth.JWTAuthMiddleware, loggerMiddleware)
	httputils.HandleWithMiddlewaresGET(mux, "/api/auth/users/me", ctrl.HandleGetUserMe, apiKeyMiddleware, loggerMiddleware)
	httputils.HandleWithMiddlewaresPOST(mux, "/auth/users/activate", ctrl.HandleActivateUser, loggerMiddleware)
	httputils.HandleWithMiddlewaresPOST(mux, "/auth/tokens", ctrl.HandleCreateJWT, loggerMiddleware)
	httputils.HandleWithMiddlewaresPOST(mux, "/auth/tokens/refresh", ctrl.HandleRefreshJWT, loggerMiddleware)
	httputils.HandleWithMiddlewaresPOST(mux, "/auth/api-keys", ctrl.HandleCreateAPIKey, auth.JWTAuthMiddleware, loggerMiddleware)
	httputils.HandleWithMiddlewaresGET(mux, "/auth/api-keys", ctrl.HandleListAPIKeys, auth.JWTAuthMiddleware, loggerMiddleware)
	httputils.HandleWithMiddlewaresDELETE(mux, "/auth/api-keys/{id}", ctrl.HandleDeleteAPIKey, auth.JWTAuthMiddleware, loggerMiddleware)

	httputils.HandleWithMiddlewaresGET(mux, "/flags", ctrl.HandleListFlags, auth.JWTAuthMiddleware, loggerMiddleware)
	httputils.HandleWithMiddlewaresPOST(mux, "/flags", ctrl.HandleCreateFlag, auth.JWTAuthMiddleware, loggerMiddleware)
	httputils.HandleWithMiddlewaresGET(mux, "/flags/{id}", ctrl.HandleGetFlagByID, auth.JWTAuthMiddleware, loggerMiddleware)
	httputils.HandleWithMiddlewaresGET(mux, "/api/flags/{name}", ctrl.HandleGetFlagByName, apiKeyMiddleware, loggerMiddleware)
	httputils.HandleWithMiddlewaresPUT(mux, "/flags/{id}", ctrl.HandleUpdateFlag, auth.JWTAuthMiddleware, loggerMiddleware)

	return mux
}
