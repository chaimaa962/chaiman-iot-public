// internal/handlers/device.go
// CHAIMAN IoT - Device Handler
// VERSION CORRIGÉE v5.5 — PREUVE DE POSSESSION OBLIGATOIRE À L'ENREGISTREMENT
//
// CORRECTIONS v5.4 → v5.5 :
//
//   [FIX-REGISTER-POP] RegisterDevice exigeait une signature ECDSA
//               UNIQUEMENT SI elle était fournie ("if req.Signature != "" && ...").
//               Un attaquant connaissant l'adresse publique d'un futur device
//               (adresse Ethereum, déductible/publique) pouvait donc
//               s'enregistrer EN PREMIER avec une clé publique Schnorr
//               fabriquée pour cette adresse, sans jamais prouver qu'il
//               contrôle la clé privée correspondante. Comme
//               RegisterSchnorrPublicKey() refuse d'écraser une clé Schnorr
//               déjà enregistrée pour une adresse (protection anti-vol),
//               ceci bloquait ensuite le VRAI device lors de son propre
//               enregistrement — un risque de squatting d'identité / déni
//               de service (pas un contournement direct de l'authentification,
//               puisque l'attaquant ne pourrait toujours pas passer
//               VerifyProof sans connaître le secret Schnorr réel — mais un
//               risque réel de blocage du device légitime).
//
//               CORRECTION : la signature ECDSA est désormais OBLIGATOIRE.
//               Le message signé N'EST PLUS fourni par le client (on ne lui
//               fait jamais confiance pour ça) mais RECALCULÉ côté serveur
//               de façon déterministe :
//                   "REGISTER:<address>:<schnorrPublicKey>"  (si clé Schnorr fournie)
//                   "REGISTER:<address>"                     (sinon, rétrocompat)
//               Cela lie explicitement la preuve de possession à LA clé
//               Schnorr soumise dans CETTE requête précise — impossible de
//               rejouer une signature d'un enregistrement précédent avec
//               une clé Schnorr différente.
//
//   Toutes les corrections v5.3/v5.4 sont conservées (FIX-NONCE-13/14,
//   FIX-NIL-1/2/3/4, FIX-ACTIVE-1/2/3, SCHNORR-1) — AUCUN changement sur
//   ActivateDevice/DeactivateDevice.

package handlers

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/mux"

	"iot-backend/internal/blockchain"
	"iot-backend/internal/models"
	"iot-backend/internal/utils"
	"iot-backend/pkg/zkp"
)

var (
	cleanRegex = regexp.MustCompile(`[\n\r\t]+`)

	registeredDevices   = []string{}
	registeredDevicesMu sync.RWMutex
)

// RegistrationRecord stocke un enregistrement d'appareil avec ses temps réels.
type RegistrationRecord struct {
	DeviceAddress  string `json:"deviceAddress"`
	DeviceID       string `json:"deviceId"`
	TxHash         string `json:"txHash"`
	BlockNumber    uint64 `json:"blockNumber"`
	TotalMs        int64  `json:"totalMs"`
	BlockchainMs   int64  `json:"blockchainMs"`
	ConfirmationMs int64  `json:"confirmationMs"`
	ECDSAMs        int64  `json:"ecdsaMs"`
	ZKPMs          int64  `json:"zkpMs"`
	Timestamp      int64  `json:"timestamp"`
}

type RegistrationResponse struct {
	Status         string `json:"status"`
	Message        string `json:"message"`
	TxHash         string `json:"txHash"`
	BlockNumber    uint64 `json:"blockNumber"`
	TotalMs        int64  `json:"totalMs"`
	BlockchainMs   int64  `json:"blockchainMs"`
	ConfirmationMs int64  `json:"confirmationMs"`
	ECDSAMs        int64  `json:"ecdsaMs"`
	ZKPMs          int64  `json:"zkpMs"`
}

var (
	registrationHistory   = make([]RegistrationRecord, 0, 200)
	registrationHistoryMu sync.RWMutex
	maxRegistrations      = 200
)

func StoreRegistrationRecord(record RegistrationRecord) {
	registrationHistoryMu.Lock()
	defer registrationHistoryMu.Unlock()

	registrationHistory = append([]RegistrationRecord{record}, registrationHistory...)
	if len(registrationHistory) > maxRegistrations {
		registrationHistory = registrationHistory[:maxRegistrations]
	}
}

func HandleRegistrationHistory(w http.ResponseWriter, r *http.Request) {
	registrationHistoryMu.RLock()
	defer registrationHistoryMu.RUnlock()

	history := make([]RegistrationRecord, len(registrationHistory))
	copy(history, registrationHistory)

	sendJSON(w, history, http.StatusOK)
}

func HandleRegistrationStats(w http.ResponseWriter, r *http.Request) {
	registrationHistoryMu.RLock()
	defer registrationHistoryMu.RUnlock()

	if len(registrationHistory) == 0 {
		sendJSON(w, map[string]interface{}{
			"count":             0,
			"avgTotalMs":        0,
			"avgBlockchainMs":   0,
			"avgConfirmationMs": 0,
			"avgECDSAMs":        0,
			"avgZKPMs":          0,
			"minTotalMs":        0,
			"maxTotalMs":        0,
		}, http.StatusOK)
		return
	}

	var totalTotalMs, totalBlockchainMs, totalConfirmationMs, totalECDSAMs, totalZKPMs int64
	var minTotalMs, maxTotalMs int64 = 1<<63 - 1, 0

	for _, reg := range registrationHistory {
		totalTotalMs += reg.TotalMs
		totalBlockchainMs += reg.BlockchainMs
		totalConfirmationMs += reg.ConfirmationMs
		totalECDSAMs += reg.ECDSAMs
		totalZKPMs += reg.ZKPMs

		if reg.TotalMs < minTotalMs {
			minTotalMs = reg.TotalMs
		}
		if reg.TotalMs > maxTotalMs {
			maxTotalMs = reg.TotalMs
		}
	}

	count := int64(len(registrationHistory))
	stats := map[string]interface{}{
		"count":             count,
		"avgTotalMs":        totalTotalMs / count,
		"avgBlockchainMs":   totalBlockchainMs / count,
		"avgConfirmationMs": totalConfirmationMs / count,
		"avgECDSAMs":        totalECDSAMs / count,
		"avgZKPMs":          totalZKPMs / count,
		"minTotalMs":        minTotalMs,
		"maxTotalMs":        maxTotalMs,
		"latest":            registrationHistory[0],
	}

	sendJSON(w, stats, http.StatusOK)
}

func validateMetadata(metadata string) error {
	if len(metadata) > 10000 {
		return fmt.Errorf("métadonnées trop longues (max 10000 caractères)")
	}

	for _, r := range metadata {
		if unicode.IsControl(r) && r != '\n' && r != '\t' && r != '\r' {
			return fmt.Errorf("caractères de contrôle non autorisés dans les métadonnées")
		}
	}

	if metadata != "" {
		var js map[string]interface{}
		if err := json.Unmarshal([]byte(metadata), &js); err != nil {
			return nil
		}
	}

	return nil
}

func validatePublicKey(pubKey string) error {
	if pubKey == "" {
		return fmt.Errorf("clé publique requise")
	}
	if len(pubKey) > 1000 {
		return fmt.Errorf("clé publique trop longue")
	}
	pubKey = strings.TrimPrefix(pubKey, "0x")
	if _, err := hex.DecodeString(pubKey); err != nil {
		return fmt.Errorf("clé publique invalide (doit être hexadécimale)")
	}
	return nil
}

// validateSchnorrPublicKey — valide le format de la clé Schnorr envoyée par
// le device, si présente. Ne bloque PAS l'enregistrement si absente
// (rétrocompatibilité), mais bloque si présente et malformée.
func validateSchnorrPublicKey(schnorrPubKeyHex string) error {
	if schnorrPubKeyHex == "" {
		return nil // absent = OK, migration progressive
	}
	if _, err := zkp.PublicKeyFromHex(schnorrPubKeyHex); err != nil {
		return fmt.Errorf("clé publique Schnorr invalide: %v", err)
	}
	return nil
}

// buildRegistrationProofMessage — [FIX-REGISTER-POP] construit le message
// canonique dont la signature ECDSA doit prouver la possession de l'adresse
// (et, si fournie, la légitimité de la clé Schnorr associée à CETTE requête).
// Calculé côté serveur uniquement — jamais fait confiance à une valeur
// envoyée par le client pour cette étape.
func buildRegistrationProofMessage(address, schnorrPublicKeyHex string) string {
	if schnorrPublicKeyHex != "" {
		return fmt.Sprintf("REGISTER:%s:%s", address, schnorrPublicKeyHex)
	}
	return fmt.Sprintf("REGISTER:%s", address)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":  "OK",
		"message": "IoT Backend opérationnel",
	}
	sendJSON(w, response, http.StatusOK)
}

// RegisterDevice enregistre un nouvel appareil IoT.
//
// [FIX-REGISTER-POP] La preuve de possession ECDSA (Signature + adresse
// contrôlée) est désormais OBLIGATOIRE, pas optionnelle. Le device doit
// signer un message canonique reconstruit côté serveur
// ("REGISTER:<address>:<schnorrPublicKey>") avec sa clé privée ECDSA AVANT
// que l'enregistrement blockchain ou l'enregistrement de la clé Schnorr
// n'ait lieu. Ceci ferme le risque de squatting d'adresse identifié lors
// de l'audit : sans cette preuve, quiconque connaissant une adresse pouvait
// soumettre une clé Schnorr arbitraire pour celle-ci avant le vrai device.
func RegisterDevice(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()

	startTotal := time.Now()

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Format JSON invalide", http.StatusBadRequest)
		return
	}

	req.Address = strings.ToLower(req.Address)
	req.PublicKey = strings.ToLower(req.PublicKey)

	if len(req.Address) < 15 {
		sendError(w, "Adresse invalide", http.StatusBadRequest)
		return
	}
	logger.Info("📝 Enregistrement appareil IoT", "address", req.Address[:15]+"...")

	if !common.IsHexAddress(req.Address) {
		sendError(w, "Adresse invalide", http.StatusBadRequest)
		return
	}

	if err := validatePublicKey(req.PublicKey); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validation du format de la clé Schnorr si fournie.
	if err := validateSchnorrPublicKey(req.SchnorrPublicKey); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validateMetadata(req.Metadata); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	deviceAddr := common.HexToAddress(req.Address)

	exists, err := blockchain.DeviceIsRegistered(deviceAddr)
	if err != nil {
		logger.Error("❌ Erreur vérification", "error", err)
	} else if exists {
		logger.Warn("⚠️ Appareil déjà enregistré", "address", req.Address[:15]+"...")
		sendError(w, "Appareil déjà enregistré", http.StatusConflict)
		return
	}

	// ── [FIX-REGISTER-POP] Preuve de possession OBLIGATOIRE ─────────────────
	//
	// AVANT (v5.4) : "if req.Signature != "" && req.MessageHash != "" { ... }"
	// — la vérification était SAUTÉE si le client n'envoyait rien. Un
	// attaquant pouvait donc s'enregistrer sans jamais prouver qu'il
	// contrôle l'adresse.
	//
	// APRÈS : la signature est requise, et le message qu'elle doit couvrir
	// est calculé ICI (jamais depuis req.MessageHash fourni par le client),
	// pour empêcher toute manipulation du contenu signé.
	if req.Signature == "" {
		logger.Warn("❌ Signature ECDSA manquante (preuve de possession requise)",
			"address", req.Address[:15]+"...")
		sendError(w, "Signature ECDSA requise pour prouver la possession de l'adresse", http.StatusBadRequest)
		return
	}

	registrationMsg := buildRegistrationProofMessage(req.Address, req.SchnorrPublicKey)
	expectedHash := crypto.Keccak256Hash([]byte(registrationMsg))
	expectedHashHex := hex.EncodeToString(expectedHash.Bytes())

	tECDSA := time.Now()
	validECDSA, err := zkp.VerifyECDSASignature(deviceAddr, expectedHashHex, req.Signature)
	ecdsaMs := time.Since(tECDSA).Milliseconds()

	if err != nil || !validECDSA {
		logger.Warn("❌ Signature ECDSA invalide — preuve de possession échouée",
			"address", req.Address[:15]+"...")
		sendError(w, "Signature ECDSA invalide — preuve de possession de l'adresse échouée", http.StatusUnauthorized)
		return
	}
	logger.Info("✅ Preuve de possession ECDSA valide", "ecdsa_ms", ecdsaMs)

	// Aucune preuve ZKP requise à l'enregistrement lui-même — Schnorr ne
	// nécessite QUE la clé publique (aucun secret n'est jamais transmis).
	// La preuve de connaissance du secret Schnorr sera exigée à chaque
	// authentification ultérieure (HandleSecureNodeMessage), pas ici.
	var zkpMs int64

	tBlockchain := time.Now()
	tx, err := blockchain.DeviceRegisterWithType(
		deviceAddr,
		req.PublicKey,
		req.Metadata,
		blockchain.EntityIoTDevice,
	)
	blockchainMs := time.Since(tBlockchain).Milliseconds()

	if err != nil {
		logger.Error("❌ Erreur enregistrement blockchain", "error", err, "blockchain_ms", blockchainMs)
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if tx == nil {
		logger.Error("❌ Transaction nil retournée par DeviceRegisterWithType")
		sendError(w, "Transaction blockchain invalide", http.StatusInternalServerError)
		return
	}

	txHash := tx.Hash()
	logger.Info("📡 Transaction envoyée", "txHash", txHash.Hex()[:20], "blockchain_ms", blockchainMs)

	// Enregistrement de la clé publique Schnorr — APRÈS confirmation
	// blockchain ET après vérification de la preuve de possession ci-dessus.
	// Non bloquant pour la réponse HTTP : si ça échoue, on logge mais on ne
	// fait pas échouer toute la requête (la tx blockchain est déjà soumise
	// et irréversible à ce stade).
	if req.SchnorrPublicKey != "" {
		if err := RegisterSchnorrPublicKey(req.Address, req.SchnorrPublicKey); err != nil {
			logger.Error("⚠️ Enregistrement clé Schnorr échoué (device blockchain OK)",
				"address", req.Address[:15]+"...", "error", err)
		} else {
			logger.Info("🔑 Clé publique Schnorr enregistrée", "address", req.Address[:15]+"...")
		}
	} else {
		logger.Warn("⚠️ Aucune clé Schnorr fournie — device non migré (ancien protocole ZKP utilisé)",
			"address", req.Address[:15]+"...")
	}

	var confirmationMs int64
	var blockNumber uint64

	if os.Getenv("WAIT_FOR_CONFIRMATION") == "true" {
		tConfirm := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		receipt, err := blockchain.Client.TransactionReceipt(ctx, txHash)
		confirmationMs = time.Since(tConfirm).Milliseconds()

		if err == nil && receipt != nil {
			blockNumber = receipt.BlockNumber.Uint64()
			logger.Info("✅ Transaction confirmée", "block", blockNumber, "confirmation_ms", confirmationMs)
		} else {
			logger.Warn("⚠️ Transaction non confirmée dans le délai", "error", err)
		}
	}

	totalMs := time.Since(startTotal).Milliseconds()

	record := RegistrationRecord{
		DeviceAddress:  req.Address,
		DeviceID:       extractDeviceIDFromMetadata(req.Metadata),
		TxHash:         txHash.Hex(),
		BlockNumber:    blockNumber,
		TotalMs:        totalMs,
		BlockchainMs:   blockchainMs,
		ConfirmationMs: confirmationMs,
		ECDSAMs:        ecdsaMs,
		ZKPMs:          zkpMs,
		Timestamp:      time.Now().Unix(),
	}
	StoreRegistrationRecord(record)

	registeredDevicesMu.Lock()
	registeredDevices = append(registeredDevices, req.Address)
	registeredDevicesMu.Unlock()

	response := RegistrationResponse{
		Status:         "success",
		Message:        "✅ Appareil IoT enregistré dans la blockchain",
		TxHash:         txHash.Hex(),
		BlockNumber:    blockNumber,
		TotalMs:        totalMs,
		BlockchainMs:   blockchainMs,
		ConfirmationMs: confirmationMs,
		ECDSAMs:        ecdsaMs,
		ZKPMs:          zkpMs,
	}

	logger.Info("✅ Appareil IoT enregistré",
		"address", req.Address[:15]+"...",
		"ecdsa_ms", ecdsaMs,
		"zkp_ms", zkpMs,
		"blockchain_ms", blockchainMs,
		"confirmation_ms", confirmationMs,
		"total_ms", totalMs,
	)

	sendJSON(w, response, http.StatusOK)
}

func extractDeviceIDFromMetadata(metadata string) string {
	if metadata == "" {
		return ""
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(metadata), &data); err != nil {
		return ""
	}
	if id, ok := data["device_id"].(string); ok {
		return id
	}
	if name, ok := data["name"].(string); ok {
		return name
	}
	return ""
}

func GetDeviceInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	addressHex := vars["address"]

	if !common.IsHexAddress(addressHex) {
		sendError(w, "Adresse invalide", http.StatusBadRequest)
		return
	}

	deviceAddr := common.HexToAddress(addressHex)

	publicKey, metadata, isActive, lastSeen, err := blockchain.DeviceGetInfo(deviceAddr)
	if err != nil {
		sendError(w, "Appareil non trouvé", http.StatusNotFound)
		return
	}

	publicKey = strings.TrimSpace(publicKey)
	metadata = cleanRegex.ReplaceAllString(metadata, " ")
	metadata = strings.TrimSpace(metadata)

	info := models.DeviceInfo{
		PublicKey:  publicKey,
		Metadata:   metadata,
		IsActive:   isActive,
		LastSeen:   lastSeen.Uint64(),
		EntityType: models.EntityIoTDevice,
	}

	sendJSON(w, info, http.StatusOK)
}

func GetAllDevices(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()

	addresses, err := blockchain.Contract.GetAllDevices(&bind.CallOpts{})
	if err != nil {
		logger.Error("❌ Erreur lecture liste blockchain", "error", err)
		sendJSON(w, []interface{}{}, http.StatusOK)
		return
	}

	devices := []map[string]interface{}{}

	for _, addr := range addresses {
		if blockchain.IsProtectedNode(addr) {
			continue
		}

		publicKey, metadata, isActive, lastSeen, err := blockchain.DeviceGetInfo(addr)
		if err != nil {
			continue
		}

		publicKey = strings.TrimSpace(publicKey)
		metadata = cleanRegex.ReplaceAllString(metadata, " ")
		metadata = strings.ReplaceAll(metadata, "  ", " ")
		metadata = strings.TrimSpace(metadata)

		authCountBig, _ := blockchain.StatsGetDeviceAuthCount(addr)
		authCount := uint64(0)
		if authCountBig != nil {
			authCount = authCountBig.Uint64()
		}

		var lastSeenUint64 uint64
		if lastSeen != nil {
			lastSeenUint64 = lastSeen.Uint64()
		}

		devices = append(devices, map[string]interface{}{
			"address":    addr.Hex(),
			"publicKey":  publicKey,
			"metadata":   metadata,
			"isActive":   isActive,
			"lastSeen":   lastSeenUint64,
			"authCount":  authCount,
			"isGethNode": false,
			"entityType": "iot_device",
		})
	}

	logger.Info("📋 Appareils IoT récupérés", "count", len(devices))
	sendJSON(w, devices, http.StatusOK)
}

func GetGethNodes(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()

	nodes := blockchain.GetAllGethNodes()

	for _, node := range nodes {
		if meta, ok := node["metadata"].(string); ok {
			node["metadata"] = cleanRegex.ReplaceAllString(meta, " ")
			node["metadata"] = strings.ReplaceAll(node["metadata"].(string), "  ", " ")
			node["metadata"] = strings.TrimSpace(node["metadata"].(string))
		}
		if pk, ok := node["publicKey"].(string); ok {
			node["publicKey"] = strings.TrimSpace(pk)
		}
	}

	logger.Info("📋 Nœuds Geth récupérés", "count", len(nodes))
	sendJSON(w, nodes, http.StatusOK)
}

func cleanAllStrings(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			result[key] = cleanAllStrings(val)
		}
		return result
	case []interface{}:
		for i, val := range v {
			v[i] = cleanAllStrings(val)
		}
		return v
	case []map[string]interface{}:
		for i, val := range v {
			v[i] = cleanAllStrings(val).(map[string]interface{})
		}
		return v
	case string:
		cleaned := strings.ReplaceAll(v, "\r", "")
		cleaned = strings.Map(func(r rune) rune {
			if r < 32 && r != '\n' && r != '\t' {
				return -1
			}
			return r
		}, cleaned)
		return strings.TrimSpace(cleaned)
	default:
		return data
	}
}

// DeactivateDevice désactive un appareil IoT. [INCHANGÉ v5.3 — voir original]
func DeactivateDevice(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()

	startTotal := time.Now()

	var req struct {
		DeviceAddress string `json:"deviceAddress"`
		Reason        string `json:"reason,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Format JSON invalide", http.StatusBadRequest)
		return
	}
	req.DeviceAddress = strings.ToLower(req.DeviceAddress)
	if !common.IsHexAddress(req.DeviceAddress) {
		sendError(w, "Adresse invalide", http.StatusBadRequest)
		return
	}
	deviceAddr := common.HexToAddress(req.DeviceAddress)

	exists, err := blockchain.DeviceIsRegistered(deviceAddr)
	if err != nil || !exists {
		sendError(w, "Appareil non trouvé", http.StatusNotFound)
		return
	}

	isActive, err := blockchain.IsDeviceActive(deviceAddr)
	if err != nil {
		logger.Warn("⚠️ Erreur vérification état actif", "error", err)
	} else if !isActive {
		sendJSON(w, map[string]interface{}{
			"success":      true,
			"txHash":       "",
			"message":      "Appareil déjà inactif",
			"blockchainMs": 0,
			"totalMs":      time.Since(startTotal).Milliseconds(),
		}, http.StatusOK)
		return
	}

	if blockchain.IsProtectedNode(deviceAddr) {
		sendError(w, "Nœud protégé - désactivation impossible", http.StatusForbidden)
		return
	}

	blockchain.TxSemaphore <- struct{}{}
	defer func() { <-blockchain.TxSemaphore }()

	tBlockchain := time.Now()
	auth, err := blockchain.GetTransactionAuth()

	if err != nil || auth == nil {
		logger.Error("❌ Erreur authentification blockchain (auth nil ou erreur)",
			"error", err,
			"auth_nil", auth == nil,
		)
		sendError(w, "Erreur blockchain: impossible d'obtenir les options de transaction", http.StatusInternalServerError)
		return
	}

	tx, err := blockchain.Contract.DeactivateDevice(auth, deviceAddr)
	blockchainMs := time.Since(tBlockchain).Milliseconds()

	if err != nil {
		errMsg := err.Error()

		if auth.Nonce != nil {
			blockchain.ReleaseNonceSync(auth.Nonce.Uint64())
		}

		if strings.Contains(errMsg, "deja inactif") ||
			strings.Contains(errMsg, "already inactive") ||
			strings.Contains(errMsg, "Appareil deja inactif") {
			sendJSON(w, map[string]interface{}{
				"success":      true,
				"txHash":       "",
				"message":      "Appareil déjà inactif",
				"blockchainMs": blockchainMs,
				"totalMs":      time.Since(startTotal).Milliseconds(),
			}, http.StatusOK)
			return
		}

		logger.Error("❌ Erreur désactivation", "error", err, "blockchain_ms", blockchainMs)
		sendError(w, "Erreur désactivation: "+errMsg, http.StatusInternalServerError)
		return
	}

	if tx == nil {
		logger.Error("❌ Transaction nil retournée par DeactivateDevice")
		sendError(w, "Transaction blockchain invalide", http.StatusInternalServerError)
		return
	}

	totalMs := time.Since(startTotal).Milliseconds()

	go func() {
		time.Sleep(5 * time.Second)
		receipt, _ := blockchain.Client.TransactionReceipt(r.Context(), tx.Hash())
		if receipt != nil && receipt.Status == 1 {
			addrShort := req.DeviceAddress
			if len(addrShort) > 15 {
				addrShort = addrShort[:15]
			}
			logger.Info("✅ Désactivation confirmée",
				"address", addrShort,
				"txHash", tx.Hash().Hex()[:20])
		}
	}()

	addrShort := req.DeviceAddress
	if len(addrShort) > 15 {
		addrShort = addrShort[:15]
	}
	logger.Warn("🔒 Appareil désactivé",
		"address", addrShort,
		"reason", req.Reason,
		"blockchain_ms", blockchainMs,
		"total_ms", totalMs,
	)

	sendJSON(w, map[string]interface{}{
		"success":      true,
		"txHash":       tx.Hash().Hex(),
		"message":      "Appareil désactivé avec succès",
		"blockchainMs": blockchainMs,
		"totalMs":      totalMs,
	}, http.StatusOK)
}

// ActivateDevice réactive un appareil IoT. [INCHANGÉ v5.3 — voir original]
func ActivateDevice(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()

	startTotal := time.Now()

	var req struct {
		DeviceAddress string `json:"deviceAddress"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Format JSON invalide", http.StatusBadRequest)
		return
	}
	req.DeviceAddress = strings.ToLower(req.DeviceAddress)
	if !common.IsHexAddress(req.DeviceAddress) {
		sendError(w, "Adresse invalide", http.StatusBadRequest)
		return
	}
	deviceAddr := common.HexToAddress(req.DeviceAddress)

	exists, err := blockchain.DeviceIsRegistered(deviceAddr)
	if err != nil || !exists {
		sendError(w, "Appareil non trouvé", http.StatusNotFound)
		return
	}

	blockchain.TxSemaphore <- struct{}{}
	defer func() { <-blockchain.TxSemaphore }()

	tBlockchain := time.Now()
	auth, err := blockchain.GetTransactionAuth()

	if err != nil || auth == nil {
		logger.Error("❌ Erreur authentification blockchain (auth nil ou erreur)",
			"error", err,
			"auth_nil", auth == nil,
		)
		sendError(w, "Erreur blockchain: impossible d'obtenir les options de transaction", http.StatusInternalServerError)
		return
	}

	tx, err := blockchain.Contract.ActivateDevice(auth, deviceAddr)
	blockchainMs := time.Since(tBlockchain).Milliseconds()

	if err != nil {
		errMsg := err.Error()

		if auth.Nonce != nil {
			blockchain.ReleaseNonceSync(auth.Nonce.Uint64())
		}

		if strings.Contains(errMsg, "deja actif") ||
			strings.Contains(errMsg, "already active") ||
			strings.Contains(errMsg, "Appareil deja actif") {
			addrShort := req.DeviceAddress
			if len(addrShort) > 15 {
				addrShort = addrShort[:15]
			}
			logger.Info("ℹ️  Appareil déjà actif (revert ignoré)",
				"address", addrShort,
				"blockchain_ms", blockchainMs,
			)
			sendJSON(w, map[string]interface{}{
				"success":      true,
				"txHash":       "",
				"message":      "Appareil déjà actif",
				"blockchainMs": blockchainMs,
				"totalMs":      time.Since(startTotal).Milliseconds(),
			}, http.StatusOK)
			return
		}

		logger.Error("❌ Erreur réactivation", "error", err, "blockchain_ms", blockchainMs)
		sendError(w, "Erreur réactivation: "+errMsg, http.StatusInternalServerError)
		return
	}

	if tx == nil {
		logger.Error("❌ Transaction nil retournée par ActivateDevice")
		sendError(w, "Transaction blockchain invalide", http.StatusInternalServerError)
		return
	}

	totalMs := time.Since(startTotal).Milliseconds()

	addrShort := req.DeviceAddress
	if len(addrShort) > 15 {
		addrShort = addrShort[:15]
	}
	logger.Info("🔓 Appareil réactivé",
		"address", addrShort,
		"txHash", tx.Hash().Hex()[:20],
		"blockchain_ms", blockchainMs,
		"total_ms", totalMs,
	)

	sendJSON(w, map[string]interface{}{
		"success":      true,
		"txHash":       tx.Hash().Hex(),
		"message":      "Appareil réactivé avec succès",
		"blockchainMs": blockchainMs,
		"totalMs":      totalMs,
	}, http.StatusOK)
}

// GetDeviceStatus retourne le statut d'un appareil. [INCHANGÉ]
func GetDeviceStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	addressHex := vars["address"]

	if !common.IsHexAddress(addressHex) {
		sendError(w, "Adresse invalide", http.StatusBadRequest)
		return
	}

	deviceAddr := common.HexToAddress(addressHex)

	exists, err := blockchain.DeviceIsRegistered(deviceAddr)
	if err != nil || !exists {
		sendError(w, "Appareil non trouvé", http.StatusNotFound)
		return
	}

	publicKey, metadata, isActive, lastSeen, err := blockchain.DeviceGetInfo(deviceAddr)
	if err != nil {
		sendError(w, "Erreur récupération infos", http.StatusInternalServerError)
		return
	}

	authCount, _ := blockchain.StatsGetDeviceAuthCount(deviceAddr)
	isGethNode := blockchain.IsProtectedNode(deviceAddr)

	var lastSeenVal uint64
	if lastSeen != nil {
		lastSeenVal = lastSeen.Uint64()
	}
	var authCountVal uint64
	if authCount != nil {
		authCountVal = authCount.Uint64()
	}

	response := map[string]interface{}{
		"address":    deviceAddr.Hex(),
		"publicKey":  strings.TrimSpace(publicKey),
		"metadata":   strings.TrimSpace(metadata),
		"isActive":   isActive,
		"lastSeen":   lastSeenVal,
		"authCount":  authCountVal,
		"isGethNode": isGethNode,
		"entityType": "iot_device",
		"status":     map[bool]string{true: "active", false: "inactive"}[isActive],
	}

	sendJSON(w, response, http.StatusOK)
}

func GetRegistrationHistoryByDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	addressHex := vars["address"]

	addressHex = strings.ToLower(addressHex)

	registrationHistoryMu.RLock()
	defer registrationHistoryMu.RUnlock()

	var deviceRegistrations []RegistrationRecord
	for _, reg := range registrationHistory {
		if strings.EqualFold(reg.DeviceAddress, addressHex) {
			deviceRegistrations = append(deviceRegistrations, reg)
		}
	}

	sendJSON(w, deviceRegistrations, http.StatusOK)
}

// NOTE : AuthenticateWithZKP (ancien handler Groth16, /api/auth/zkp) est
// VOLONTAIREMENT retiré de cette version — il dépendait de
// zkp.GenerateSecureSecretFromMaster() qui n'existe plus dans le nouveau
// modèle Schnorr. Si vous avez besoin d'une route de test manuel équivalente,
// utilisez /api/node/message-secure (HandleSecureNodeMessage, messages_secure.go)
// qui intègre désormais la vérification Schnorr en conditions réelles.
