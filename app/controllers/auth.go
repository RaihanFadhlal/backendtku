package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"backendtku/app/helpers"
	"backendtku/app/middleware"
	"backendtku/app/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var Request struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var Response struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Email string `json:"email"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&Request); err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}
	defer r.Body.Close()

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(Request.Password), bcrypt.DefaultCost)
	if err != nil {
		Response.Status = false
		Response.Message = "Error : Gagal Hashing Password"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	token, err := middleware.GenerateToken(Request.Email, 24*time.Hour)
	if err != nil {
		Response.Status = false
		Response.Message = "Error : Gagal Membuat Token Verifikasi"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	user := &models.User{
		ID:        uuid.New(),
		Name:      Request.Name,
		Email:     Request.Email,
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
	Response.Data.Email = user.Email
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		helpers.ResponseJSON(w, http.StatusBadRequest, map[string]string{"message": "Token is required"})
		return
	}

	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return middleware.JwtKey, nil
	})

	if err != nil || !token.Valid {
		helpers.ResponseJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid token"})
		return
	}

	user, err := h.UserRepo.FindByEmailAndVerificationToken(claims.Email, tokenString)
	if err != nil {
			helpers.ResponseJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid token"})
			return
	}

	user.IsVerified = true
	user.VerificationToken = ""
	if err := h.UserRepo.Save(user); err != nil {
		helpers.ResponseJSON(w, http.StatusInternalServerError, map[string]string{"message": "Error verifying email"})
		return
	}

	http.Redirect(w, r, "http://localhost:5173/login?verified=true", http.StatusSeeOther)
	// http.Redirect(w, r, "https://tkfl.my.id/login?verified=true", http.StatusSeeOther)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var Request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var Response struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&Request); err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}
	defer r.Body.Close()

	user, err := h.UserRepo.FindByEmail(Request.Email);
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

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(Request.Password)); err != nil {
		Response.Status = false
		Response.Message = "Invalid email or password"
		helpers.ResponseJSON(w, http.StatusUnauthorized, Response)
		return
	}

	accessToken, err := middleware.GenerateToken(Request.Email, 15*time.Minute)
	if err != nil {
		Response.Status = false
		Response.Message = "Error generating access token"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	refreshToken, err := middleware.GenerateToken(Request.Email, 7*24*time.Hour)
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

	if user.Type == "admin" {
		Response.Status = true
		Response.Message = "admin"
		Response.Data.AccessToken = accessToken
		helpers.ResponseJSON(w, http.StatusOK, Response)
		return
	}

	Response.Status = true
	Response.Message = "Login successful"
	Response.Data.AccessToken = accessToken
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		if err == http.ErrNoCookie {
			helpers.ResponseJSON(w, http.StatusUnauthorized, map[string]string{"message": "Refresh token required"})
			return
		}
		helpers.ResponseJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	tokenString := cookie.Value
	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return middleware.JwtKey, nil
	})

	if err != nil || !token.Valid {
		helpers.ResponseJSON(w, http.StatusUnauthorized, map[string]string{"message": "Invalid refresh token"})
		return
	}

	newAccessToken, err := middleware.GenerateToken(claims.Email, 15*time.Minute)
	if err != nil {
		helpers.ResponseJSON(w, http.StatusInternalServerError, map[string]string{"message": "Error generating access token"})
		return
	}

	helpers.ResponseJSON(w, http.StatusOK, map[string]string{"access_token": newAccessToken})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var Response struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
	}

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
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var Request struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	var Response struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&Request); err != nil {
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

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(Request.OldPassword)); err != nil {
		Response.Status = false
		Response.Message = "Old password is incorrect"
		helpers.ResponseJSON(w, http.StatusUnauthorized, Response)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(Request.NewPassword), bcrypt.DefaultCost)
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
	helpers.ResponseJSON(w, http.StatusOK, Response)
}
