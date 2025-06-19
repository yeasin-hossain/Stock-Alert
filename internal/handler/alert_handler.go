package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hello-api/internal/common"
	"github.com/hello-api/internal/domain"
	"github.com/hello-api/internal/handler/dto"
)

type AlertHandler struct {
	alertService domain.AlertService
}

func NewAlertHandler(alertService domain.AlertService) *AlertHandler {
	return &AlertHandler{alertService: alertService}
}

func (h *AlertHandler) CreateAlert(w http.ResponseWriter, r *http.Request) {
	var req dto.AlertCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format")
		return
	}
	alert, err := h.alertService.CreateAlert(req)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	common.RespondWithSuccess(w, http.StatusCreated, alert)
}

func (h *AlertHandler) GetAlert(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	alert, err := h.alertService.GetAlertByID(id)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	if alert == nil {
		common.RespondWithError(w, http.StatusNotFound, "NOT_FOUND", "Alert not found")
		return
	}
	common.RespondWithSuccess(w, http.StatusOK, alert)
}

func (h *AlertHandler) GetAlertsByUser(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["userId"]
	alerts, err := h.alertService.GetAlertsByUser(userId)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	common.RespondWithSuccess(w, http.StatusOK, alerts)
}

func (h *AlertHandler) UpdateAlert(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var req dto.AlertCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format")
		return
	}
	alert, err := h.alertService.UpdateAlert(id, req)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	common.RespondWithSuccess(w, http.StatusOK, alert)
}

func (h *AlertHandler) DeleteAlert(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.alertService.DeleteAlert(id); err != nil {
		common.HandleError(w, err)
		return
	}
	common.RespondWithSuccess(w, http.StatusOK, map[string]string{"message": "Alert deleted"})
}
