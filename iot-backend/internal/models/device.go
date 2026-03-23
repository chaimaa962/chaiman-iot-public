package models

type Device struct {
	Address   string `json:"address"`
	PublicKey string `json:"publicKey"`
	Metadata  string `json:"metadata"`
}

type DeviceInfo struct {
	PublicKey string `json:"publicKey"`
	Metadata  string `json:"metadata"`
	IsActive  bool   `json:"isActive"`
	LastSeen  uint64 `json:"lastSeen"` // uint64 pour le timestamp
}

type AuthRequest struct {
	DeviceAddress string `json:"deviceAddress"`
	Message       string `json:"message"`
	Signature     string `json:"signature"`
}

type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type RegisterResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	TxHash  string `json:"txHash"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
