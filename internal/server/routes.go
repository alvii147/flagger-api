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

	mux.Handle(
		"POST /auth/users",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				ctrl.HandleCreateUser,
				ctrl.logger,
			),
		),
	)

	mux.Handle(
		"GET /auth/users/me",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				auth.JWTAuthMiddleware(
					ctrl.HandleGetUserMe,
				),
				ctrl.logger,
			),
		),
	)

	mux.Handle(
		"GET /api/auth/users/me",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				auth.APIKeyAuthMiddleware(
					ctrl.HandleGetUserMe,
					ctrl.authService,
				),
				ctrl.logger,
			),
		),
	)

	mux.Handle(
		"POST /auth/users/activate",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				ctrl.HandleActivateUser,
				ctrl.logger,
			),
		),
	)

	mux.Handle(
		"POST /auth/tokens",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				ctrl.HandleCreateJWT,
				ctrl.logger,
			),
		),
	)

	mux.Handle(
		"POST /auth/tokens/refresh",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				ctrl.HandleRefreshJWT,
				ctrl.logger,
			),
		),
	)

	mux.Handle(
		"POST /auth/api-keys",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				auth.JWTAuthMiddleware(
					ctrl.HandleCreateAPIKey,
				),
				ctrl.logger,
			),
		),
	)

	mux.Handle(
		"GET /auth/api-keys",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				auth.JWTAuthMiddleware(
					ctrl.HandleListAPIKeys,
				),
				ctrl.logger,
			),
		),
	)

	mux.Handle(
		"DELETE /auth/api-keys/{id}",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				auth.JWTAuthMiddleware(
					ctrl.HandleDeleteAPIKey,
				),
				ctrl.logger,
			),
		),
	)

	mux.Handle(
		"GET /flags",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				auth.JWTAuthMiddleware(
					ctrl.HandleListFlags,
				),
				ctrl.logger,
			),
		),
	)

	mux.Handle(
		"POST /flags",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				auth.JWTAuthMiddleware(
					ctrl.HandleCreateFlag,
				),
				ctrl.logger,
			),
		),
	)

	mux.Handle(
		"GET /flags/{id}",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				auth.JWTAuthMiddleware(
					ctrl.HandleGetFlagByID,
				),
				ctrl.logger,
			),
		),
	)

	mux.Handle(
		"GET /api/flags/{name}",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				auth.APIKeyAuthMiddleware(
					ctrl.HandleGetFlagByName,
					ctrl.authService,
				),
				ctrl.logger,
			),
		),
	)

	mux.Handle(
		"PUT /flags/{id}",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				auth.JWTAuthMiddleware(
					ctrl.HandleUpdateFlag,
				),
				ctrl.logger,
			),
		),
	)

	return mux
}
