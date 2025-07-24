package dto

import "backendtku/app/repositories"

type GetProductsResponseItemDTO struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Price int    `json:"price"`
	Image string `json:"image"`
}

type GetProductDetailResponseDTO struct {
	Name        string            `json:"name"`
	Description string            `json:"desc"`
	Image       string            `json:"image"`
	Terms       string            `json:"tnc"`
	Types       []string          `json:"type"`
	Countries   string            `json:"countries"`
	Price       ProductPriceDetailsDTO `json:"price"`
	Benefits    []BenefitCategoryDTO `json:"benefits"`
}

type ProductPriceDetailsDTO struct {
	Basic    []repositories.PricePeriod `json:"basic"`
	Gold     []repositories.PricePeriod `json:"gold"`
	Platinum []repositories.PricePeriod `json:"platinum"`
	Titanium []repositories.PricePeriod `json:"titanium"`
}

type BenefitCategoryDTO struct {
	Desc   string                       `json:"desc"`
	Detail []repositories.BenefitDetail `json:"detail"`
}

type GetAbrorDetailResponseDTO struct {
	Name        string            `json:"name"`
	Description string            `json:"desc"`
	Image       string            `json:"image"`
	Terms       string            `json:"terms"`
	Cars        []CarDetailsDTO   `json:"cars"`
	Price       AbrorPriceDetailsDTO `json:"price"`
	Benefits    []AbrorBenefitCategoryDTO `json:"benefits"`
}

type CarDetailsDTO struct {
	Brand string   `json:"brand"`
	Type  []string `json:"type"`
}

type AbrorPriceDetailsDTO struct {
	Standard []AbrorPricePeriodDTO `json:"standard"`
	Premium  []AbrorPricePeriodDTO `json:"premium"`
}

type AbrorPricePeriodDTO struct {
	C          string  `json:"c"`
	RangePrice string  `json:"range_price"`
	A1         float32 `json:"a1"`
	A2         float32 `json:"a2"`
	A3         float32 `json:"a3"`
}

type AbrorBenefitCategoryDTO struct {
	Desc   string                `json:"desc"`
	Type   string                `json:"type"`
	Detail []AbrorBenefitDetailDTO `json:"detail"`
}

type AbrorBenefitDetailDTO struct {
	Standard string `json:"standard"`
	Premium  string `json:"premium"`
}

type GetAbrorPriceRequestDTO struct {
	Contribution string `json:"contribution"`
	Type         string `json:"type"`
	PlatCode     string `json:"plat_code"`
}

type GetAbrorPriceResponseDTO struct {
	ProductCode  string  `json:"product_code"`
	Price        int     `json:"price"`
	Percentage   float32 `json:"percentage"`
	VehiclePrice int     `json:"vehicle_price"`
}

type GetDayMaxRequestDTO struct {
	GroupCode string `json:"group_code"`
}

type GetDayMaxResponseDTO struct {
	DayMax int `json:"day_max"`
}

type GetCountriesResponseDTO struct {
	Countries string `json:"countries"`
}

type GetCarsResponseDTO struct {
	CarNames []string `json:"car_names"`
}

type GetSafariPriceRequestDTO struct {
	GroupCode string `json:"group_code"`
	Type      string `json:"type"`
	Period    int    `json:"period"`
}

type GetSafariPriceResponseDTO struct {
	Code  string `json:"code"`
	Price int    `json:"price"`
	Name  string `json:"name"`
}