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
	"time"

	"github.com/google/uuid"
)

func (h *Handler) RequestClaim(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.RequestClaimRequestDTO

	var Response dto.BaseResponse

	email := r.Context().Value(middleware.UserEmailKey).(string)

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	reportDate, err := time.Parse("2006-01-02", requestDTO.DateReport)
	if err != nil {
		Response.Status = false
		Response.Message = "Invalid start date format"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	accDate, err := time.Parse("2006-01-02", requestDTO.DateAccident)
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

	count, err := h.ClaimRepo.FindEnrollmentForClaim(email, requestDTO.PolicyId)
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
		claimId = "C-" + requestDTO.PolicyId + "-" + helpers.RandomString(5)
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

	productCode, err := h.ClaimRepo.GetProductCodeByPolicy(requestDTO.PolicyId)
	if err != nil {
		Response.Status = false
		Response.Message = "Product code not found"
		helpers.ResponseJSON(w, http.StatusNotFound, Response)
		return
	}

	enrollmentId, err := h.ClaimRepo.GetEnrollmentIDByPolicy(requestDTO.PolicyId)
	if err != nil {
		Response.Status = false
		Response.Message = "Enrollment ID not found"
		helpers.ResponseJSON(w, http.StatusNotFound, Response)
		return
	}

	var imageName string
	if requestDTO.Evidence != "" {
		imageFormat := helpers.GetTypeBase64(requestDTO.Evidence)
		imageName = "ClaimEv-" + requestDTO.PolicyId + imageFormat

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
			imageName = fmt.Sprintf("ClaimEv-%s%d%s", requestDTO.PolicyId, 1, imageFormat)
		}

		decodedImage, err := base64.StdEncoding.DecodeString(requestDTO.Evidence)
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
		ProductName:  requestDTO.ProductName,
		Status:       "Diproses",
		DateReport:   requestDTO.DateReport,
		DateAccident: requestDTO.DateAccident,
		Location:     requestDTO.Location,
		Evidence:     imageName,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		PolicyId:     requestDTO.PolicyId,
		Detail:       requestDTO.Detail,
	}

	if err := h.ClaimRepo.CreateClaim(&claim); err != nil {
		Response.Status = false
		Response.Message = "Error saving claim: " + err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Claim successfull"
	Response.Data = claimId
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetClaim(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.GetClaimRequestDTO
	var Response dto.BaseResponse

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)
	claims, err := h.ClaimRepo.GetClaims(email, requestDTO.ProductName, requestDTO.DateReport)
	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving claims"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	var claimItems []dto.ClaimResponseItemDTO
	for _, claim := range claims {
		image, err := h.ProductRepo.FindProductImageByName(claim.ProductName)
		if err != nil {
			Response.Status = false
			Response.Message = "Error retrieving product image"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}
		claimItems = append(claimItems, dto.ClaimResponseItemDTO{
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
	Response.Data = claimItems
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) RequestClaimAbror(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.RequestClaimRequestDTO

	var Response dto.BaseResponse

	email := r.Context().Value(middleware.UserEmailKey).(string)

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	reportDate, err := time.Parse("2006-01-02", requestDTO.DateReport)
	if err != nil {
		Response.Status = false
		Response.Message = "Invalid start date format"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	accDate, err := time.Parse("2006-01-02", requestDTO.DateAccident)
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

	count, err := h.ClaimRepo.FindEnrollmentAbrorForClaim(email, requestDTO.PolicyId)
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
		claimId = "C-" + requestDTO.PolicyId + "-" + helpers.RandomString(5)
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

	productCode, err := h.ClaimRepo.GetProductCodeAbrorByPolicy(requestDTO.PolicyId)
	if err != nil {
		Response.Status = false
		Response.Message = "Product code not found"
		helpers.ResponseJSON(w, http.StatusNotFound, Response)
		return
	}

	enrollmentId, err := h.ClaimRepo.GetEnrollmentIDAbrorByPolicy(requestDTO.PolicyId)
	if err != nil {
		Response.Status = false
		Response.Message = "Enrollment ID not found"
		helpers.ResponseJSON(w, http.StatusNotFound, Response)
		return
	}

	var imageName string
	if requestDTO.Evidence != "" {
		imageFormat := helpers.GetTypeBase64(requestDTO.Evidence)
		imageName = "ClaimEv-" + requestDTO.PolicyId + imageFormat

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
			imageName = fmt.Sprintf("ClaimEv-%s%d%s", requestDTO.PolicyId, 1, imageFormat)
		}

		decodedImage, err := base64.StdEncoding.DecodeString(requestDTO.Evidence)
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
		ProductName:  requestDTO.ProductName,
		Status:       "Diproses",
		DateReport:   requestDTO.DateReport,
		DateAccident: requestDTO.DateAccident,
		Location:     requestDTO.Location,
		Evidence:     imageName,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		PolicyId:     requestDTO.PolicyId,
		Detail:       requestDTO.Detail,
	}

	if err := h.ClaimRepo.CreateClaimAbror(&claim); err != nil {
		Response.Status = false
		Response.Message = "Error saving claim: " + err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Enrollment successfull"
	Response.Data = map[string]string{"claim_id": claimId}
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetClaimAbror(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.GetClaimAbrorRequestDTO
	var Response dto.BaseResponse

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)
	claims, err := h.ClaimRepo.GetClaimsAbror(email, requestDTO.CarType, requestDTO.DateReport, h.DB)
	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving claims"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	var claimItems []dto.ClaimAbrorResponseItemDTO
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
		if requestDTO.CarType != "" {
			query3 = query3.Where("car_type = ?", requestDTO.CarType)
		}

		if err := query3.Find(&enroll).Error; err != nil {
			Response.Status = false
			Response.Message = "Error retrieving products"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}

		if requestDTO.CarType != "" && enroll.CarType == "" {
			Response.Status = false
			Response.Message = "Error retrieving products"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}

		claimItems = append(claimItems, dto.ClaimAbrorResponseItemDTO{
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
	Response.Data = claimItems
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetClaimDetail(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.GetClaimDetailRequestDTO
	var Response dto.BaseResponse

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)

	var claim interface{}
	var err error

	switch requestDTO.Type {
	case "safari":
		claim, err = h.ClaimRepo.GetClaimDetail(email, requestDTO.ClaimId)
	case "abror":
		claim, err = h.ClaimRepo.GetClaimAbrorDetail(email, requestDTO.ClaimId)
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

	var claimDetail dto.GetClaimDetailResponseDTO
	switch c := claim.(type) {
	case *models.ClaimSafari:
		claimDetail.Message = c.Message
		claimDetail.Status = c.Status
		claimDetail.CoverCost = c.CoverCost
		claimDetail.PayProof = h.Config.BaseUrl + "/upload/claim/" + c.PayProof
	case *models.ClaimAbror:
		claimDetail.Message = c.Message
		claimDetail.Status = c.Status
		claimDetail.CoverCost = c.CoverCost
		claimDetail.PayProof = h.Config.BaseUrl + "/upload/claim/" + c.PayProof
	}

	Response.Status = true
	Response.Message = "Claim retrieved successfully"
	Response.Data = claimDetail
	helpers.ResponseJSON(w, http.StatusOK, Response)
}
