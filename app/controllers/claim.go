package controllers

import (
	"encoding/json"
	"net/http"

	"backendtku/app/dto"
	"backendtku/app/helpers"
	"backendtku/app/middleware"
)

func (h *Handler) RequestClaim(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.RequestClaimRequestDTO
	var Response dto.BaseResponse

	email := r.Context().Value(middleware.UserEmailKey).(string)

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	claimId, err := h.ClaimUseCase.RequestClaim(requestDTO, email)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Claim successfull"
	Response.Data = claimId
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetClaim(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.GetClaimRequestDTO
	var Response dto.BaseResponse

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)
	claimItems, err := h.ClaimUseCase.GetClaim(requestDTO, email)
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

func (h *Handler) RequestClaimAbror(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.RequestClaimRequestDTO

	var Response dto.BaseResponse

	email := r.Context().Value(middleware.UserEmailKey).(string)

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	claimId, err := h.ClaimUseCase.RequestClaimAbror(requestDTO, email)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Enrollment successfull"
	Response.Data = claimId
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetClaimAbror(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.GetClaimAbrorRequestDTO
	var Response dto.BaseResponse

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)
	claimItems, err := h.ClaimUseCase.GetClaimAbror(requestDTO, email)
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

func (h *Handler) GetClaimDetail(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.GetClaimDetailRequestDTO
	var Response dto.BaseResponse

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)

	claimDetail, err := h.ClaimUseCase.GetClaimDetail(requestDTO, email)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Claim retrieved successfully"
	Response.Data = claimDetail
	helpers.ResponseJSON(w, http.StatusOK, Response)
}