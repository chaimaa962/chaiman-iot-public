// internal/handlers/stats.go
package handlers

import (
    "net/http"
    "strconv"
    "time"

    "iot-backend/internal/blockchain"
    "iot-backend/internal/utils"
)

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
    Latency       int    `json:"latency"`
    ECDSAValid    bool   `json:"ecdsaValid"`
    ZKPValid      bool   `json:"zkpValid"`
    AuthType      string `json:"authType"`
    Sequence      uint64 `json:"sequence"`
    GasUsed       uint64 `json:"gasUsed"`
}

func HandleRecentTransactions(w http.ResponseWriter, r *http.Request) {
    logger := utils.GetLogger()

    limitStr := r.URL.Query().Get("limit")
    limit := 100
    if limitStr != "" {
        if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
            limit = l
        }
    }

    logger.Info("📋 Récupération transactions récentes", "limit", limit)

    transactions, err := blockchain.GetRecentAuthentications(limit)
    if err != nil {
        logger.Error("❌ Erreur récupération transactions", "error", err)
        sendJSON(w, []TransactionResponse{}, http.StatusOK)
        return
    }

    logger.Info("✅ Transactions récupérées", "count", len(transactions))

    response := make([]TransactionResponse, len(transactions))
    for i, tx := range transactions {
        response[i] = TransactionResponse{
            Hash:          tx.TxHash,
            TxHash:        tx.TxHash,
            DeviceID:      tx.DeviceID,
            DeviceAddress: tx.DeviceAddress,
            Address:       tx.DeviceAddress,
            Type:          "authentication",
            Status:        "success",
            Success:       tx.Success,
            Message:       tx.Message,
            Timestamp:     tx.Timestamp, // ← DÉJÀ EN MS
            BlockNumber:   tx.BlockNumber,
            Block:         tx.BlockNumber,
            Latency:       int(time.Now().Unix()*1000 - tx.Timestamp), // ← Calcul en ms
            ECDSAValid:    tx.Success,
            ZKPValid:      tx.Success,
            AuthType:      "ECDSA+ZKP",
            Sequence:      tx.Sequence,
            GasUsed:       21000,
        }
    }

    sendJSON(w, response, http.StatusOK)
}
