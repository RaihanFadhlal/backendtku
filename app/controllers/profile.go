package controllers

import (
	"encoding/json"
	"net/http"

	"backendtku/app/dto"
	"backendtku/app/helpers"
	"backendtku/app/middleware"
)

func (h *Handler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	var Response dto.BaseResponse

	email := r.Context().Value(middleware.UserEmailKey).(string)

	userDTO, err := h.ProfileUseCase.GetUserProfile(email)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusNotFound, Response)
		return
	}

	Response.Status = true
	Response.Message = "User profile fetched successfully"
	Response.Data = userDTO

	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.UpdateUserProfileRequestDTO

	var Response dto.BaseResponse

	email := r.Context().Value(middleware.UserEmailKey).(string)
	
	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request payload"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	err := h.ProfileUseCase.UpdateUserProfile(requestDTO, email)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "User profile updated successfully"
	helpers.ResponseJSON(w, http.StatusOK, Response)
}