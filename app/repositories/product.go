package repositories

import (
	"backendtku/app/models"
	"gorm.io/gorm"
)

type PricePeriod struct {
	DayMin int `json:"day_min"`
	DayMax int `json:"day_max"`
	Price  int `json:"price"`
}

type BenefitDetail struct {
	Detail   string `json:"detail"`
	Basic    string `json:"basic"`
	Gold     string `json:"gold"`
	Platinum string `json:"platinum"`
	Titanium string `json:"titanium"`
}

type AbrorPricePeriod struct {
	C          string  `json:"c"`
	RangePrice string  `json:"range_price"`
	A1         float32 `json:"a1"`
	A2         float32 `json:"a2"`
	A3         float32 `json:"a3"`
}

type ProductRepository interface { // Untuk GetProducts
	FindAllSafari(country string) ([]models.ProductSafari, error)

	// GetProductDetail
	FindSafariByCode(code string) (*models.ProductSafari, error)
	FindSafariContributionsByGroupCode(groupCode string) ([]string, error)
	FindSafariPriceDetails(groupCode, category string) ([]PricePeriod, error)
	FindDistinctSafariBenefitDescriptions(groupCode string) ([]string, error)
	FindSafariBenefitDetails(groupCode, desc string) ([]BenefitDetail, error)

	// GetAbrorDetail
	FindFirstAbror() (*models.ProductAbror, error)
	FindAllVehicleTypes() ([]models.VehicleType, error)
	FindAbrorProductsByTypeName(typeName string) ([]models.ProductAbror, error)
	FindAllCarBrands() ([]string, error)
	FindCarTypesByBrand(brand string) ([]string, error)
	FindAllAbrorBenefits() ([]models.ProductBenefitAbror, error)

	// GetCars
	FindAllCarNames() ([]string, error)

	// GetAbrorPrice
	FindCarByName(name string) (*models.Car, error)
	FindVehicleTypeByPrice(price int64) (*models.VehicleType, error)
	FindRegionByPlat(platCode string) (*models.Region, error)
	FindAbrorProductByCriteria(contribution, regionCode, vehicleCode string) (*models.ProductAbror, error)

	// GetDayMax
	FindSafariWithMaxDay(groupCode string) (*models.ProductSafari, error)

	// GetCountries
	FindLongestCountries() (string, error)

	// GetSafariPrice
	FindSafariPrice(groupCode, contribution string, period int) (*models.ProductSafari, error)

	FindSafariProductByGroupCode(groupCode string) (*models.ProductSafari, error)
	FindSafariBenefitsByGroupCode(groupCode string) ([]models.ProductBenefitSafari, error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) FindAllSafari(country string) ([]models.ProductSafari, error) {
	var products []models.ProductSafari
	query := r.db.Where("code LIKE ?", "%B1")

	if country != "" {
		countryFilter := "%" + country + "%"
		query = query.Where("countries LIKE ?", countryFilter)
	}

	if err := query.Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *productRepository) FindSafariByCode(code string) (*models.ProductSafari, error) {
	var product models.ProductSafari
	err := r.db.Where("code = ?", code).First(&product).Error
	return &product, err
}

func (r *productRepository) FindSafariContributionsByGroupCode(groupCode string) ([]string, error) {
	var contributions []string
	err := r.db.Model(&models.ProductSafari{}).Distinct("contribution").Where("group_code = ?", groupCode).Pluck("contribution", &contributions).Error
	return contributions, err
}

func (r *productRepository) FindSafariPriceDetails(groupCode, category string) ([]PricePeriod, error) {
	var prices []PricePeriod
	err := r.db.Table("product_safaris").Select("day_min, day_max, price").Where("group_code = ? AND contribution = ?", groupCode, category).Order("day_min ASC").Scan(&prices).Error
	return prices, err
}

func (r *productRepository) FindDistinctSafariBenefitDescriptions(groupCode string) ([]string, error) {
	var distinctDesc []string
	err := r.db.Table("product_benefit_safaris").Select("distinct COALESCE(description, '-') as desc").Where("group_code = ?", groupCode).Scan(&distinctDesc).Error
	return distinctDesc, err
}

func (r *productRepository) FindSafariBenefitDetails(groupCode, desc string) ([]BenefitDetail, error) {
	var benefitDetails []BenefitDetail
	err := r.db.Table("product_benefit_safaris").Select("COALESCE(detail, '-') as detail, COALESCE(basic, '-') as basic, COALESCE(gold, '-') as gold, COALESCE(platinum, '-') as platinum, COALESCE(titanium, '-') as titanium").Where("group_code = ? AND description = ?", groupCode, desc).Scan(&benefitDetails).Error
	return benefitDetails, err
}

func (r *productRepository) FindFirstAbror() (*models.ProductAbror, error) {
	var product models.ProductAbror
	err := r.db.Where("id = '1'").First(&product).Error
	return &product, err
}

func (r *productRepository) FindAllVehicleTypes() ([]models.VehicleType, error) {
	var vehicleTypes []models.VehicleType
	err := r.db.Find(&vehicleTypes).Error
	return vehicleTypes, err
}

func (r *productRepository) FindAbrorProductsByTypeName(typeName string) ([]models.ProductAbror, error) {
	var productAbrors []models.ProductAbror
	err := r.db.Where("type_name = ?", typeName).Find(&productAbrors).Error
	return productAbrors, err
}

func (r *productRepository) FindAllCarBrands() ([]string, error) {
	var brands []string
	err := r.db.Model(&models.Car{}).Distinct().Pluck("brand", &brands).Error
	return brands, err
}

func (r *productRepository) FindCarTypesByBrand(brand string) ([]string, error) {
	var types []string
	err := r.db.Model(&models.Car{}).Where("brand = ?", brand).Pluck("name", &types).Error
	return types, err
}

func (r *productRepository) FindAllAbrorBenefits() ([]models.ProductBenefitAbror, error) {
	var benefits []models.ProductBenefitAbror
	err := r.db.Find(&benefits).Error
	return benefits, err
}

func (r *productRepository) FindCarByName(name string) (*models.Car, error) {
	var vehicle models.Car
	err := r.db.Where("name = ?", name).First(&vehicle).Error
	return &vehicle, err
}

func (r *productRepository) FindVehicleTypeByPrice(price int64) (*models.VehicleType, error) {
	var vehicleType models.VehicleType
	err := r.db.Where("min <= ? AND max >= ?", price, price).First(&vehicleType).Error
	return &vehicleType, err
}

func (r *productRepository) FindRegionByPlat(platCode string) (*models.Region, error) {
	var region models.Region
	query := `SELECT * FROM regions WHERE ? = ANY(string_to_array(plat, ','))`
	err := r.db.Raw(query, platCode).Scan(&region).Error
	return &region, err
}

func (r *productRepository) FindAbrorProductByCriteria(contribution, regionCode, vehicleCode string) (*models.ProductAbror, error) {
	var product models.ProductAbror
	err := r.db.Where("type_name = ? AND region_code = ? AND vehicle_code = ? ", contribution, regionCode, vehicleCode).First(&product).Error
	return &product, err
}

func (r *productRepository) FindSafariWithMaxDay(groupCode string) (*models.ProductSafari, error) {
	var price models.ProductSafari
	err := r.db.Where("group_code = ?", groupCode).Order("day_max desc").First(&price).Error
	return &price, err
}

func (r *productRepository) FindLongestCountries() (string, error) {
	var result struct {
		Countries string
	}
	err := r.db.Model(&models.ProductSafari{}).Select("countries").Order("LENGTH(countries) DESC").Limit(1).Scan(&result).Error
	return result.Countries, err
}

func (r *productRepository) FindAllCarNames() ([]string, error) {
	var carNames []string
	err := r.db.Model(&models.Car{}).Select("name").Find(&carNames).Error
	return carNames, err
}

func (r *productRepository) FindSafariPrice(groupCode, contribution string, period int) (*models.ProductSafari, error) {
	var product models.ProductSafari
	err := r.db.Where("group_code = ? AND contribution = ? AND day_min <= ? AND day_max >= ?", groupCode, contribution, period, period).First(&product).Error
	return &product, err
}

func (r *productRepository) FindSafariProductByGroupCode(groupCode string) (*models.ProductSafari, error) {
	var product models.ProductSafari
	err := r.db.Where("group_code = ?", groupCode).First(&product).Error
	return &product, err
}

func (r *productRepository) FindSafariBenefitsByGroupCode(groupCode string) ([]models.ProductBenefitSafari, error) {
	var benefits []models.ProductBenefitSafari
	err := r.db.Where("group_code = ?", groupCode).Find(&benefits).Error
	return benefits, err
}