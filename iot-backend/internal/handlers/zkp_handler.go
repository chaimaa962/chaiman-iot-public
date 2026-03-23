package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"iot-backend/internal/utils"
	"iot-backend/pkg/zkp"
)

type ZKPGenerateRequest struct {
	Signature   string `json:"signature"`
	MessageHash string `json:"message_hash"`
	Address     string `json:"address"`
	Sequence    uint64 `json:"sequence"`
	Timestamp   int64  `json:"timestamp"`
	PrivateKey  string `json:"private_key"`
}

type ZKPGenerateResponse struct {
	Proof        string `json:"proof"`
	ExpectedHash string `json:"expected_hash"`
	Challenge    string `json:"challenge"`
}

// BN254 field modulus
var bn254Modulus, _ = new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)

func HandleZKPGenerate(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()

	var req ZKPGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Format JSON invalide", http.StatusBadRequest)
		return
	}

	// Normaliser
	req.Signature = strings.ToLower(req.Signature)
	req.MessageHash = strings.ToLower(req.MessageHash)
	req.Address = strings.ToLower(req.Address)

	logger.Info("📥 Génération preuve ZKP demandée", "address", req.Address[:15])

	// Récupérer le service ZKP
	zkpService, err := zkp.GetZKPService()
	if err != nil {
		logger.Error("❌ Erreur service ZKP", "error", err)
		sendError(w, "Service ZKP indisponible", http.StatusInternalServerError)
		return
	}

	// Créer le secret à partir de la signature (réduit modulo BN254)
	secret := hashToSecret(req.Signature, req.MessageHash)
	
	// Créer le défi (réduit modulo BN254)
	challenge := new(big.Int).SetInt64(req.Timestamp + int64(req.Sequence))
	challenge.Mod(challenge, bn254Modulus)

	// Adresse (déjà en format big.Int via common.Address)
	address := common.HexToAddress(req.Address)

	// Générer la preuve
	proofBytes, expectedHash, err := zkpService.GenerateProof(secret, address, challenge)
	if err != nil {
		logger.Error("❌ Erreur génération preuve", "error", err)
		sendError(w, "Erreur génération preuve: "+err.Error(), http.StatusInternalServerError)
		return
	}

	proofHex := hex.EncodeToString(proofBytes)
	expectedHashHex := hex.EncodeToString(expectedHash.Bytes())

	response := ZKPGenerateResponse{
		Proof:        proofHex,
		ExpectedHash: expectedHashHex,
		Challenge:    challenge.String(),
	}

	logger.Info("✅ Preuve ZKP générée", "taille", len(proofBytes), "hash", expectedHashHex[:20])
	sendJSON(w, response, http.StatusOK)
}

// hashToSecret crée un secret déterministe réduit modulo BN254
func hashToSecret(signature, messageHash string) *big.Int {
	data := signature + messageHash
	hash := sha256.Sum256([]byte(data))
	
	// Convertir en big.Int et réduire modulo BN254
	result := new(big.Int).SetBytes(hash[:])
	result.Mod(result, bn254Modulus)
	
	return result
}
