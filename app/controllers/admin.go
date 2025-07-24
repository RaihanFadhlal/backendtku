package controllers

import (
	"encoding/json"
	"net/http"
	"backendtku/app/dto"
	"backendtku/app/helpers"
	"backendtku/app/middleware"

	"github.com/gorilla/mux"
)



func (h *Handler) GetClaimSafariAll(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.GetClaimSafariAllRequestDTO
	var Response dto.BaseResponse

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)

	claimItems, err := h.AdminUseCase.GetClaimSafariAll(requestDTO, email)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Products retrieved successfully"
	Response.Data = claimItems
	helpers.ResponseJSON(w, http.StatusOK, Response)
}


func (h *Handler) GetClaimAbrorAll(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.GetClaimAbrorAllRequestDTO
	var Response dto.BaseResponse

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)

	claimItems, err := h.AdminUseCase.GetClaimAbrorAll(requestDTO, email)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Products retrieved successfully"
	Response.Data = claimItems
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) UpdateClaim(w http.ResponseWriter, r *http.Request) {
    claimType := mux.Vars(r)["type"]
	
	var requestDTO dto.UpdateClaimRequestDTO

	var Response dto.BaseResponse

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request payload"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)

	err := h.AdminUseCase.UpdateClaim(claimType, requestDTO, email)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Claim updated successfully"
	Response.Data = nil // No specific data to return for a successful update
	helpers.ResponseJSON(w, http.StatusOK, Response)
}
