// internal/models/transaction.go
// [FIX] AuthTimings ajouté dans NodeMessageResponse
// pour retourner les vrais temps ECDSA, ZKP et blockchain au dashboard.

package models

// AuthTimings contient les temps réels mesurés pour chaque étape d'authentification.
// ECDSAMs      : temps réel de vérification ECDSA en millisecondes
// ZKPMs        : temps réel de vérification ZKP Groth16 en millisecondes
// BlockchainMs : temps d'enregistrement de la transaction blockchain
// TotalMs      : temps total de toute la séquence d'authentification
type AuthTimings struct {
	ECDSAMs      int64 `json:"ecdsa_ms"`
	ZKPMs        int64 `json:"zkp_ms"`
	BlockchainMs int64 `json:"blockchain_ms"`
	TotalMs      int64 `json:"total_ms"`
}

// NodeMessageRequest est le payload envoyé par le simulateur IoT.
type NodeMessageRequest struct {
	DeviceID     string `json:"device_id"`
	Address      string `json:"address"`
	Message      string `json:"message"`
	MessageHash  string `json:"message_hash"`
	Signature    string `json:"signature"`
	Proof        string `json:"proof"`
	ExpectedHash string `json:"expected_hash"`
	Timestamp    int64  `json:"timestamp"`
	Sequence     uint64 `json:"sequence"`
	Nonce        string `json:"nonce"`
}

// NodeMessageResponse est la réponse du backend incluant les timings réels.
type NodeMessageResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	TxHash  string      `json:"txHash"`
	Block   uint64      `json:"block"`
	Timings AuthTimings `json:"timings"` // [FIX] Temps réels mesurés
}
