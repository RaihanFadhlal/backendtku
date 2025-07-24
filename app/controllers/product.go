package controllers

import (
	"backendtku/app/dto"
	"backendtku/app/helpers"
	"encoding/json"
	"net/http"
)

func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	var Response dto.BaseResponse

	country := r.URL.Query().Get("country")
	productItems, err := h.ProductUseCase.GetProducts(country)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Products retrieved successfully"
	Response.Data = productItems
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetProductDetail(w http.ResponseWriter, r *http.Request) {
	var Response dto.BaseResponse

	id := r.URL.Query().Get("id")
	productDetail, err := h.ProductUseCase.GetProductDetail(id)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Product details retrieved successfully"
	Response.Data = productDetail
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetAbrorDetail(w http.ResponseWriter, r *http.Request) {
	var Response dto.BaseResponse

	abrorDetail, err := h.ProductUseCase.GetAbrorDetail()
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Product details retrieved successfully"
	Response.Data = abrorDetail
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetAbrorPrice(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.GetAbrorPriceRequestDTO
	var Response dto.BaseResponse

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request payload"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}
	defer r.Body.Close()

	abrorPriceResponse, err := h.ProductUseCase.GetAbrorPrice(requestDTO)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Product retrieved successfully"
	Response.Data = abrorPriceResponse
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetDayMax(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.GetDayMaxRequestDTO
	var Response dto.BaseResponse

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request payload"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}
	defer r.Body.Close()

	dayMaxResponse, err := h.ProductUseCase.GetDayMax(requestDTO.GroupCode)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Product retrieved successfully"
	Response.Data = dayMaxResponse
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetCountries(w http.ResponseWriter, r *http.Request) {
	var Response dto.BaseResponse

	countriesResponse, err := h.ProductUseCase.GetCountries()
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Country retrieved successfully"
	Response.Data = countriesResponse
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetCars(w http.ResponseWriter, r *http.Request) {
	var Response dto.BaseResponse

	carNames, err := h.ProductUseCase.GetCars()
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Car names retrieved successfully"
	Response.Data = carNames
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetSafariPrice(w http.ResponseWriter, r *http.Request) {
	var requestDTO dto.GetSafariPriceRequestDTO
	var Response dto.BaseResponse

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestDTO); err != nil {
		Response.Status = false
		Response.Message = "Invalid request payload"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}
	defer r.Body.Close()
	product, err := h.ProductUseCase.GetSafariPrice(requestDTO)
	if err != nil {
		Response.Status = false
		Response.Message = err.Error()
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Status = true
	Response.Message = "Product retrieved successfully"
	Response.Data = product
	helpers.ResponseJSON(w, http.StatusOK, Response)
}