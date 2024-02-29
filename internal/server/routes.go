package server

import (
	"github.com/alvii147/flagger-api/internal/auth"
	"github.com/alvii147/flagger-api/pkg/httputils"
	"github.com/alvii147/flagger-api/pkg/logging"
)

// Route sets up routes for the controller.
func (ctrl *controller) Route() {
	loggerMiddleware := func(next httputils.HandlerFunc) httputils.HandlerFunc {
		return logging.LoggerMiddleware(next, ctrl.logger)
	}
	jwtMiddleware := func(next httputils.HandlerFunc) httputils.HandlerFunc {
		return auth.JWTAuthMiddleware(next, ctrl.config.SecretKey)
	}
	apiKeyMiddleware := func(next httputils.HandlerFunc) httputils.HandlerFunc {
		return auth.APIKeyAuthMiddleware(next, ctrl.authService)
	}

	ctrl.router.POST("/auth/users", ctrl.HandleCreateUser, loggerMiddleware)
	ctrl.router.GET("/auth/users/me", ctrl.HandleGetUserMe, jwtMiddleware, loggerMiddleware)
	ctrl.router.GET("/api/auth/users/me", ctrl.HandleGetUserMe, apiKeyMiddleware, loggerMiddleware)
	ctrl.router.POST("/auth/users/activate", ctrl.HandleActivateUser, loggerMiddleware)
	ctrl.router.POST("/auth/tokens", ctrl.HandleCreateJWT, loggerMiddleware)
	ctrl.router.POST("/auth/tokens/refresh", ctrl.HandleRefreshJWT, loggerMiddleware)
	ctrl.router.POST("/auth/api-keys", ctrl.HandleCreateAPIKey, jwtMiddleware, loggerMiddleware)
	ctrl.router.GET("/auth/api-keys", ctrl.HandleListAPIKeys, jwtMiddleware, loggerMiddleware)
	ctrl.router.DELETE("/auth/api-keys/{id}", ctrl.HandleDeleteAPIKey, jwtMiddleware, loggerMiddleware)

	ctrl.router.GET("/flags", ctrl.HandleListFlags, jwtMiddleware, loggerMiddleware)
	ctrl.router.POST("/flags", ctrl.HandleCreateFlag, jwtMiddleware, loggerMiddleware)
	ctrl.router.GET("/flags/{id}", ctrl.HandleGetFlagByID, jwtMiddleware, loggerMiddleware)
	ctrl.router.GET("/api/flags/{name}", ctrl.HandleGetFlagByName, apiKeyMiddleware, loggerMiddleware)
	ctrl.router.PUT("/flags/{id}", ctrl.HandleUpdateFlag, jwtMiddleware, loggerMiddleware)
}
