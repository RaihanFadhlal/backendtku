package repositories

import (
	"backendtku/app/models"
	"time"

	"gorm.io/gorm"
)

type TransactionWithImage struct {
	TransactionId string
	ProductName   string
	TotalPrice    int
	CreatedAt     time.Time
	ExpiredAt     time.Time
	Status        string
	Image         string
}

type EnrollmentRepository interface {
	// Transaction
	FindTransactionByID(trxID, email string) (*models.Transaction, error)
	FindTransactionAbrorByID(trxID, email string) (*models.TransactionAbror, error)
	CreateTransaction(trx *models.Transaction) error
	CreateTransactionAbror(trx *models.TransactionAbror) error
	SaveTransaction(trx *models.Transaction) error
	SaveTransactionAbror(trx *models.TransactionAbror) error
	FindSafariTransactions(email, status, productName string) ([]TransactionWithImage, error)
	FindAbrorTransactions(email, status, productName string) ([]TransactionWithImage, error)
	FindTransactionForStatusUpdate(trxID, email string) (*models.Transaction, error)
	FindTransactionAbrorForStatusUpdate(trxID, email string) (*models.TransactionAbror, error)
	IsTransactionIDTaken(id string) (bool, error)
	IsTransactionAbrorIDTaken(id string) (bool, error)

	// Enrollment
	CreateEnrollment(enrollment *models.EnrollmentSafari) error
	CreateEnrollmentAbror(enrollment *models.EnrollmentAbror) error
	FindMainEnrollmentByTxID(txID, email string) (*models.EnrollmentSafari, error)
	FindMainEnrollmentAbrorByTxID(txID, email string) (*models.EnrollmentAbror, error)
	FindAllEnrollmentsByTxID(txID, email string) ([]models.EnrollmentSafari, error)
	UpdateEnrollmentsPolicyID(txID, email, policyID string) error
	UpdateEnrollmentsAbrorPolicyID(txID, email, policyID string) error
	FindAllPolicies(email, productName, destination string) ([]models.EnrollmentSafari, error)
	FindAllPoliciesAbror(email, carType, dateStart string) ([]models.EnrollmentAbror, error)
	IsEnrollmentIDTaken(id string) (bool, error)
	FindPolicyForDownload(policyID, email string) (*models.EnrollmentSafari, error)
	FindPolicyAbrorForDownload(policyID, email string) (*models.EnrollmentAbror, error)

	// Validation
	CountMatchingSafariProduct(code, contribution string, price int64, totalDays int) (int64, error)
}

type enrollmentRepository struct {
	db *gorm.DB
}

func NewEnrollmentRepository(db *gorm.DB) EnrollmentRepository {
	return &enrollmentRepository{db: db}
}

func (r *enrollmentRepository) FindTransactionByID(trxID, email string) (*models.Transaction, error) {
	var trx models.Transaction
	err := r.db.Where("registrant_id = ? AND transaction_id = ? AND status = 'Menunggu Pembayaran'", email, trxID).First(&trx).Error
	return &trx, err
}

func (r *enrollmentRepository) FindTransactionAbrorByID(trxID, email string) (*models.TransactionAbror, error) {
	var trx models.TransactionAbror
	err := r.db.Where("registrant_id = ? AND transaction_id = ? AND status = 'Menunggu Pembayaran'", email, trxID).First(&trx).Error
	return &trx, err
}

func (r *enrollmentRepository) CreateTransaction(trx *models.Transaction) error {
	return r.db.Create(trx).Error
}

func (r *enrollmentRepository) CreateTransactionAbror(trx *models.TransactionAbror) error {
	return r.db.Create(trx).Error
}

func (r *enrollmentRepository) SaveTransaction(trx *models.Transaction) error {
	return r.db.Save(trx).Error
}

func (r *enrollmentRepository) SaveTransactionAbror(trx *models.TransactionAbror) error {
	return r.db.Save(trx).Error
}

func (r *enrollmentRepository) FindSafariTransactions(email, status, productName string) ([]TransactionWithImage, error) {
	var results []TransactionWithImage
	query := r.db.Model(&models.Transaction{}).
		Select("transactions.transaction_id, transactions.product_name, transactions.total_price, transactions.created_at, transactions.expired_at, transactions.status, product_safaris.image").
		Joins("LEFT JOIN product_safaris ON product_safaris.name = transactions.product_name").
		Where("transactions.registrant_id = ?", email)

	if status != "" {
		query = query.Where("transactions.status = ?", status)
	}
	if productName != "" {
		query = query.Where("transactions.product_name = ?", productName)
	}

	err := query.Scan(&results).Error
	return results, err
}

func (r *enrollmentRepository) FindAbrorTransactions(email, status, productName string) ([]TransactionWithImage, error) {
	var results []TransactionWithImage
	query := r.db.Model(&models.TransactionAbror{}).
		Select("transaction_abrors.transaction_id, transaction_abrors.product_name, transaction_abrors.total_price, transaction_abrors.created_at, transaction_abrors.expired_at, transaction_abrors.status, product_abrors.image").
		Joins("LEFT JOIN product_abrors ON product_abrors.name = transaction_abrors.product_name").
		Where("transaction_abrors.registrant_id = ?", email)

	if status != "" {
		query = query.Where("transaction_abrors.status = ?", status)
	}
	if productName != "" {
		query = query.Where("transaction_abrors.product_name = ?", productName)
	}

	err := query.Scan(&results).Error
	return results, err
}

func (r *enrollmentRepository) CreateEnrollment(enrollment *models.EnrollmentSafari) error {
	return r.db.Create(enrollment).Error
}

func (r *enrollmentRepository) CreateEnrollmentAbror(enrollment *models.EnrollmentAbror) error {
	return r.db.Create(enrollment).Error
}

func (r *enrollmentRepository) FindMainEnrollmentByTxID(txID, email string) (*models.EnrollmentSafari, error) {
	var enroll models.EnrollmentSafari
	err := r.db.Where("registrant_id = ? AND transaction_id = ? AND LENGTH(phone) > 0", email, txID).First(&enroll).Error
	return &enroll, err
}

func (r *enrollmentRepository) FindMainEnrollmentAbrorByTxID(txID, email string) (*models.EnrollmentAbror, error) {
	var enroll models.EnrollmentAbror
	err := r.db.Where("registrant_id = ? AND transaction_id = ? AND LENGTH(phone) > 0", email, txID).First(&enroll).Error
	return &enroll, err
}

func (r *enrollmentRepository) FindAllEnrollmentsByTxID(txID, email string) ([]models.EnrollmentSafari, error) {
	var others []models.EnrollmentSafari
	err := r.db.Where("registrant_id = ? AND transaction_id = ?", email, txID).Find(&others).Error
	return others, err
}

func (r *enrollmentRepository) UpdateEnrollmentsPolicyID(txID, email, policyID string) error {
	return r.db.Model(&models.EnrollmentSafari{}).Where("registrant_id = ? AND transaction_id = ?", email, txID).Update("policy_id", policyID).Error
}

func (r *enrollmentRepository) UpdateEnrollmentsAbrorPolicyID(txID, email, policyID string) error {
	return r.db.Model(&models.EnrollmentAbror{}).Where("registrant_id = ? AND transaction_id = ?", email, txID).Update("policy_id", policyID).Error
}

func (r *enrollmentRepository) FindAllPolicies(email, productName, destination string) ([]models.EnrollmentSafari, error) {
	query := r.db.Where("registrant_id = ? AND LENGTH(phone) > 0 AND LENGTH(policy_id) > 0", email)
	if destination != "" {
		query = query.Where("destination = ?", destination)
	}
	if productName != "" {
		query = query.Where("product_name = ?", productName)
	}
	var enrolls []models.EnrollmentSafari
	err := query.Order("created_at DESC").Find(&enrolls).Error
	return enrolls, err
}

func (r *enrollmentRepository) FindAllPoliciesAbror(email, carType, dateStart string) ([]models.EnrollmentAbror, error) {
	query := r.db.Where("registrant_id = ? AND LENGTH(phone) > 0 AND LENGTH(policy_id) > 0", email)
	if carType != "" {
		query = query.Where("car_type = ?", carType)
	}
	if dateStart != "" {
		query = query.Where("date_start = ?", dateStart)
	}
	var enrolls []models.EnrollmentAbror
	err := query.Order("created_at DESC").Find(&enrolls).Error
	return enrolls, err
}

func (r *enrollmentRepository) CountMatchingSafariProduct(code, contribution string, price int64, totalDays int) (int64, error) {
	var count int64
	err := r.db.Model(&models.ProductSafari{}).Where("code = ? AND contribution = ? AND price = ? AND day_min <= ? AND day_max >= ?",
		code, contribution, price, totalDays, totalDays).Count(&count).Error
	return count, err
}

func (r *enrollmentRepository) FindTransactionForStatusUpdate(trxID, email string) (*models.Transaction, error) {
	var trx models.Transaction
	err := r.db.Where("registrant_id = ? AND transaction_id = ?", email, trxID).First(&trx).Error
	return &trx, err
}

func (r *enrollmentRepository) FindPolicyForDownload(policyID, email string) (*models.EnrollmentSafari, error) {
	var enroll models.EnrollmentSafari
	err := r.db.Where("registrant_id = ? AND policy_id = ?", email, policyID).First(&enroll).Error
	return &enroll, err
}

func (r *enrollmentRepository) FindPolicyAbrorForDownload(policyID, email string) (*models.EnrollmentAbror, error) {
	var enrollAbror models.EnrollmentAbror
	err := r.db.Where("registrant_id = ? AND policy_id = ?", email, policyID).First(&enrollAbror).Error
	return &enrollAbror, err
}

func (r *enrollmentRepository) IsTransactionIDTaken(id string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Transaction{}).Where("transaction_id = ?", id).Count(&count).Error
	if err != nil {
		return true, err
	}
	return count > 0, nil
}

func (r *enrollmentRepository) IsTransactionAbrorIDTaken(id string) (bool, error) {
	var count int64
	err := r.db.Model(&models.TransactionAbror{}).Where("transaction_id = ?", id).Count(&count).Error
	if err != nil {
		return true, err
	}
	return count > 0, nil
}

func (r *enrollmentRepository) IsEnrollmentIDTaken(id string) (bool, error) {
	var count int64
	err := r.db.Model(&models.EnrollmentSafari{}).Where("enrollment_id = ?", id).Count(&count).Error
	if err != nil {
		return true, err
	}
	return count > 0, nil
}

func (r *enrollmentRepository) FindTransactionAbrorForStatusUpdate(trxID, email string) (*models.TransactionAbror, error) {
	var trx models.TransactionAbror
	err := r.db.Where("registrant_id = ? AND transaction_id = ?", email, trxID).First(&trx).Error
	return &trx, err
}
