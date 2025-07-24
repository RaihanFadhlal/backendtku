package dto

type RegisterRequestDTO struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequestDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponseDTO struct {
	AccessToken string `json:"access_token"`
}

type ChangePasswordRequestDTO struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type ForgotPasswordRequestDTO struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequestDTO struct {
	Email       string `json:"email" validate:"required,email"`
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type UserProfileResponseDTO struct {
	Name       string `json:"name"`
	Gender     string `json:"gender"`
	Phone      string `json:"phone"`
	Birthplace string `json:"birthplace"`
	Birthdate  string `json:"birthdate"`
	Address    string `json:"address"`
	Email      string `json:"email"`
	Image      string `json:"image"`
	ImageName  string `json:"image_name,omitempty"`
}

type UpdateUserProfileRequestDTO struct {
	Name       string `json:"name"`
	Gender     string `json:"gender"`
	Phone      string `json:"phone"`
	Birthplace string `json:"birthplace"`
	Birthdate  string `json:"birthdate"`
	Address    string `json:"address"`
	Image      string `json:"image"`
}