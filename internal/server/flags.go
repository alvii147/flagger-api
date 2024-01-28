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
	"github.com/alvii147/flagger-api/pkg/logging"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	FlagIDParamKey   = "id"
	FlagNameParamKey = "name"
)

func getFlagIDParam(vars map[string]string) (int, error) {
	param, ok := vars[FlagIDParamKey]
	if !ok {
		return 0, errors.New("getFlagIDParam failed, no param found")
	}

	flagID, err := strconv.Atoi(param)
	if err != nil {
		return 0, fmt.Errorf("getFlagIDParam failed to strconv.Atoi: %v", err)
	}

	return flagID, nil
}

func getFlagNameParam(vars map[string]string) (string, error) {
	param, ok := vars[FlagNameParamKey]
	if !ok {
		return "", errors.New("getFlagNameParam failed, no param found")
	}

	return param, nil
}

// HandleCreateFlag handles creation of new User Flag.
// Methods: POST
// URL: /flags
func (ctrl *controller) HandleCreateFlag(w *httputils.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()

	var req api.CreateFlagRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.LogWarn("HandleCreateFlag failed to Decode:", err)
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInvalidRequest,
				Detail: api.ErrDetailInvalidRequestData,
			},
			http.StatusBadRequest,
		)
		return
	}

	flag, err := ctrl.flagsService.CreateFlag(r.Context(), string(req.Name))
	if err != nil {
		logger.LogWarn("HandleCreateFlag failed to ctrl.flagsService.CreateFlag:", err)
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

// HandleGetFlagByID handles retrieval of Flag of currently authenticated User using Flag ID.
// Methods: GET
// URL: /flags/{id}
func (ctrl *controller) HandleGetFlagByID(w *httputils.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()

	flagID, err := getFlagIDParam(mux.Vars(r))
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
		logger.LogError("HandleGetFlagByID failed to ctrl.flagsService.GetFlagByID:", err)
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

// HandleGetFlagByName handles retrieval of Flag of currently authenticated User using Flag ID.
// Methods: GET
// URL: /api/flags/{name}
func (ctrl *controller) HandleGetFlagByName(w *httputils.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()

	flagName, err := getFlagNameParam(mux.Vars(r))
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
		logger.LogError("HandleGetFlagByName failed to ctrl.flagsService.GetFlagByName:", err)
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
				Valid: true,
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

// HandleListFlags handles retrieval of all Flags of currently authenticated User.
// Methods: GET
// URL: /flags
func (ctrl *controller) HandleListFlags(w *httputils.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()

	flags, err := ctrl.flagsService.ListFlags(r.Context())
	if err != nil {
		logger.LogWarn("HandleListFlags failed to ctrl.flagsService.ListFlags:", err)
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

// HandleUpdateFlag handles updating of Flag of currently authenticated User.
// Methods: PUT
// URL: /flags/{id}
func (ctrl *controller) HandleUpdateFlag(w *httputils.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()

	flagID, err := getFlagIDParam(mux.Vars(r))
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
		logger.LogWarn("HandleUpdateFlag failed to Decode:", err)
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInvalidRequest,
				Detail: api.ErrDetailInvalidRequestData,
			},
			http.StatusBadRequest,
		)
		return
	}

	flag, err := ctrl.flagsService.UpdateFlag(r.Context(), flagID, string(req.Name), req.IsEnabled)
	if err != nil {
		logger.LogError("HandleUpdateFlag failed to ctrl.flagsService.UpdateFlag:", err)
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
