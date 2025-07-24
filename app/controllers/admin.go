package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"backendtku/app/dto"
	"backendtku/app/helpers"
	"backendtku/app/middleware"
	"backendtku/app/models"

	"github.com/gorilla/mux"
)

func (h *Handler) GetClaimSafariAll(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.GetClaimSafariAllRequestDTO
	var Response dto.BaseResponse

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)

	user, err := h.UserRepo.FindByEmail(email)
	if err != nil {
		Response.Status = false
		Response.Message = "Invalid email or password"
		helpers.ResponseJSON(w, http.StatusUnauthorized, Response)
		return
	}

	if user.Type != "admin" {
		Response.Status = false
		Response.Message = "unauthorized"
		helpers.ResponseJSON(w, http.StatusUnauthorized, Response)
		return
	}

	filters := map[string]string{
		"policy_id":      requestDTO.PolicyId,
		"product_name":   requestDTO.ProductName,
		"date_report":    requestDTO.DateReport,
		"registrant_id":  requestDTO.RegistrantId,
		"status":         requestDTO.Status,
	}

	claims, err := h.ClaimRepo.GetAdminClaimsSafari(filters)
	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving claims"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	var claimItems []dto.GetClaimSafariAllResponseItemDTO
	for _, claim := range claims {
		image, err := h.ProductRepo.FindProductImageByName(claim.ProductName)
		if err != nil {
			Response.Status = false
			Response.Message = "Error retrieving product image"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}
		claimItems = append(claimItems, dto.GetClaimSafariAllResponseItemDTO{
			ClaimId:      claim.ClaimId,
			PolicyId:     claim.PolicyId,
			ProductName:  claim.ProductName,
			DateReport:   claim.DateReport,
			DateAccident: claim.DateAccident,
			Status:       claim.Status,
			Image:        h.Config.BaseUrl + "/upload/product/" + image,
			Evidence:     h.Config.BaseUrl + "/upload/claim/" + claim.Evidence,
			Detail:       claim.Detail,
			PolicyPdf:    h.Config.BaseUrl + "/upload/policy/pdfs/" + claim.PolicyId + ".pdf",
			RegistrantId: claim.RegistrantId,
		})
	}

	Response.Status = true
	Response.Message = "Products retrieved successfully"
	Response.Data = claimItems
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetClaimAbrorAll(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.GetClaimAbrorAllRequestDTO
	var Response dto.BaseResponse

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)

	user, err := h.UserRepo.FindByEmail(email)
	if err != nil {
		Response.Status = false
		Response.Message = "Invalid email or password"
		helpers.ResponseJSON(w, http.StatusUnauthorized, Response)
		return
	}

	if user.Type != "admin" {
		Response.Status = false
		Response.Message = "unauthorized"
		helpers.ResponseJSON(w, http.StatusUnauthorized, Response)
		return
	}

	filters := map[string]string{
		"policy_id":     requestDTO.PolicyId,
		"date_report":   requestDTO.DateReport,
		"registrant_id": requestDTO.RegistrantId,
		"status":        requestDTO.Status,
	}

	claims, err := h.ClaimRepo.GetAdminClaimsAbror(filters)
	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving claims"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	var claimItems []dto.GetClaimAbrorAllResponseItemDTO
	for _, claim := range claims {
		image, err := h.ProductRepo.FindProductAbrorImageByName(claim.ProductName)
		if err != nil {
			Response.Status = false
			Response.Message = "Error retrieving product image"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}
		claimItems = append(claimItems, dto.GetClaimAbrorAllResponseItemDTO{
			ClaimId:      claim.ClaimId,
			PolicyId:     claim.PolicyId,
			ProductName:  claim.ProductName,
			DateReport:   claim.DateReport,
			DateAccident: claim.DateAccident,
			Status:       claim.Status,
			Image:        h.Config.BaseUrl + "/upload/product/" + image,
			Evidence:     h.Config.BaseUrl + "/upload/claim/" + claim.Evidence,
			Detail:       claim.Detail,
			PolicyPdf:    h.Config.BaseUrl + "/upload/policy/pdfs/" + claim.PolicyId + ".pdf",
			RegistrantId: claim.RegistrantId,
		})
	}

	Response.Status = true
	Response.Message = "Products retrieved successfully"
	Response.Data = claimItems
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) UpdateClaim(w http.ResponseWriter, r *http.Request) {
    claimType := mux.Vars(r)["type"]
	
	var requestDTO dto.UpdateClaimRequestDTO

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
		Response.Message = "Invalid email or password"
		helpers.ResponseJSON(w, http.StatusUnauthorized, Response)
		return
	}

	if user.Type != "admin" {
		Response.Status = false
		Response.Message = "Unauthorized"
		helpers.ResponseJSON(w, http.StatusUnauthorized, Response)
		return
	}

	var claim interface{}

	switch claimType {
	case "safari":
		claim, err = h.ClaimRepo.FindClaimSafariByID(requestDTO.ClaimId)
	case "abror":
		claim, err = h.ClaimRepo.FindClaimAbrorByID(requestDTO.ClaimId)
	default:
		Response.Status = false
		Response.Message = "Invalid claim type"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	if err != nil {
		Response.Status = false
		Response.Message = "Claim not found"
		helpers.ResponseJSON(w, http.StatusNotFound, Response)
		return
	}

	switch c := claim.(type) {
	case *models.ClaimSafari:
		c.Status = requestDTO.Status
		c.Message = requestDTO.Message
		c.CoverCost = requestDTO.CoverCost
	case *models.ClaimAbror:
		c.Status = requestDTO.Status
		c.Message = requestDTO.Message
		c.CoverCost = requestDTO.CoverCost
	}

	if requestDTO.PayProof != "" {
		imageFormat := helpers.GetTypeBase64(requestDTO.PayProof)
		imageName := "ClaimProof-" + requestDTO.ClaimId + imageFormat

		var count int64
		if claimType == "safari" {
			for {
				if err := h.DB.Model(&models.ClaimSafari{}).Where("pay_proof = ?", imageName).Count(&count).Error; err != nil {
					Response.Status = false
					Response.Message = "Error checking image name"
					helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
					return
				}
				if count == 0 {
					break
				}
				count++
				imageName = fmt.Sprintf("ClaimProof-%s%d%s", requestDTO.ClaimId, count, imageFormat)
			}
		} else if claimType == "abror" {
			for {
				if err := h.DB.Model(&models.ClaimAbror{}).Where("pay_proof = ?", imageName).Count(&count).Error; err != nil {
					Response.Status = false
					Response.Message = "Error checking image name"
					helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
					return
				}
				if count == 0 {
					break
				}
				count++
				imageName = fmt.Sprintf("ClaimProof-%s%d%s", requestDTO.ClaimId, count, imageFormat)
			}
		}

		decodedImage, err := base64.StdEncoding.DecodeString(requestDTO.PayProof)
		if err != nil {
			Response.Status = false
			Response.Message = "Failed to decode image"
			helpers.ResponseJSON(w, http.StatusBadRequest, Response)
			return
		}

		imagePath := filepath.Join("upload/claim", imageName)
		if err := os.WriteFile(imagePath, decodedImage, 0644); err != nil {
			Response.Status = false
			Response.Message = "Failed to save image"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}

		switch c := claim.(type) {
		case *models.ClaimSafari:
			c.PayProof = imageName
		case *models.ClaimAbror:
			c.PayProof = imageName
		}
	}

	if err := h.ClaimRepo.UpdateClaim(claim); err != nil {
		Response.Status = false
		Response.Message = "Failed to update claim"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Claim updated successfully"
	Response.Data = nil // No specific data to return for a successful update
	helpers.ResponseJSON(w, http.StatusOK, Response)
}
