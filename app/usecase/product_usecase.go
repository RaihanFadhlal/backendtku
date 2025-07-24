package usecase

import (
	"backendtku/app/dto"
	"backendtku/app/helpers"
	"backendtku/app/repositories"
	"backendtku/config"
	"errors"
	"fmt"
)

type ProductUseCase interface {
	GetProducts(country string) ([]dto.GetProductsResponseItemDTO, error)
	GetProductDetail(id string) (dto.GetProductDetailResponseDTO, error)
	GetAbrorDetail() (dto.GetAbrorDetailResponseDTO, error)
	GetAbrorPrice(requestDTO dto.GetAbrorPriceRequestDTO) (dto.GetAbrorPriceResponseDTO, error)
	GetDayMax(groupCode string) (dto.GetDayMaxResponseDTO, error)
	GetCountries() (dto.GetCountriesResponseDTO, error)
	GetCars() (dto.GetCarsResponseDTO, error)
	GetSafariPrice(requestDTO dto.GetSafariPriceRequestDTO) (dto.GetSafariPriceResponseDTO, error)
}

type productUseCase struct {
	productRepo repositories.ProductRepository
	config      *config.Config
}

func NewProductUseCase(productRepo repositories.ProductRepository, cfg *config.Config) ProductUseCase {
	return &productUseCase{
		productRepo: productRepo,
		config:      cfg,
	}
}

func (uc *productUseCase) GetProducts(country string) ([]dto.GetProductsResponseItemDTO, error) {
	products, err := uc.productRepo.FindAllSafari(country)
	if err != nil {
		return nil, errors.New("Error retrieving products")
	}

	var productItems []dto.GetProductsResponseItemDTO
	for _, product := range products {
		productItems = append(productItems, dto.GetProductsResponseItemDTO{
			Code:  product.Code,
			Name:  product.Name,
			Price: product.Price,
			Image: uc.config.BaseUrl + "/upload/product/" + product.Image,
		})
	}

	return productItems, nil
}

func (uc *productUseCase) GetProductDetail(id string) (dto.GetProductDetailResponseDTO, error) {
	if id == "" {
		return dto.GetProductDetailResponseDTO{}, errors.New("Product ID is required")
	}

	product, err := uc.productRepo.FindSafariByCode(id)
	if err != nil {
		return dto.GetProductDetailResponseDTO{}, errors.New("Product not found")
	}

	groupCode := id[:len(id)-2]

	contributions, err := uc.productRepo.FindSafariContributionsByGroupCode(groupCode)
	if err != nil {
		return dto.GetProductDetailResponseDTO{}, errors.New("Failed to retrieve contributions")
	}

	var priceDetails dto.ProductPriceDetailsDTO
	priceCategories := []string{"Basic", "Gold", "Platinum", "Titanium"}
	for _, category := range priceCategories {
		prices, _ := uc.productRepo.FindSafariPriceDetails(groupCode, category)
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
	distinctDesc, _ := uc.productRepo.FindDistinctSafariBenefitDescriptions(groupCode)
	for _, desc := range distinctDesc {
		benefitDetails, _ := uc.productRepo.FindSafariBenefitDetails(groupCode, desc)
		benefits = append(benefits, dto.BenefitCategoryDTO{
			Desc:   desc,
			Detail: benefitDetails,
		})
	}

	productDetail := dto.GetProductDetailResponseDTO{
		Name:        product.Name,
		Description: product.Description,
		Image:       uc.config.BaseUrl + "/upload/product/" + product.Image,
		Terms:       product.Terms,
		Countries:   product.Countries,
		Price:       priceDetails,
		Benefits:    benefits,
		Types:       contributions,
	}

	return productDetail, nil
}

func (uc *productUseCase) GetAbrorDetail() (dto.GetAbrorDetailResponseDTO, error) {
	var abrorDetail dto.GetAbrorDetailResponseDTO

	product, err := uc.productRepo.FindFirstAbror()
	if err != nil {
		return dto.GetAbrorDetailResponseDTO{}, errors.New("Error retrieving base product")
	}
	abrorDetail.Name = product.Name
	abrorDetail.Description = product.Description
	abrorDetail.Image = uc.config.BaseUrl + "/upload/product/" + product.Image
	abrorDetail.Terms = product.AllowedVehicle

	vehicleTypes, err := uc.productRepo.FindAllVehicleTypes()
	if err != nil {
		return dto.GetAbrorDetailResponseDTO{}, errors.New("Error fetching vehicle types")
	}

	standardProducts, err := uc.productRepo.FindAbrorProductsByTypeName("Standard")
	if err != nil {
		return dto.GetAbrorDetailResponseDTO{}, errors.New("Error fetching standard products")
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

	premiumProducts, err := uc.productRepo.FindAbrorProductsByTypeName("Premium")
	if err != nil {
		return dto.GetAbrorDetailResponseDTO{}, errors.New("Error fetching premium products")
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

	brands, err := uc.productRepo.FindAllCarBrands()
	if err != nil {
		return dto.GetAbrorDetailResponseDTO{}, errors.New("Error fetching car brands")
	}
	for _, brand := range brands {
		types, err := uc.productRepo.FindCarTypesByBrand(brand)
		if err != nil {
			return dto.GetAbrorDetailResponseDTO{}, errors.New("Error fetching car types for brand " + brand)
		}
		abrorDetail.Cars = append(abrorDetail.Cars, dto.CarDetailsDTO{
			Brand: brand,
			Type:  types,
		})
	}

	benefits, err := uc.productRepo.FindAllAbrorBenefits()
	if err != nil {
		return dto.GetAbrorDetailResponseDTO{}, errors.New("Error fetching benefits")
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

	return abrorDetail, nil
}

func (uc *productUseCase) GetAbrorPrice(requestDTO dto.GetAbrorPriceRequestDTO) (dto.GetAbrorPriceResponseDTO, error) {
	vehicle, err := uc.productRepo.FindCarByName(requestDTO.Type)
	if err != nil {
		return dto.GetAbrorPriceResponseDTO{}, errors.New("Error retrieving car")
	}

	vehicleType, err := uc.productRepo.FindVehicleTypeByPrice(int64(vehicle.Price))
	if err != nil {
		return dto.GetAbrorPriceResponseDTO{}, errors.New("Error retrieving car type")
	}

	plat := helpers.ExtractPlateCode(requestDTO.PlatCode)
	region, err := uc.productRepo.FindRegionByPlat(plat)
	if err != nil {
		return dto.GetAbrorPriceResponseDTO{}, errors.New("Error retrieving region")
	}

	product, err := uc.productRepo.FindAbrorProductByCriteria(requestDTO.Contribution, region.Code, vehicleType.Code)
	if err != nil {
		return dto.GetAbrorPriceResponseDTO{}, errors.New("Error retrieving product")
	}

	price := product.Percentage / 100 * float32(vehicle.Price)

	abrorPriceResponse := dto.GetAbrorPriceResponseDTO{
		ProductCode: product.Code,
		Price:       int(price),
		Percentage:  product.Percentage,
		VehiclePrice: vehicle.Price,
	}

	return abrorPriceResponse, nil
}

func (uc *productUseCase) GetDayMax(groupCode string) (dto.GetDayMaxResponseDTO, error) {
	price, err := uc.productRepo.FindSafariWithMaxDay(groupCode)
	if err != nil {
		return dto.GetDayMaxResponseDTO{}, errors.New("Error retrieving product")
	}

	dayMaxResponse := dto.GetDayMaxResponseDTO{
		DayMax: price.DayMax,
	}

	return dayMaxResponse, nil
}

func (uc *productUseCase) GetCountries() (dto.GetCountriesResponseDTO, error) {
	countries, err := uc.productRepo.FindLongestCountries()

	if err != nil {
		return dto.GetCountriesResponseDTO{}, errors.New("Error retrieving country")
	}

	countriesResponse := dto.GetCountriesResponseDTO{
		Countries: countries,
	}

	return countriesResponse, nil
}

func (uc *productUseCase) GetCars() (dto.GetCarsResponseDTO, error) {
	carNames, err := uc.productRepo.FindAllCarNames()
	if err != nil {
		return dto.GetCarsResponseDTO{}, errors.New("Error retrieving car names")
	}

	return dto.GetCarsResponseDTO{
		CarNames: carNames,
	}, nil
}

func (uc *productUseCase) GetSafariPrice(requestDTO dto.GetSafariPriceRequestDTO) (dto.GetSafariPriceResponseDTO, error) {
	product, err := uc.productRepo.FindSafariPrice(requestDTO.GroupCode, requestDTO.Type, requestDTO.Period)
	if err != nil {
		return dto.GetSafariPriceResponseDTO{}, errors.New("Error retrieving product")
	}

	safariPriceResponse := dto.GetSafariPriceResponseDTO{
		Code:  product.Code,
		Price: product.Price,
		Name:  product.Name,
	}

	return safariPriceResponse, nil
}
