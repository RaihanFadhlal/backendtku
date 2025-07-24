package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"backendtku/app/dto"
	"backendtku/app/helpers"
	"backendtku/app/middleware"
)

func (h *Handler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	var Response dto.BaseResponse

	email := r.Context().Value(middleware.UserEmailKey).(string)

	user, err := h.UserRepo.FindByEmail(email)
	if err != nil {
		Response.Status = false
		Response.Message = "User not found"
		helpers.ResponseJSON(w, http.StatusNotFound, Response)
		return
	}

	userDTO := dto.UserProfileResponseDTO{
		Name:       user.Name,
		Gender:     user.Gender,
		Phone:      user.Phone,
		Birthplace: user.Birthplace,
		Birthdate:  user.Birthdate,
		Address:    user.Address,
		Email:      user.Email,
		Image:      h.Config.BaseUrl + "/upload/users/" + user.Image,
		ImageName:  user.Image,
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
	
	user, err := h.UserRepo.FindByEmail(email)
	if err != nil {
		Response.Status = false
		Response.Message = "User not found"
		helpers.ResponseJSON(w, http.StatusNotFound, Response)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request payload"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
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
			isTaken, err := h.UserRepo.IsImageNameTaken(imageName)
			if err != nil {
					Response.Status = false
					Response.Message = "Error checking image name"
					helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
					return
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
			Response.Status = false
			Response.Message = "Failed to decode image"
			helpers.ResponseJSON(w, http.StatusBadRequest, Response)
			return
		}

		imagePath := filepath.Join("upload/users", imageName)
		if err := os.WriteFile(imagePath, decodedImage, 0644); err != nil {
			Response.Status = false
			Response.Message = "Failed to save image"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}
	}

	if err := h.UserRepo.Save(user); err != nil {
		Response.Status = false
		Response.Message = "Failed to update user"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "User profile updated successfully"
	helpers.ResponseJSON(w, http.StatusOK, Response)
}
