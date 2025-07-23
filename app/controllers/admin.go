package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"backendtku/app/helpers"
	"backendtku/app/middleware"
	"backendtku/app/models"

	"github.com/gorilla/mux"
)

func (h *Handler) GetClaimSafariAll(w http.ResponseWriter, r *http.Request) {
	var Request struct {
		PolicyId     string `json:"policy_id"`
		ProductName  string `json:"product_name"`
		DateReport   string `json:"date_report"`
		RegistrantId string `json:"registrant_id"`
		Status       string `json:"status"`
	}
	var Response struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    []struct {
			ClaimId      string `json:"claim_id"`
			PolicyId     string `json:"policy_id"`
			ProductName  string `json:"product_name"`
			DateReport   string `json:"date_report"`
			DateAccident string `json:"date_accident"`
			Status       string `json:"status"`
			Image        string `json:"image"`
			Evidence     string `json:"evidence"`
			Detail       string `json:"detail"`
			PolicyPdf    string `json:"policy_pdf"`
			RegistrantId string `json:"registrant_id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&Request); err != nil {
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
		"policy_id":      Request.PolicyId,
		"product_name":   Request.ProductName,
		"date_report":    Request.DateReport,
		"registrant_id":  Request.RegistrantId,
		"status":         Request.Status,
	}

	claims, err := h.ClaimRepo.GetAdminClaimsSafari(filters)
	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving claims"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	for _, claim := range claims {
		image, err := h.ProductRepo.FindProductImageByName(claim.ProductName)
		if err != nil {
			Response.Status = false
			Response.Message = "Error retrieving product image"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}
		Response.Data = append(Response.Data, struct {
			ClaimId      string `json:"claim_id"`
			PolicyId     string `json:"policy_id"`
			ProductName  string `json:"product_name"`
			DateReport   string `json:"date_report"`
			DateAccident string `json:"date_accident"`
			Status       string `json:"status"`
			Image        string `json:"image"`
			Evidence     string `json:"evidence"`
			Detail       string `json:"detail"`
			PolicyPdf    string `json:"policy_pdf"`
			RegistrantId string `json:"registrant_id"`
		}{
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
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetClaimAbrorAll(w http.ResponseWriter, r *http.Request) {
	var Request struct {
		PolicyId   string `json:"policy_id"`
		DateReport string `json:"date_report"`
		RegistrantId string `json:"registrant_id"`
		Status     string `json:"status"`
	}
	var Response struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    []struct {
			ClaimId      string `json:"claim_id"`
			PolicyId     string `json:"policy_id"`
			ProductName  string `json:"product_name"`
			DateReport   string `json:"date_report"`
			DateAccident string `json:"date_accident"`
			Status       string `json:"status"`
			Image        string `json:"image"`
			Evidence     string `json:"evidence"`
			Detail       string `json:"detail"`
			PolicyPdf    string `json:"policy_pdf"`
			RegistrantId string `json:"registrant_id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&Request); err != nil {
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
		"policy_id":     Request.PolicyId,
		"date_report":   Request.DateReport,
		"registrant_id": Request.RegistrantId,
		"status":        Request.Status,
	}

	claims, err := h.ClaimRepo.GetAdminClaimsAbror(filters)
	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving claims"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	for _, claim := range claims {
		image, err := h.ProductRepo.FindProductAbrorImageByName(claim.ProductName)
		if err != nil {
			Response.Status = false
			Response.Message = "Error retrieving product image"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}
		Response.Data = append(Response.Data, struct {
			ClaimId      string `json:"claim_id"`
			PolicyId     string `json:"policy_id"`
			ProductName  string `json:"product_name"`
			DateReport   string `json:"date_report"`
			DateAccident string `json:"date_accident"`
			Status       string `json:"status"`
			Image        string `json:"image"`
			Evidence     string `json:"evidence"`
			Detail       string `json:"detail"`
			PolicyPdf    string `json:"policy_pdf"`
			RegistrantId string `json:"registrant_id"`
		}{
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
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) UpdateClaim(w http.ResponseWriter, r *http.Request) {
    claimType := mux.Vars(r)["type"]
	
	var Request struct {
		Status    string `json:"status"`
		Message   string `json:"message"`
		CoverCost int    `json:"cover_cost"`
		PayProof  string `json:"pay_proof"`
		ClaimId   string `json:"claim_id"`
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
		claim, err = h.ClaimRepo.FindClaimSafariByID(Request.ClaimId)
	case "abror":
		claim, err = h.ClaimRepo.FindClaimAbrorByID(Request.ClaimId)
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
		c.Status = Request.Status
		c.Message = Request.Message
		c.CoverCost = Request.CoverCost
	case *models.ClaimAbror:
		c.Status = Request.Status
		c.Message = Request.Message
		c.CoverCost = Request.CoverCost
	}

	if Request.PayProof != "" {
		imageFormat := helpers.GetTypeBase64(Request.PayProof)
		imageName := "ClaimProof-" + Request.ClaimId + imageFormat

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
				imageName = fmt.Sprintf("ClaimProof-%s%d%s", Request.ClaimId, count, imageFormat)
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
				imageName = fmt.Sprintf("ClaimProof-%s%d%s", Request.ClaimId, count, imageFormat)
			}
		}

		decodedImage, err := base64.StdEncoding.DecodeString(Request.PayProof)
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
	helpers.ResponseJSON(w, http.StatusOK, Response)
}
