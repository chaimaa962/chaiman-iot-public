// internal/handlers/messages_secure.go
package handlers

import (
    "encoding/hex"
    "encoding/json"
    "math/big"
    "net/http"
    "strings"
    "sync"
    "time"

    "github.com/ethereum/go-ethereum/common"
    "iot-backend/internal/blockchain"
    "iot-backend/internal/models"
    "iot-backend/internal/utils"
    "iot-backend/pkg/zkp"
)

// ============================================
// VARIABLES POUR ANTI-REJEU
// ============================================
var (
    lastSequences   = make(map[string]uint64)
    lastSequencesMu sync.RWMutex
    
    usedNonces   = make(map[string]bool)
    usedNoncesMu sync.RWMutex
)

type SecureNodeMessageRequest struct {
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

// HandleSecureNodeMessage - Handler sécurisé avec vérification ECDSA + ZKP
func HandleSecureNodeMessage(w http.ResponseWriter, r *http.Request) {
    logger := utils.GetLogger()

    var req SecureNodeMessageRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, "Format JSON invalide", http.StatusBadRequest)
        return
    }

    req.Address = strings.ToLower(req.Address)

    logger.Info("🔐 Requête sécurisée reçue", "device", req.DeviceID, "address", req.Address[:15]+"...")

    // ============================================
    // 1. VÉRIFICATIONS ANTI-REJEU
    // ============================================

    // Vérifier timestamp (5 minutes max)
    if time.Now().Unix()-req.Timestamp > 300 {
        sendError(w, "Message trop ancien (>5min)", http.StatusBadRequest)
        return
    }

    // Vérifier timestamp dans le futur (tolerance 30s)
    if req.Timestamp > time.Now().Unix()+30 {
        sendError(w, "Timestamp dans le futur", http.StatusBadRequest)
        return
    }

    // Vérification séquence
    lastSequencesMu.Lock()
    lastSeq := lastSequences[req.DeviceID]
    if req.Sequence <= lastSeq {
        lastSequencesMu.Unlock()
        sendError(w, "Séquence invalide", http.StatusBadRequest)
        return
    }
    lastSequences[req.DeviceID] = req.Sequence
    lastSequencesMu.Unlock()

    // Vérification nonce
    nonceKey := req.DeviceID + ":" + req.Nonce
    usedNoncesMu.Lock()
    if usedNonces[nonceKey] {
        usedNoncesMu.Unlock()
        sendError(w, "Nonce déjà utilisé", http.StatusBadRequest)
        return
    }
    usedNonces[nonceKey] = true
    usedNoncesMu.Unlock()

    // ============================================
    // 2. VÉRIFIER QUE L'APPAREIL EXISTE
    // ============================================

    deviceAddr := common.HexToAddress(req.Address)
    exists, err := blockchain.DeviceIsRegistered(deviceAddr)
    if err != nil || !exists {
        logger.Warn("Appareil non enregistré", "address", req.Address[:15])
        sendError(w, "Appareil non enregistré", http.StatusNotFound)
        return
    }

    // ============================================
    // 3. VÉRIFICATION ECDSA
    // ============================================
    logger.Info("🔑 Vérification ECDSA...", "device", req.DeviceID)

    if req.Signature == "" || req.MessageHash == "" {
        sendError(w, "Signature ou hash manquant", http.StatusBadRequest)
        return
    }

    validECDSA, err := zkp.VerifyECDSASignature(deviceAddr, req.MessageHash, req.Signature)
    if err != nil || !validECDSA {
        logger.Warn("❌ Signature ECDSA invalide", "device", req.DeviceID, "error", err)
        sendError(w, "Authentification ECDSA échouée", http.StatusUnauthorized)
        return
    }

    logger.Info("✅ Signature ECDSA valide", "device", req.DeviceID)

    // ============================================
    // 4. VÉRIFICATION ZKP
    // ============================================
    logger.Info("🔐 Vérification ZKP...", "device", req.DeviceID)

    zkpService, err := zkp.GetSecureZKPManager()
    if err != nil {
        logger.Error("❌ Service ZKP indisponible", "error", err)
        sendError(w, "Service ZKP indisponible", http.StatusInternalServerError)
        return
    }

    proofBytes, err := hex.DecodeString(req.Proof)
    if err != nil || len(proofBytes) == 0 {
        sendError(w, "Preuve ZKP invalide", http.StatusBadRequest)
        return
    }

    expectedHashBytes, err := hex.DecodeString(req.ExpectedHash)
    if err != nil {
        sendError(w, "Hash attendu invalide", http.StatusBadRequest)
        return
    }
    
    expectedHash := new(big.Int).SetBytes(expectedHashBytes)
    if expectedHash.Sign() == 0 {
        sendError(w, "Hash attendu invalide (zéro)", http.StatusBadRequest)
        return
    }

    challenge := zkp.GenerateChallenge(req.Timestamp, req.Sequence, req.Nonce)

    validZKP, err := zkpService.VerifyProof(proofBytes, deviceAddr, challenge, expectedHash)
    if err != nil || !validZKP {
        logger.Warn("❌ Preuve ZKP invalide", "device", req.DeviceID)
        sendError(w, "Authentification ZKP échouée", http.StatusUnauthorized)
        return
    }

    logger.Info("✅ Preuve ZKP valide", "device", req.DeviceID)
    logger.Info("🎉 Double authentification réussie (ECDSA + ZKP)", "device", req.DeviceID)

    // ============================================
    // 5. ENREGISTRER DANS LA BLOCKCHAIN (CORRIGÉ)
    // ============================================
    
    // 🔴 CORRECTION CRITIQUE: Vérifier la longueur avant de copier
    var proofHash [32]byte
    
    if len(expectedHashBytes) >= 32 {
        // Cas normal: on a 32 bytes ou plus
        copy(proofHash[:], expectedHashBytes[:32])
        logger.Info("📦 Hash de preuve copié normalement", "length", len(expectedHashBytes))
    } else {
        // Cas où le hash est trop court: on copie ce qu'on a et on complète avec des zéros
        logger.Warn("⚠️ Hash attendu trop court", "length", len(expectedHashBytes), "attendu", 32)
        copy(proofHash[:], expectedHashBytes)
        // Le reste est déjà à zéro car proofHash est initialisé avec des zéros
    }

    // Alternative plus simple et sûre:
    // var proofHash [32]byte
    // copy(proofHash[:], expectedHashBytes) // Go copie seulement jusqu'à len(expectedHashBytes)

    txHash, err := blockchain.AuthRecord(deviceAddr, true, proofHash)
    if err != nil {
        logger.Error("❌ Erreur enregistrement blockchain", "error", err)
        sendError(w, "Erreur blockchain: "+err.Error(), http.StatusInternalServerError)
        return
    }

    logger.Info("✅ Authentification enregistrée dans blockchain",
        "device", req.DeviceID,
        "txHash", txHash.Hex())

    // ============================================
    // 6. RÉCUPÉRER LE NUMÉRO DE BLOC
    // ============================================
    blockNum := uint64(0)
    
    // Attendre un peu que la transaction soit minée
    for i := 0; i < 30; i++ {
        receipt, err := blockchain.Client.TransactionReceipt(r.Context(), txHash)
        if err == nil && receipt != nil {
            blockNum = receipt.BlockNumber.Uint64()
            logger.Info("✅ Transaction confirmée", "block", blockNum)
            break
        }
        time.Sleep(1 * time.Second)
        if i%5 == 0 {
            logger.Info("⏳ Attente confirmation...", "tentative", i+1)
        }
    }

    // ============================================
    // 7. RÉPONSE
    // ============================================
    response := models.NodeMessageResponse{
        Success: true,
        Message: "✅ Authentifié (ECDSA + ZKP) et enregistré",
        TxHash:  txHash.Hex(),
        Block:   blockNum,
    }

    logger.Info("📨 Message enregistré", 
        "device", req.DeviceID, 
        "block", blockNum,
        "txHash", txHash.Hex()[:20])
    
    sendJSON(w, response, http.StatusOK)
}

// HandleSecureZKPGenerate - Génère preuve ZKP (nécessite ECDSA valide)
func HandleSecureZKPGenerate(w http.ResponseWriter, r *http.Request) {
    logger := utils.GetLogger()

    var req struct {
        Signature   string `json:"signature"`
        MessageHash string `json:"message_hash"`
        Address     string `json:"address"`
        Sequence    uint64 `json:"sequence"`
        Timestamp   int64  `json:"timestamp"`
        Nonce       string `json:"nonce"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, "Format JSON invalide", http.StatusBadRequest)
        return
    }

    req.Address = strings.ToLower(req.Address)

    deviceAddr := common.HexToAddress(req.Address)
    validECDSA, err := zkp.VerifyECDSASignature(deviceAddr, req.MessageHash, req.Signature)
    if err != nil || !validECDSA {
        logger.Warn("❌ Tentative génération ZKP sans ECDSA valide")
        sendError(w, "Authentification ECDSA requise", http.StatusUnauthorized)
        return
    }

    sigBytes, _ := hex.DecodeString(req.Signature)
    secret := zkp.GenerateSecureSecret(sigBytes)
    challenge := zkp.GenerateChallenge(req.Timestamp, req.Sequence, req.Nonce)

    zkpService, _ := zkp.GetSecureZKPManager()
    proofBytes, expectedHash, err := zkpService.GenerateProof(secret, deviceAddr, challenge)
    if err != nil {
        logger.Error("❌ Erreur génération preuve", "error", err)
        sendError(w, "Erreur génération preuve", http.StatusInternalServerError)
        return
    }

    response := map[string]interface{}{
        "proof":         hex.EncodeToString(proofBytes),
        "expected_hash": hex.EncodeToString(expectedHash.Bytes()),
        "challenge":     challenge.String(),
    }

    logger.Info("✅ Preuve ZKP générée", "address", req.Address[:15])
    sendJSON(w, response, http.StatusOK)
}
