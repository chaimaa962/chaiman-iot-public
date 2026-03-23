package handlers

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
        "time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	"iot-backend/internal/blockchain"
	"iot-backend/internal/models"
	"iot-backend/internal/utils"
	"iot-backend/pkg/zkp"
)

var registeredDevices = []string{}

// ============================================
// FONCTIONS DE BASE
// ============================================

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":  "OK",
		"message": "IoT Backend opérationnel",
	}
	sendJSON(w, response, http.StatusOK)
}

// ============================================
// ENREGISTREMENT DES APPAREILS
// ============================================

func RegisterDevice(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()

	var device models.Device
	if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
		sendError(w, "Format JSON invalide", http.StatusBadRequest)
		return
	}

	device.Address = strings.ToLower(device.Address)
	device.PublicKey = strings.ToLower(device.PublicKey)
	
	logger.Info("📝 Enregistrement appareil", "address", device.Address[:15]+"...")

	if !common.IsHexAddress(device.Address) {
		sendError(w, "Adresse invalide", http.StatusBadRequest)
		return
	}

	deviceAddr := common.HexToAddress(device.Address)

	exists, err := blockchain.DeviceIsRegistered(deviceAddr)
	if err != nil {
		logger.Error("❌ Erreur vérification", "error", err)
	} else if exists {
		logger.Warn("⚠️ Appareil déjà enregistré", "address", device.Address[:15]+"...")
		sendError(w, "Appareil déjà enregistré", http.StatusConflict)
		return
	}

	txHash, err := blockchain.DeviceRegister(deviceAddr, device.PublicKey, device.Metadata)
	if err != nil {
		logger.Error("❌ Erreur enregistrement blockchain", "error", err)
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	registeredDevices = append(registeredDevices, device.Address)

	response := models.RegisterResponse{
		Status:  "success",
		Message: "✅ Appareil enregistré dans la blockchain",
		TxHash:  txHash.Hex(),
	}

	logger.Info("✅ Appareil enregistré", "address", device.Address[:15]+"...")
	sendJSON(w, response, http.StatusOK)
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

	info := models.DeviceInfo{
		PublicKey: publicKey,
		Metadata:  metadata,
		IsActive:  isActive,
		LastSeen:  lastSeen.Uint64(),
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
		publicKey, metadata, isActive, lastSeen, err := blockchain.DeviceGetInfo(addr)
		if err != nil {
			continue
		}
		
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
			"isGethNode": blockchain.IsGethNode(addr),
		})
	}
	
	logger.Info("📋 Appareils récupérés", "count", len(devices))
	sendJSON(w, devices, http.StatusOK)
}

// ============================================
// AUTHENTIFICATION ZKP
// ============================================

func AuthenticateWithZKP(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()

	var req zkp.ZKPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("❌ Erreur décodage JSON", "error", err)
		sendError(w, "Format JSON invalide", http.StatusBadRequest)
		return
	}

	req.DeviceAddress = strings.ToLower(req.DeviceAddress)
	req.PublicKey = strings.ToLower(req.PublicKey)
	req.Message = strings.ToLower(req.Message)
	req.Signature = strings.ToLower(req.Signature)
	
	safeAddr := req.DeviceAddress
	if len(safeAddr) > 10 {
		safeAddr = safeAddr[:10] + "..."
	}
	
	logger.Info("📥 Requête ZKP reçue", "device", safeAddr)

	if !common.IsHexAddress(req.DeviceAddress) {
		logger.Error("❌ Adresse invalide", "address", req.DeviceAddress)
		sendError(w, "Adresse invalide", http.StatusBadRequest)
		return
	}
	
	deviceAddr := common.HexToAddress(req.DeviceAddress)

	exists, err := blockchain.DeviceIsRegistered(deviceAddr)
	if err != nil || !exists {
		logger.Warn("⚠️ Appareil non enregistré", "adresse", req.DeviceAddress)
		sendError(w, "Appareil non enregistré", http.StatusNotFound)
		return
	}

	zkpService, err := zkp.GetZKPService()
	if err != nil {
		logger.Error("❌ Erreur service ZKP", "error", err)
		sendError(w, "Service ZKP indisponible", http.StatusInternalServerError)
		return
	}

	// Extraire le secret (simplifié)
	secret := big.NewInt(42)
	
	// Créer un défi (timestamp actuel par exemple)
	challenge := big.NewInt(time.Now().Unix())
	
	// 🔴 CORRECTION: GenerateProof retourne maintenant 3 valeurs
	proofBytes, expectedHash, err := zkpService.GenerateProof(secret, deviceAddr, challenge)
	if err != nil {
		logger.Error("❌ Erreur génération preuve", "error", err)
		sendError(w, "Erreur génération preuve", http.StatusInternalServerError)
		return
	}

	// 🔴 CORRECTION: VerifyProof prend maintenant 4 paramètres
	valid, err := zkpService.VerifyProof(proofBytes, deviceAddr, challenge, expectedHash)
	if err != nil {
		logger.Error("❌ Erreur vérification", "error", err)
		sendError(w, "Erreur vérification", http.StatusInternalServerError)
		return
	}

	var proofHash [32]byte
	if len(proofBytes) >= 32 {
		copy(proofHash[:], proofBytes[:32])
	}

	tx, _ := blockchain.AuthRecord(deviceAddr, valid, proofHash)
	_ = tx // Ignorer tx non utilisé

	if !valid {
		logger.Warn("❌ Preuve invalide", "device", safeAddr)
		response := zkp.ZKPResponse{
			Success: false,
			Message: "Authentification ZKP échouée",
		}
		sendJSON(w, response, http.StatusUnauthorized)
		return
	}

	logger.Info("🎉 Authentification réussie", "device", safeAddr)
	response := zkp.ZKPResponse{
		Success: true,
		Message: fmt.Sprintf("✅ Authentification ZKP réussie pour %s", safeAddr),
	}

	sendJSON(w, response, http.StatusOK)
}

// ============================================
// NŒUDS GETH
// ============================================

func GetGethNodes(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()
	
	nodes := blockchain.GetAllGethNodes()
	
	logger.Info("📋 Nœuds Geth récupérés", "count", len(nodes))
	sendJSON(w, nodes, http.StatusOK)
}

// ============================================
// FONCTIONS UTILITAIRES
// ============================================

func sendJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, message string, status int) {
	response := models.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	}
	sendJSON(w, response, status)
}
