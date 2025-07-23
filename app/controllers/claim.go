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
	"time"

	"github.com/google/uuid"
)

func (h *Handler) RequestClaim(w http.ResponseWriter, r *http.Request) {
	var Request struct {
		PolicyId     string `json:"policy_id"`
		ProductName  string `json:"product_name"`
		DateReport   string `json:"date_report"`
		DateAccident string `json:"date_acc"`
		Location     string `json:"location"`
		Detail       string `json:"detail"`
		Evidence     string `json:"evidence"`
	}

	var Response struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		ClaimId string `json:"claim_id"`
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)

	if err := json.NewDecoder(r.Body).Decode(&Request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	reportDate, err := time.Parse("2006-01-02", Request.DateReport)
	if err != nil {
		Response.Status = false
		Response.Message = "Invalid start date format"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	accDate, err := time.Parse("2006-01-02", Request.DateAccident)
	if err != nil {
		Response.Status = false
		Response.Message = "Invalid end date format"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	totalDays := int(reportDate.Sub(accDate).Hours() / 24)
	if totalDays < 0 {
		Response.Status = false
		Response.Message = "Tanggal kejadian tidak boleh lebih dari tanggal laporan"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	count, err := h.ClaimRepo.FindEnrollmentForClaim(email, Request.PolicyId)
	if err != nil {
		Response.Status = false
		Response.Message = "Error checking enrollment"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	if count == 0 {
		Response.Status = false
		Response.Message = "Polis Tidak Terdaftar"
		helpers.ResponseJSON(w, http.StatusNotFound, Response)
		return
	}

	var claimId string
	for {
		claimId = "C-" + Request.PolicyId + "-" + helpers.RandomString(5)
		isTaken, err := h.ClaimRepo.IsClaimIDTaken(claimId)
		if err != nil {
			Response.Status = false
			Response.Message = "Error checking claim ID"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}
		if !isTaken {
			break
		}
	}

	productCode, err := h.ClaimRepo.GetProductCodeByPolicy(Request.PolicyId)
	if err != nil {
		Response.Status = false
		Response.Message = "Product code not found"
		helpers.ResponseJSON(w, http.StatusNotFound, Response)
		return
	}

	enrollmentId, err := h.ClaimRepo.GetEnrollmentIDByPolicy(Request.PolicyId)
	if err != nil {
		Response.Status = false
		Response.Message = "Enrollment ID not found"
		helpers.ResponseJSON(w, http.StatusNotFound, Response)
		return
	}

	var imageName string
	if Request.Evidence != "" {
		imageFormat := helpers.GetTypeBase64(Request.Evidence)
		imageName = "ClaimEv-" + Request.PolicyId + imageFormat

		for {
			isTaken, err := h.ClaimRepo.IsClaimImageNameTaken(imageName)
			if err != nil {
				Response.Status = false
				Response.Message = "Error checking image name"
				helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
				return
			}
			if !isTaken {
				break
			}
			imageName = fmt.Sprintf("ClaimEv-%s%d%s", Request.PolicyId, 1, imageFormat)
		}

		decodedImage, err := base64.StdEncoding.DecodeString(Request.Evidence)
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
	}

	claim := models.ClaimSafari{
		ID:           uuid.New(),
		ClaimId:      claimId,
		RegistrantId: email,
		EnrollmentId: enrollmentId,
		ProductCode:  productCode,
		ProductName:  Request.ProductName,
		Status:       "Diproses",
		DateReport:   Request.DateReport,
		DateAccident: Request.DateAccident,
		Location:     Request.Location,
		Evidence:     imageName,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		PolicyId:     Request.PolicyId,
		Detail:       Request.Detail,
	}

	if err := h.ClaimRepo.CreateClaim(&claim); err != nil {
		Response.Status = false
		Response.Message = "Error saving claim: " + err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Claim successfull"
	Response.ClaimId = claimId
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetClaim(w http.ResponseWriter, r *http.Request) {
	var Request struct {
		ProductName string `json:"product_name"`
		DateReport  string `json:"date_report"`
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
		} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&Request); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)
	claims, err := h.ClaimRepo.GetClaims(email, Request.ProductName, Request.DateReport)
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
		})
	}

	Response.Status = true
	Response.Message = "Products retrieved successfully"
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) RequestClaimAbror(w http.ResponseWriter, r *http.Request) {
	var Request struct {
		PolicyId     string `json:"policy_id"`
		ProductName  string `json:"product_name"`
		DateReport   string `json:"date_report"`
		DateAccident string `json:"date_acc"`
		Location     string `json:"location"`
		Detail       string `json:"detail"`
		Evidence     string `json:"evidence"`
	}

	var Response struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		ClaimId string `json:"claim_id"`
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)

	if err := json.NewDecoder(r.Body).Decode(&Request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	reportDate, err := time.Parse("2006-01-02", Request.DateReport)
	if err != nil {
		Response.Status = false
		Response.Message = "Invalid start date format"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	accDate, err := time.Parse("2006-01-02", Request.DateAccident)
	if err != nil {
		Response.Status = false
		Response.Message = "Invalid end date format"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	totalDays := int(reportDate.Sub(accDate).Hours() / 24)
	if totalDays < 0 {
		Response.Status = false
		Response.Message = "Tanggal kejadian tidak boleh lebih dari tanggal laporan"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	count, err := h.ClaimRepo.FindEnrollmentAbrorForClaim(email, Request.PolicyId)
	if err != nil {
		Response.Status = false
		Response.Message = "Error checking enrollment"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	if count == 0 {
		Response.Status = false
		Response.Message = "Polis Tidak Terdaftar"
		helpers.ResponseJSON(w, http.StatusNotFound, Response)
		return
	}

	var claimId string
	for {
		claimId = "C-" + Request.PolicyId + "-" + helpers.RandomString(5)
		isTaken, err := h.ClaimRepo.IsClaimAbrorIDTaken(claimId)
		if err != nil {
			Response.Status = false
			Response.Message = "Error checking claim ID"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}
		if !isTaken {
			break
		}
	}

	productCode, err := h.ClaimRepo.GetProductCodeAbrorByPolicy(Request.PolicyId)
	if err != nil {
		Response.Status = false
		Response.Message = "Product code not found"
		helpers.ResponseJSON(w, http.StatusNotFound, Response)
		return
	}

	enrollmentId, err := h.ClaimRepo.GetEnrollmentIDAbrorByPolicy(Request.PolicyId)
	if err != nil {
		Response.Status = false
		Response.Message = "Enrollment ID not found"
		helpers.ResponseJSON(w, http.StatusNotFound, Response)
		return
	}

	var imageName string
	if Request.Evidence != "" {
		imageFormat := helpers.GetTypeBase64(Request.Evidence)
		imageName = "ClaimEv-" + Request.PolicyId + imageFormat

		for {
			isTaken, err := h.ClaimRepo.IsClaimAbrorImageNameTaken(imageName)
			if err != nil {
				Response.Status = false
				Response.Message = "Error checking image name"
				helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
				return
			}
			if !isTaken {
				break
			}
			imageName = fmt.Sprintf("ClaimEv-%s%d%s", Request.PolicyId, 1, imageFormat)
		}

		decodedImage, err := base64.StdEncoding.DecodeString(Request.Evidence)
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
	}

	claim := models.ClaimAbror{
		ID:           uuid.New(),
		ClaimId:      claimId,
		RegistrantId: email,
		EnrollmentId: enrollmentId,
		ProductCode:  productCode,
		ProductName:  Request.ProductName,
		Status:       "Diproses",
		DateReport:   Request.DateReport,
		DateAccident: Request.DateAccident,
		Location:     Request.Location,
		Evidence:     imageName,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		PolicyId:     Request.PolicyId,
		Detail:       Request.Detail,
	}

	if err := h.ClaimRepo.CreateClaimAbror(&claim); err != nil {
		Response.Status = false
		Response.Message = "Error saving claim: " + err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Enrollment successfull"
	Response.ClaimId = claimId
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetClaimAbror(w http.ResponseWriter, r *http.Request) {
	var Request struct {
		CarType    string `json:"car_type"`
		DateReport string `json:"date_report"`
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
			CarType      string `json:"car_type"`
		} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&Request); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)
	claims, err := h.ClaimRepo.GetClaimsAbror(email, Request.CarType, Request.DateReport, h.DB)
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

		var enroll models.EnrollmentAbror
		query3 := h.DB.Where("policy_id = ?", claim.PolicyId)
		if Request.CarType != "" {
			query3 = query3.Where("car_type = ?", Request.CarType)
		}

		if err := query3.Find(&enroll).Error; err != nil {
			Response.Status = false
			Response.Message = "Error retrieving products"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}

		if Request.CarType != "" && enroll.CarType == "" {
			Response.Status = false
			Response.Message = "Error retrieving products"
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
			CarType      string `json:"car_type"`
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
			CarType:      enroll.CarType,
		})
	}

	Response.Status = true
	Response.Message = "Products retrieved successfully"
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetClaimDetail(w http.ResponseWriter, r *http.Request) {
	var Request struct {
		ClaimId string `json:"claim_id"`
		Type    string `json:"type"`
	}
	var Response struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Status    string `json:"status"`
			Message   string `json:"message"`
			PayProof  string `json:"pay_proof"`
			CoverCost int    `json:"cover_cost"`
		} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&Request); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)

	var claim interface{}
	var err error

	switch Request.Type {
	case "safari":
		claim, err = h.ClaimRepo.GetClaimDetail(email, Request.ClaimId)
	case "abror":
		claim, err = h.ClaimRepo.GetClaimAbrorDetail(email, Request.ClaimId)
	default:
		Response.Status = false
		Response.Message = "Invalid claim type"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving claim"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	switch c := claim.(type) {
	case *models.ClaimSafari:
		Response.Data.Message = c.Message
		Response.Data.Status = c.Status
		Response.Data.CoverCost = c.CoverCost
		Response.Data.PayProof = h.Config.BaseUrl + "/upload/claim/" + c.PayProof
	case *models.ClaimAbror:
		Response.Data.Message = c.Message
		Response.Data.Status = c.Status
		Response.Data.CoverCost = c.CoverCost
		Response.Data.PayProof = h.Config.BaseUrl + "/upload/claim/" + c.PayProof
	}

	Response.Status = true
	Response.Message = "Claim retrieved successfully"
	helpers.ResponseJSON(w, http.StatusOK, Response)
}
