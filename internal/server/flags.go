package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/alvii147/flagger-api/pkg/api"
	"github.com/alvii147/flagger-api/pkg/errutils"
	"github.com/alvii147/flagger-api/pkg/httputils"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	FlagIDParamKey   = "id"
	FlagNameParamKey = "name"
)

func getFlagIDParam(r *http.Request) (int, error) {
	param := r.PathValue(FlagIDParamKey)
	flagID, err := strconv.Atoi(param)
	if err != nil {
		return 0, fmt.Errorf("getFlagIDParam failed to strconv.Atoi: %v", err)
	}

	return flagID, nil
}

func getFlagNameParam(r *http.Request) (string, error) {
	param := r.PathValue(FlagNameParamKey)
	if param == "" {
		return "", errors.New("getFlagNameParam failed, no param found")
	}

	return param, nil
}

// handleCreateFlag handles creation of new User Flag.
// Methods: POST
// URL: /flags
func (ctrl *controller) handleCreateFlag(w *httputils.ResponseWriter, r *http.Request) {
	var req api.CreateFlagRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ctrl.logger.LogWarn("handleCreateFlag failed to Decode:", err)
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
		ctrl.logger.LogWarn("handleCreateFlag failed to Validate:", validationFailures)
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

	flag, err := ctrl.flagsService.CreateFlag(r.Context(), string(req.Name))
	if err != nil {
		ctrl.logger.LogWarn("handleCreateFlag failed to ctrl.flagsService.CreateFlag:", err)
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInternalServerError,
				Detail: api.ErrDetailInternalServerError,
			},
			http.StatusInternalServerError,
		)
		return
	}

	responseBody := &api.CreateFlagResponse{
		ID:        flag.ID,
		UserUUID:  flag.UserUUID,
		Name:      flag.Name,
		IsEnabled: flag.IsEnabled,
		CreatedAt: flag.CreatedAt,
		UpdatedAt: flag.UpdatedAt,
	}

	w.WriteJSON(responseBody, http.StatusCreated)
}

// handleGetFlagByID handles retrieval of Flag of currently authenticated User using Flag ID.
// Methods: GET
// URL: /flags/{id}
func (ctrl *controller) handleGetFlagByID(w *httputils.ResponseWriter, r *http.Request) {
	flagID, err := getFlagIDParam(r)
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

	flag, err := ctrl.flagsService.GetFlagByID(r.Context(), flagID)
	if err != nil {
		ctrl.logger.LogError("handleGetFlagByID failed to ctrl.flagsService.GetFlagByID:", err)
		switch {
		case errors.Is(err, errutils.ErrFlagNotFound):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeResourceNotFound,
					Detail: api.ErrDetailFlagNotFound,
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

	resp := &api.GetFlagByIDResponse{
		ID:        flag.ID,
		UserUUID:  flag.UserUUID,
		Name:      flag.Name,
		IsEnabled: flag.IsEnabled,
		CreatedAt: flag.CreatedAt,
		UpdatedAt: flag.UpdatedAt,
	}

	w.WriteJSON(resp, http.StatusOK)
}

// handleGetFlagByName handles retrieval of Flag of currently authenticated User using Flag ID.
// Methods: GET
// URL: /api/flags/{name}
func (ctrl *controller) handleGetFlagByName(w *httputils.ResponseWriter, r *http.Request) {
	flagName, err := getFlagNameParam(r)
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

	flag, err := ctrl.flagsService.GetFlagByName(r.Context(), flagName)
	if err != nil {
		ctrl.logger.LogError("handleGetFlagByName failed to ctrl.flagsService.GetFlagByName:", err)
		switch {
		case errors.Is(err, errutils.ErrFlagNotFound):
			resp := &api.GetFlagByNameResponse{
				ID:        nil,
				UserUUID:  nil,
				Name:      flagName,
				IsEnabled: false,
				CreatedAt: pgtype.Timestamp{
					Valid: false,
				},
				UpdatedAt: pgtype.Timestamp{
					Valid: false,
				},
				Valid: false,
			}
			w.WriteJSON(resp, http.StatusOK)
			return
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

	resp := &api.GetFlagByNameResponse{
		ID:        &flag.ID,
		UserUUID:  &flag.UserUUID,
		Name:      flag.Name,
		IsEnabled: flag.IsEnabled,
		CreatedAt: pgtype.Timestamp{
			Time:  flag.CreatedAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamp{
			Time:  flag.UpdatedAt,
			Valid: true,
		},
		Valid: true,
	}

	w.WriteJSON(resp, http.StatusOK)
}

// handleListFlags handles retrieval of all Flags of currently authenticated User.
// Methods: GET
// URL: /flags
func (ctrl *controller) handleListFlags(w *httputils.ResponseWriter, r *http.Request) {
	flags, err := ctrl.flagsService.ListFlags(r.Context())
	if err != nil {
		ctrl.logger.LogWarn("handleListFlags failed to ctrl.flagsService.ListFlags:", err)
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInternalServerError,
				Detail: api.ErrDetailInternalServerError,
			},
			http.StatusInternalServerError,
		)
		return
	}

	responseBody := &api.ListFlagsResponse{
		Flags: make([]*api.GetFlagByIDResponse, len(flags)),
	}

	for i, flag := range flags {
		responseBody.Flags[i] = &api.GetFlagByIDResponse{
			ID:        flag.ID,
			UserUUID:  flag.UserUUID,
			Name:      flag.Name,
			IsEnabled: flag.IsEnabled,
			CreatedAt: flag.CreatedAt,
			UpdatedAt: flag.UpdatedAt,
		}
	}

	w.WriteJSON(responseBody, http.StatusOK)
}

// handleUpdateFlag handles updating of Flag of currently authenticated User.
// Methods: PUT
// URL: /flags/{id}
func (ctrl *controller) handleUpdateFlag(w *httputils.ResponseWriter, r *http.Request) {
	flagID, err := getFlagIDParam(r)
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

	var req api.UpdateFlagRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ctrl.logger.LogWarn("handleUpdateFlag failed to Decode:", err)
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
		ctrl.logger.LogWarn("handleUpdateFlag failed to Validate:", validationFailures)
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

	flag, err := ctrl.flagsService.UpdateFlag(r.Context(), flagID, req.IsEnabled)
	if err != nil {
		ctrl.logger.LogError("handleUpdateFlag failed to ctrl.flagsService.UpdateFlag:", err)
		switch {
		case errors.Is(err, errutils.ErrFlagNotFound):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeResourceNotFound,
					Detail: api.ErrDetailFlagNotFound,
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

	resp := &api.UpdateFlagResponse{
		ID:        flag.ID,
		UserUUID:  flag.UserUUID,
		Name:      flag.Name,
		IsEnabled: flag.IsEnabled,
		CreatedAt: flag.CreatedAt,
		UpdatedAt: flag.UpdatedAt,
	}

	w.WriteJSON(resp, http.StatusOK)
}
