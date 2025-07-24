package controllers

import (
	"net/http"
	"backendtku/app/helpers"
)

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	Response := h.HomeUseCase.GetHomeMessage()
	helpers.ResponseJSON(w, http.StatusOK, Response)
}
