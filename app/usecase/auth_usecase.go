package usecase

import (
	"fmt"
	"strings"
	"time"

	"backendtku/app/dto"
	"backendtku/app/helpers"
	"backendtku/app/middleware"
	"backendtku/app/models"
	"backendtku/app/repositories"
	"backendtku/config"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"errors"
)

type AuthUseCase interface {
	Register(requestDTO dto.RegisterRequestDTO) (map[string]string, error)
	VerifyEmail(tokenString string) error
	Login(requestDTO dto.LoginRequestDTO) (dto.LoginResponseDTO, string, string, error)
	RefreshToken(tokenString string) (string, error)
	Logout(refreshToken string, accessToken string) error
	ChangePassword(requestDTO dto.ChangePasswordRequestDTO, userEmail string) error
}

type authUseCase struct {
	userRepo repositories.UserRepository
	config   *config.Config
	DB       *gorm.DB
}

func NewAuthUseCase(userRepo repositories.UserRepository, cfg *config.Config, db *gorm.DB) AuthUseCase {
	return &authUseCase{
		userRepo: userRepo,
		config:   cfg,
		DB:       db,
	}
}

func (uc *authUseCase) Register(requestDTO dto.RegisterRequestDTO) (map[string]string, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(requestDTO.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("Error : Gagal Hashing Password")
	}

	token, err := middleware.GenerateToken(requestDTO.Email, 24*time.Hour)
	if err != nil {
		return nil, errors.New("Error : Gagal Membuat Token Verifikasi")
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

	if err := uc.userRepo.Create(user); err != nil {
		return nil, err
	}

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
		return nil, errors.New("Error : Email Verifikasi Gagal Terkirim")
	}

	return map[string]string{"email": user.Email}, nil
}

func (uc *authUseCase) VerifyEmail(tokenString string) error {
	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return middleware.JwtKey, nil
	})

	if err != nil || !token.Valid {
		return errors.New("Invalid token")
	}

	user, err := uc.userRepo.FindByEmailAndVerificationToken(claims.Email, tokenString)
	if err != nil {
		return errors.New("Invalid token")
	}

	user.IsVerified = true
	user.VerificationToken = ""
	if err := uc.userRepo.Save(user); err != nil {
		return errors.New("Error verifying email")
	}

	return nil
}

func (uc *authUseCase) Login(requestDTO dto.LoginRequestDTO) (dto.LoginResponseDTO, string, string, error) {
	user, err := uc.userRepo.FindByEmail(requestDTO.Email)
	if err != nil {
		return dto.LoginResponseDTO{}, "", "", errors.New("Invalid email or password")
	}

	if !user.IsVerified {
		return dto.LoginResponseDTO{}, "", "", errors.New("Akun belum diverifikasi")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(requestDTO.Password)); err != nil {
		return dto.LoginResponseDTO{}, "", "", errors.New("Invalid email or password")
	}

	accessToken, err := middleware.GenerateToken(requestDTO.Email, 15*time.Minute)
	if err != nil {
		return dto.LoginResponseDTO{}, "", "", errors.New("Error generating access token")
	}

	refreshToken, err := middleware.GenerateToken(requestDTO.Email, 7*24*time.Hour)
	if err != nil {
		return dto.LoginResponseDTO{}, "", "", errors.New("Error generating refresh token")
	}

	user.RefreshToken = refreshToken
	if err := uc.userRepo.Save(user); err != nil {
		return dto.LoginResponseDTO{}, "", "", errors.New("Error saving refresh token")
	}

	loginResponseData := dto.LoginResponseDTO{
		AccessToken: accessToken,
	}

	return loginResponseData, refreshToken, user.Type, nil
}

func (uc *authUseCase) RefreshToken(tokenString string) (string, error) {
	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return middleware.JwtKey, nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("Invalid refresh token")
	}

	newAccessToken, err := middleware.GenerateToken(claims.Email, 15*time.Minute)
	if err != nil {
		return "", errors.New("Error generating access token")
	}

	return newAccessToken, nil
}

func (uc *authUseCase) Logout(refreshToken string, accessToken string) error {
	user, err := uc.userRepo.FindByRefreshToken(refreshToken)
	if err != nil {
		return errors.New("Invalid refresh token")
	}

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
					return errors.New("Error blacklisting access token")
				}
			}
		}
	}

	user.RefreshToken = ""
	if err := uc.userRepo.Save(user); err != nil {
		return errors.New("Error logging out")
	}

	return nil
}

func (uc *authUseCase) ChangePassword(requestDTO dto.ChangePasswordRequestDTO, userEmail string) error {
	user, err := uc.userRepo.FindByEmail(userEmail)
	if err != nil {
		return errors.New("User not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(requestDTO.OldPassword)); err != nil {
		return errors.New("Old password is incorrect")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestDTO.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("Error hashing password")
	}

	user.Password = string(hashedPassword)
	if err := uc.userRepo.Save(user); err != nil {
		return errors.New("Error updating password")
	}

	return nil
}
