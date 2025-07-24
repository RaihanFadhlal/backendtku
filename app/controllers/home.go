package controllers

import (
	"net/http"
	"backendtku/app/dto"
	"backendtku/app/helpers"
)

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	var Response dto.BaseResponse
	Response.Status = true
	Response.Message = "Welcome to Takaful Umum Back End"
	Response.Data = nil
	helpers.ResponseJSON(w, http.StatusOK, Response)
}