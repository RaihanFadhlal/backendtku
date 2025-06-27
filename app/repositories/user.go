package repositories

import (
	"backendtku/app/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindByEmail(email string) (*models.User, error)
	FindByRefreshToken(token string) (*models.User, error)
	Create(user *models.User) error
	Save(user *models.User) error
	IsImageNameTaken(imageName string) (bool, error)
	FindByEmailAndVerificationToken(email, token string) (*models.User, error)
	FindNameByEmail(email string) (string, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
			return nil, err
	}
	return &user, nil
}

func (r *userRepository) Save(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByRefreshToken(token string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("refresh_token = ?", token).First(&user).Error; err != nil {
			return nil, err
	}
	return &user, nil
}

func (r *userRepository) IsImageNameTaken(imageName string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.User{}).Where("image = ?", imageName).Count(&count).Error; err != nil {
			return false, err
	}
	return count > 0, nil
}

func (r *userRepository) FindByEmailAndVerificationToken(email, token string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ? AND verification_token = ?", email, token).First(&user).Error; err != nil {
			return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindNameByEmail(email string) (string, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).Select("name").First(&user).Error; err != nil {
		return "", err
	}
	return user.Name, nil
}
