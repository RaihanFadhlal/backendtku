package usecase

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"backendtku/app/dto"
	"backendtku/app/helpers"
	"backendtku/app/middleware"
	"backendtku/app/models"
	"backendtku/app/repositories"
	"backendtku/config"

	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go/snap"
	"gorm.io/gorm"
	"github.com/midtrans/midtrans-go"
)

type EnrollmentUseCase interface {
	CreateTransaction(requestDTO dto.CreateTransactionRequestDTO, userEmail string) (interface{}, error)
	DownloadPdf(policyId string, userEmail string) (string, error)
	RequestProduct(requestDTO dto.RequestProductRequestDTO, userEmail string) (map[string]string, error)
	PaymentStatus(requestDTO dto.PaymentStatusRequestDTO, userEmail string) (map[string]string, error)
	GetPolicies(requestDTO dto.GetPoliciesRequestDTO, userEmail string) ([]dto.PolicyResponseItemDTO, error)
	RequestAbror(requestDTO dto.RequestAbrorRequestDTO, userEmail string) (map[string]string, error)
	PaymentStatusAbror(requestDTO dto.PaymentStatusRequestDTO, userEmail string) (map[string]string, error)
	GetPoliciesAbror(requestDTO dto.GetPoliciesAbrorRequestDTO, userEmail string) ([]dto.PolicyAbrorResponseItemDTO, error)
	GetTrx(requestDTO dto.GetTrxRequestDTO, userEmail string) ([]dto.TransactionResponseItemDTO, error)
}

type enrollmentUseCase struct {
	enrollRepo    repositories.EnrollmentRepository
	userRepo    repositories.UserRepository
	productRepo repositories.ProductRepository
	config      *config.Config
	DB          *gorm.DB
}

func NewEnrollmentUseCase(enrollRepo repositories.EnrollmentRepository, userRepo repositories.UserRepository, productRepo repositories.ProductRepository, cfg *config.Config, db *gorm.DB) EnrollmentUseCase {
	return &enrollmentUseCase{
		enrollRepo:    enrollRepo,
		userRepo:    userRepo,
		productRepo: productRepo,
		config:      cfg,
		DB:          db,
	}
}

func (uc *enrollmentUseCase) CreateTransaction(requestDTO dto.CreateTransactionRequestDTO, userEmail string) (interface{}, error) {
	fullname, err := uc.userRepo.FindNameByEmail(userEmail)
	if err != nil {
		return nil, errors.New("User not found")
	}

	createSnapRequest := func(orderID string, totalPrice int64, productCode string, productPrice int64, capacity int32, productName string) (interface{}, error) {
		req := &snap.Request{
			TransactionDetails: midtrans.TransactionDetails{
				OrderID:  orderID,
				GrossAmt: totalPrice,
			},
			CreditCard: &snap.CreditCardDetails{
				Secure: true,
			},
			CustomerDetail: &midtrans.CustomerDetails{
				FName: fullname,
				Email: userEmail,
			},
			Items: &[]midtrans.ItemDetails{
				{
					ID:    productCode,
					Price: productPrice,
					Qty:   capacity,
					Name:  productName,
				},
			},
		}

		snapResp, err := middleware.SnapClient.CreateTransaction(req)
		if err != nil {
			log.Println("Error creating transaction:", err)
			return nil, errors.New("Failed to create transaction")
		}
		return snapResp, nil
	}

	if requestDTO.TrxId[3] == 'S' {
		trx, err := uc.enrollRepo.FindTransactionByID(requestDTO.TrxId, userEmail)
		if err != nil {
			return nil, errors.New("Transaction not found")
		}
		return createSnapRequest(trx.TransactionId, int64(trx.TotalPrice), trx.ProductCode, int64(trx.ProductPrice), int32(trx.Capacity), trx.ProductName)
	} else {
		trxa, err := uc.enrollRepo.FindTransactionAbrorByID(requestDTO.TrxId, userEmail)
		if err != nil {
			return nil, errors.New("Transaction not found")
		}
		return createSnapRequest(trxa.TransactionId, int64(trxa.TotalPrice), trxa.ProductCode, int64(trxa.ProductPrice), int32(trxa.Capacity), trxa.ProductName)
	}
}

func (uc *enrollmentUseCase) DownloadPdf(policyId string, userEmail string) (string, error) {
	policyID := ""
	enroll, err := uc.enrollRepo.FindPolicyForDownload(policyId, userEmail)
	if err != nil {
		enrollAbror, err := uc.enrollRepo.FindPolicyAbrorForDownload(policyId, userEmail)
		if err != nil {
			return "", errors.New("Policy not found or unauthorized")
		}
		policyID = enrollAbror.PolicyId
	} else {
		policyID = enroll.PolicyId
	}

	filePath := "./upload/policy/pdfs/" + policyID + ".pdf"

	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		return "", errors.New("File not found")
	}

	return filePath, nil
}

func (uc *enrollmentUseCase) RequestProduct(requestDTO dto.RequestProductRequestDTO, userEmail string) (map[string]string, error) {
	startDate, err := time.Parse("2006-01-02", requestDTO.DateStart)
	if err != nil {
		return nil, errors.New("Invalid start date format")
	}

	endDate, err := time.Parse("2006-01-02", requestDTO.DateEnd)
	if err != nil {
		return nil, errors.New("Invalid end date format")
	}

	totalDays := int(endDate.Sub(startDate).Hours() / 24)
	if totalDays <= 0 {
		return nil, errors.New("End date must be after start date")
	}

	count, err := uc.enrollRepo.CountMatchingSafariProduct(requestDTO.ProductCode, requestDTO.Contribution, requestDTO.ProductPrice, totalDays)
	if err != nil || count == 0 {
		return nil, errors.New("No matching product found")
	}

	grossAmt := int64(requestDTO.Capacity) * requestDTO.ProductPrice

	var transactionId string
	for {
		transactionId = "T-" + requestDTO.ProductCode + "-" + helpers.RandomString(5)
		isTaken, err := uc.enrollRepo.IsTransactionIDTaken(transactionId)
		if err != nil {
			return nil, errors.New("Error checking transaction ID")
		}
		if !isTaken {
			break
		}
	}

	transaction := &models.Transaction{
		ID:            uuid.New(),
		TransactionId: transactionId,
		RegistrantId:  userEmail,
		ProductCode:   requestDTO.ProductCode,
		ProductName:   requestDTO.ProductName,
		ProductPrice:  int(requestDTO.ProductPrice),
		Capacity:      int(requestDTO.Capacity),
		TotalPrice:    int(grossAmt),
		Status:        "Menunggu Pembayaran",
		CreatedAt:     time.Now(),
		ExpiredAt:     time.Now().Add(24 * time.Hour),
	}

	if err := uc.enrollRepo.CreateTransaction(transaction); err != nil {
		return nil, errors.New("Error saving transaction: " + err.Error())
	}

	var enrollmentId string
	for {
		enrollmentId = "E-" + requestDTO.ProductCode + "-" + helpers.RandomString(5)
		isTaken, err := uc.enrollRepo.IsEnrollmentIDTaken(enrollmentId)
		if err != nil {
			return nil, errors.New("Error checking enrollment ID")
		}
		if !isTaken {
			break
		}
	}

	enrollment := &models.EnrollmentSafari{
		ID:            uuid.New(),
		EnrollmentId:  enrollmentId,
		RegistrantId:  userEmail,
		TransactionId: transactionId,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Phone:         requestDTO.Phone,
		ProductCode:   requestDTO.ProductCode,
		ProductName:   requestDTO.ProductName,
		From:          requestDTO.From,
		Destination:   requestDTO.Destination,
		DateStart:     requestDTO.DateStart,
		DateEnd:       requestDTO.DateEnd,
		Contribution:  requestDTO.Contribution,
		Capacity:      requestDTO.Capacity,
		Name:          requestDTO.FullName,
		Birthdate:     requestDTO.Birthdate,
		Birthplace:    requestDTO.Birthplace,
		Gender:        helpers.FormatGender(requestDTO.Gender),
		Passport:      requestDTO.Passport,
	}

	if err := uc.enrollRepo.CreateEnrollment(enrollment); err != nil {
		return nil, errors.New("Error saving enrollment: " + err.Error())
	}

	for _, other := range requestDTO.Others {
		var otherEnrollmentId string
		for {
			otherEnrollmentId = "E-" + requestDTO.ProductCode + "-" + helpers.RandomString(5)
			isTaken, err := uc.enrollRepo.IsEnrollmentIDTaken(otherEnrollmentId)
			if err != nil {
				return nil, errors.New("Error checking enrollment ID")
			}
			if !isTaken {
				break
			}
		}

		otherEnrollment := &models.EnrollmentSafari{
			ID:            uuid.New(),
			EnrollmentId:  otherEnrollmentId,
			RegistrantId:  userEmail,
			TransactionId: transactionId,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			ProductCode:   requestDTO.ProductCode,
			ProductName:   requestDTO.ProductName,
			From:          requestDTO.From,
			Destination:   requestDTO.Destination,
			DateStart:     requestDTO.DateStart,
			DateEnd:       requestDTO.DateEnd,
			Contribution:  requestDTO.Contribution,
			Capacity:      requestDTO.Capacity,
			Name:          other.Fullname,
			Birthdate:     other.Birthdate,
		}

		if err := uc.enrollRepo.CreateEnrollment(otherEnrollment); err != nil {
			return nil, errors.New("Error saving additional enrollment: " + err.Error())
		}
	}

	return map[string]string{"trx_id": transactionId}, nil
}

func (uc *enrollmentUseCase) PaymentStatus(requestDTO dto.PaymentStatusRequestDTO, userEmail string) (map[string]string, error) {
	trx, err := uc.enrollRepo.FindTransactionForStatusUpdate(requestDTO.TrxId, userEmail)
	if err != nil {
		return nil, errors.New("Transaction not found")
	}

	midtransResponse, err := middleware.VerifyMidtransTrx(requestDTO.TrxId)
	if err != nil {
		return nil, errors.New("Failed to verify transaction status")
	}

	status := midtransResponse.TransactionStatus
	log.Println("Midtrans Transaction Status:", status)

	if status == "pending" {
		return map[string]string{"message": "Menunggu Pembayaran"}, nil
	}

	if status == "expire" || status == "deny" || status == "cancel" {
		trx.Status = "Gagal"
		if err := uc.enrollRepo.SaveTransaction(trx); err != nil {
			return nil, errors.New("Failed to update transaction status")
		}
		return map[string]string{"message": "Pembayaran Gagal"}, nil
	}

	if status == "settlement" {
		trx.Status = "Berhasil"
		if err := uc.enrollRepo.SaveTransaction(trx); err != nil {
			return nil, errors.New("Failed to update transaction status")
		}

		enroll, err := uc.enrollRepo.FindMainEnrollmentByTxID(requestDTO.TrxId, userEmail)
		if err != nil {
			return nil, errors.New("Transaction not found")
		}

		uniqueCode := trx.TransactionId
		policyId := "policy-" + uniqueCode[len(uniqueCode)-5:]

		group := enroll.ProductCode[:len(enroll.ProductCode)-2]

		product, err := uc.productRepo.FindSafariProductByGroupCode(group)
		if err != nil {
			return nil, errors.New("Product not found")
		}

		benefits, err := uc.productRepo.FindSafariBenefitsByGroupCode(product.GroupCode)
		if err != nil {
			return nil, errors.New("Benefits not found")
		}

		var result []string
		for _, benefit := range benefits {
			var contributionField string
			switch product.Contribution {
			case "Basic":
				contributionField = benefit.Basic
			case "Gold":
				contributionField = benefit.Gold
			case "Platinum":
				contributionField = benefit.Platinum
			case "Titanium":
				contributionField = benefit.Titanium
			default:
				contributionField = ""
			}

			if contributionField != "" {
				result = append(result, fmt.Sprintf("%s : %s", benefit.Detail, contributionField))
			}
		}

		others, err := uc.enrollRepo.FindAllEnrollmentsByTxID(requestDTO.TrxId, userEmail)
		if err != nil {
			return nil, errors.New("Transaction not found")
		}
		log.Println(others)

		m := helpers.GetMaroto(policyId, trx.ProductName, enroll.Name, enroll.Contribution, enroll.DateStart, enroll.DateEnd, trx.TotalPrice, result, others)
		document, err := m.Generate()
		if err != nil {
			log.Fatal(err.Error())
		}
		err = document.Save("upload/policy/pdfs/" + policyId + ".pdf")
		if err != nil {
			log.Fatal(err.Error())
		}

		if err := uc.enrollRepo.UpdateEnrollmentsPolicyID(requestDTO.TrxId, userEmail, policyId); err != nil {
			return nil, errors.New("Failed to save policies")
		}

		return map[string]string{"policy_id": policyId}, nil
	}
	return nil, errors.New("Unknown transaction status")
}

func (uc *enrollmentUseCase) GetPolicies(requestDTO dto.GetPoliciesRequestDTO, userEmail string) ([]dto.PolicyResponseItemDTO, error) {
	enrolls, err := uc.enrollRepo.FindAllPolicies(userEmail, requestDTO.ProductName, requestDTO.Destination)
	if err != nil {
		return nil, errors.New("Error retrieving products")
	}

	var policyItems []dto.PolicyResponseItemDTO
	for _, enroll := range enrolls {
		image, err := uc.productRepo.FindProductImageByName(enroll.ProductName)
		if err != nil {
			return nil, errors.New("Error retrieving product image")
		}
		policyItems = append(policyItems, dto.PolicyResponseItemDTO{
			PolicyId:     enroll.PolicyId,
			ProductName:  enroll.ProductName,
			Contribution: enroll.Contribution,
			Destination:  enroll.Destination,
			DateStart:    enroll.DateStart,
			DateEnd:      enroll.DateEnd,
			Image:        uc.config.BaseUrl + "/upload/product/" + image,
		})
	}

	return policyItems, nil
}

func (uc *enrollmentUseCase) RequestAbror(requestDTO dto.RequestAbrorRequestDTO, userEmail string) (map[string]string, error) {
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, requestDTO.DateStart)
	if err != nil {
		return nil, errors.New("Invalid DateStart format")
	}

	endDate, err := time.Parse(layout, requestDTO.DateEnd)
	if err != nil {
		return nil, errors.New("Invalid DateEnd format")
	}

	duration := endDate.Sub(startDate)
	if duration.Hours() > 366*24 {
		return nil, errors.New("Date range exceeds 366 days")
	}

	vehicle, err := uc.productRepo.FindCarByName(requestDTO.CarType)
	if err != nil {
		return nil, errors.New("Error retrieving car")
	}

	vehicleType, err := uc.productRepo.FindVehicleTypeByPrice(int64(vehicle.Price))
	if err != nil {
		return nil, errors.New("Error retrieving car type")
	}

	plat := helpers.ExtractPlateCode(requestDTO.Plat)

	region, err := uc.productRepo.FindRegionByPlat(plat)
	if err != nil {
		return nil, errors.New("Error retrieving region")
	}

	product, err := uc.productRepo.FindAbrorProductByCriteria(requestDTO.Contribution, region.Code, vehicleType.Code)
	if err != nil {
		return nil, errors.New("Error retrieving product")
	}

	price := product.Percentage / 100 * float32(vehicle.Price)

	var transactionId string
	for {
		transactionId = "T-" + requestDTO.ProductCode + "-" + helpers.RandomString(5)
		isTaken, err := uc.enrollRepo.IsTransactionAbrorIDTaken(transactionId)
		if err != nil {
			return nil, errors.New("Error checking transaction ID")
		}
		if !isTaken {
			break
		}
	}

	transaction := &models.TransactionAbror{
		ID:            uuid.New(),
		TransactionId: transactionId,
		RegistrantId:  userEmail,
		ProductCode:   requestDTO.ProductCode,
		ProductName:   requestDTO.ProductName,
		ProductPrice:  int(price),
		Capacity:      1,
		TotalPrice:    int(price),
		Status:        "Menunggu Pembayaran",
		CreatedAt:     time.Now(),
		ExpiredAt:     time.Now().Add(24 * time.Hour),
	}

	if err := uc.enrollRepo.CreateTransactionAbror(transaction); err != nil {
		return nil, errors.New("Error saving transaction: " + err.Error())
	}

	var enrollmentId string
	for {
		enrollmentId = "E-" + requestDTO.ProductCode + "-" + helpers.RandomString(5)
		isTaken, err := uc.enrollRepo.IsEnrollmentIDTaken(enrollmentId)
		if err != nil {
			return nil, errors.New("Error checking enrollment ID")
		}
		if !isTaken {
			break
		}
	}

	basePath := "upload/enroll"
	imageNames := []string{}
	for _, image := range []struct {
		Base64   string
		BaseName string
	}{
		{requestDTO.Image1, enrollmentId},
		{requestDTO.Image2, enrollmentId},
		{requestDTO.Image3, enrollmentId},
		{requestDTO.Image4, enrollmentId},
		{requestDTO.IdUser, enrollmentId},
	} {
		imageName, err := saveImage(image.Base64, image.BaseName, basePath)
		if err != nil {
			return nil, err
		}
		imageNames = append(imageNames, imageName)
	}

	enrollment := &models.EnrollmentAbror{
		ID:            uuid.New(),
		EnrollmentId:  enrollmentId,
		RegistrantId:  userEmail,
		TransactionId: transactionId,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Phone:         requestDTO.Phone,
		ProductCode:   requestDTO.ProductCode,
		ProductName:   requestDTO.ProductName,
		DateStart:     requestDTO.DateStart,
		DateEnd:       requestDTO.DateEnd,
		Contribution:  requestDTO.Contribution,
		CarBrand:      requestDTO.CarBrand,
		CarType:       requestDTO.CarType,
		Year:          requestDTO.Year,
		Plat:          requestDTO.Plat,
		Chassis:       requestDTO.Chassis,
		Engine:        requestDTO.Engine,
		Image1:        imageNames[0],
		Image2:        imageNames[1],
		Image3:        imageNames[2],
		Image4:        imageNames[3],
		Name:          requestDTO.Fullname,
		Birthdate:     requestDTO.Birthdate,
		Gender:        requestDTO.Gender,
		IdUser:        imageNames[4],
	}

	if err := uc.enrollRepo.CreateEnrollmentAbror(enrollment); err != nil {
		return nil, errors.New("Error saving enrollment: " + err.Error())
	}

	return map[string]string{"trx_id": transactionId}, nil
}

func (uc *enrollmentUseCase) PaymentStatusAbror(requestDTO dto.PaymentStatusRequestDTO, userEmail string) (map[string]string, error) {
	trx, err := uc.enrollRepo.FindTransactionAbrorForStatusUpdate(requestDTO.TrxId, userEmail)
	if err != nil {
		return nil, errors.New("Transaction not found")
	}

	midtransResponse, err := middleware.VerifyMidtransTrx(requestDTO.TrxId)
	if err != nil {
		return nil, errors.New("Failed to verify transaction status")
	}

	status := midtransResponse.TransactionStatus
	log.Println("Midtrans Transaction Status:", status)

	if status == "pending" {
		return map[string]string{"message": "Menunggu Pembayaran"}, nil
	}

	if status == "expire" || status == "deny" || status == "cancel" {
		trx.Status = "Gagal"
		if err := uc.enrollRepo.SaveTransactionAbror(trx); err != nil {
			return nil, errors.New("Failed to update transaction status")
		}
		return map[string]string{"message": "Pembayaran Gagal"}, nil
	}

	if status == "settlement" {
		trx.Status = "Berhasil"
		if err := uc.enrollRepo.SaveTransactionAbror(trx); err != nil {
			return nil, errors.New("Failed to update transaction status")
		}

		enroll, err := uc.enrollRepo.FindMainEnrollmentAbrorByTxID(requestDTO.TrxId, userEmail)
		if err != nil {
			return nil, errors.New("Transaction not found")
		}

		uniqueCode := trx.TransactionId
		policyId := "policy-" + uniqueCode[len(uniqueCode)-5:]

		char6 := requestDTO.TrxId[5]
		benefits, err := uc.productRepo.FindAllAbrorBenefit()
		if err != nil {
			return nil, errors.New("Error retrieving benefits")
		}

		var result []string
		for _, benefit := range benefits {
			var value string
			switch char6 {
			case 'S':
				value = benefit.Standard
			case 'P':
				value = benefit.Premium
			default:
				continue
			}

			if value != "" {
				result = append(result, fmt.Sprintf("%s : %s", benefit.Description, value))
			}
		}

		m := helpers.GetMarotoAbror(policyId, enroll.CarBrand, enroll.CarType, enroll.Plat, enroll.Name, enroll.Contribution, enroll.DateStart, enroll.DateEnd, trx.TotalPrice, result, enroll.Chassis, enroll.Engine, enroll.Image1, enroll.Image2, enroll.Image3, enroll.Image4)
		document, err := m.Generate()
		if err != nil {
			log.Fatal(err.Error())
		}
		err = document.Save("upload/policy/pdfs/" + policyId + ".pdf")
		if err != nil {
			log.Fatal(err.Error())
		}

		if err := uc.enrollRepo.UpdateEnrollmentsAbrorPolicyID(requestDTO.TrxId, userEmail, policyId); err != nil {
			return nil, errors.New("Failed to save policies")
		}

		return map[string]string{"policy_id": policyId}, nil
	}
	return nil, errors.New("Unknown transaction status")
}

func (uc *enrollmentUseCase) GetPoliciesAbror(requestDTO dto.GetPoliciesAbrorRequestDTO, userEmail string) ([]dto.PolicyAbrorResponseItemDTO, error) {
	enrolls, err := uc.enrollRepo.FindAllPoliciesAbror(userEmail, requestDTO.CarType, requestDTO.DateStart)
	if err != nil {
		return nil, errors.New("Error retrieving products")
	}

	var policyItems []dto.PolicyAbrorResponseItemDTO
	for _, enroll := range enrolls {
		image, err := uc.productRepo.FindProductAbrorImageByName(enroll.ProductName)
		if err != nil {
			return nil, errors.New("Error retrieving product image")
		}
		policyItems = append(policyItems, dto.PolicyAbrorResponseItemDTO{
			PolicyId:     enroll.PolicyId,
			ProductName:  enroll.ProductName,
			Contribution: enroll.Contribution,
			CarType:      enroll.CarType,
			DateStart:    enroll.DateStart,
			DateEnd:      enroll.DateEnd,
			Image:        uc.config.BaseUrl + "/upload/product/" + image,
		})
	}

	return policyItems, nil
}

func (uc *enrollmentUseCase) GetTrx(requestDTO dto.GetTrxRequestDTO, userEmail string) ([]dto.TransactionResponseItemDTO, error) {
	safariTrx, err1 := uc.enrollRepo.FindSafariTransactions(userEmail, requestDTO.Status, requestDTO.ProductName)
	abrorTrx, err2 := uc.enrollRepo.FindAbrorTransactions(userEmail, requestDTO.Status, requestDTO.ProductName)

	if err1 != nil || err2 != nil {
		return nil, errors.New("Error retrieving transactions")
	}

	allTrx := append(safariTrx, abrorTrx...)
	sort.Slice(allTrx, func(i, j int) bool {
		return allTrx[i].CreatedAt.After(allTrx[j].CreatedAt)
	})

	var transactionItems []dto.TransactionResponseItemDTO
	for _, trx := range allTrx {
		transactionItems = append(transactionItems, dto.TransactionResponseItemDTO{
			TransactionId: trx.TransactionId,
			ProductName:   trx.ProductName,
			TotalPrice:    trx.TotalPrice,
			CreatedAt:     trx.CreatedAt.Format("2006-01-02 15:04:05"),
			ExpiredAt:     trx.ExpiredAt.Format("2006-01-02 15:04:05"),
			Status:        trx.Status,
			Image:         uc.config.BaseUrl + "/upload/product/" + trx.Image,
		})
	}

	return transactionItems, nil
}

func saveImage(imageBase64, baseName, path string) (string, error) {
	if imageBase64 == "" {
		return "", nil
	}

	imageFormat := helpers.GetTypeBase64(imageBase64)
	timestamp := time.Now().UnixNano()
	imageName := fmt.Sprintf("%s_%d%s", baseName, timestamp, imageFormat)

	decodedImage, err := base64.StdEncoding.DecodeString(imageBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	imagePath := filepath.Join(path, imageName)
	if err := os.WriteFile(imagePath, decodedImage, 0644); err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	return imageName, nil
}
