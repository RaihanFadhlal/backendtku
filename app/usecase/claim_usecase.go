package usecase

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"backendtku/app/dto"
	"backendtku/app/helpers"
	"backendtku/app/models"
	"backendtku/app/repositories"
	"backendtku/config"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClaimUseCase interface {
	RequestClaim(requestDTO dto.RequestClaimRequestDTO, userEmail string) (string, error)
	GetClaim(requestDTO dto.GetClaimRequestDTO, userEmail string) ([]dto.ClaimResponseItemDTO, error)
	RequestClaimAbror(requestDTO dto.RequestClaimRequestDTO, userEmail string) (map[string]string, error)
	GetClaimAbror(requestDTO dto.GetClaimAbrorRequestDTO, userEmail string) ([]dto.ClaimAbrorResponseItemDTO, error)
	GetClaimDetail(requestDTO dto.GetClaimDetailRequestDTO, userEmail string) (dto.GetClaimDetailResponseDTO, error)
}

type claimUseCase struct {
	claimRepo   repositories.ClaimRepository
	productRepo repositories.ProductRepository
	config      *config.Config
	DB          *gorm.DB
}

func NewClaimUseCase(claimRepo repositories.ClaimRepository, productRepo repositories.ProductRepository, cfg *config.Config, db *gorm.DB) ClaimUseCase {
	return &claimUseCase{
		claimRepo:   claimRepo,
		productRepo: productRepo,
		config:      cfg,
		DB:          db,
	}
}

func (uc *claimUseCase) RequestClaim(requestDTO dto.RequestClaimRequestDTO, userEmail string) (string, error) {
	reportDate, err := time.Parse("2006-01-02", requestDTO.DateReport)
	if err != nil {
		return "", errors.New("Invalid start date format")
	}

	accDate, err := time.Parse("2006-01-02", requestDTO.DateAccident)
	if err != nil {
		return "", errors.New("Invalid end date format")
	}

	totalDays := int(reportDate.Sub(accDate).Hours() / 24)
	if totalDays < 0 {
		return "", errors.New("Tanggal kejadian tidak boleh lebih dari tanggal laporan")
	}

	count, err := uc.claimRepo.FindEnrollmentForClaim(userEmail, requestDTO.PolicyId)
	if err != nil {
		return "", errors.New("Error checking enrollment")
	}

	if count == 0 {
		return "", errors.New("Polis Tidak Terdaftar")
	}

	var claimId string
	for {
		claimId = "C-" + requestDTO.PolicyId + "-" + helpers.RandomString(5)
		isTaken, err := uc.claimRepo.IsClaimIDTaken(claimId)
		if err != nil {
			return "", errors.New("Error checking claim ID")
		}
		if !isTaken {
			break
		}
	}

	productCode, err := uc.claimRepo.GetProductCodeByPolicy(requestDTO.PolicyId)
	if err != nil {
		return "", errors.New("Product code not found")
	}

	enrollmentId, err := uc.claimRepo.GetEnrollmentIDByPolicy(requestDTO.PolicyId)
	if err != nil {
		return "", errors.New("Enrollment ID not found")
	}

	var imageName string
	if requestDTO.Evidence != "" {
		imageFormat := helpers.GetTypeBase64(requestDTO.Evidence)
		imageName = "ClaimEv-" + requestDTO.PolicyId + imageFormat

		for {
			isTaken, err := uc.claimRepo.IsClaimImageNameTaken(imageName)
			if err != nil {
				return "", errors.New("Error checking image name")
			}
			if !isTaken {
				break
			}
			imageName = fmt.Sprintf("ClaimEv-%s%d%s", requestDTO.PolicyId, 1, imageFormat)
		}

		decodedImage, err := base64.StdEncoding.DecodeString(requestDTO.Evidence)
		if err != nil {
			return "", errors.New("Failed to decode image")
		}

		imagePath := filepath.Join("upload/claim", imageName)
		if err := os.WriteFile(imagePath, decodedImage, 0644); err != nil {
			return "", errors.New("Failed to save image")
		}
	}

	claim := models.ClaimSafari{
		ID:           uuid.New(),
		ClaimId:      claimId,
		RegistrantId: userEmail,
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

	if err := uc.claimRepo.CreateClaim(&claim); err != nil {
		return "", errors.New("Error saving claim: " + err.Error())
	}

	return claimId, nil
}

func (uc *claimUseCase) GetClaim(requestDTO dto.GetClaimRequestDTO, userEmail string) ([]dto.ClaimResponseItemDTO, error) {
	claims, err := uc.claimRepo.GetClaims(userEmail, requestDTO.ProductName, requestDTO.DateReport)
	if err != nil {
		return nil, errors.New("Error retrieving claims")
	}

	var claimItems []dto.ClaimResponseItemDTO
	for _, claim := range claims {
		image, err := uc.productRepo.FindProductImageByName(claim.ProductName)
		if err != nil {
			return nil, errors.New("Error retrieving product image")
		}
		claimItems = append(claimItems, dto.ClaimResponseItemDTO{
			ClaimId:      claim.ClaimId,
			PolicyId:     claim.PolicyId,
			ProductName:  claim.ProductName,
			DateReport:   claim.DateReport,
			DateAccident: claim.DateAccident,
			Status:       claim.Status,
			Image:        uc.config.BaseUrl + "/upload/product/" + image,
			Evidence:     uc.config.BaseUrl + "/upload/claim/" + claim.Evidence,
			Detail:       claim.Detail,
		})
	}

	return claimItems, nil
}

func (uc *claimUseCase) RequestClaimAbror(requestDTO dto.RequestClaimRequestDTO, userEmail string) (map[string]string, error) {
	reportDate, err := time.Parse("2006-01-02", requestDTO.DateReport)
	if err != nil {
		return nil, errors.New("Invalid start date format")
	}

	accDate, err := time.Parse("2006-01-02", requestDTO.DateAccident)
	if err != nil {
		return nil, errors.New("Invalid end date format")
	}

	totalDays := int(reportDate.Sub(accDate).Hours() / 24)
	if totalDays < 0 {
		return nil, errors.New("Tanggal kejadian tidak boleh lebih dari tanggal laporan")
	}

	count, err := uc.claimRepo.FindEnrollmentAbrorForClaim(userEmail, requestDTO.PolicyId)
	if err != nil {
		return nil, errors.New("Error checking enrollment")
	}

	if count == 0 {
		return nil, errors.New("Polis Tidak Terdaftar")
	}

	var claimId string
	for {
		claimId = "C-" + requestDTO.PolicyId + "-" + helpers.RandomString(5)
		isTaken, err := uc.claimRepo.IsClaimAbrorIDTaken(claimId)
		if err != nil {
			return nil, errors.New("Error checking claim ID")
		}
		if !isTaken {
			break
		}
	}

	productCode, err := uc.claimRepo.GetProductCodeAbrorByPolicy(requestDTO.PolicyId)
	if err != nil {
		return nil, errors.New("Product code not found")
	}

	enrollmentId, err := uc.claimRepo.GetEnrollmentIDAbrorByPolicy(requestDTO.PolicyId)
	if err != nil {
		return nil, errors.New("Enrollment ID not found")
	}

	var imageName string
	if requestDTO.Evidence != "" {
		imageFormat := helpers.GetTypeBase64(requestDTO.Evidence)
		imageName = "ClaimEv-" + requestDTO.PolicyId + imageFormat

		for {
			isTaken, err := uc.claimRepo.IsClaimAbrorImageNameTaken(imageName)
			if err != nil {
				return nil, errors.New("Error checking image name")
			}
			if !isTaken {
				break
			}
			imageName = fmt.Sprintf("ClaimEv-%s%d%s", requestDTO.PolicyId, 1, imageFormat)
		}

		decodedImage, err := base64.StdEncoding.DecodeString(requestDTO.Evidence)
		if err != nil {
			return nil, errors.New("Failed to decode image")
		}

		imagePath := filepath.Join("upload/claim", imageName)
		if err := os.WriteFile(imagePath, decodedImage, 0644); err != nil {
			return nil, errors.New("Failed to save image")
		}
	}

	claim := models.ClaimAbror{
		ID:           uuid.New(),
		ClaimId:      claimId,
		RegistrantId: userEmail,
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

	if err := uc.claimRepo.CreateClaimAbror(&claim); err != nil {
		return nil, errors.New("Error saving claim: " + err.Error())
	}

	return map[string]string{"claim_id": claimId}, nil
}

func (uc *claimUseCase) GetClaimAbror(requestDTO dto.GetClaimAbrorRequestDTO, userEmail string) ([]dto.ClaimAbrorResponseItemDTO, error) {
	claims, err := uc.claimRepo.GetClaimsAbror(userEmail, requestDTO.CarType, requestDTO.DateReport, uc.DB)
	if err != nil {
		return nil, errors.New("Error retrieving claims")
	}

	var claimItems []dto.ClaimAbrorResponseItemDTO
	for _, claim := range claims {
		image, err := uc.productRepo.FindProductAbrorImageByName(claim.ProductName)
		if err != nil {
			return nil, errors.New("Error retrieving product image")
		}

		var enroll models.EnrollmentAbror
		query3 := uc.DB.Where("policy_id = ?", claim.PolicyId)
		if requestDTO.CarType != "" {
			query3 = query3.Where("car_type = ?", requestDTO.CarType)
		}

		if err := query3.Find(&enroll).Error; err != nil {
			return nil, errors.New("Error retrieving products")
		}

		if requestDTO.CarType != "" && enroll.CarType == "" {
			return nil, errors.New("Error retrieving products")
		}

		claimItems = append(claimItems, dto.ClaimAbrorResponseItemDTO{
			ClaimId:      claim.ClaimId,
			PolicyId:     claim.PolicyId,
			ProductName:  claim.ProductName,
			DateReport:   claim.DateReport,
			DateAccident: claim.DateAccident,
			Status:       claim.Status,
			Image:        uc.config.BaseUrl + "/upload/product/" + image,
			Evidence:     uc.config.BaseUrl + "/upload/claim/" + claim.Evidence,
			Detail:       claim.Detail,
			CarType:      enroll.CarType,
		})
	}

	return claimItems, nil
}

func (uc *claimUseCase) GetClaimDetail(requestDTO dto.GetClaimDetailRequestDTO, userEmail string) (dto.GetClaimDetailResponseDTO, error) {
	var claim interface{}
	var err error

	switch requestDTO.Type {
	case "safari":
		claim, err = uc.claimRepo.GetClaimDetail(userEmail, requestDTO.ClaimId)
	case "abror":
		claim, err = uc.claimRepo.GetClaimAbrorDetail(userEmail, requestDTO.ClaimId)
	default:
		return dto.GetClaimDetailResponseDTO{}, errors.New("Invalid claim type")
	}

	if err != nil {
		return dto.GetClaimDetailResponseDTO{}, errors.New("Error retrieving claim")
	}

	var claimDetail dto.GetClaimDetailResponseDTO
	switch c := claim.(type) {
	case *models.ClaimSafari:
		claimDetail.Message = c.Message
		claimDetail.Status = c.Status
		claimDetail.CoverCost = c.CoverCost
		claimDetail.PayProof = uc.config.BaseUrl + "/upload/claim/" + c.PayProof
	case *models.ClaimAbror:
		claimDetail.Message = c.Message
		claimDetail.Status = c.Status
		claimDetail.CoverCost = c.CoverCost
		claimDetail.PayProof = uc.config.BaseUrl + "/upload/claim/" + c.PayProof
	}

	return claimDetail, nil
}
