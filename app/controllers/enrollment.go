package controllers

import (
	"encoding/json"
	"net/http"
	"os"

	"backendtku/app/dto"
	"backendtku/app/helpers"
	"backendtku/app/middleware"
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

	snapResp, err := h.EnrollmentUseCase.CreateTransaction(requestDTO, email)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Transaction created successfully"
	Response.Data = snapResp
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) DownloadPdf(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.DownloadPdfRequestDTO

	email := r.Context().Value(middleware.UserEmailKey).(string)

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	filePath, err := h.EnrollmentUseCase.DownloadPdf(requestDTO.PolicyId, email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+requestDTO.PolicyId+".pdf")
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

	data, err := h.EnrollmentUseCase.RequestProduct(requestDTO, email)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Enrollment successfull"
	Response.Data = data
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

	data, err := h.EnrollmentUseCase.PaymentStatus(requestDTO, email)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = data["message"]
	Response.Data = data
	helpers.ResponseJSON(w, http.StatusOK, Response)
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
	policyItems, err := h.EnrollmentUseCase.GetPolicies(requestDTO, email)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
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

	data, err := h.EnrollmentUseCase.RequestAbror(requestDTO, email)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Enrollment successfull"
	Response.Data = data
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

	data, err := h.EnrollmentUseCase.PaymentStatusAbror(requestDTO, email)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = data["message"]
	Response.Data = data
	helpers.ResponseJSON(w, http.StatusOK, Response)
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
	policyItems, err := h.EnrollmentUseCase.GetPoliciesAbror(requestDTO, email)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
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

	transactionItems, err := h.EnrollmentUseCase.GetTrx(requestDTO, email)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Transactions retrieved successfully"
	Response.Data = transactionItems
	helpers.ResponseJSON(w, http.StatusOK, Response)
}
