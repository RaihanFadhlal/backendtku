package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"backendtku/app/helpers"
	"backendtku/app/middleware"
)

func (h *Handler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	var Response struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Name       string `json:"name"`
			Gender     string `json:"gender"`
			Phone      string `json:"phone"`
			Birthplace string `json:"birthplace"`
			Birthdate  string `json:"birthdate"`
			Address    string `json:"address"`
			Email      string `json:"email"`
			Image      string `json:"image"`
			ImageName  string `json:"image_name"`
		} `json:"data"`
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)

	user, err := h.UserRepo.FindByEmail(email)
	if err != nil {
			Response.Status = false
			Response.Message = "User not found"
			helpers.ResponseJSON(w, http.StatusNotFound, Response)
			return
	}

	Response.Status = true
	Response.Message = "User profile fetched successfully"

	Response.Data.Name = user.Name
	Response.Data.Gender = user.Gender
	Response.Data.Phone = user.Phone
	Response.Data.Birthplace = user.Birthplace
	Response.Data.Birthdate = user.Birthdate
	Response.Data.Address = user.Address
	Response.Data.Email = user.Email
	Response.Data.Image = h.Config.BaseUrl + "/upload/users/" + user.Image
	Response.Data.ImageName = user.Image

	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	var Request struct {
		Name       string `json:"name"`
		Gender     string `json:"gender"`
		Phone      string `json:"phone"`
		Birthplace string `json:"birthplace"`
		Birthdate  string `json:"birthdate"`
		Address    string `json:"address"`
		Image      string `json:"image"`
	}

	var Response struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)
	
	user, err := h.UserRepo.FindByEmail(email)
	if err != nil {
		Response.Status = false
		Response.Message = "User not found"
		helpers.ResponseJSON(w, http.StatusNotFound, Response)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&Request); err != nil {
		Response.Status = false
		Response.Message = "Invalid request payload"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	user.Name = Request.Name
	user.Gender = Request.Gender
	user.Phone = Request.Phone
	user.Birthplace = Request.Birthplace
	user.Birthdate = Request.Birthdate
	user.Address = Request.Address

	if Request.Image == "delete" {
		user.Image = ""
	} else if Request.Image != "" {
		imageFormat := helpers.GetTypeBase64(Request.Image)
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

		decodedImage, err := base64.StdEncoding.DecodeString(Request.Image)
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
