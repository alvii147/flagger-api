package server

import (
	"github.com/alvii147/flagger-api/internal/auth"
	"github.com/alvii147/flagger-api/pkg/httputils"
	"github.com/alvii147/flagger-api/pkg/logging"
)

// route sets up routes for the controller.
func (ctrl *controller) route() {
	loggerMiddleware := func(next httputils.HandlerFunc) httputils.HandlerFunc {
		return logging.LoggerMiddleware(next, ctrl.logger)
	}
	jwtMiddleware := func(next httputils.HandlerFunc) httputils.HandlerFunc {
		return auth.JWTAuthMiddleware(next, ctrl.config.SecretKey)
	}
	apiKeyMiddleware := func(next httputils.HandlerFunc) httputils.HandlerFunc {
		return auth.APIKeyAuthMiddleware(next, ctrl.authService)
	}

	ctrl.router.POST("/auth/users", ctrl.handleCreateUser, loggerMiddleware)
	ctrl.router.GET("/auth/users/me", ctrl.handleGetUserMe, jwtMiddleware, loggerMiddleware)
	ctrl.router.GET("/api/auth/users/me", ctrl.handleGetUserMe, apiKeyMiddleware, loggerMiddleware)
	ctrl.router.POST("/auth/users/activate", ctrl.handleActivateUser, loggerMiddleware)
	ctrl.router.POST("/auth/tokens", ctrl.handleCreateJWT, loggerMiddleware)
	ctrl.router.POST("/auth/tokens/refresh", ctrl.handleRefreshJWT, loggerMiddleware)
	ctrl.router.POST("/auth/api-keys", ctrl.handleCreateAPIKey, jwtMiddleware, loggerMiddleware)
	ctrl.router.GET("/auth/api-keys", ctrl.handleListAPIKeys, jwtMiddleware, loggerMiddleware)
	ctrl.router.DELETE("/auth/api-keys/{id}", ctrl.handleDeleteAPIKey, jwtMiddleware, loggerMiddleware)

	ctrl.router.GET("/flags", ctrl.handleListFlags, jwtMiddleware, loggerMiddleware)
	ctrl.router.POST("/flags", ctrl.handleCreateFlag, jwtMiddleware, loggerMiddleware)
	ctrl.router.GET("/flags/{id}", ctrl.handleGetFlagByID, jwtMiddleware, loggerMiddleware)
	ctrl.router.GET("/api/flags/{name}", ctrl.handleGetFlagByName, apiKeyMiddleware, loggerMiddleware)
	ctrl.router.PUT("/flags/{id}", ctrl.handleUpdateFlag, jwtMiddleware, loggerMiddleware)
}
