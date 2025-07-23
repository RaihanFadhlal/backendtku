package repositories

import (
	"backendtku/app/models"

	"gorm.io/gorm"
)

type ClaimRepository interface {

	// Safari
	FindEnrollmentForClaim(email, policyID string) (int64, error)
	IsClaimIDTaken(claimID string) (bool, error)
	GetProductCodeByPolicy(policyID string) (string, error)
	GetEnrollmentIDByPolicy(policyID string) (string, error)
	CreateClaim(claim *models.ClaimSafari) error
	GetClaims(email, productName, dateReport string) ([]models.ClaimSafari, error)
	GetClaimDetail(email, claimID string) (*models.ClaimSafari, error)
	IsClaimImageNameTaken(imageName string) (bool, error)

	// Abror
	FindEnrollmentAbrorForClaim(email, policyID string) (int64, error)
	IsClaimAbrorIDTaken(claimID string) (bool, error)
	GetProductCodeAbrorByPolicy(policyID string) (string, error)
	GetEnrollmentIDAbrorByPolicy(policyID string) (string, error)
	CreateClaimAbror(claim *models.ClaimAbror) error
	GetClaimsAbror(email, carType, dateReport string, db *gorm.DB) ([]models.ClaimAbror, error)
	GetClaimAbrorDetail(email, claimID string) (*models.ClaimAbror, error)
	IsClaimAbrorImageNameTaken(imageName string) (bool, error)

	// Admin
	GetAdminClaimsSafari(filters map[string]string) ([]models.ClaimSafari, error)
	GetAdminClaimsAbror(filters map[string]string) ([]models.ClaimAbror, error)
	FindClaimSafariByID(claimID string) (*models.ClaimSafari, error)
	FindClaimAbrorByID(claimID string) (*models.ClaimAbror, error)
	UpdateClaim(claim interface{}) error
}

type claimRepository struct {
	db *gorm.DB
}

func NewClaimRepository(db *gorm.DB) ClaimRepository {
	return &claimRepository{db: db}
}

func (r *claimRepository) FindEnrollmentForClaim(email, policyID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.EnrollmentSafari{}).Where("registrant_id = ? AND policy_id = ?", email, policyID).Count(&count).Error
	return count, err
}

func (r *claimRepository) FindEnrollmentAbrorForClaim(email, policyID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.EnrollmentAbror{}).Where("registrant_id = ? AND policy_id = ?", email, policyID).Count(&count).Error
	return count, err
}

func (r *claimRepository) IsClaimIDTaken(claimID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.ClaimSafari{}).Where("claim_id = ?", claimID).Count(&count).Error
	if err != nil {
		return true, err
	}
	return count > 0, nil
}

func (r *claimRepository) IsClaimAbrorIDTaken(claimID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.ClaimAbror{}).Where("claim_id = ?", claimID).Count(&count).Error
	if err != nil {
		return true, err
	}
	return count > 0, nil
}

func (r *claimRepository) GetProductCodeByPolicy(policyID string) (string, error) {
	var productCode string
	err := r.db.Model(&models.EnrollmentSafari{}).Where("policy_id = ?", policyID).Select("product_code").First(&productCode).Error
	return productCode, err
}

func (r *claimRepository) GetEnrollmentIDByPolicy(policyID string) (string, error) {
	var enrollmentId string
	err := r.db.Model(&models.EnrollmentSafari{}).Where("policy_id = ?", policyID).Select("enrollment_id").First(&enrollmentId).Error
	return enrollmentId, err
}

func (r *claimRepository) GetProductCodeAbrorByPolicy(policyID string) (string, error) {
	var productCode string
	err := r.db.Model(&models.EnrollmentAbror{}).Where("policy_id = ?", policyID).Select("product_code").First(&productCode).Error
	return productCode, err
}

func (r *claimRepository) GetEnrollmentIDAbrorByPolicy(policyID string) (string, error) {
	var enrollmentId string
	err := r.db.Model(&models.EnrollmentAbror{}).Where("policy_id = ?", policyID).Select("enrollment_id").First(&enrollmentId).Error
	return enrollmentId, err
}

func (r *claimRepository) CreateClaim(claim *models.ClaimSafari) error {
	return r.db.Create(claim).Error
}

func (r *claimRepository) CreateClaimAbror(claim *models.ClaimAbror) error {
	return r.db.Create(claim).Error
}

func (r *claimRepository) GetClaims(email, productName, dateReport string) ([]models.ClaimSafari, error) {
	var claims []models.ClaimSafari
	query := r.db.Where("registrant_id = ?", email)
	if productName != "" {
		query = query.Where("product_name = ?", productName)
	}
	if dateReport != "" {
		query = query.Where("date_report = ?", dateReport)
	}
	err := query.Order("created_at DESC").Find(&claims).Error
	return claims, err
}

func (r *claimRepository) GetClaimsAbror(email, carType, dateReport string, db *gorm.DB) ([]models.ClaimAbror, error) {
	var claims []models.ClaimAbror
	query := r.db.Where("registrant_id = ?", email)

	if dateReport != "" {
		query = query.Where("date_report = ?", dateReport)
	}

	if err := query.Order("created_at DESC").Find(&claims).Error; err != nil {
		return nil, err
	}

	if carType != "" {
		var filteredClaims []models.ClaimAbror
		for i := range claims {
			var enroll models.EnrollmentAbror
			if err := db.Where("policy_id = ? AND car_type = ?", claims[i].PolicyId, carType).First(&enroll).Error; err == nil {
				filteredClaims = append(filteredClaims, claims[i])
			}
		}
		return filteredClaims, nil
	}

	return claims, nil
}

func (r *claimRepository) GetClaimDetail(email, claimID string) (*models.ClaimSafari, error) {
	var claim models.ClaimSafari
	err := r.db.Where("registrant_id = ? AND claim_id = ?", email, claimID).First(&claim).Error
	return &claim, err
}

func (r *claimRepository) GetClaimAbrorDetail(email, claimID string) (*models.ClaimAbror, error) {
	var claim models.ClaimAbror
	err := r.db.Where("registrant_id = ? AND claim_id = ?", email, claimID).First(&claim).Error
	return &claim, err
}

func (r *claimRepository) IsClaimImageNameTaken(imageName string) (bool, error) {
	var count int64
	err := r.db.Model(&models.ClaimSafari{}).Where("evidence = ?", imageName).Count(&count).Error
	if err != nil {
		return true, err
	}
	return count > 0, nil
}

func (r *claimRepository) IsClaimAbrorImageNameTaken(imageName string) (bool, error) {
	var count int64
	err := r.db.Model(&models.ClaimAbror{}).Where("evidence = ?", imageName).Count(&count).Error
	if err != nil {
		return true, err
	}
	return count > 0, nil
}


// Admin-facing implementations
func (r *claimRepository) GetAdminClaimsSafari(filters map[string]string) ([]models.ClaimSafari, error) {
	var claims []models.ClaimSafari
	query := r.db.Model(&models.ClaimSafari{})
	for key, value := range filters {
		if value != "" {
			query = query.Where(key+" = ?", value)
		}
	}
	err := query.Order("created_at DESC").Find(&claims).Error
	return claims, err
}

func (r *claimRepository) GetAdminClaimsAbror(filters map[string]string) ([]models.ClaimAbror, error) {
	var claims []models.ClaimAbror
	query := r.db.Model(&models.ClaimAbror{})
	for key, value := range filters {
		if value != "" {
			query = query.Where(key+" = ?", value)
		}
	}
	err := query.Order("created_at DESC").Find(&claims).Error
	return claims, err
}

func (r *claimRepository) FindClaimSafariByID(claimID string) (*models.ClaimSafari, error) {
	var claim models.ClaimSafari
	err := r.db.Where("claim_id = ?", claimID).First(&claim).Error
	return &claim, err
}

func (r *claimRepository) FindClaimAbrorByID(claimID string) (*models.ClaimAbror, error) {
	var claim models.ClaimAbror
	err := r.db.Where("claim_id = ?", claimID).First(&claim).Error
	return &claim, err
}

func (r *claimRepository) UpdateClaim(claim interface{}) error {
	return r.db.Save(claim).Error
}
