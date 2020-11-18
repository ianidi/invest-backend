package otp

type GatewayAPIRes struct {
	Ids   []interface{} `json:"ids"`
	Usage struct {
		Countries struct {
			DK int `json:"DK"`
		} `json:"countries"`
		Currency  string  `json:"currency"`
		TotalCost float64 `json:"total_cost"`
	} `json:"usage"`
}

type GatewayAPIRecipient struct {
	Msisdn uint64 `json:"msisdn"`
}
type GatewayAPIRequest struct {
	Sender     string                `json:"sender"`
	Message    string                `json:"message"`
	Recipients []GatewayAPIRecipient `json:"recipients"`
}
type GatewayAPIResponse struct {
	Ids []uint64
}
