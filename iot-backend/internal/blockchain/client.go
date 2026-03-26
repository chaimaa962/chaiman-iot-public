package blockchain

import (
    "context"
    "crypto/ecdsa"
    "fmt"
    "log"
    "math/big"
    "strings"
    "sync"
    "time"
    "sort"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
)

// ============================================
// VARIABLES GLOBALES
// ============================================

var (
    Client           *ethclient.Client
    PrivateKey       *ecdsa.PrivateKey
    Contract         *IoTDeviceManager
    ExtendedContract *Blockchain
    ChainID          *big.Int
    ContractAddress  common.Address

    // Cache pour les transactions confirmées
    confirmedTxCache     = make(map[common.Hash]*types.Receipt)
    confirmedTxCacheMu   sync.RWMutex
    pendingTransactions  = make(map[common.Hash]bool)
    pendingTxMu          sync.Mutex
)

// LES 4 NŒUDS GETH
var GethNodes = []struct {
    Address  string
    DeviceID string
    NodeID   int
    Interval int
    Message  string
}{
    {
        Address:  "0xc6a0a22e356ce0b09246dd7a1b0c3b223d39e0ff",
        DeviceID: "GETH_NODE_1",
        NodeID:   1,
        Interval: 5,
        Message:  "Bonjour je suis le nœud Geth 1 (autorité PoA)",
    },
    {
        Address:  "0x7d2b6759153a1625a757f36b63f7034dd8c095ed",
        DeviceID: "GETH_NODE_2",
        NodeID:   2,
        Interval: 10,
        Message:  "Bonjour je suis le nœud Geth 2 (autorité PoA)",
    },
    {
        Address:  "0x56f1df7fd9da1b90616c9fc7e4faf303a1cf6830",
        DeviceID: "GETH_NODE_3",
        NodeID:   3,
        Interval: 15,
        Message:  "Bonjour je suis le nœud Geth 3 (autorité PoA)",
    },
    {
        Address:  "0xb1a91f75437f01983a843719ec421532f5e044ed",
        DeviceID: "GETH_NODE_4",
        NodeID:   4,
        Interval: 25,
        Message:  "Bonjour je suis le nœud Geth 4 (autorité PoA)",
    },
}

// ============================================
// INITIALISATION
// ============================================

func InitBlockchain(nodeURL, privateKeyHex, contractAddr string) error {
    var err error

    Client, err = ethclient.Dial(nodeURL)
    if err != nil {
        return fmt.Errorf("❌ erreur connexion: %v", err)
    }

    ChainID, err = Client.NetworkID(context.Background())
    if err != nil {
        return fmt.Errorf("❌ erreur chainID: %v", err)
    }

    PrivateKey, err = crypto.HexToECDSA(privateKeyHex)
    if err != nil {
        return fmt.Errorf("❌ erreur clé privée: %v", err)
    }

    ContractAddress = common.HexToAddress(contractAddr)

    Contract, err = NewIoTDeviceManager(ContractAddress, Client)
    if err != nil {
        return fmt.Errorf("❌ erreur contrat principal: %v", err)
    }

    extContract, err := NewBlockchain(ContractAddress, Client)
    if err == nil {
        ExtendedContract = extContract
        log.Println("✅ Contrat étendu chargé avec succès")
    } else {
        log.Printf("⚠️ Contrat étendu non disponible: %v", err)
        ExtendedContract = nil
    }

    log.Println("✅ Blockchain connectée - ChainID:", ChainID)
    log.Printf("✅ Contrat chargé à l'adresse: %s", ContractAddress.Hex())

    go cacheCleanupRoutine()
    go asyncConfirmationRoutine()

    InitGethNodes()

    return nil
}

func cacheCleanupRoutine() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        confirmedTxCacheMu.Lock()
        if len(confirmedTxCache) > 1000 {
            newCache := make(map[common.Hash]*types.Receipt)
            count := 0
            for hash, receipt := range confirmedTxCache {
                if count >= len(confirmedTxCache)-1000 {
                    newCache[hash] = receipt
                }
                count++
            }
            confirmedTxCache = newCache
        }
        confirmedTxCacheMu.Unlock()
    }
}

func asyncConfirmationRoutine() {
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        pendingTxMu.Lock()
        pendingList := make([]common.Hash, 0, len(pendingTransactions))
        for hash := range pendingTransactions {
            pendingList = append(pendingList, hash)
        }
        pendingTxMu.Unlock()
        
        for _, txHash := range pendingList {
            receipt, err := Client.TransactionReceipt(context.Background(), txHash)
            if err == nil && receipt != nil {
                confirmedTxCacheMu.Lock()
                confirmedTxCache[txHash] = receipt
                confirmedTxCacheMu.Unlock()
                
                pendingTxMu.Lock()
                delete(pendingTransactions, txHash)
                pendingTxMu.Unlock()
                
                if receipt.Status == 1 {
                    log.Printf("✅ [ASYNC] TX confirmée: %s au bloc %d", 
                        txHash.Hex()[:20], receipt.BlockNumber)
                } else {
                    log.Printf("❌ [ASYNC] TX échouée: %s", txHash.Hex()[:20])
                }
            }
        }
    }
}

func InitGethNodes() {
    log.Println("\n============================================================")
    log.Println("🔧 VÉRIFICATION DES 4 NŒUDS GETH")
    log.Println("============================================================")

    for _, node := range GethNodes {
        deviceAddr := common.HexToAddress(node.Address)

        exists, err := DeviceIsRegistered(deviceAddr)
        if err != nil {
            log.Printf("❌ Erreur vérification %s: %v", node.DeviceID, err)
            continue
        }

        if exists {
            log.Printf("✅ %s déjà enregistré: %s", node.DeviceID, node.Address[:20])
        } else {
            log.Printf("📝 %s non trouvé, enregistrement...", node.DeviceID)

            publicKey := generatePublicKeyFromAddress(node.Address)

            metadata := fmt.Sprintf(`{
                "device_id": "%s",
                "node_id": %d,
                "type": "geth_node",
                "is_authority": true,
                "interval": %d,
                "message_template": "%s",
                "initialized_at": %d,
                "permanent": true
            }`, node.DeviceID, node.NodeID, node.Interval, node.Message, time.Now().Unix())

            txHash, err := DeviceRegister(deviceAddr, publicKey, metadata)
            if err != nil {
                log.Printf("❌ Erreur enregistrement %s: %v", node.DeviceID, err)
            } else {
                log.Printf("✅ %s enregistré avec succès!", node.DeviceID)
                log.Printf("   TX: %s", txHash.Hex()[:30])
            }

            time.Sleep(2 * time.Second)
        }
    }

    log.Println("============================================================")
}

func generatePublicKeyFromAddress(address string) string {
    hash := crypto.Keccak256Hash([]byte(address))
    return "0x" + common.Bytes2Hex(hash.Bytes()[:32])
}

func RecordAuthenticationWithRetry(deviceAddress common.Address, success bool, proofHash [32]byte) (common.Hash, error) {
    var lastErr error

    for attempt := 0; attempt < 3; attempt++ {
        auth, err := GetTransactionAuth()
        if err != nil {
            lastErr = err
            time.Sleep(2 * time.Second)
            continue
        }

        tx, err := Contract.RecordAuthentication(auth, deviceAddress, success, proofHash)
        if err != nil {
            lastErr = err
            log.Printf("⚠️ Tentative %d échouée: %v", attempt+1, err)
            time.Sleep(3 * time.Second)
            continue
        }

        log.Printf("✅ Transaction envoyée: %s", tx.Hash().Hex())
        return tx.Hash(), nil
    }

    return common.Hash{}, fmt.Errorf("échec après 3 tentatives: %v", lastErr)
}

func DeviceRegister(deviceAddress common.Address, publicKey, metadata string) (common.Hash, error) {
    auth, err := GetTransactionAuth()
    if err != nil {
        return common.Hash{}, err
    }

    tx, err := Contract.RegisterDevice(auth, deviceAddress, publicKey, metadata)
    if err != nil {
        return common.Hash{}, fmt.Errorf("❌ erreur enregistrement: %v", err)
    }

    log.Printf("✅ Appareil %s enregistré - TX: %s",
        deviceAddress.Hex()[:15], tx.Hash().Hex()[:15])

    WaitForTransaction(tx.Hash())

    return tx.Hash(), nil
}

func DeviceGetInfo(deviceAddress common.Address) (string, string, bool, *big.Int, error) {
    publicKey, metadata, isActive, lastSeen, err := Contract.GetDeviceInfo(&bind.CallOpts{}, deviceAddress)
    if err != nil {
        return "", "", false, nil, fmt.Errorf("❌ erreur getDeviceInfo: %v", err)
    }
    return publicKey, metadata, isActive, lastSeen, nil
}

func DeviceIsRegistered(deviceAddress common.Address) (bool, error) {
    return Contract.IsDeviceRegistered(&bind.CallOpts{}, deviceAddress)
}

func DeviceDeactivate(deviceAddress common.Address) (common.Hash, error) {
    auth, err := GetTransactionAuth()
    if err != nil {
        return common.Hash{}, err
    }

    tx, err := Contract.DeactivateDevice(auth, deviceAddress)
    if err != nil {
        return common.Hash{}, fmt.Errorf("❌ erreur désactivation: %v", err)
    }

    log.Printf("✅ Appareil désactivé - TX: %s", tx.Hash().Hex()[:15])
    return tx.Hash(), nil
}

func DeviceActivate(deviceAddress common.Address) (common.Hash, error) {
    auth, err := GetTransactionAuth()
    if err != nil {
        return common.Hash{}, err
    }

    tx, err := Contract.ActivateDevice(auth, deviceAddress)
    if err != nil {
        return common.Hash{}, fmt.Errorf("❌ erreur réactivation: %v", err)
    }

    log.Printf("✅ Appareil réactivé - TX: %s", tx.Hash().Hex()[:15])
    return tx.Hash(), nil
}

func AuthRecord(deviceAddress common.Address, success bool, proofHash [32]byte) (common.Hash, error) {
    log.Printf("🔐 Enregistrement authentification pour %s", deviceAddress.Hex()[:15])

    txHash, err := RecordAuthenticationWithRetry(deviceAddress, success, proofHash)
    if err != nil {
        return common.Hash{}, err
    }

    if IsGethNode(deviceAddress) {
        log.Printf("🔐 Nœud Geth authentifié: %s", deviceAddress.Hex()[:20])
    }

    pendingTxMu.Lock()
    pendingTransactions[txHash] = true
    pendingTxMu.Unlock()

    log.Printf("⏳ Attente confirmation TX: %s", txHash.Hex()[:15])
    
    for i := 0; i < 60; i++ {
        confirmedTxCacheMu.RLock()
        receipt, found := confirmedTxCache[txHash]
        confirmedTxCacheMu.RUnlock()
        
        if found {
            if receipt.Status == 1 {
                log.Printf("✅ Transaction confirmée au bloc %d (depuis cache)", receipt.BlockNumber)
                pendingTxMu.Lock()
                delete(pendingTransactions, txHash)
                pendingTxMu.Unlock()
                return txHash, nil
            }
        }
        
        receipt, err = Client.TransactionReceipt(context.Background(), txHash)
        if err == nil && receipt != nil {
            confirmedTxCacheMu.Lock()
            confirmedTxCache[txHash] = receipt
            confirmedTxCacheMu.Unlock()
            
            if receipt.Status == 1 {
                log.Printf("✅ Transaction confirmée au bloc %d", receipt.BlockNumber)
                pendingTxMu.Lock()
                delete(pendingTransactions, txHash)
                pendingTxMu.Unlock()
                return txHash, nil
            }
        }
        time.Sleep(1 * time.Second)
    }
    
    log.Printf("⚠️ Transaction non confirmée après 60s: %s", txHash.Hex()[:15])
    return txHash, nil
}

func GetGethNodeAddresses() []common.Address {
    addresses := make([]common.Address, len(GethNodes))
    for i, node := range GethNodes {
        addresses[i] = common.HexToAddress(node.Address)
    }
    return addresses
}

func IsGethNode(addr common.Address) bool {
    addrLower := strings.ToLower(addr.Hex())
    for _, node := range GethNodes {
        if strings.ToLower(node.Address) == addrLower {
            return true
        }
    }
    return false
}

func GetGethNodeInfo(addr common.Address) *struct {
    Address  string
    DeviceID string
    NodeID   int
    Interval int
    Message  string
} {
    addrLower := strings.ToLower(addr.Hex())
    for i := range GethNodes {
        if strings.ToLower(GethNodes[i].Address) == addrLower {
            return &GethNodes[i]
        }
    }
    return nil
}

func GetAllGethNodes() []map[string]interface{} {
    nodes := []map[string]interface{}{}

    for _, node := range GethNodes {
        addr := common.HexToAddress(node.Address)

        publicKey, metadata, isActive, lastSeen, err := DeviceGetInfo(addr)
        if err != nil {
            publicKey = generatePublicKeyFromAddress(node.Address)
            metadata = fmt.Sprintf(`{"device_id":"%s","node_id":%d,"type":"geth_node"}`,
                node.DeviceID, node.NodeID)
            isActive = true
            lastSeen = big.NewInt(0)
        }

        authCountBig, _ := Contract.AuthCount(&bind.CallOpts{}, addr)
        authCount := uint64(0)
        if authCountBig != nil {
            authCount = authCountBig.Uint64()
        }

        nodes = append(nodes, map[string]interface{}{
            "address":    addr.Hex(),
            "device_id":  node.DeviceID,
            "node_id":    node.NodeID,
            "publicKey":  publicKey,
            "metadata":   metadata,
            "isActive":   isActive,
            "lastSeen":   lastSeen.Uint64(),
            "authCount":  authCount,
            "interval":   node.Interval,
            "message":    node.Message,
            "isGethNode": true,
            "permanent":  true,
        })
    }

    return nodes
}

func HistoryGetAll(deviceAddress common.Address) ([]IoTDeviceManagerAuthRecord, error) {
    return Contract.GetAuthHistory(&bind.CallOpts{}, deviceAddress)
}

func HistoryGetRecent(deviceAddress common.Address, limit uint64) ([]IoTDeviceManagerAuthRecord, error) {
    return Contract.GetRecentAuthHistory(&bind.CallOpts{}, deviceAddress, new(big.Int).SetUint64(limit))
}

func StatsGetGlobal() (map[string]interface{}, error) {
    result, err := Contract.GetGlobalStats(&bind.CallOpts{})
    if err != nil {
        return nil, fmt.Errorf("❌ erreur stats globales: %v", err)
    }

    inactiveDevices := new(big.Int).Sub(result.TotalDevices, result.ActiveDevices)

    stats := map[string]interface{}{
        "totalDevices":    result.TotalDevices.Uint64(),
        "totalAuths":      result.TotalAuthentications.Uint64(),
        "activeDevices":   result.ActiveDevices.Uint64(),
        "inactiveDevices": inactiveDevices.Uint64(),
        "gethNodesCount":  len(GethNodes),
        "gethNodesActive": countActiveGethNodes(),
        "successRate":     0,
    }

    return stats, nil
}

func StatsGetDeviceAuthCount(deviceAddress common.Address) (*big.Int, error) {
    return Contract.AuthCount(&bind.CallOpts{}, deviceAddress)
}

func countActiveGethNodes() int {
    active := 0
    for _, node := range GethNodes {
        addr := common.HexToAddress(node.Address)
        exists, _ := DeviceIsRegistered(addr)
        if exists {
            active++
        }
    }
    return active
}

func WaitForTransaction(txHash common.Hash) error {
    log.Printf("⏳ Attente confirmation TX: %s", txHash.Hex()[:15])

    for i := 0; i < 60; i++ {
        confirmedTxCacheMu.RLock()
        receipt, found := confirmedTxCache[txHash]
        confirmedTxCacheMu.RUnlock()
        
        if found {
            if receipt.Status == 1 {
                log.Printf("✅ Transaction confirmée au bloc %d (cache)", receipt.BlockNumber)
                return nil
            }
            return fmt.Errorf("❌ transaction échouée status: %d", receipt.Status)
        }
        
        receipt, err := Client.TransactionReceipt(context.Background(), txHash)
        if err == nil && receipt != nil {
            confirmedTxCacheMu.Lock()
            confirmedTxCache[txHash] = receipt
            confirmedTxCacheMu.Unlock()
            
            if receipt.Status == 1 {
                log.Printf("✅ Transaction confirmée au bloc %d", receipt.BlockNumber)
                return nil
            }
            return fmt.Errorf("❌ transaction échouée status: %d", receipt.Status)
        }
        time.Sleep(1 * time.Second)
    }
    return fmt.Errorf("❌ timeout après 60s")
}

func GetBalance(address common.Address) (*big.Int, error) {
    return Client.BalanceAt(context.Background(), address, nil)
}

func GetBlockNumber() (uint64, error) {
    return Client.BlockNumber(context.Background())
}

func GetAllDevices() ([]common.Address, error) {
    return Contract.GetAllDevices(&bind.CallOpts{})
}

func GetAllDevicesWithDetails() ([]map[string]interface{}, error) {
    addresses, err := GetAllDevices()
    if err != nil {
        return nil, err
    }

    devices := []map[string]interface{}{}

    for _, addr := range addresses {
        publicKey, metadata, isActive, lastSeen, err := DeviceGetInfo(addr)
        if err != nil {
            continue
        }

        authCountBig, _ := StatsGetDeviceAuthCount(addr)
        authCount := uint64(0)
        if authCountBig != nil {
            authCount = authCountBig.Uint64()
        }

        var lastSeenUint64 uint64
        if lastSeen != nil {
            lastSeenUint64 = lastSeen.Uint64()
        }

        device := map[string]interface{}{
            "address":    addr.Hex(),
            "publicKey":  publicKey,
            "metadata":   metadata,
            "isActive":   isActive,
            "lastSeen":   lastSeenUint64,
            "authCount":  authCount,
            "isGethNode": IsGethNode(addr),
        }

        if IsGethNode(addr) {
            if info := GetGethNodeInfo(addr); info != nil {
                device["device_id"] = info.DeviceID
                device["node_id"] = info.NodeID
                device["interval"] = info.Interval
                device["message"] = info.Message
                device["permanent"] = true
            }
        }

        devices = append(devices, device)
    }

    return devices, nil
}

type Transaction struct {
    TxHash        string
    DeviceID      string
    DeviceAddress string
    Message       string
    Timestamp     int64
    BlockNumber   uint64
    Sequence      uint64
    Success       bool
    Data          []byte
}

func RecordTransaction(deviceID, deviceAddr string, data []byte, messageHash string, sequence uint64, timestamp int64) (common.Hash, uint64, error) {

    addr := common.HexToAddress(deviceAddr)
    isGeth := IsGethNode(addr)

    if isGeth {
        log.Printf("🔷 Transaction nœud Geth: %s", deviceID)
    }

    var tx *types.Transaction
    var err error
    var receipt *types.Receipt

    for attempt := 0; attempt < 3; attempt++ {
        auth, authErr := GetTransactionAuth()
        if authErr != nil {
            if attempt == 2 {
                return common.Hash{}, 0, fmt.Errorf("❌ erreur auth après %d tentatives: %v", attempt+1, authErr)
            }
            time.Sleep(1 * time.Second)
            continue
        }

        if ExtendedContract != nil {
            tx, err = ExtendedContract.RecordTransaction(
                auth,
                addr,
                deviceID,
                data,
                messageHash,
                new(big.Int).SetUint64(sequence),
                big.NewInt(timestamp),
            )
        } else {
            var proofHash [32]byte
            copy(proofHash[:], messageHash[:32])
            tx, err = Contract.RecordAuthentication(auth, addr, true, proofHash)
        }

        if err != nil {
            if strings.Contains(err.Error(), "nonce") || strings.Contains(err.Error(), "underpriced") {
                log.Printf("⚠️ Erreur nonce (tentative %d): %v", attempt+1, err)
                time.Sleep(2 * time.Second)
                continue
            }
            return common.Hash{}, 0, fmt.Errorf("❌ erreur transaction: %v", err)
        }

        break
    }

    if err != nil {
        return common.Hash{}, 0, err
    }

    WaitForTransaction(tx.Hash())

    receipt, _ = Client.TransactionReceipt(context.Background(), tx.Hash())
    blockNum := uint64(0)
    if receipt != nil {
        blockNum = receipt.BlockNumber.Uint64()
    }

    nodeType := "Appareil"
    if isGeth {
        nodeType = "Nœud Geth"
    }

    log.Printf("✅ %s %s - Bloc: %d", nodeType, deviceID, blockNum)

    return tx.Hash(), blockNum, nil
}

func PrintGethNodesStatus() {
    log.Println("\n============================================================")
    log.Println("📊 STATUT DES 4 NŒUDS GETH PERMANENTS")
    log.Println("============================================================")

    for i, node := range GethNodes {
        addr := common.HexToAddress(node.Address)
        exists, _ := DeviceIsRegistered(addr)

        status := "❌ Non enregistré"
        if exists {
            status = "✅ Enregistré"

            _, _, _, lastSeen, _ := DeviceGetInfo(addr)
            if lastSeen != nil && lastSeen.Uint64() > 0 {
                status += fmt.Sprintf(" (actif il y a %ds)", time.Now().Unix()-int64(lastSeen.Uint64()))
            }
        }

        log.Printf("Nœud %d: %s", i+1, node.Address[:30])
        log.Printf("   • Device ID: %s", node.DeviceID)
        log.Printf("   • Intervalle: %ds", node.Interval)
        log.Printf("   • Statut: %s", status)
        log.Printf("   • Message: \"%s\"", node.Message)
        if i < len(GethNodes)-1 {
            log.Println("   " + strings.Repeat("-", 40))
        }
    }

    log.Println("============================================================")
}

func GetRecentTransactions(limit int) ([]Transaction, error) {
    transactions := []Transaction{}

    devices, err := GetAllDevices()
    if err != nil {
        return transactions, err
    }

    for _, addr := range devices {
        history, err := HistoryGetRecent(addr, uint64(limit/len(devices)+1))
        if err != nil {
            continue
        }

        for _, record := range history {
            deviceID := addr.Hex()[:10]
            if node := GetGethNodeInfo(addr); node != nil {
                deviceID = node.DeviceID
            }

            txHash := common.BytesToHash(record.ProofHash[:])
            blockNumber := uint64(0)
            
            confirmedTxCacheMu.RLock()
            receipt, found := confirmedTxCache[txHash]
            confirmedTxCacheMu.RUnlock()
            
            if found {
                blockNumber = receipt.BlockNumber.Uint64()
            } else {
                receipt, err := Client.TransactionReceipt(context.Background(), txHash)
                if err == nil && receipt != nil {
                    blockNumber = receipt.BlockNumber.Uint64()
                    confirmedTxCacheMu.Lock()
                    confirmedTxCache[txHash] = receipt
                    confirmedTxCacheMu.Unlock()
                }
            }

            transactions = append(transactions, Transaction{
                TxHash:        "0x" + common.Bytes2Hex(record.ProofHash[:]),
                DeviceID:      deviceID,
                DeviceAddress: addr.Hex(),
                Message:       "Authentication",
                Timestamp:     record.Timestamp.Int64() * 1000,
                BlockNumber:   blockNumber,
                Sequence:      0,
                Success:       record.Success,
                Data:          nil,
            })
        }
    }

    sort.Slice(transactions, func(i, j int) bool {
        return transactions[i].Timestamp > transactions[j].Timestamp
    })

    if len(transactions) > limit {
        transactions = transactions[:limit]
    }

    return transactions, nil
}

func GetRecentAuthentications(limit int) ([]Transaction, error) {
    transactions := []Transaction{}

    addresses, err := GetAllDevices()
    if err != nil {
        return transactions, err
    }

    log.Printf("📋 Récupération historique pour %d appareils", len(addresses))

    for _, addr := range addresses {
        history, err := Contract.GetAuthHistory(&bind.CallOpts{}, addr)
        if err != nil {
            log.Printf("⚠️ Erreur historique %s: %v", addr.Hex(), err)
            continue
        }

        log.Printf("📊 Appareil %s a %d authentifications", addr.Hex()[:10], len(history))

        deviceID := addr.Hex()[:10]
        if node := GetGethNodeInfo(addr); node != nil {
            deviceID = node.DeviceID
        }

        for i, record := range history {
            txHash := common.BytesToHash(record.ProofHash[:])
            
            blockNumber := uint64(0)
            
            confirmedTxCacheMu.RLock()
            receipt, found := confirmedTxCache[txHash]
            confirmedTxCacheMu.RUnlock()
            
            if found {
                blockNumber = receipt.BlockNumber.Uint64()
            } else {
                receipt, err := Client.TransactionReceipt(context.Background(), txHash)
                if err == nil && receipt != nil {
                    blockNumber = receipt.BlockNumber.Uint64()
                    confirmedTxCacheMu.Lock()
                    confirmedTxCache[txHash] = receipt
                    confirmedTxCacheMu.Unlock()
                }
            }

            transactions = append(transactions, Transaction{
                TxHash:        txHash.Hex(),
                DeviceID:      deviceID,
                DeviceAddress: addr.Hex(),
                Message:       fmt.Sprintf("Authentication #%d", i+1),
                Timestamp:     record.Timestamp.Int64() * 1000,
                BlockNumber:   blockNumber,
                Sequence:      uint64(i + 1),
                Success:       record.Success,
                Data:          nil,
            })
        }
    }

    log.Printf("📊 Total transactions trouvées: %d", len(transactions))

    sort.Slice(transactions, func(i, j int) bool {
        return transactions[i].Timestamp > transactions[j].Timestamp
    })

    if len(transactions) > limit {
        transactions = transactions[:limit]
    }

    return transactions, nil
}

func VerifyTransactionStatus(txHash common.Hash) (*types.Receipt, error) {
    confirmedTxCacheMu.RLock()
    receipt, found := confirmedTxCache[txHash]
    confirmedTxCacheMu.RUnlock()
    
    if found {
        return receipt, nil
    }
    
    receipt, err := Client.TransactionReceipt(context.Background(), txHash)
    if err != nil {
        return nil, err
    }
    
    if receipt != nil {
        confirmedTxCacheMu.Lock()
        confirmedTxCache[txHash] = receipt
        confirmedTxCacheMu.Unlock()
    }
    
    return receipt, nil
}

func GetPendingTransactionsCount() int {
    pendingTxMu.Lock()
    defer pendingTxMu.Unlock()
    return len(pendingTransactions)
}

func GetTransactionAuth() (*bind.TransactOpts, error) {
    auth, err := bind.NewKeyedTransactorWithChainID(PrivateKey, ChainID)
    if err != nil {
        return nil, fmt.Errorf("❌ erreur création auth: %v", err)
    }

    nonce, err := Client.PendingNonceAt(context.Background(), auth.From)
    if err != nil {
        return nil, fmt.Errorf("❌ erreur récupération nonce: %v", err)
    }

    auth.Nonce = big.NewInt(int64(nonce))
    auth.GasLimit = uint64(6721975)
    auth.Value = big.NewInt(0)
    auth.NoSend = false

    // 🔥 FORCER UN GAS PRICE ÉLEVÉ (50 Gwei)
    auth.GasPrice = big.NewInt(50000000000) // 50 Gwei

    log.Printf("💰 Gas price forcé: %d wei (50 Gwei)", auth.GasPrice)
    return auth, nil
}
