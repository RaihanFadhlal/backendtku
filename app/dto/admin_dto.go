package dto

type GetClaimSafariAllRequestDTO struct {
	PolicyId     string `json:"policy_id"`
	ProductName  string `json:"product_name"`
	DateReport   string `json:"date_report"`
	RegistrantId string `json:"registrant_id"`
	Status       string `json:"status"`
}

type GetClaimSafariAllResponseItemDTO struct {
	ClaimId      string `json:"claim_id"`
	PolicyId     string `json:"policy_id"`
	ProductName  string `json:"product_name"`
	DateReport   string `json:"date_report"`
	DateAccident string `json:"date_accident"`
	Status       string `json:"status"`
	Image        string `json:"image"`
	Evidence     string `json:"evidence"`
	Detail       string `json:"detail"`
	PolicyPdf    string `json:"policy_pdf"`
	RegistrantId string `json:"registrant_id"`
}

type GetClaimAbrorAllRequestDTO struct {
	PolicyId     string `json:"policy_id"`
	DateReport   string `json:"date_report"`
	RegistrantId string `json:"registrant_id"`
	Status       string `json:"status"`
}

type GetClaimAbrorAllResponseItemDTO struct {
	ClaimId      string `json:"claim_id"`
	PolicyId     string `json:"policy_id"`
	ProductName  string `json:"product_name"`
	DateReport   string `json:"date_report"`
	DateAccident string `json:"date_accident"`
	Status       string `json:"status"`
	Image        string `json:"image"`
	Evidence     string `json:"evidence"`
	Detail       string `json:"detail"`
	PolicyPdf    string `json:"policy_pdf"`
	RegistrantId string `json:"registrant_id"`
}

type UpdateClaimRequestDTO struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	CoverCost int    `json:"cover_cost"`
	PayProof  string `json:"pay_proof"`
	ClaimId   string `json:"claim_id"`
}