package models

type EntityType string

const (
    EntityIoTDevice EntityType = "iot_device"
    EntityValidator EntityType = "validator"
    EntityFullnode  EntityType = "fullnode"
)

type Device struct {
    Address    string     `json:"address"`
    EntityType EntityType `json:"entityType"`
    PublicKey  string     `json:"publicKey"`
    Metadata   string     `json:"metadata"`
}

// RegisterRequest — [MODIFIÉ] ajout de SchnorrPublicKey.
//
// SchnorrPublicKey est la clé publique Schnorr (Y = x·G) générée LOCALEMENT
// par le device (voir pkg/zkp/schnorr_device.go). Elle remplace la
// dérivation côté serveur (GenerateSecureSecretFromMaster) : le serveur ne
// fait que STOCKER cette valeur, il ne peut rien recalculer ni deviner.
//
// Format attendu : hex string, produit par SchnorrPublicKey.Hex().
type RegisterRequest struct {
    Address          string     `json:"address"`
    EntityType       EntityType `json:"entityType"`
    PublicKey        string     `json:"publicKey"`
    Metadata         string     `json:"metadata"`
    Signature        string     `json:"signature,omitempty"`
    MessageHash      string     `json:"message_hash,omitempty"`
    Proof            string     `json:"proof,omitempty"`
    ExpectedHash     string     `json:"expected_hash,omitempty"`
    Nonce            string     `json:"nonce,omitempty"`
    SchnorrPublicKey string     `json:"schnorrPublicKey,omitempty"` // [NOUVEAU]
}

type DeviceInfo struct {
    PublicKey  string     `json:"publicKey"`
    Metadata   string     `json:"metadata"`
    IsActive   bool       `json:"isActive"`
    LastSeen   uint64     `json:"lastSeen"`
    EntityType EntityType `json:"entityType"`
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
