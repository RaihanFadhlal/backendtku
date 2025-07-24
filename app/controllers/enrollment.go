package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"backendtku/app/dto"
	"backendtku/app/helpers"
	"backendtku/app/middleware"
	"backendtku/app/models"

	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

// safari
func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.CreateTransactionRequestDTO

	var Response dto.BaseResponse

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)

	fullname, err := h.UserRepo.FindNameByEmail(email)
	if err != nil {
		Response.Status = false
		Response.Message = "User not found"
		helpers.ResponseJSON(w, http.StatusNotFound, Response)
		return
	}

	createSnapRequest := func(orderID string, totalPrice int64, productCode string, productPrice int64, capacity int32, productName string) {
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
				Email: email,
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
			Response.Status = false
			Response.Message = "Failed to create transaction"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}

		Response.Status = true
		Response.Message = "Transaction created successfully"
		Response.Data = snapResp
		helpers.ResponseJSON(w, http.StatusOK, Response)
	}

	if requestDTO.TrxId[3] == 'S' {
		trx, err := h.EnrollRepo.FindTransactionByID(requestDTO.TrxId, email)
		if err != nil {
			Response.Status = false
			Response.Message = "Transaction not found"
			helpers.ResponseJSON(w, http.StatusNotFound, Response)
			return
		}
		createSnapRequest(trx.TransactionId, int64(trx.TotalPrice), trx.ProductCode, int64(trx.ProductPrice), int32(trx.Capacity), trx.ProductName)
	} else {
		trxa, err := h.EnrollRepo.FindTransactionAbrorByID(requestDTO.TrxId, email)
		if err != nil {
			Response.Status = false
			Response.Message = "Transaction not found"
			helpers.ResponseJSON(w, http.StatusNotFound, Response)
			return
		}
		createSnapRequest(trxa.TransactionId, int64(trxa.TotalPrice), trxa.ProductCode, int64(trxa.ProductPrice), int32(trxa.Capacity), trxa.ProductName)
	}
}

func (h *Handler) DownloadPdf(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.DownloadPdfRequestDTO

	email := r.Context().Value(middleware.UserEmailKey).(string)

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	policyID := ""
	enroll, err := h.EnrollRepo.FindPolicyForDownload(requestDTO.PolicyId, email)
	if err != nil {
		enrollAbror, err := h.EnrollRepo.FindPolicyAbrorForDownload(requestDTO.PolicyId, email)
		if err != nil {
			http.Error(w, "Policy not found or unauthorized", http.StatusNotFound)
			return
		}
		policyID = enrollAbror.PolicyId
	} else {
		policyID = enroll.PolicyId
	}

	filePath := "./upload/policy/pdfs/" + policyID + ".pdf"

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+policyID+".pdf")
	w.Header().Set("Content-Type", "application/pdf")
	http.ServeFile(w, r, filePath)
}

func (h *Handler) RequestProduct(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.RequestProductRequestDTO

	var Response dto.BaseResponse

	email := r.Context().Value(middleware.UserEmailKey).(string)

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	startDate, err := time.Parse("2006-01-02", requestDTO.DateStart)
	if err != nil {
		Response.Status = false
		Response.Message = "Invalid start date format"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	endDate, err := time.Parse("2006-01-02", requestDTO.DateEnd)
	if err != nil {
		Response.Status = false
		Response.Message = "Invalid end date format"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	totalDays := int(endDate.Sub(startDate).Hours() / 24)
	if totalDays <= 0 {
		Response.Status = false
		Response.Message = "End date must be after start date"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	count, err := h.EnrollRepo.CountMatchingSafariProduct(requestDTO.ProductCode, requestDTO.Contribution, requestDTO.ProductPrice, totalDays)
	if err != nil || count == 0 {
		Response.Status = false
		Response.Message = "No matching product found"
		helpers.ResponseJSON(w, http.StatusNotFound, Response)
		return
	}

	grossAmt := int64(requestDTO.Capacity) * requestDTO.ProductPrice

	var transactionId string
	for {
		transactionId = "T-" + requestDTO.ProductCode + "-" + helpers.RandomString(5)
		isTaken, err := h.EnrollRepo.IsTransactionIDTaken(transactionId)
		if err != nil {
			Response.Status = false
			Response.Message = "Error checking transaction ID"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}
		if !isTaken {
			break
		}
	}

	transaction := &models.Transaction{
		ID:            uuid.New(),
		TransactionId: transactionId,
		RegistrantId:  email,
		ProductCode:   requestDTO.ProductCode,
		ProductName:   requestDTO.ProductName,
		ProductPrice:  int(requestDTO.ProductPrice),
		Capacity:      int(requestDTO.Capacity),
		TotalPrice:    int(grossAmt),
		Status:        "Menunggu Pembayaran",
		CreatedAt:     time.Now(),
		ExpiredAt:     time.Now().Add(24 * time.Hour),
	}

	if err := h.EnrollRepo.CreateTransaction(transaction); err != nil {
		Response.Status = false
		Response.Message = "Error saving transaction: " + err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	var enrollmentId string
	for {
		enrollmentId = "E-" + requestDTO.ProductCode + "-" + helpers.RandomString(5)
		isTaken, err := h.EnrollRepo.IsEnrollmentIDTaken(enrollmentId)
		if err != nil {
			Response.Status = false
			Response.Message = "Error checking enrollment ID"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}
		if !isTaken {
			break
		}
	}

	enrollment := &models.EnrollmentSafari{
		ID:            uuid.New(),
		EnrollmentId:  enrollmentId,
		RegistrantId:  email,
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
		Gender:        requestDTO.Gender,
		Passport:      requestDTO.Passport,
	}

	if err := h.EnrollRepo.CreateEnrollment(enrollment); err != nil {
		Response.Status = false
		Response.Message = "Error saving enrollment: " + err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	for _, other := range requestDTO.Others {
		var otherEnrollmentId string
		for {
			otherEnrollmentId = "E-" + requestDTO.ProductCode + "-" + helpers.RandomString(5)
			isTaken, err := h.EnrollRepo.IsEnrollmentIDTaken(otherEnrollmentId)
			if err != nil {
				Response.Status = false
				Response.Message = "Error checking enrollment ID"
				helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
				return
			}
			if !isTaken {
				break
			}
		}

		otherEnrollment := &models.EnrollmentSafari{
			ID:            uuid.New(),
			EnrollmentId:  otherEnrollmentId,
			RegistrantId:  email,
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

		if err := h.EnrollRepo.CreateEnrollment(otherEnrollment); err != nil {
			Response.Status = false
			Response.Message = "Error saving additional enrollment: " + err.Error()
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}
	}

	Response.Status = true
	Response.Message = "Enrollment successfull"
	Response.Data = map[string]string{"trx_id": transactionId}
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) PaymentStatus(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.PaymentStatusRequestDTO
	var Response dto.BaseResponse

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)

	trx, err := h.EnrollRepo.FindTransactionForStatusUpdate(requestDTO.TrxId, email)
	if err != nil {
		Response.Status = false
		Response.Message = "Transaction not found"
		helpers.ResponseJSON(w, http.StatusNotFound, Response)
		return
	}

	midtransResponse, err := middleware.VerifyMidtransTrx(requestDTO.TrxId)
	if err != nil {
		Response.Status = false
		Response.Message = "Failed to verify transaction status"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	status := midtransResponse.TransactionStatus
	log.Println("Midtrans Transaction Status:", status)

	if status == "pending" {
		Response.Status = true
		Response.Message = "Menunggu Pembayaran"
		helpers.ResponseJSON(w, http.StatusOK, Response)
		return
	}

	if status == "expire" || status == "deny" || status == "cancel" {
		trx.Status = "Gagal"
		if err := h.EnrollRepo.SaveTransaction(trx); err != nil {
			Response.Status = false
			Response.Message = "Failed to update transaction status"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}
		Response.Status = true
		Response.Message = "Pembayaran Gagal"
		helpers.ResponseJSON(w, http.StatusOK, Response)
		return
	}

	if status == "settlement" {
		trx.Status = "Berhasil"
		if err := h.EnrollRepo.SaveTransaction(trx); err != nil {
			Response.Status = false
			Response.Message = "Failed to update transaction status"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}

		enroll, err := h.EnrollRepo.FindMainEnrollmentByTxID(requestDTO.TrxId, email)
		if err != nil {
			Response.Status = false
			Response.Message = "Transaction not found"
			helpers.ResponseJSON(w, http.StatusNotFound, Response)
			return
		}

		uniqueCode := trx.TransactionId
		policyId := "policy-" + uniqueCode[len(uniqueCode)-5:]

		group := enroll.ProductCode[:len(enroll.ProductCode)-2]

		product, err := h.ProductRepo.FindSafariProductByGroupCode(group)
		if err != nil {
			Response.Status = false
			Response.Message = "Product not found"
			helpers.ResponseJSON(w, http.StatusNotFound, Response)
			return
		}

		benefits, err := h.ProductRepo.FindSafariBenefitsByGroupCode(product.GroupCode)
		if err != nil {
			Response.Status = false
			Response.Message = "Benefits not found"
			helpers.ResponseJSON(w, http.StatusNotFound, Response)
			return
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

		others, err := h.EnrollRepo.FindAllEnrollmentsByTxID(requestDTO.TrxId, email)
		if err != nil {
			Response.Status = false
			Response.Message = "Transaction not found"
			helpers.ResponseJSON(w, http.StatusNotFound, Response)
			return
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

		if err := h.EnrollRepo.UpdateEnrollmentsPolicyID(requestDTO.TrxId, email, policyId); err != nil {
			Response.Status = false
			Response.Message = "Failed to save policies"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}

		Response.Status = true
		Response.Message = "Pembayaran Berhasil"
		Response.Data = map[string]string{"policy_id": policyId}
		helpers.ResponseJSON(w, http.StatusOK, Response)
	}
}

func (h *Handler) GetPolicies(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.GetPoliciesRequestDTO
	var Response dto.BaseResponse

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)
	enrolls, err := h.EnrollRepo.FindAllPolicies(email, requestDTO.ProductName, requestDTO.Destination)
	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving products"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	var policyItems []dto.PolicyResponseItemDTO
	for _, enroll := range enrolls {
		image, err := h.ProductRepo.FindProductImageByName(enroll.ProductName)
		if err != nil {
			Response.Status = false
			Response.Message = "Error retrieving product image"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}
		policyItems = append(policyItems, dto.PolicyResponseItemDTO{
			PolicyId:     enroll.PolicyId,
			ProductName:  enroll.ProductName,
			Contribution: enroll.Contribution,
			Destination:  enroll.Destination,
			DateStart:    enroll.DateStart,
			DateEnd:      enroll.DateEnd,
			Image:        h.Config.BaseUrl + "/upload/product/" + image,
		})
	}

	Response.Status = true
	Response.Message = "Products retrieved successfully"
	Response.Data = policyItems
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

// abror
func (h *Handler) RequestAbror(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.RequestAbrorRequestDTO

	var Response dto.BaseResponse

	email := r.Context().Value(middleware.UserEmailKey).(string)
	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, requestDTO.DateStart)
	if err != nil {
		Response.Status = false
		Response.Message = "Invalid DateStart format"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	endDate, err := time.Parse(layout, requestDTO.DateEnd)
	if err != nil {
		Response.Status = false
		Response.Message = "Invalid DateEnd format"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	duration := endDate.Sub(startDate)
	if duration.Hours() > 366*24 {
		Response.Status = false
		Response.Message = "Date range exceeds 366 days"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	vehicle, err := h.ProductRepo.FindCarByName(requestDTO.CarType)
	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving car"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	vehicleType, err := h.ProductRepo.FindVehicleTypeByPrice(int64(vehicle.Price))
	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving car type"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	plat := helpers.ExtractPlateCode(requestDTO.Plat)

	region, err := h.ProductRepo.FindRegionByPlat(plat)
	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving region"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	product, err := h.ProductRepo.FindAbrorProductByCriteria(requestDTO.Contribution, region.Code, vehicleType.Code)
	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving product"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	price := product.Percentage / 100 * float32(vehicle.Price)

	var transactionId string
	for {
		transactionId = "T-" + requestDTO.ProductCode + "-" + helpers.RandomString(5)
		isTaken, err := h.EnrollRepo.IsTransactionAbrorIDTaken(transactionId)
		if err != nil {
			Response.Status = false
			Response.Message = "Error checking transaction ID"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}
		if !isTaken {
			break
		}
	}

	transaction := &models.TransactionAbror{
		ID:            uuid.New(),
		TransactionId: transactionId,
		RegistrantId:  email,
		ProductCode:   requestDTO.ProductCode,
		ProductName:   requestDTO.ProductName,
		ProductPrice:  int(price),
		Capacity:      1,
		TotalPrice:    int(price),
		Status:        "Menunggu Pembayaran",
		CreatedAt:     time.Now(),
		ExpiredAt:     time.Now().Add(24 * time.Hour),
	}

	if err := h.EnrollRepo.CreateTransactionAbror(transaction); err != nil {
		Response.Status = false
		Response.Message = "Error saving transaction: " + err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	var enrollmentId string
	for {
		enrollmentId = "E-" + requestDTO.ProductCode + "-" + helpers.RandomString(5)
		isTaken, err := h.EnrollRepo.IsEnrollmentIDTaken(enrollmentId)
		if err != nil {
			Response.Status = false
			Response.Message = "Error checking enrollment ID"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
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
			Response.Status = false
			Response.Message = err.Error()
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}
		imageNames = append(imageNames, imageName)
	}

	enrollment := &models.EnrollmentAbror{
		ID:            uuid.New(),
		EnrollmentId:  enrollmentId,
		RegistrantId:  email,
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

	if err := h.EnrollRepo.CreateEnrollmentAbror(enrollment); err != nil {
		Response.Status = false
		Response.Message = "Error saving enrollment: " + err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Enrollment successfull"
	Response.Data = map[string]string{"trx_id": transactionId}
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) PaymentStatusAbror(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.PaymentStatusRequestDTO
	var Response dto.BaseResponse

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)

	trx, err := h.EnrollRepo.FindTransactionAbrorForStatusUpdate(requestDTO.TrxId, email)
	if err != nil {
		Response.Status = false
		Response.Message = "Transaction not found"
		helpers.ResponseJSON(w, http.StatusNotFound, Response)
		return
	}

	midtransResponse, err := middleware.VerifyMidtransTrx(requestDTO.TrxId)
	if err != nil {
		Response.Status = false
		Response.Message = "Failed to verify transaction status"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	status := midtransResponse.TransactionStatus
	log.Println("Midtrans Transaction Status:", status)

	if status == "pending" {
		Response.Status = true
		Response.Message = "Menunggu Pembayaran"
		helpers.ResponseJSON(w, http.StatusOK, Response)
		return
	}

	if status == "expire" || status == "deny" || status == "cancel" {
		trx.Status = "Gagal"
		if err := h.EnrollRepo.SaveTransactionAbror(trx); err != nil {
			Response.Status = false
			Response.Message = "Failed to update transaction status"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}
		Response.Status = true
		Response.Message = "Pembayaran Gagal"
		helpers.ResponseJSON(w, http.StatusOK, Response)
		return
	}

	if status == "settlement" {
		trx.Status = "Berhasil"
		if err := h.EnrollRepo.SaveTransactionAbror(trx); err != nil {
			Response.Status = false
			Response.Message = "Failed to update transaction status"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}

		enroll, err := h.EnrollRepo.FindMainEnrollmentAbrorByTxID(requestDTO.TrxId, email)
		if err != nil {
			Response.Status = false
			Response.Message = "Transaction not found"
			helpers.ResponseJSON(w, http.StatusNotFound, Response)
			return
		}

		uniqueCode := trx.TransactionId
		policyId := "policy-" + uniqueCode[len(uniqueCode)-5:]

		char6 := requestDTO.TrxId[5]
		benefits, err := h.ProductRepo.FindAllAbrorBenefit()
		if err != nil {
			Response.Status = false
			Response.Message = "Error retrieving benefits"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
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

		if err := h.EnrollRepo.UpdateEnrollmentsAbrorPolicyID(requestDTO.TrxId, email, policyId); err != nil {
			Response.Status = false
			Response.Message = "Failed to save policies"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}

		Response.Status = true
		Response.Message = "Pembayaran Berhasil"
		Response.Data = map[string]string{"policy_id": policyId}
		helpers.ResponseJSON(w, http.StatusOK, Response)
	}
}

func (h *Handler) GetPoliciesAbror(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.GetPoliciesAbrorRequestDTO
	var Response dto.BaseResponse

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)
	enrolls, err := h.EnrollRepo.FindAllPoliciesAbror(email, requestDTO.CarType, requestDTO.DateStart)
	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving products"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	var policyItems []dto.PolicyAbrorResponseItemDTO
	for _, enroll := range enrolls {
		image, err := h.ProductRepo.FindProductAbrorImageByName(enroll.ProductName)
		if err != nil {
			Response.Status = false
			Response.Message = "Error retrieving product image"
			helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
			return
		}
		policyItems = append(policyItems, dto.PolicyAbrorResponseItemDTO{
			PolicyId:     enroll.PolicyId,
			ProductName:  enroll.ProductName,
			Contribution: enroll.Contribution,
			CarType:      enroll.CarType,
			DateStart:    enroll.DateStart,
			DateEnd:      enroll.DateEnd,
			Image:        h.Config.BaseUrl + "/upload/product/" + image,
		})
	}

	Response.Status = true
	Response.Message = "Products retrieved successfully"
	Response.Data = policyItems
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetTrx(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.GetTrxRequestDTO
	var Response dto.BaseResponse

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request body"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	email := r.Context().Value(middleware.UserEmailKey).(string)

	safariTrx, err1 := h.EnrollRepo.FindSafariTransactions(email, requestDTO.Status, requestDTO.ProductName)
	abrorTrx, err2 := h.EnrollRepo.FindAbrorTransactions(email, requestDTO.Status, requestDTO.ProductName)

	if err1 != nil || err2 != nil {
		Response.Status = false
		Response.Message = "Error retrieving transactions"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
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
			Image:         h.Config.BaseUrl + "/upload/product/" + trx.Image,
		})
	}

	Response.Status = true
	Response.Message = "Transactions retrieved successfully"
	Response.Data = transactionItems
	helpers.ResponseJSON(w, http.StatusOK, Response)
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