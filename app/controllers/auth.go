package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"backendtku/app/dto"
	"backendtku/app/helpers"
	"backendtku/app/middleware"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.RegisterRequestDTO
	var Response dto.BaseResponse

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}
	defer r.Body.Close()

	data, err := h.AuthUseCase.Register(requestDTO)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Registrasi Berhasil, Cek Email Untuk Verifikasi Akun!"
	Response.Data = data
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var Response dto.BaseResponse

	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		Response.Status = false
		Response.Message = "Token is required"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	err := h.AuthUseCase.VerifyEmail(tokenString)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	http.Redirect(w, r, "http://localhost:5173/login?verified=true", http.StatusSeeOther)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.LoginRequestDTO
	var Response dto.BaseResponse

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}
	defer r.Body.Close()

	loginResponseData, refreshToken, userType, err := h.AuthUseCase.Login(requestDTO)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusUnauthorized, Response)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "user_type",
		Value:    userType,
		Expires:  time.Now().Add(1 * 24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	if userType == "admin" {
		Response.Status = true
		Response.Message = "admin"
		Response.Data = loginResponseData
		helpers.ResponseJSON(w, http.StatusOK, Response)
		return
	}

	Response.Status = true
	Response.Message = "Login successful"
	Response.Data = loginResponseData
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var Response dto.BaseResponse

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		if err == http.ErrNoCookie {
			Response.Status = false
			Response.Message = "Refresh token required"
			helpers.ResponseJSON(w, http.StatusUnauthorized, Response)
			return
		}
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	tokenString := cookie.Value

	newAccessToken, err := h.AuthUseCase.RefreshToken(tokenString)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusUnauthorized, Response)
		return
	}

	Response.Status = true
	Response.Message = "Access token refreshed successfully"
	Response.Data = map[string]string{"access_token": newAccessToken}
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var Response dto.BaseResponse

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		Response.Status = false
		Response.Message = "No refresh token provided"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	refreshToken := cookie.Value
	accessToken := r.Header.Get("Authorization")

	err = h.AuthUseCase.Logout(refreshToken, accessToken)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})

	Response.Status = true
	Response.Message = "Logout successful"
	Response.Data = nil // No specific data to return for a successful logout
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.ChangePasswordRequestDTO
	var Response dto.BaseResponse

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request payload"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)

	err := h.AuthUseCase.ChangePassword(requestDTO, email)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Password changed successfully"
	Response.Data = nil // No specific data to return for a successful password change
	helpers.ResponseJSON(w, http.StatusOK, Response)
}