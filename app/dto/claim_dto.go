package dto

type RequestClaimRequestDTO struct {
	PolicyId     string `json:"policy_id"`
	ProductName  string `json:"product_name"`
	DateReport   string `json:"date_report"`
	DateAccident string `json:"date_acc"`
	Location     string `json:"location"`
	Detail       string `json:"detail"`
	Evidence     string `json:"evidence"`
}

type GetClaimRequestDTO struct {
	ProductName string `json:"product_name"`
	DateReport  string `json:"date_report"`
}

type ClaimResponseItemDTO struct {
	ClaimId      string `json:"claim_id"`
	PolicyId     string `json:"policy_id"`
	ProductName  string `json:"product_name"`
	DateReport   string `json:"date_report"`
	DateAccident string `json:"date_accident"`
	Status       string `json:"status"`
	Image        string `json:"image"`
	Evidence     string `json:"evidence"`
	Detail       string `json:"detail"`
}

type GetClaimAbrorRequestDTO struct {
	CarType    string `json:"car_type"`
	DateReport string `json:"date_report"`
}

type ClaimAbrorResponseItemDTO struct {
	ClaimId      string `json:"claim_id"`
	PolicyId     string `json:"policy_id"`
	ProductName  string `json:"product_name"`
	DateReport   string `json:"date_report"`
	DateAccident string `json:"date_accident"`
	Status       string `json:"status"`
	Image        string `json:"image"`
	Evidence     string `json:"evidence"`
	Detail       string `json:"detail"`
	CarType      string `json:"car_type"`
}

type GetClaimDetailRequestDTO struct {
	ClaimId string `json:"claim_id"`
	Type    string `json:"type"`
}

type GetClaimDetailResponseDTO struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	PayProof  string `json:"pay_proof"`
	CoverCost int    `json:"cover_cost"`
}

type GeneralClaimResponseDTO struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	ClaimId string `json:"claim_id"`
}