package usecase

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"backendtku/app/dto"
	"backendtku/app/helpers"
	"backendtku/app/repositories"
	"backendtku/config"
)

type ProfileUseCase interface {
	GetUserProfile(userEmail string) (dto.UserProfileResponseDTO, error)
	UpdateUserProfile(requestDTO dto.UpdateUserProfileRequestDTO, userEmail string) error
}

type profileUseCase struct {
	userRepo repositories.UserRepository
	config   *config.Config
}

func NewProfileUseCase(userRepo repositories.UserRepository, cfg *config.Config) ProfileUseCase {
	return &profileUseCase{
		userRepo: userRepo,
		config:   cfg,
	}
}

func (uc *profileUseCase) GetUserProfile(userEmail string) (dto.UserProfileResponseDTO, error) {
	user, err := uc.userRepo.FindByEmail(userEmail)
	if err != nil {
		return dto.UserProfileResponseDTO{}, errors.New("User not found")
	}

	userDTO := dto.UserProfileResponseDTO{
		Name:       user.Name,
		Gender:     user.Gender,
		Phone:      user.Phone,
		Birthplace: user.Birthplace,
		Birthdate:  user.Birthdate,
		Address:    user.Address,
		Email:      user.Email,
		Image:      uc.config.BaseUrl + "/upload/users/" + user.Image,
		ImageName:  user.Image,
	}

	return userDTO, nil
}

func (uc *profileUseCase) UpdateUserProfile(requestDTO dto.UpdateUserProfileRequestDTO, userEmail string) error {
	user, err := uc.userRepo.FindByEmail(userEmail)
	if err != nil {
		return errors.New("User not found")
	}

	user.Name = requestDTO.Name
	user.Gender = requestDTO.Gender
	user.Phone = requestDTO.Phone
	user.Birthplace = requestDTO.Birthplace
	user.Birthdate = requestDTO.Birthdate
	user.Address = requestDTO.Address

	if requestDTO.Image == "delete" {
		user.Image = ""
	} else if requestDTO.Image != "" {
		imageFormat := helpers.GetTypeBase64(requestDTO.Image)
		imageName := "ProfilePict-" + strings.ReplaceAll(user.Name, " ", "") + imageFormat

		var count int64
		for {
			isTaken, err := uc.userRepo.IsImageNameTaken(imageName)
			if err != nil {
				return errors.New("Error checking image name")
			}
			if !isTaken {
				break
			}
			count++
			imageName = fmt.Sprintf("ProfilePict-%s%d%s", strings.ReplaceAll(user.Name, " ", ""), count, imageFormat)
		}

		user.Image = imageName

		decodedImage, err := base64.StdEncoding.DecodeString(requestDTO.Image)
		if err != nil {
			return errors.New("Failed to decode image")
		}

		imagePath := filepath.Join("upload/users", imageName)
		if err := os.WriteFile(imagePath, decodedImage, 0644); err != nil {
			return errors.New("Failed to save image")
		}
	}

	if err := uc.userRepo.Save(user); err != nil {
		return errors.New("Failed to update user")
	}

	return nil
}
