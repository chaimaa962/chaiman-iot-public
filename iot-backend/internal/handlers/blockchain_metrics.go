// internal/handlers/blockchain_metrics.go
// VERSION CORRIGÉE v2.0
//
// CORRECTIONS :
//   [FIX-1]  URL ML hardcodée "http://localhost:5001" remplacée par ML_SERVICE_URL
//            depuis les variables d'environnement (cohérence avec messages_secure.go).
//   [FIX-2]  Timeout explicite de 5 secondes sur le client HTTP (déjà présent, conservé).
//   [FIX-3]  En production (ENV=production), https:// obligatoire.
//   [FIX-4]  Réponse JSON structurée en cas d'erreur (déjà présent, conservé).
//   [FIX-5]  Body limité à 64KB (déjà présent, conservé).
//   [FIX-6]  Code HTTP upstream propagé (déjà présent, conservé).

package handlers

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"iot-backend/internal/blockchain"
	"iot-backend/internal/utils"
)

func HandleBlockchainMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":  "ok",
		"message": "Blockchain metrics endpoint",
	}
	json.NewEncoder(w).Encode(response)
}

// HandleAIMetrics proxy vers le service ML avec URL configurable.
// [FIX-1] ML_SERVICE_URL depuis env — plus de localhost hardcodé.
// [FIX-3] HTTPS obligatoire en production.
func HandleAIMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// [FIX-1] URL depuis variable d'environnement
	mlURL := os.Getenv("ML_SERVICE_URL")
	if mlURL == "" {
		mlURL = "http://localhost:5001"
	}

	// [FIX-3] Bloquer http:// en production
	if os.Getenv("ENV") == "production" && strings.HasPrefix(mlURL, "http://") {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "configuration_error",
			"message": fmt.Sprintf("ML_SERVICE_URL doit utiliser https:// en production (actuel: %s)", mlURL),
		})
		return
	}

	// [FIX-2] Client avec timeout explicite
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(mlURL + "/v4/health")
	if err != nil {
		// [FIX-4] Réponse JSON structurée
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "unavailable",
			"message": "Service AI inaccessible",
			"error":   err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// [FIX-5] Limiter la taille de la réponse
	limitedBody := io.LimitReader(resp.Body, 64*1024)
	body, err := io.ReadAll(limitedBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Erreur lecture réponse AI",
		})
		return
	}

	// [FIX-6] Propager le code HTTP upstream
	w.WriteHeader(resp.StatusCode)
	w.Write(body) //nolint:errcheck
}


// HandleGlobalStats retourne les statistiques globales du système.
func HandleGlobalStats(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()
	logger.Info("📊 Récupération stats globales")

	if blockchain.Contract == nil {
		sendError(w, "Contrat non initialisé", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stats, err := blockchain.Contract.GetGlobalStats(&bind.CallOpts{Context: ctx})
	if err != nil {
		logger.Error("❌ Erreur stats globales", "error", err)
		sendError(w, "Erreur blockchain", http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"totalDevices":         stats.TotalDevices.Uint64(),
		"totalAuthentications": stats.TotalAuthentications.Uint64(),
		"activeDevices":        stats.ActiveDevices.Uint64(),
		"inactiveDevices":      stats.InactiveDevices.Uint64(),
	}

	sendJSON(w, result, http.StatusOK)
}

// HandleDeviceHistory retourne l'historique d'authentification d'un device.
func HandleDeviceHistory(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()
	vars := mux.Vars(r)
	addressHex := vars["address"]

	if !common.IsHexAddress(addressHex) {
		sendError(w, "Adresse invalide", http.StatusBadRequest)
		return
	}

	deviceAddr := common.HexToAddress(addressHex)

	if blockchain.Contract == nil {
		sendError(w, "Contrat non initialisé", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	history, err := blockchain.Contract.GetAuthHistory(&bind.CallOpts{Context: ctx}, deviceAddr)
	if err != nil {
		logger.Error("❌ Erreur historique", "error", err)
		sendError(w, "Erreur blockchain", http.StatusInternalServerError)
		return
	}

	result := make([]map[string]interface{}, len(history))
	for i, record := range history {
		result[i] = map[string]interface{}{
			"timestamp": record.Timestamp.Uint64(),
			"success":   record.Success,
			"proofHash":  hex.EncodeToString(record.ProofHash[:]),
		}
	}

	sendJSON(w, result, http.StatusOK)
}

// HandleBlockchainInfo retourne les informations sur la blockchain.
func HandleBlockchainInfo(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()

	if blockchain.Client == nil {
		sendError(w, "Client blockchain non initialisé", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	blockNumber, err := blockchain.Client.BlockNumber(ctx)
	if err != nil {
		logger.Error("❌ Erreur block number", "error", err)
		sendError(w, "Erreur blockchain", http.StatusInternalServerError)
		return
	}

	chainID := blockchain.ChainID
	if chainID == nil {
		chainID = big.NewInt(0)
	}

	nodeAddress := ""
	if blockchain.SecureKey != nil {
		nodeAddress = blockchain.SecureKey.GetAddress().Hex()
	}

	result := map[string]interface{}{
		"blockNumber":     blockNumber,
		"chainID":         chainID.String(),
		"contractAddress": blockchain.ContractAddress.Hex(),
		"nodeAddress":     nodeAddress,
	}

	sendJSON(w, result, http.StatusOK)
}

// HandleTransactionByHash retourne les détails d'une transaction par son hash.
func HandleTransactionByHash(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()
	vars := mux.Vars(r)
	txHash := vars["hash"]

	if blockchain.Client == nil {
		sendError(w, "Client blockchain non initialisé", http.StatusServiceUnavailable)
		return
	}

	hash := common.HexToHash(txHash)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, isPending, err := blockchain.Client.TransactionByHash(ctx, hash)
	if err != nil {
		logger.Error("❌ Erreur transaction", "error", err)
		sendError(w, "Transaction non trouvée", http.StatusNotFound)
		return
	}

	receipt, err := blockchain.Client.TransactionReceipt(ctx, hash)
	var status string
	var blockNumber uint64
	var gasUsed uint64
	if err == nil && receipt != nil {
		status = map[bool]string{true: "success", false: "failed"}[receipt.Status == 1]
		blockNumber = receipt.BlockNumber.Uint64()
		gasUsed = receipt.GasUsed
	} else {
		status = "pending"
	}

	result := map[string]interface{}{
		"hash":        txHash,
		"from":        "",
		"to":          tx.To().Hex(),
		"value":       tx.Value().String(),
		"gas":         tx.Gas(),
		"gasPrice":    tx.GasPrice().String(),
		"nonce":       tx.Nonce(),
		"pending":     isPending,
		"status":      status,
		"blockNumber": blockNumber,
		"gasUsed":     gasUsed,
	}

	sendJSON(w, result, http.StatusOK)
}

// proxyToML est un helper générique qui proxifie une requête GET vers le service ML.
func proxyToML(w http.ResponseWriter, mlPath string, rawQuery string) {
	w.Header().Set("Content-Type", "application/json")

	mlURL := os.Getenv("ML_SERVICE_URL")
	if mlURL == "" {
		mlURL = "http://localhost:5001"
	}

	if os.Getenv("ENV") == "production" && strings.HasPrefix(mlURL, "http://") {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "configuration_error",
			"message": fmt.Sprintf("ML_SERVICE_URL doit utiliser https:// en production (actuel: %s)", mlURL),
		})
		return
	}

	targetURL := mlURL + mlPath
	if rawQuery != "" {
		targetURL += "?" + rawQuery
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(targetURL)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "unavailable",
			"message": "Service AI inaccessible",
			"error":   err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "message": "Erreur lecture réponse AI"})
		return
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(body) //nolint:errcheck
}

// HandleAIAnomalies proxifie /api/ai/anomalies → ML /v4/anomalies
func HandleAIAnomalies(w http.ResponseWriter, r *http.Request) {
	proxyToML(w, "/v4/anomalies", r.URL.RawQuery)
}

// HandleAIStats proxifie /api/ai/stats → ML /v4/stats (global)
func HandleAIStats(w http.ResponseWriter, r *http.Request) {
	proxyToML(w, "/v4/stats", r.URL.RawQuery)
}

// HandleAIActions proxifie /api/ai/actions → ML /v4/actions
func HandleAIActions(w http.ResponseWriter, r *http.Request) {
	proxyToML(w, "/v4/actions", r.URL.RawQuery)
}

// HandleAIDevices proxifie /api/ai/devices → ML /v4/devices
func HandleAIDevices(w http.ResponseWriter, r *http.Request) {
	proxyToML(w, "/v4/devices", r.URL.RawQuery)
}
