// internal/handlers/blockchain_metrics.go
package handlers

import (
	"net/http"
	"time"

	"iot-backend/internal/blockchain"
	"iot-backend/internal/utils"
)

// HandleBlockchainMetrics retourne les métriques détaillées de la blockchain
func HandleBlockchainMetrics(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()

	// Récupérer le dernier block
	blockNumber, err := blockchain.GetBlockNumber()
	if err != nil {
		blockNumber = 0
	}

	// Récupérer les transactions récentes pour calculer les métriques
	transactions, err := blockchain.GetRecentTransactions(100)
	if err != nil {
		transactions = []blockchain.Transaction{}
	}

	// Calculer les métriques
	var blockTimes []float64
	var validationTimes []float64
	lastBlockTime := time.Now().Unix()

	// Si on a des transactions avec des blocks, on peut estimer les temps
	if len(transactions) > 1 {
		// Calcul approximatif du temps entre les blocks
		// (à améliorer avec des vraies données de la blockchain)
		blockTimes = append(blockTimes, 5.0) // Valeur par défaut
		validationTimes = append(validationTimes, 2.5)
	}

	metrics := map[string]interface{}{
		"blockTimes":       blockTimes,
		"validationTimes":  validationTimes,
		"lastBlockTime":    lastBlockTime,
		"averageBlockTime": 5.0, // Par défaut pour PoA
		"blockNumber":      blockNumber,
		"totalTx":          len(transactions),
		"timestamp":        time.Now().Unix(),
	}

	logger.Info("📊 Métriques blockchain", "block", blockNumber)
	sendJSON(w, metrics, http.StatusOK)
}

// HandleAIMetrics retourne les métriques IA (placeholder)
func HandleAIMetrics(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()

	// Pour l'instant, retourner des métriques vides
	// À implémenter quand le module IA sera ajouté
	metrics := map[string]interface{}{
		"detections":      []interface{}{},
		"accuracy":        0,
		"totalDetections": 0,
		"falsePositives":  0,
		"analysisTime":    0,
		"enabled":         false,
		"timestamp":       time.Now().Unix(),
	}

	logger.Info("🤖 Métriques IA (non configurées)")
	sendJSON(w, metrics, http.StatusOK)
}
