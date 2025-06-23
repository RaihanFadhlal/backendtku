package controllers

import (
	"backendtku/app/helpers"
	"backendtku/app/repositories"
	"encoding/json"
	"fmt"
	"net/http"
)

func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	var Response struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    []struct {
			Code  string `json:"code"`
			Name  string `json:"name"`
			Price int    `json:"price"`
			Image string `json:"image"`
		} `json:"data"`
	}

	country := r.URL.Query().Get("country")
	products, err := h.ProductRepo.FindAllSafari(country)
	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving products"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	for _, product := range products {
		Response.Data = append(Response.Data, struct {
			Code  string `json:"code"`
			Name  string `json:"name"`
			Price int    `json:"price"`
			Image string `json:"image"`
		}{
			Code:  product.Code,
			Name:  product.Name,
			Price: product.Price,
			Image: h.Config.BaseUrl + "/upload/product/" + product.Image,
		})
	}

	Response.Status = true
	Response.Message = "Products retrieved successfully"
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetProductDetail(w http.ResponseWriter, r *http.Request) {
	type PriceDetails struct {
		Basic    []repositories.PricePeriod `json:"basic"`
		Gold     []repositories.PricePeriod `json:"gold"`
		Platinum []repositories.PricePeriod `json:"platinum"`
		Titanium []repositories.PricePeriod `json:"titanium"`
	}
	type BenefitCategory struct {
		Desc   string                       `json:"desc"`
		Detail []repositories.BenefitDetail `json:"detail"`
	}
	type Response struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Name        string            `json:"name"`
			Description string            `json:"desc"`
			Image       string            `json:"image"`
			Terms       string            `json:"tnc"`
			Types       []string          `json:"type"`
			Countries   string            `json:"countries"`
			Price       PriceDetails      `json:"price"`
			Benefits    []BenefitCategory `json:"benefits"`
		} `json:"data"`
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		helpers.ResponseJSON(w, http.StatusBadRequest, map[string]string{"message": "Product ID is required"})
		return
	}

	product, err := h.ProductRepo.FindSafariByCode(id)
	if err != nil {
		helpers.ResponseJSON(w, http.StatusNotFound, map[string]string{"message": "Product not found"})
		return
	}

	groupCode := id[:len(id)-2]

	contributions, err := h.ProductRepo.FindSafariContributionsByGroupCode(groupCode)
	if err != nil {
		http.Error(w, "Failed to retrieve contributions", http.StatusInternalServerError)
		return
	}

	var priceDetails PriceDetails
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

	var benefits []BenefitCategory
	distinctDesc, _ := h.ProductRepo.FindDistinctSafariBenefitDescriptions(groupCode)
	for _, desc := range distinctDesc {
		benefitDetails, _ := h.ProductRepo.FindSafariBenefitDetails(groupCode, desc)
		benefits = append(benefits, BenefitCategory{
			Desc:   desc,
			Detail: benefitDetails,
		})
	}

	response := Response{
		Status:  true,
		Message: "Product details retrieved successfully",
	}
	response.Data.Name = product.Name
	response.Data.Description = product.Description
	response.Data.Image = h.Config.BaseUrl + "/upload/product/" + product.Image
	response.Data.Terms = product.Terms
	response.Data.Countries = product.Countries
	response.Data.Price = priceDetails
	response.Data.Benefits = benefits
	response.Data.Types = contributions

	helpers.ResponseJSON(w, http.StatusOK, response)
}

func (h *Handler) GetAbrorDetail(w http.ResponseWriter, r *http.Request) {
	type PricePeriod struct {
		C          string  `json:"c"`
		RangePrice string  `json:"range_price"`
		A1         float32 `json:"a1"`
		A2         float32 `json:"a2"`
		A3         float32 `json:"a3"`
	}
	type PriceDetails struct {
		Standard []PricePeriod `json:"standard"`
		Premium  []PricePeriod `json:"premium"`
	}
	type BenefitDetail struct {
		Standard string `json:"standard"`
		Premium  string `json:"premium"`
	}
	type BenefitCategory struct {
		Desc   string          `json:"desc"`
		Type   string          `json:"type"`
		Detail []BenefitDetail `json:"detail"`
	}
	type Cars struct {
		Brand string   `json:"brand"`
		Type  []string `json:"type"`
	}
	type ResponseData struct {
		Name        string            `json:"name"`
		Description string            `json:"desc"`
		Image       string            `json:"image"`
		Terms       string            `json:"terms"`
		Cars        []Cars            `json:"cars"`
		Price       PriceDetails      `json:"price"`
		Benefits    []BenefitCategory `json:"benefits"`
	}
	var Response struct {
		Status  bool         `json:"status"`
		Message string       `json:"message"`
		Data    ResponseData `json:"data"`
	}

	handleError := func(msg string) {
		Response.Status = false
		Response.Message = msg
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
	}

	product, err := h.ProductRepo.FindFirstAbror()
	if err != nil {
		handleError("Error retrieving base product")
		return
	}
	Response.Data.Name = product.Name
	Response.Data.Description = product.Description
	Response.Data.Image = h.Config.BaseUrl + "/upload/product/" + product.Image
	Response.Data.Terms = product.AllowedVehicle

	vehicleTypes, err := h.ProductRepo.FindAllVehicleTypes()
	if err != nil {
		handleError("Error fetching vehicle types")
		return
	}

	standardProducts, err := h.ProductRepo.FindAbrorProductsByTypeName("Standard")
	if err != nil {
		handleError("Error fetching standard products")
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
		Response.Data.Price.Standard = append(Response.Data.Price.Standard, PricePeriod{
			C:          vt.Code,
			RangePrice: fmt.Sprintf("(%d-%d)", vt.Min, vt.Max),
			A1:         standardProductMap[vt.Code]["A1"],
			A2:         standardProductMap[vt.Code]["A2"],
			A3:         standardProductMap[vt.Code]["A3"],
		})
	}

	premiumProducts, err := h.ProductRepo.FindAbrorProductsByTypeName("Premium")
	if err != nil {
		handleError("Error fetching premium products")
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
		Response.Data.Price.Premium = append(Response.Data.Price.Premium, PricePeriod{
			C:          vt.Code,
			RangePrice: fmt.Sprintf("(%d-%d)", vt.Min, vt.Max),
			A1:         premiumProductMap[vt.Code]["A1"],
			A2:         premiumProductMap[vt.Code]["A2"],
			A3:         premiumProductMap[vt.Code]["A3"],
		})
	}

	brands, err := h.ProductRepo.FindAllCarBrands()
	if err != nil {
		handleError("Error fetching car brands")
		return
	}
	for _, brand := range brands {
		types, err := h.ProductRepo.FindCarTypesByBrand(brand)
		if err != nil {
			handleError("Error fetching car types for brand " + brand)
			return
		}
		Response.Data.Cars = append(Response.Data.Cars, Cars{
			Brand: brand,
			Type:  types,
		})
	}

	benefits, err := h.ProductRepo.FindAllAbrorBenefits()
	if err != nil {
		handleError("Error fetching benefits")
		return
	}
	benefitMap := make(map[string]BenefitCategory)
	for _, benefit := range benefits {
		key := benefit.Description + "_" + benefit.Type
		if category, exists := benefitMap[key]; exists {
			category.Detail = append(category.Detail, BenefitDetail{Standard: benefit.Standard, Premium: benefit.Premium})
			benefitMap[key] = category
		} else {
			benefitMap[key] = BenefitCategory{
				Desc:   benefit.Description,
				Type:   benefit.Type,
				Detail: []BenefitDetail{{Standard: benefit.Standard, Premium: benefit.Premium}},
			}
		}
	}
	for _, category := range benefitMap {
		Response.Data.Benefits = append(Response.Data.Benefits, category)
	}

	Response.Status = true
	Response.Message = "Product details retrieved successfully"
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetAbrorPrice(w http.ResponseWriter, r *http.Request) {
	var Request struct {
		Contribution string `json:"contribution"`
		Type         string `json:"type"`
		PlatCode     string `json:"plat_code"`
	}
	var Response struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			ProductCode  string  `json:"product_code"`
			Price        int     `json:"price"`
			Percentage   float32 `json:"percentage"`
			VehiclePrice int     `json:"vehicle_price"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&Request); err != nil {
		Response.Status = false
		Response.Message = "Invalid request payload"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}
	defer r.Body.Close()

	vehicle, err := h.ProductRepo.FindCarByName(Request.Type)
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

	plat := helpers.ExtractPlateCode(Request.PlatCode)
	region, err := h.ProductRepo.FindRegionByPlat(plat)
	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving region"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	product, err := h.ProductRepo.FindAbrorProductByCriteria(Request.Contribution, region.Code, vehicleType.Code)
	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving product"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	price := product.Percentage / 100 * float32(vehicle.Price)

	Response.Data.ProductCode = product.Code
	Response.Data.Price = int(price)
	Response.Data.Percentage = product.Percentage
	Response.Data.VehiclePrice = vehicle.Price
	Response.Status = true
	Response.Message = "Product retrieved successfully"
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetDayMax(w http.ResponseWriter, r *http.Request) {
	var Request struct {
		GroupCode string `json:"group_code"`
	}
	var Response struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			DayMax int `json:"day_max"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&Request); err != nil {
		Response.Status = false
		Response.Message = "Invalid request payload"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}
	defer r.Body.Close()

	price, err := h.ProductRepo.FindSafariWithMaxDay(Request.GroupCode)
	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving product"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Data.DayMax = price.DayMax
	Response.Status = true
	Response.Message = "Product retrieved successfully"
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetCountries(w http.ResponseWriter, r *http.Request) {
	var Response struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Countries string `json:"countries"`
		} `json:"data"`
	}

	countries, err := h.ProductRepo.FindLongestCountries()

	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving country"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Data.Countries = countries
	Response.Status = true
	Response.Message = "Country retrieved successfully"
	helpers.ResponseJSON(w, http.StatusOK, Response)
}

func (h *Handler) GetCars(w http.ResponseWriter, r *http.Request) {
	var Response struct {
		Status  bool     `json:"status"`
		Message string   `json:"message"`
		Data    []string `json:"data"`
	}

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
	var Request struct {
		GroupCode string `json:"group_code"`
		Type      string `json:"type"`
		Period    int    `json:"period"`
	}
	var Response struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Code  string `json:"code"`
			Price int    `json:"price"`
			Name  string `json:"name"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&Request); err != nil {
		Response.Status = false
		Response.Message = "Invalid request payload"
		helpers.ResponseJSON(w, http.StatusBadRequest, Response)
		return
	}
	defer r.Body.Close()
	product, err := h.ProductRepo.FindSafariPrice(Request.GroupCode, Request.Type, Request.Period)
	if err != nil {
		Response.Status = false
		Response.Message = "Error retrieving product"
		helpers.ResponseJSON(w, http.StatusInternalServerError, Response)
		return
	}

	Response.Data.Code = product.Code
	Response.Data.Price = product.Price
	Response.Data.Name = product.Name
	Response.Status = true
	Response.Message = "Product retrieved successfully"
	helpers.ResponseJSON(w, http.StatusOK, Response)
}
