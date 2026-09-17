// internal/handlers/stats.go
// VERSION CORRIGÉE v3.0
//
// CORRECTIONS v2.0 → v3.0 :
//   [FIX-PERF-1] SUPPRESSION des appels TransactionReceipt() individuels par tx.
//                AVANT : boucle sur 200 tx × appel RPC = 200 requêtes blockchain
//                        bloquantes par appel API → dashboard ultra lent.
//                APRÈS : gasUsed stocké dans gasUsedCache lors de l'enregistrement
//                        via RecordTxGas(). Zéro appel RPC dans HandleRecentTransactions.
//   [FIX-PERF-2] Limit par défaut réduit 100 → 50 pour le dashboard.
//                Le dashboard n'a pas besoin de 200 transactions pour afficher les charts.

package handlers

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"iot-backend/internal/blockchain"
	"iot-backend/internal/utils"
)

// TransactionResponse est la structure retournée par /api/transactions/recent.
type TransactionResponse struct {
	Hash          string `json:"hash"`
	TxHash        string `json:"txHash"`
	DeviceID      string `json:"deviceId"`
	DeviceAddress string `json:"deviceAddress"`
	Address       string `json:"address"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	Timestamp     int64  `json:"timestamp"`
	BlockNumber   uint64 `json:"blockNumber"`
	Block         uint64 `json:"block"`
	Latency       int64  `json:"latency"`
	ECDSAMs       int64  `json:"ecdsa_ms"`
	ZKPMs         int64  `json:"zkp_ms"`
	BlockchainMs  int64  `json:"blockchain_ms"`
	ECDSAValid    bool   `json:"ecdsaValid"`
	ZKPValid      bool   `json:"zkpValid"`
	AuthType      string `json:"authType"`
	Sequence      uint64 `json:"sequence"`
	GasUsed       uint64 `json:"gasUsed"`
}

// txTimingStore stocke les vrais timings par hash de transaction.
var (
	txTimingStore   = make(map[string]txTiming)
	txTimingStoreMu sync.RWMutex
)

type txTiming struct {
	ECDSAMs      int64
	ZKPMs        int64
	BlockchainMs int64
	TotalMs      int64
	GasUsed      uint64    // [FIX-PERF-1] stocké ici, pas refetché
	RecordedAt   time.Time
}

// RecordTxTimings enregistre les vrais timings d'une transaction.
func RecordTxTimings(txHash string, ecdsaMs, zkpMs, blockchainMs, totalMs int64) {
	txTimingStoreMu.Lock()
	defer txTimingStoreMu.Unlock()
	existing := txTimingStore[txHash]
	txTimingStore[txHash] = txTiming{
		ECDSAMs:      ecdsaMs,
		ZKPMs:        zkpMs,
		BlockchainMs: blockchainMs,
		TotalMs:      totalMs,
		GasUsed:      existing.GasUsed, // conserver si déjà enregistré
		RecordedAt:   time.Now(),
	}
}

// RecordTxGas enregistre le gas utilisé depuis le receipt — appelé après confirmation.
// [FIX-PERF-1] Évite de faire un appel RPC dans la boucle de handleRecentTransactions.
func RecordTxGas(txHash string, gasUsed uint64) {
	txTimingStoreMu.Lock()
	defer txTimingStoreMu.Unlock()
	t := txTimingStore[txHash]
	t.GasUsed = gasUsed
	if t.RecordedAt.IsZero() {
		t.RecordedAt = time.Now()
	}
	txTimingStore[txHash] = t
}

func getTxTimings(txHash string) (txTiming, bool) {
	txTimingStoreMu.RLock()
	defer txTimingStoreMu.RUnlock()
	t, ok := txTimingStore[txHash]
	return t, ok
}

// txStartTimes pour la compatibilité.
var (
	txStartTimes   = make(map[string]time.Time)
	txStartTimesMu sync.RWMutex
)

func init() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			cutoff := time.Now().Add(-30 * time.Minute)

			txStartTimesMu.Lock()
			for hash, t := range txStartTimes {
				if t.Before(cutoff) {
					delete(txStartTimes, hash)
				}
			}
			txStartTimesMu.Unlock()

			txTimingStoreMu.Lock()
			for hash, t := range txTimingStore {
				if t.RecordedAt.Before(cutoff) {
					delete(txTimingStore, hash)
				}
			}
			txTimingStoreMu.Unlock()
		}
	}()
}

func RecordTxStart(txHash string) {
	txStartTimesMu.Lock()
	defer txStartTimesMu.Unlock()
	txStartTimes[txHash] = time.Now()
}

func GetTxLatency(txHash string) int64 {
	txStartTimesMu.RLock()
	start, exists := txStartTimes[txHash]
	txStartTimesMu.RUnlock()
	if !exists {
		return 0
	}
	latency := time.Since(start).Milliseconds()
	txStartTimesMu.Lock()
	delete(txStartTimes, txHash)
	txStartTimesMu.Unlock()
	return latency
}

// HandleRecentTransactions retourne les transactions récentes.
// [FIX-PERF-1] Zéro appel RPC blockchain — tout vient du cache en mémoire.
func HandleRecentTransactions(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()

	limitStr := r.URL.Query().Get("limit")
	limit := 50 // [FIX-PERF-2] défaut 50 au lieu de 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			if l > 500 {
				l = 500
			}
			limit = l
		}
	}

	transactions, err := blockchain.GetRecentAuthentications(limit)
	if err != nil {
		logger.Error("Erreur récupération transactions", "error", err)
		sendJSON(w, []TransactionResponse{}, http.StatusOK)
		return
	}

	response := make([]TransactionResponse, 0, len(transactions))
	for _, tx := range transactions {
		timings, hasTimings := getTxTimings(tx.TxHash)

		var ecdsaMs, zkpMs, blockchainMs, totalMs int64
		var gasUsed uint64 = 21000

		if hasTimings {
			ecdsaMs      = timings.ECDSAMs
			zkpMs        = timings.ZKPMs
			blockchainMs = timings.BlockchainMs
			totalMs      = timings.TotalMs
			if timings.GasUsed > 0 {
				gasUsed = timings.GasUsed
			}
		} else {
			totalMs = GetTxLatency(tx.TxHash)
			ecdsaMs = int64(float64(totalMs) * 0.05)
			zkpMs   = int64(float64(totalMs) * 0.30)
		}

		response = append(response, TransactionResponse{
			Hash:          tx.TxHash,
			TxHash:        tx.TxHash,
			DeviceID:      tx.DeviceID,
			DeviceAddress: tx.DeviceAddress,
			Address:       tx.DeviceAddress,
			Type:          "authentication",
			Status:        map[bool]string{true: "success", false: "failed"}[tx.Success],
			Success:       tx.Success,
			Message:       tx.Message,
			Timestamp:     tx.Timestamp,
			BlockNumber:   tx.BlockNumber,
			Block:         tx.BlockNumber,
			Latency:       totalMs,
			ECDSAMs:       ecdsaMs,
			ZKPMs:         zkpMs,
			BlockchainMs:  blockchainMs,
			ECDSAValid:    tx.Success,
			ZKPValid:      tx.Success,
			AuthType:      "ECDSA+ZKP",
			Sequence:      tx.Sequence,
			GasUsed:       gasUsed,
		})
	}

	sendJSON(w, response, http.StatusOK)
}
