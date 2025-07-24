package controllers

import (
	"backendtku/app/dto"
	"backendtku/app/helpers"
	"encoding/json"
	"fmt"
	"net/http"
)

func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	var Response dto.BaseResponse

	country := r.URL.Query().Get("country")
	products, err := h.ProductRepo.FindAllSafari(country)
	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving products"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	var productItems []dto.GetProductsResponseItemDTO
	for _, product := range products {
		productItems = append(productItems, dto.GetProductsResponseItemDTO{
			Code:  product.Code,
			Name:  product.Name,
			Price: product.Price,
			Image: h.Config.BaseUrl + "/upload/product/" + product.Image,
		})
	}

	Response.Status = true
	Response.Message = "Products retrieved successfully"
	Response.Data = productItems
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetProductDetail(w http.ResponseWriter, r *http.Request) {
	var Response dto.BaseResponse

	id := r.URL.Query().Get("id")
	if id == "" {
		Response.Status = false
		Response.Message = "Product ID is required"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}

	product, err := h.ProductRepo.FindSafariByCode(id)
	if err != nil {
		Response.Status = false
		Response.Message = "Product not found"
		helpers.ResponseJSON(w, http.StatusNotFound, Response)
		return
	}

	groupCode := id[:len(id)-2]

	contributions, err := h.ProductRepo.FindSafariContributionsByGroupCode(groupCode)
	if err != nil {
		Response.Status = false
		Response.Message = "Failed to retrieve contributions"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	var priceDetails dto.ProductPriceDetailsDTO
	priceCategories := []string{"Basic", "Gold", "Platinum", "Titanium"}
	for _, category := range priceCategories {
		prices, _ := h.ProductRepo.FindSafariPriceDetails(groupCode, category)
		switch category {
		case "Basic":
			priceDetails.Basic = prices
		case "Gold":
			priceDetails.Gold = prices
		case "Platinum":
			priceDetails.Platinum = prices
		case "Titanium":
			priceDetails.Titanium = prices
		}
	}

	var benefits []dto.BenefitCategoryDTO
	distinctDesc, _ := h.ProductRepo.FindDistinctSafariBenefitDescriptions(groupCode)
	for _, desc := range distinctDesc {
		benefitDetails, _ := h.ProductRepo.FindSafariBenefitDetails(groupCode, desc)
		benefits = append(benefits, dto.BenefitCategoryDTO{
			Desc:   desc,
			Detail: benefitDetails,
		})
	}

	productDetail := dto.GetProductDetailResponseDTO{
		Name:        product.Name,
		Description: product.Description,
		Image:       h.Config.BaseUrl + "/upload/product/" + product.Image,
		Terms:       product.Terms,
		Countries:   product.Countries,
		Price:       priceDetails,
		Benefits:    benefits,
		Types:       contributions,
	}

	Response.Status = true
	Response.Message = "Product details retrieved successfully"
	Response.Data = productDetail
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetAbrorDetail(w http.ResponseWriter, r *http.Request) {
	var Response dto.BaseResponse
	var abrorDetail dto.GetAbrorDetailResponseDTO

	handleError := func(msg string, statusCode int) {
		Response.Status = false
		Response.Message = msg
		helpers.ResponseJSON(w, statusCode, Response)
	}

	product, err := h.ProductRepo.FindFirstAbror()
	if err != nil {
		handleError("Error retrieving base product", http.StatusInternalServerError)
		return
	}
	abrorDetail.Name = product.Name
	abrorDetail.Description = product.Description
	abrorDetail.Image = h.Config.BaseUrl + "/upload/product/" + product.Image
	abrorDetail.Terms = product.AllowedVehicle

	vehicleTypes, err := h.ProductRepo.FindAllVehicleTypes()
	if err != nil {
		handleError("Error fetching vehicle types", http.StatusInternalServerError)
		return
	}

	standardProducts, err := h.ProductRepo.FindAbrorProductsByTypeName("Standard")
	if err != nil {
		handleError("Error fetching standard products", http.StatusInternalServerError)
		return
	}
	standardProductMap := make(map[string]map[string]float32)
	for _, pa := range standardProducts {
		if _, exists := standardProductMap[pa.VehicleCode]; !exists {
			standardProductMap[pa.VehicleCode] = make(map[string]float32)
		}
		standardProductMap[pa.VehicleCode][pa.RegionCode] = pa.Percentage
	}
	for _, vt := range vehicleTypes {
		abrorDetail.Price.Standard = append(abrorDetail.Price.Standard, dto.AbrorPricePeriodDTO{
			C:          vt.Code,
			RangePrice: fmt.Sprintf("(%d-%d)", vt.Min, vt.Max),
			A1:         standardProductMap[vt.Code]["A1"],
			A2:         standardProductMap[vt.Code]["A2"],
			A3:         standardProductMap[vt.Code]["A3"],
		})
	}

	premiumProducts, err := h.ProductRepo.FindAbrorProductsByTypeName("Premium")
	if err != nil {
		handleError("Error fetching premium products", http.StatusInternalServerError)
		return
	}
	premiumProductMap := make(map[string]map[string]float32)
	for _, pa := range premiumProducts {
		if _, exists := premiumProductMap[pa.VehicleCode]; !exists {
			premiumProductMap[pa.VehicleCode] = make(map[string]float32)
		}
		premiumProductMap[pa.VehicleCode][pa.RegionCode] = pa.Percentage
	}
	for _, vt := range vehicleTypes {
		abrorDetail.Price.Premium = append(abrorDetail.Price.Premium, dto.AbrorPricePeriodDTO{
			C:          vt.Code,
			RangePrice: fmt.Sprintf("(%d-%d)", vt.Min, vt.Max),
			A1:         premiumProductMap[vt.Code]["A1"],
			A2:         premiumProductMap[vt.Code]["A2"],
			A3:         premiumProductMap[vt.Code]["A3"],
		})
	}

	brands, err := h.ProductRepo.FindAllCarBrands()
	if err != nil {
		handleError("Error fetching car brands", http.StatusInternalServerError)
		return
	}
	for _, brand := range brands {
		types, err := h.ProductRepo.FindCarTypesByBrand(brand)
		if err != nil {
			handleError("Error fetching car types for brand "+brand, http.StatusInternalServerError)
			return
		}
		abrorDetail.Cars = append(abrorDetail.Cars, dto.CarDetailsDTO{
			Brand: brand,
			Type:  types,
		})
	}

	benefits, err := h.ProductRepo.FindAllAbrorBenefits()
	if err != nil {
		handleError("Error fetching benefits", http.StatusInternalServerError)
		return
	}
	benefitMap := make(map[string]dto.AbrorBenefitCategoryDTO)
	for _, benefit := range benefits {
		key := benefit.Description + "_" + benefit.Type
		if category, exists := benefitMap[key]; exists {
			category.Detail = append(category.Detail, dto.AbrorBenefitDetailDTO{Standard: benefit.Standard, Premium: benefit.Premium})
			benefitMap[key] = category
		} else {
			benefitMap[key] = dto.AbrorBenefitCategoryDTO{
				Desc:   benefit.Description,
				Type:   benefit.Type,
				Detail: []dto.AbrorBenefitDetailDTO{{Standard: benefit.Standard, Premium: benefit.Premium}},
			}
		}
	}
	for _, category := range benefitMap {
		abrorDetail.Benefits = append(abrorDetail.Benefits, category)
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

	vehicle, err := h.ProductRepo.FindCarByName(requestDTO.Type)
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

	plat := helpers.ExtractPlateCode(requestDTO.PlatCode)
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

	abrorPriceResponse := dto.GetAbrorPriceResponseDTO{
		ProductCode: product.Code,
		Price:       int(price),
		Percentage:  product.Percentage,
		VehiclePrice: vehicle.Price,
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

	price, err := h.ProductRepo.FindSafariWithMaxDay(requestDTO.GroupCode)
	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving product"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	dayMaxResponse := dto.GetDayMaxResponseDTO{
		DayMax: price.DayMax,
	}

	Response.Status = true
	Response.Message = "Product retrieved successfully"
	Response.Data = dayMaxResponse
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetCountries(w http.ResponseWriter, r *http.Request) {
	var Response dto.BaseResponse

	countries, err := h.ProductRepo.FindLongestCountries()

	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving country"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	countriesResponse := dto.GetCountriesResponseDTO{
		Countries: countries,
	}

	Response.Status = true
	Response.Message = "Country retrieved successfully"
	Response.Data = countriesResponse
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetCars(w http.ResponseWriter, r *http.Request) {
	var Response dto.BaseResponse

	carNames, err := h.ProductRepo.FindAllCarNames()
	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving car names"
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
	product, err := h.ProductRepo.FindSafariPrice(requestDTO.GroupCode, requestDTO.Type, requestDTO.Period)
	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving product"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	safariPriceResponse := dto.GetSafariPriceResponseDTO{
		Code:  product.Code,
		Price: product.Price,
		Name:  product.Name,
	}

	Response.Status = true
	Response.Message = "Product retrieved successfully"
	Response.Data = safariPriceResponse
	helpers.ResponseJSON(w, http.StatusOK, Response)
}
