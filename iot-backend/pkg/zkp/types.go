package zkp

type ZKPRequest struct {
	DeviceAddress string `json:"deviceAddress"`
	PublicKey     string `json:"publicKey"`
	Message       string `json:"message"`
	Signature     string `json:"signature"`
}

type ZKPResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
