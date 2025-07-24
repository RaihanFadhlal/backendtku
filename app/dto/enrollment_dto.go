package dto


type CreateTransactionRequestDTO struct {
	TrxId string `json:"trx_id"`
}

type DownloadPdfRequestDTO struct {
	PolicyId string `json:"policy_id"`
}

type RequestProductRequestDTO struct {
	ProductCode  string `json:"product_code"`
	ProductName  string `json:"product_name"`
	Capacity     int    `json:"capacity"`
	ProductPrice int64  `json:"product_price"`
	Phone        string `json:"phone"`
	From         string `json:"from"`
	Destination  string `json:"destination"`
	DateStart    string `json:"date_start"`
	DateEnd      string `json:"date_end"`
	Contribution string `json:"contribution"`
	FullName     string `json:"fullname"`
	Birthdate    string `json:"birthdate"`
	Birthplace   string `json:"birthplace"`
	Gender       string `json:"gender"`
	Passport     string `json:"passport"`
	Others       []struct {
		Fullname  string `json:"fullname"`
		Birthdate string `json:"birthdate"`
	} `json:"others"`
}

type RequestAbrorRequestDTO struct {
	Contribution string `json:"contribution"`
	ProductCode  string `json:"product_code"`
	ProductName  string `json:"product_name"`
	CarBrand     string `json:"car_brand"`
	CarType      string `json:"car_type"`
	Year         string `json:"year"`
	DateStart    string `json:"date_start"`
	DateEnd      string `json:"date_end"`
	Price        int    `json:"price"`
	Plat         string `json:"plat"`
	Chassis      string `json:"chassis"`
	Engine       string `json:"engine"`
	Image1       string `json:"image1"`
	Image2       string `json:"image2"`
	Image3       string `json:"image3"`
	Image4       string `json:"image4"`
	Fullname     string `json:"fullname"`
	Birthdate    string `json:"birtdate"`
	Gender       string `json:"gender"`
	Phone        string `json:"phone"`
	IdUser       string `json:"id_user"`
}

type PaymentStatusRequestDTO struct {
	TrxId string `json:"trx_id"`
}

type GetPoliciesRequestDTO struct {
	ProductName string `json:"product_name"`
	Destination string `json:"destination"`
}

type PolicyResponseItemDTO struct {
	PolicyId     string `json:"policy_id"`
	ProductName  string `json:"product_name"`
	Contribution string `json:"contribution"`
	Destination  string `json:"destination"`
	DateStart    string `json:"sdate"`
	DateEnd      string `json:"edate"`
	Image        string `json:"image"`
}

type GetPoliciesAbrorRequestDTO struct {
	CarType   string `json:"car_type"`
	DateStart string `json:"date_start"`
}

type PolicyAbrorResponseItemDTO struct {
	PolicyId     string `json:"policy_id"`
	ProductName  string `json:"product_name"`
	Contribution string `json:"contribution"`
	CarType      string `json:"car_type"`
	DateStart    string `json:"date_start"`
	DateEnd      string `json:"date_end"`
	Image        string `json:"image"`
}

type GetTrxRequestDTO struct {
	ProductName string `json:"product_name"`
	Status      string `json:"status"`
}

type TransactionResponseItemDTO struct {
	TransactionId string `json:"transaction_id"`
	ProductName   string `json:"product_name"`
	TotalPrice    int    `json:"total_price"`
	CreatedAt     string `json:"created_at"`
	ExpiredAt     string `json:"expired_at"`
	Status        string `json:"status"`
	Image         string `json:"image"`
}

type GeneralResponseDTO struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type TransactionIDResponseDTO struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	TrxId   string `json:"trx_id"`
}

type SnapResponseDTO struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}