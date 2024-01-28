package server

import (
	"net/http"

	"github.com/alvii147/flagger-api/internal/auth"
	"github.com/alvii147/flagger-api/pkg/httputils"
	"github.com/alvii147/flagger-api/pkg/logging"
	"github.com/gorilla/mux"
)

// Route sets up routes for the controller and returns a router.
func (ctrl *controller) Route() *mux.Router {
	router := mux.NewRouter()

	router.Handle(
		"/auth/users",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				ctrl.HandleCreateUser,
			),
		),
	).Methods(http.MethodPost)

	router.Handle(
		"/auth/users/me",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				auth.JWTAuthMiddleware(
					ctrl.HandleGetUserMe,
				),
			),
		),
	).Methods(http.MethodGet)

	router.Handle(
		"/api/auth/users/me",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				auth.APIKeyAuthMiddleware(
					ctrl.HandleGetUserMe,
					ctrl.authService,
				),
			),
		),
	).Methods(http.MethodGet)

	router.Handle(
		"/auth/users/activate",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				ctrl.HandleActivateUser,
			),
		),
	).Methods(http.MethodPost)

	router.Handle(
		"/auth/tokens",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				ctrl.HandleCreateJWT,
			),
		),
	).Methods(http.MethodPost)

	router.Handle(
		"/auth/tokens/refresh",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				ctrl.HandleRefreshJWT,
			),
		),
	).Methods(http.MethodPost)

	router.Handle(
		"/auth/api-keys",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				auth.JWTAuthMiddleware(
					ctrl.HandleCreateAPIKey,
				),
			),
		),
	).Methods(http.MethodPost)

	router.Handle(
		"/auth/api-keys",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				auth.JWTAuthMiddleware(
					ctrl.HandleListAPIKeys,
				),
			),
		),
	).Methods(http.MethodGet)

	router.Handle(
		"/auth/api-keys/{id:[0-9]+}",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				auth.JWTAuthMiddleware(
					ctrl.HandleDeleteAPIKey,
				),
			),
		),
	).Methods(http.MethodDelete)

	router.Handle(
		"/flags",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				auth.JWTAuthMiddleware(
					ctrl.HandleListFlags,
				),
			),
		),
	).Methods(http.MethodGet)

	router.Handle(
		"/flags",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				auth.JWTAuthMiddleware(
					ctrl.HandleCreateFlag,
				),
			),
		),
	).Methods(http.MethodPost)

	router.Handle(
		"/flags/{id:[0-9]+}",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				auth.JWTAuthMiddleware(
					ctrl.HandleGetFlagByID,
				),
			),
		),
	).Methods(http.MethodGet)

	router.Handle(
		"/api/flags/{name:[a-z0-9]+(?:-[a-z0-9]+)*}",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				auth.APIKeyAuthMiddleware(
					ctrl.HandleGetFlagByName,
					ctrl.authService,
				),
			),
		),
	).Methods(http.MethodGet)

	router.Handle(
		"/flags/{id:[0-9]+}",
		httputils.ResponseWriterMiddleware(
			logging.LoggerMiddleware(
				auth.JWTAuthMiddleware(
					ctrl.HandleUpdateFlag,
				),
			),
		),
	).Methods(http.MethodPut)

	return router
}
