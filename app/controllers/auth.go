package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"backendtku/app/dto"
	"backendtku/app/helpers"
	"backendtku/app/middleware"
	"backendtku/app/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(requestDTO.Password), bcrypt.DefaultCost)
	if err != nil {
		Response.Status = false
		Response.Message = "Error : Gagal Hashing Password"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	token, err := middleware.GenerateToken(requestDTO.Email, 24*time.Hour)
	if err != nil {
		Response.Status = false
		Response.Message = "Error : Gagal Membuat Token Verifikasi"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	user := &models.User{
		ID:        uuid.New(),
		Name:      requestDTO.Name,
		Email:     requestDTO.Email,
		Password:  string(hashPassword),
		VerificationToken: token,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.UserRepo.Create(user); err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	// verificationLink := fmt.Sprintf("https://api.tkfl.my.id/takafulumum/verify?token=%s", token)
	verificationLink := fmt.Sprintf("http://localhost:9000/verify?token=%s", token)
	content := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <body width="100%%" style="background-color: #f1f1f1; font-family: 'Lato', sans-serif; font-weight: 400; font-size: 15px; line-height: 1.8; color: rgba(0, 0, 0, 0.4);">
      <div style="max-width: 600px; margin: 0 auto">
        <table>
          <tr>
            <td style="padding: 1em 2.5em 0 2.5em">
              <table border="0" cellpadding="0" cellspacing="0" width="100%%">
                <tr>
                  <td style="text-align: center">
                    <h1 style="color: #17bebb;">Takaful Umum</h1>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          <tr>
            <td style="padding: 0 0 4em 0">
              <table>
                <tr>
                  <td style="padding: 0 2.5em; text-align: center; padding-bottom: 1em;">
                    <div>
                      <h2 style="font-family: 'Lato', sans-serif; color: #000000; margin-top: 0; font-weight: 400;">
                        Halo %s, 
                        <br>Sebelum melakukan login, klik tombol dibawah untuk verifikasi email kamu
                      </h2>
                    </div>
                  </td>
                </tr>
                <tr>
                  <td style="text-align: center">
                    <div>
                      <a href="%s"><button style="text-decoration: none; color: #ffffff; background: #17bebb; padding: 10px 15px; border-radius: 5px; display: inline-block;">Verifikasi</button></a>
                    </div>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
        </table>
      </div>
    </body>
    </html>
    `, user.Name, verificationLink)

	err = helpers.SendEmail(user.Email, "Verifikasi Akun Takaful Umum", content)
	if err != nil {
		Response.Status = false
		Response.Message = "Error : Email Verifikasi Gagal Terkirim"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Registrasi Berhasil, Cek Email Untuk Verifikasi Akun!"
	Response.Data = map[string]string{"email": user.Email}
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

	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return middleware.JwtKey, nil
	})

	if err != nil || !token.Valid {
		Response.Status = false
		Response.Message = "Invalid token"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	user, err := h.UserRepo.FindByEmailAndVerificationToken(claims.Email, tokenString)
	if err != nil {
		Response.Status = false
		Response.Message = "Invalid token"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	user.IsVerified = true
	user.VerificationToken = ""
	if err := h.UserRepo.Save(user); err != nil {
		Response.Status = false
		Response.Message = "Error verifying email"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	http.Redirect(w, r, "http://localhost:5173/login?verified=true", http.StatusSeeOther)
	// http.Redirect(w, r, "https://tkfl.my.id/login?verified=true", http.StatusSeeOther)
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

	user, err := h.UserRepo.FindByEmail(requestDTO.Email)
	if err != nil {
		Response.Status = false
		Response.Message = "Invalid email or password"
		helpers.ResponseJSON(w, http.StatusUnauthorized, Response)
		return
	}

	if !user.IsVerified {
		Response.Status = false
		Response.Message = "Akun belum diverifikasi"
		helpers.ResponseJSON(w, http.StatusUnauthorized, Response)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(requestDTO.Password)); err != nil {
		Response.Status = false
		Response.Message = "Invalid email or password"
		helpers.ResponseJSON(w, http.StatusUnauthorized, Response)
		return
	}

	accessToken, err := middleware.GenerateToken(requestDTO.Email, 15*time.Minute)
	if err != nil {
		Response.Status = false
		Response.Message = "Error generating access token"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	refreshToken, err := middleware.GenerateToken(requestDTO.Email, 7*24*time.Hour)
	if err != nil {
		Response.Status = false
		Response.Message = "Error generating refresh token"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	user.RefreshToken = refreshToken
	if err := h.UserRepo.Save(user); err != nil {
		Response.Status = false
		Response.Message = "Error saving refresh token"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
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
		Value:    user.Type,
		Expires:  time.Now().Add(1 * 24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	loginResponseData := dto.LoginResponseDTO{
		AccessToken: accessToken,
	}

	if user.Type == "admin" {
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
	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return middleware.JwtKey, nil
	})

	if err != nil || !token.Valid {
		Response.Status = false
		Response.Message = "Invalid refresh token"
		helpers.ResponseJSON(w, http.StatusUnauthorized, Response)
		return
	}

	newAccessToken, err := middleware.GenerateToken(claims.Email, 15*time.Minute)
	if err != nil {
		Response.Status = false
		Response.Message = "Error generating access token"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
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

	user, err:= h.UserRepo.FindByRefreshToken(refreshToken)
	if err != nil {
		Response.Status = false
		Response.Message = "Invalid refresh token"
		helpers.ResponseJSON(w, http.StatusUnauthorized, Response)
		return
	}

	accessToken := r.Header.Get("Authorization")
	if accessToken != "" {
		tokenString := strings.TrimPrefix(accessToken, "Bearer ")
		claims := &middleware.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return middleware.JwtKey, nil
		})
		if err == nil && token.Valid {
			exp := time.Unix(claims.StandardClaims.ExpiresAt, 0)
			duration := time.Until(exp)
			if duration > 0 {
				err = middleware.BlacklistToken(tokenString, duration)
				if err != nil {
					Response.Status = false
					Response.Message = "Error blacklisting access token"
					helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
					return
				}
			}
		}
	}

	user.RefreshToken = ""
	if err := h.UserRepo.Save(user); err != nil {
		Response.Status = false
		Response.Message = "Error logging out"
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
	user, err := h.UserRepo.FindByEmail(email)
	if err != nil {
			Response.Status = false
			Response.Message = "User not found"
			helpers.ResponseJSON(w, http.StatusNotFound, Response)
			return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(requestDTO.OldPassword)); err != nil {
		Response.Status = false
		Response.Message = "Old password is incorrect"
		helpers.ResponseJSON(w, http.StatusUnauthorized, Response)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestDTO.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		Response.Status = false
		Response.Message = "Error hashing password"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	user.Password = string(hashedPassword)
	if err := h.UserRepo.Save(user); err != nil {
		Response.Status = false
		Response.Message = "Error updating password"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Password changed successfully"
	Response.Data = nil // No specific data to return for a successful password change
	helpers.ResponseJSON(w, http.StatusOK, Response)
}
