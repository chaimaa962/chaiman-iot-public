// internal/handlers/dashboard.go
package handlers

import (
    "math/big"
    "net/http"
    "strconv"
    "time"

    "github.com/ethereum/go-ethereum/common"
    "github.com/gorilla/mux"
    "iot-backend/internal/blockchain"
    "iot-backend/internal/utils"
)

// HandleGlobalStats retourne les statistiques globales
func HandleGlobalStats(w http.ResponseWriter, r *http.Request) {
    logger := utils.GetLogger()

    stats, err := blockchain.StatsGetGlobal()
    if err != nil {
        logger.Error("❌ Erreur stats globales", "error", err)
        stats = map[string]interface{}{
            "totalDevices":    0,
            "totalAuths":      0,
            "activeDevices":   0,
            "inactiveDevices": 0,
            "gethNodesCount":  4,
            "gethNodesActive": 0,
            "successRate":     100,
        }
    }

    blockNumber, _ := blockchain.GetBlockNumber()
    stats["latestBlock"] = blockNumber
    stats["timestamp"] = time.Now().Unix()

    sendJSON(w, stats, http.StatusOK)
}

// HandleDeviceHistory retourne l'historique d'un appareil
func HandleDeviceHistory(w http.ResponseWriter, r *http.Request) {
    logger := utils.GetLogger()
    vars := mux.Vars(r)
    addressHex := vars["address"]

    limitStr := r.URL.Query().Get("limit")
    limit := uint64(50)
    if limitStr != "" {
        if l, err := strconv.ParseUint(limitStr, 10, 64); err == nil {
            limit = l
        }
    }

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

    history, err := blockchain.HistoryGetRecent(deviceAddr, limit)
    if err != nil {
        logger.Error("❌ Erreur historique", "error", err)
        sendJSON(w, []interface{}{}, http.StatusOK)
        return
    }

    formattedHistory := make([]map[string]interface{}, len(history))
    for i, record := range history {
        formattedHistory[i] = map[string]interface{}{
            "timestamp": record.Timestamp.Int64(),
            "success":   record.Success,
            "proofHash": common.Bytes2Hex(record.ProofHash[:]),
            "index":     i,
        }
    }

    logger.Info("📋 Historique récupéré", "address", addressHex[:15], "count", len(formattedHistory))
    sendJSON(w, formattedHistory, http.StatusOK)
}

// HandleBlockchainInfo retourne les informations de la blockchain
func HandleBlockchainInfo(w http.ResponseWriter, r *http.Request) {
    logger := utils.GetLogger()

    blockNumber, err := blockchain.GetBlockNumber()
    if err != nil {
        logger.Error("❌ Erreur récupération block number", "error", err)
        blockNumber = 0
    }

    gasPrice := big.NewInt(0)
    if blockchain.Client != nil {
        price, err := blockchain.Client.SuggestGasPrice(r.Context())
        if err == nil {
            gasPrice = price
        }
    }

    peers := 4

    info := map[string]interface{}{
        "latestBlock":  blockNumber,
        "gasPrice":     gasPrice.String(),
        "networkId":    blockchain.ChainID.String(),
        "peers":        peers,
        "syncing":      false,
        "chainId":      blockchain.ChainID.Uint64(),
        "nodeUrl":      "http://localhost:8545",
        "contractAddr": blockchain.ContractAddress.Hex(),
        "timestamp":    time.Now().Unix(),
    }

    logger.Info("ℹ️ Infos blockchain", "block", blockNumber, "peers", peers)
    sendJSON(w, info, http.StatusOK)
}

// HandleTransactionByHash retourne les détails d'une transaction
func HandleTransactionByHash(w http.ResponseWriter, r *http.Request) {
    logger := utils.GetLogger()
    vars := mux.Vars(r)
    hash := vars["hash"]

    devices, err := blockchain.GetAllDevices()
    if err != nil {
        sendError(w, "Transaction non trouvée", http.StatusNotFound)
        return
    }

    for _, addr := range devices {
        history, err := blockchain.HistoryGetAll(addr)
        if err != nil {
            continue
        }

        for _, record := range history {
            recordHash := common.Bytes2Hex(record.ProofHash[:])
            if recordHash == hash || "0x"+recordHash == hash {
                nodeInfo := blockchain.GetGethNodeInfo(addr)
                deviceID := addr.Hex()[:10]
                if nodeInfo != nil {
                    deviceID = nodeInfo.DeviceID
                }

                tx := map[string]interface{}{
                    "hash":          "0x" + recordHash,
                    "deviceId":      deviceID,
                    "deviceAddress": addr.Hex(),
                    "type":          "authentication",
                    "status":        "success",
                    "success":       record.Success,
                    "message":       "Authentification ZKP",
                    "timestamp":     record.Timestamp.Int64() * 1000,
                    "blockNumber":   0,
                    "block":         0,
                    "latency":       int(time.Now().Unix() - record.Timestamp.Int64()),
                    "ecdsaValid":    record.Success,
                    "zkpValid":      record.Success,
                    "authType":      "ECDSA+ZKP",
                    "sequence":      0,
                    "proofHash":     "0x" + recordHash,
                }

                logger.Info("✅ Transaction trouvée", "hash", hash[:20])
                sendJSON(w, tx, http.StatusOK)
                return
            }
        }
    }

    sendError(w, "Transaction non trouvée", http.StatusNotFound)
}
