package models

// Dans internal/models/transaction.go, ajouter ExpectedHash
type NodeMessageRequest struct {
    DeviceID     string `json:"device_id"`
    Address      string `json:"address"`
    Message      string `json:"message"`
    MessageHash  string `json:"message_hash"`
    Signature    string `json:"signature"`
    Proof        string `json:"proof"`
    ExpectedHash string `json:"expected_hash"` // ← AJOUTER
    Timestamp    int64  `json:"timestamp"`
    Sequence     uint64 `json:"sequence"`
    Nonce        string `json:"nonce"`
}

type NodeMessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	TxHash  string `json:"txHash"`
	Block   uint64 `json:"block"`
}


