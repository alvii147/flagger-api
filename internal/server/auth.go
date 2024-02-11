package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/alvii147/flagger-api/pkg/api"
	"github.com/alvii147/flagger-api/pkg/errutils"
	"github.com/alvii147/flagger-api/pkg/httputils"
	"github.com/alvii147/flagger-api/pkg/logging"
	"github.com/gorilla/mux"
)

const APIKeyIDParamKey = "id"

func getAPIKeyIDParam(vars map[string]string) (int, error) {
	param, ok := vars[APIKeyIDParamKey]
	if !ok {
		return 0, errors.New("getAPIKeyIDParam failed, no param found")
	}

	apiKeyID, err := strconv.Atoi(param)
	if err != nil {
		return 0, fmt.Errorf("getAPIKeyIDParam failed to strconv.Atoi: %v", err)
	}

	return apiKeyID, nil
}

// HandleCreateUser handles creation of new Users.
// Methods: POST
// URL: /auth/users
func (ctrl *controller) HandleCreateUser(w *httputils.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()

	var req api.CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.LogWarn("HandleCreateUser failed to Decode:", err)
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInvalidRequest,
				Detail: api.ErrDetailInvalidRequestData,
			},
			http.StatusBadRequest,
		)
		return
	}

	validationPassed, validationFailures := req.Validate()
	if !validationPassed {
		logger.LogWarn("HandleCreateUser failed to Validate:", validationFailures)
		w.WriteJSON(
			api.ErrorResponse{
				Code:               api.ErrCodeInvalidRequest,
				Detail:             api.ErrDetailInvalidRequestData,
				ValidationFailures: validationFailures,
			},
			http.StatusBadRequest,
		)
		return
	}

	var wg sync.WaitGroup
	user, err := ctrl.authService.CreateUser(
		r.Context(),
		&wg,
		string(req.Email),
		string(req.Password),
		string(req.FirstName),
		string(req.LastName),
	)
	if err != nil {
		logger.LogError("HandleCreateUser failed to ctrl.authService.CreateUser:", err)
		switch {
		case errors.Is(err, errutils.ErrUserAlreadyExists):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeResourceExists,
					Detail: api.ErrDetailUserExists,
				},
				http.StatusConflict,
			)
		default:
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeInternalServerError,
					Detail: api.ErrDetailInternalServerError,
				},
				http.StatusInternalServerError,
			)
		}
		return
	}

	resp := &api.CreateUserResponse{
		UUID:      user.UUID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
	}

	w.WriteJSON(resp, http.StatusCreated)
}

// HandleActivateUser handles activation of Users.
// Methods: POST
// URL: /auth/users/activate
func (ctrl *controller) HandleActivateUser(w *httputils.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()

	var req api.ActivateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.LogWarn("HandleActivateUser failed to Decode:", err)
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInvalidRequest,
				Detail: api.ErrDetailInvalidRequestData,
			},
			http.StatusBadRequest,
		)
		return
	}

	validationPassed, validationFailures := req.Validate()
	if !validationPassed {
		logger.LogWarn("HandleActivateUser failed to Validate:", validationFailures)
		w.WriteJSON(
			api.ErrorResponse{
				Code:               api.ErrCodeInvalidRequest,
				Detail:             api.ErrDetailInvalidRequestData,
				ValidationFailures: validationFailures,
			},
			http.StatusBadRequest,
		)
		return
	}

	err = ctrl.authService.ActivateUser(
		r.Context(),
		string(req.Token),
	)
	if err != nil {
		logger.LogError("HandleActivateUser failed to ctrl.authService.ActivateUser:", err)
		switch {
		case errors.Is(err, errutils.ErrInvalidToken):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeInvalidRequest,
					Detail: api.ErrDetailInvalidToken,
				},
				http.StatusBadRequest,
			)
		case errors.Is(err, errutils.ErrUserNotFound):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeResourceNotFound,
					Detail: api.ErrDetailUserNotFound,
				},
				http.StatusNotFound,
			)
		default:
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeInternalServerError,
					Detail: api.ErrDetailInternalServerError,
				},
				http.StatusInternalServerError,
			)
		}
		return
	}

	w.WriteJSON(nil, http.StatusOK)
}

// HandleGetUserMe handles retrieval of currently authenticated User.
// Methods: GET
// URL: /auth/users/me, /api/auth/users/me
func (ctrl *controller) HandleGetUserMe(w *httputils.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()

	user, err := ctrl.authService.GetCurrentUser(r.Context())
	if err != nil {
		logger.LogError("HandleGetUserMe failed to ctrl.authService.GetCurrentUser:", err)
		switch {
		case errors.Is(err, errutils.ErrUserNotFound):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeResourceNotFound,
					Detail: api.ErrDetailUserNotFound,
				},
				http.StatusNotFound,
			)
		default:
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeInternalServerError,
					Detail: api.ErrDetailInternalServerError,
				},
				http.StatusInternalServerError,
			)
		}
		return
	}

	resp := &api.GetUserMeResponse{
		UUID:      user.UUID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
	}

	w.WriteJSON(resp, http.StatusOK)
}

// HandleCreateJWT handles authentication of User and creation of authentication JWTs.
// Methods: POST
// URL: /auth/tokens
func (ctrl *controller) HandleCreateJWT(w *httputils.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()

	var req api.CreateTokenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.LogWarn("HandleCreateJWT failed to Decode:", err)
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInvalidRequest,
				Detail: api.ErrDetailInvalidRequestData,
			},
			http.StatusBadRequest,
		)
		return
	}

	validationPassed, validationFailures := req.Validate()
	if !validationPassed {
		logger.LogWarn("HandleCreateJWT failed to Validate:", validationFailures)
		w.WriteJSON(
			api.ErrorResponse{
				Code:               api.ErrCodeInvalidRequest,
				Detail:             api.ErrDetailInvalidRequestData,
				ValidationFailures: validationFailures,
			},
			http.StatusBadRequest,
		)
		return
	}

	accessToken, refreshToken, err := ctrl.authService.CreateJWT(
		r.Context(),
		string(req.Email),
		string(req.Password),
	)
	if err != nil {
		logger.LogWarn("HandleCreateJWT failed to ctrl.authService.CreateJWT:", err)
		switch {
		case errors.Is(err, errutils.ErrInvalidCredentials):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeInvalidCredentials,
					Detail: api.ErrDetailInvalidEmailOrPassword,
				},
				http.StatusUnauthorized,
			)
		default:
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeInternalServerError,
					Detail: api.ErrDetailInternalServerError,
				},
				http.StatusInternalServerError,
			)
		}
		return
	}

	responseBody := &api.CreateTokenResponse{
		Access:  accessToken,
		Refresh: refreshToken,
	}

	w.WriteJSON(responseBody, http.StatusCreated)
}

// HandleRefreshJWT handles creation of new access JWT.
// Methods: POST
// URL: /auth/tokens/refresh
func (ctrl *controller) HandleRefreshJWT(w *httputils.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()

	var req api.RefreshTokenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.LogWarn("HandleRefreshJWT failed to Decode:", err)
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInvalidRequest,
				Detail: api.ErrDetailInvalidRequestData,
			},
			http.StatusBadRequest,
		)
		return
	}

	validationPassed, validationFailures := req.Validate()
	if !validationPassed {
		logger.LogWarn("HandleRefreshJWT failed to Validate:", validationFailures)
		w.WriteJSON(
			api.ErrorResponse{
				Code:               api.ErrCodeInvalidRequest,
				Detail:             api.ErrDetailInvalidRequestData,
				ValidationFailures: validationFailures,
			},
			http.StatusBadRequest,
		)
		return
	}

	accessToken, err := ctrl.authService.RefreshJWT(
		r.Context(),
		string(req.Refresh),
	)
	if err != nil {
		logger.LogWarn("HandleRefreshJWT failed to ctrl.authService.RefreshJWT:", err)
		switch {
		case errors.Is(err, errutils.ErrInvalidToken):
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
					Code:   api.ErrCodeInternalServerError,
					Detail: api.ErrDetailInternalServerError,
				},
				http.StatusInternalServerError,
			)
		}
		return
	}

	responseBody := &api.RefreshTokenResponse{
		Access: accessToken,
	}

	w.WriteJSON(responseBody, http.StatusCreated)
}

// HandleCreateAPIKey handles creation of new User API Key.
// Methods: POST
// URL: /auth/api-keys
func (ctrl *controller) HandleCreateAPIKey(w *httputils.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()

	var req api.CreateAPIKeyRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.LogWarn("HandleCreateAPIKey failed to Decode:", err)
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInvalidRequest,
				Detail: api.ErrDetailInvalidRequestData,
			},
			http.StatusBadRequest,
		)
		return
	}

	validationPassed, validationFailures := req.Validate()
	if !validationPassed {
		logger.LogWarn("HandleCreateAPIKey failed to Validate:", validationFailures)
		w.WriteJSON(
			api.ErrorResponse{
				Code:               api.ErrCodeInvalidRequest,
				Detail:             api.ErrDetailInvalidRequestData,
				ValidationFailures: validationFailures,
			},
			http.StatusBadRequest,
		)
		return
	}

	apiKey, key, err := ctrl.authService.CreateAPIKey(r.Context(), string(req.Name), req.ExpiresAt)
	if err != nil {
		logger.LogWarn("HandleCreateAPIKey failed to ctrl.authService.CreateAPIKey:", err)
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInternalServerError,
				Detail: api.ErrDetailInternalServerError,
			},
			http.StatusInternalServerError,
		)
		return
	}

	responseBody := &api.CreateAPIKeyResponse{
		ID:        apiKey.ID,
		RawKey:    key,
		UserUUID:  apiKey.UserUUID,
		Name:      apiKey.Name,
		CreatedAt: apiKey.CreatedAt,
		ExpiresAt: apiKey.ExpiresAt,
	}

	w.WriteJSON(responseBody, http.StatusCreated)
}

// HandleListAPIKeys handles retrieval of API Keys for currently authenticated User.
// Methods: GET
// URL: /auth/api-keys
func (ctrl *controller) HandleListAPIKeys(w *httputils.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()

	apiKeys, err := ctrl.authService.ListAPIKeys(r.Context())
	if err != nil {
		logger.LogWarn("HandleGetAPIKeys failed to ctrl.authService.ListAPIKeys:", err)
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInternalServerError,
				Detail: api.ErrDetailInternalServerError,
			},
			http.StatusInternalServerError,
		)
		return
	}

	responseBody := &api.ListAPIKeysResponse{
		Keys: make([]*api.GetAPIKeyResponse, len(apiKeys)),
	}

	for i, apiKey := range apiKeys {
		responseBody.Keys[i] = &api.GetAPIKeyResponse{
			ID:        apiKey.ID,
			UserUUID:  apiKey.UserUUID,
			Prefix:    apiKey.Prefix,
			Name:      apiKey.Name,
			CreatedAt: apiKey.CreatedAt,
			ExpiresAt: apiKey.ExpiresAt,
		}
	}

	w.WriteJSON(responseBody, http.StatusOK)
}

// HandleDeleteAPIKey handles deletion of API Keys.
// Methods: DELETE
// URL: /auth/api-keys/{id}
func (ctrl *controller) HandleDeleteAPIKey(w *httputils.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()

	apiKeyID, err := getAPIKeyIDParam(mux.Vars(r))
	if err != nil {
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInvalidRequest,
				Detail: api.ErrDetailInvalidRequestData,
			},
			http.StatusBadRequest,
		)
		return
	}

	err = ctrl.authService.DeleteAPIKey(r.Context(), apiKeyID)
	if err != nil {
		logger.LogError("HandleDeleteAPIKey failed to ctrl.authService.DeleteAPIKey:", err)
		switch {
		case errors.Is(err, errutils.ErrAPIKeyNotFound):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeResourceNotFound,
					Detail: api.ErrDetailAPIKeyNotFound,
				},
				http.StatusNotFound,
			)
		default:
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeInternalServerError,
					Detail: api.ErrDetailInternalServerError,
				},
				http.StatusInternalServerError,
			)
		}
		return
	}

	w.WriteJSON(nil, http.StatusNoContent)
}
