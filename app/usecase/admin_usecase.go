package usecase

import (
	"backendtku/app/dto"
	"backendtku/app/models"
	"backendtku/app/repositories"
	"backendtku/app/helpers"
	"backendtku/config"
	"errors"
	"fmt"
	"encoding/base64"
	"path/filepath"
	"os"
)

type AdminUseCase interface {
	GetClaimSafariAll(requestDTO dto.GetClaimSafariAllRequestDTO, userEmail string) ([]dto.GetClaimSafariAllResponseItemDTO, error)
	GetClaimAbrorAll(requestDTO dto.GetClaimAbrorAllRequestDTO, userEmail string) ([]dto.GetClaimAbrorAllResponseItemDTO, error)
	UpdateClaim(claimType string, requestDTO dto.UpdateClaimRequestDTO, userEmail string) error
}

type adminUseCase struct {
	userRepo    repositories.UserRepository
	claimRepo   repositories.ClaimRepository
	productRepo repositories.ProductRepository
	config      *config.Config
}

func NewAdminUseCase(userRepo repositories.UserRepository, claimRepo repositories.ClaimRepository, productRepo repositories.ProductRepository, cfg *config.Config) AdminUseCase {
	return &adminUseCase{
		userRepo:    userRepo,
		claimRepo:   claimRepo,
		productRepo: productRepo,
		config:      cfg,
	}
}

func (uc *adminUseCase) GetClaimSafariAll(requestDTO dto.GetClaimSafariAllRequestDTO, userEmail string) ([]dto.GetClaimSafariAllResponseItemDTO, error) {
	user, err := uc.userRepo.FindByEmail(userEmail)
	if err != nil {
		return nil, errors.New("Invalid email or password")
	}

	if user.Type != "admin" {
		return nil, errors.New("unauthorized")
	}

	filters := map[string]string{
		"policy_id":      requestDTO.PolicyId,
		"product_name":   requestDTO.ProductName,
		"date_report":    requestDTO.DateReport,
		"registrant_id":  requestDTO.RegistrantId,
		"status":         requestDTO.Status,
	}

	claims, err := uc.claimRepo.GetAdminClaimsSafari(filters)
	if err != nil {
		return nil, errors.New("Error retrieving claims")
	}

	var claimItems []dto.GetClaimSafariAllResponseItemDTO
	for _, claim := range claims {
		image, err := uc.productRepo.FindProductImageByName(claim.ProductName)
		if err != nil {
			return nil, errors.New("Error retrieving product image")
		}
		claimItems = append(claimItems, dto.GetClaimSafariAllResponseItemDTO{
			ClaimId:      claim.ClaimId,
			PolicyId:     claim.PolicyId,
			ProductName:  claim.ProductName,
			DateReport:   claim.DateReport,
			DateAccident: claim.DateAccident,
			Status:       claim.Status,
			Image:        uc.config.BaseUrl + "/upload/product/" + image,
			Evidence:     uc.config.BaseUrl + "/upload/claim/" + claim.Evidence,
			Detail:       claim.Detail,
			PolicyPdf:    uc.config.BaseUrl + "/upload/policy/pdfs/" + claim.PolicyId + ".pdf",
			RegistrantId: claim.RegistrantId,
		})
	}

	return claimItems, nil
}

func (uc *adminUseCase) GetClaimAbrorAll(requestDTO dto.GetClaimAbrorAllRequestDTO, userEmail string) ([]dto.GetClaimAbrorAllResponseItemDTO, error) {
	user, err := uc.userRepo.FindByEmail(userEmail)
	if err != nil {
		return nil, errors.New("Invalid email or password")
	}

	if user.Type != "admin" {
		return nil, errors.New("unauthorized")
	}

	filters := map[string]string{
		"policy_id":     requestDTO.PolicyId,
		"date_report":   requestDTO.DateReport,
		"registrant_id": requestDTO.RegistrantId,
		"status":        requestDTO.Status,
	}

	claims, err := uc.claimRepo.GetAdminClaimsAbror(filters)
	if err != nil {
		return nil, errors.New("Error retrieving claims")
	}

	var claimItems []dto.GetClaimAbrorAllResponseItemDTO
	for _, claim := range claims {
		image, err := uc.productRepo.FindProductAbrorImageByName(claim.ProductName)
		if err != nil {
			return nil, errors.New("Error retrieving product image")
		}
		claimItems = append(claimItems, dto.GetClaimAbrorAllResponseItemDTO{
			ClaimId:      claim.ClaimId,
			PolicyId:     claim.PolicyId,
			ProductName:  claim.ProductName,
			DateReport:   claim.DateReport,
			DateAccident: claim.DateAccident,
			Status:       claim.Status,
			Image:        uc.config.BaseUrl + "/upload/product/" + image,
			Evidence:     uc.config.BaseUrl + "/upload/claim/" + claim.Evidence,
			Detail:       claim.Detail,
			PolicyPdf:    uc.config.BaseUrl + "/upload/policy/pdfs/" + claim.PolicyId + ".pdf",
			RegistrantId: claim.RegistrantId,
		})
	}

	return claimItems, nil
}

func (uc *adminUseCase) UpdateClaim(claimType string, requestDTO dto.UpdateClaimRequestDTO, userEmail string) error {
	user, err := uc.userRepo.FindByEmail(userEmail)
	if err != nil {
		return errors.New("Invalid email or password")
	}

	if user.Type != "admin" {
		return errors.New("Unauthorized")
	}

	var claim interface{}

	switch claimType {
	case "safari":
		claim, err = uc.claimRepo.FindClaimSafariByID(requestDTO.ClaimId)
	case "abror":
		claim, err = uc.claimRepo.FindClaimAbrorByID(requestDTO.ClaimId)
	default:
		return errors.New("Invalid claim type")
	}

	if err != nil {
		return errors.New("Claim not found")
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
				count, err = uc.claimRepo.CountClaimSafariPayProof(imageName)
				if err != nil {
					return errors.New("Error checking image name")
				}
				if count == 0 {
					break
				}
				count++
				imageName = fmt.Sprintf("ClaimProof-%s%d%s", requestDTO.ClaimId, count, imageFormat)
			}
		} else if claimType == "abror" {
			for {
				count, err = uc.claimRepo.CountClaimAbrorPayProof(imageName)
				if err != nil {
					return errors.New("Error checking image name")
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
			return errors.New("Failed to decode image")
		}

		imagePath := filepath.Join("upload/claim", imageName)
		if err := os.WriteFile(imagePath, decodedImage, 0644); err != nil {
			return errors.New("Failed to save image")
		}

		switch c := claim.(type) {
		case *models.ClaimSafari:
			c.PayProof = imageName
		case *models.ClaimAbror:
			c.PayProof = imageName
		}
	}

	if err := uc.claimRepo.UpdateClaim(claim); err != nil {
		return errors.New("Failed to update claim")
	}

	return nil
}
