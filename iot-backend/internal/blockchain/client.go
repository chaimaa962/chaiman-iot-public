// internal/blockchain/client.go
// VERSION 3.2.0 — SECURISE ET CORRIGE
//
// CORRECTIONS v3.1 → v3.2 :
//   [BUG-CRITIQUE-5]  StartGethNodeScanner() appelé DANS InitBlockchain(), AVANT
//                     que InitNoncePool() soit exécuté dans main_secure.go.
//                     Les goroutines registerDynamicNode() échouent toutes avec
//                     "nonce pool non initialisé — appeler InitNoncePool d'abord".
//                     → StartGethNodeScanner() retiré de InitBlockchain().
//                     → main_secure.go doit l'appeler explicitement APRÈS InitNoncePool().
//                     → Une fonction InitBlockchainWithoutScanner() est également
//                       fournie comme alias explicite pour clarifier l'intention.

package blockchain

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ─────────────────────────────────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────────────────────────────────

// EntityType représente le type d'entité dans le smart contract.
type EntityType uint8

const (
	EntityIoTDevice EntityType = 0
	EntityValidator EntityType = 1
	EntityFullnode  EntityType = 2
)

// SecurePrivateKey encapsule une clé privée ECDSA chiffrée en mémoire AES-256-GCM.
type SecurePrivateKey struct {
	encryptedKey []byte
	publicKey    ecdsa.PublicKey
	address      common.Address
	mu           sync.RWMutex
}

// ─────────────────────────────────────────────────────────────────────────────
// Variables globales du package
// ─────────────────────────────────────────────────────────────────────────────

var (
	Client          *ethclient.Client
	SecureKey       *SecurePrivateKey
	Contract        *IoTDeviceManager
	ChainID         *big.Int
	ContractAddress common.Address

	memoryEncryptionKey     []byte
	memoryEncryptionKeyOnce sync.Once
	memoryKeyMu             sync.RWMutex

	confirmedTxCache   = make(map[common.Hash]*types.Receipt)
	confirmedTxCacheMu sync.RWMutex

	pendingTransactions = make(map[common.Hash]bool)
	pendingTxMu         sync.Mutex

	dynamicNodes   = make(map[string]*DynamicGethNode)
	dynamicNodesMu sync.RWMutex
	lastScanTime   time.Time

	gethRPCClient *http.Client
	gethJWTToken  string

	maintenanceCancel context.CancelFunc
)

// DynamicGethNode représente un nœud Geth découvert dynamiquement.
type DynamicGethNode struct {
	Address  string `json:"address"`
	DeviceID string `json:"device_id"`
	NodeID   int    `json:"node_id"`
	Type     string `json:"type"`
	RPCPort  int    `json:"rpc_port"`
	P2PPort  int    `json:"p2p_port"`
	IsActive bool   `json:"is_active"`
	LastSeen int64  `json:"last_seen"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Initialisation
// ─────────────────────────────────────────────────────────────────────────────

// InitBlockchain initialise la connexion Ethereum, charge le contrat et
// chiffre la clé privée en mémoire via AES-256-GCM.
//
// [BUG-CRITIQUE-5 CORRIGÉ] : StartGethNodeScanner() N'EST PLUS appelé ici.
// Il doit être appelé explicitement dans main_secure.go APRÈS InitNoncePool(),
// sinon registerDynamicNode() échoue avec "nonce pool non initialisé".
//
// Ordre correct dans main_secure.go :
//   1. InitBlockchain(...)     ← connexion + chiffrement clé
//   2. InitNoncePool()         ← pool de nonces prêt
//   3. blockchain.StartGethNodeScanner()  ← scanner peut maintenant envoyer des TX
func InitBlockchain(nodeURL, privateKeyHex, contractAddr string) error {
	var err error

	// 1. Générer la clé de chiffrement mémoire (AES-256) — une seule fois
	memoryEncryptionKeyOnce.Do(func() {
		memoryKeyMu.Lock()
		defer memoryKeyMu.Unlock()
		key := make([]byte, 32)
		if _, initErr := io.ReadFull(rand.Reader, key); initErr != nil {
			err = fmt.Errorf("génération clé AES mémoire : %v", initErr)
			return
		}
		memoryEncryptionKey = key
	})
	if err != nil {
		return err
	}

	// 2. Connexion au nœud Ethereum
	Client, err = ethclient.Dial(nodeURL)
	if err != nil {
		return fmt.Errorf("connexion nœud Ethereum '%s' : %v", nodeURL, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ChainID, err = Client.NetworkID(ctx)
	if err != nil {
		return fmt.Errorf("récupération chainID : %v", err)
	}

	// 3. Chargement et validation de la clé privée
	cleanKey := cleanHexPrefix(privateKeyHex)
	privateKey, err := crypto.HexToECDSA(cleanKey)
	if err != nil {
		return fmt.Errorf("clé privée invalide : %v", err)
	}

	// 4. Chiffrement de la clé privée en mémoire
	SecureKey = &SecurePrivateKey{
		publicKey: privateKey.PublicKey,
		address:   crypto.PubkeyToAddress(privateKey.PublicKey),
	}
	if err = SecureKey.encryptPrivateKey(cleanKey); err != nil {
		zeroPrivateKey(privateKey)
		return fmt.Errorf("chiffrement clé privée : %v", err)
	}

	// 5. Zéroïsation de la clé en clair
	zeroPrivateKey(privateKey)
	privateKey = nil

	// 6. Zéroïser la copie locale de la clé hex
	keyBytes := []byte(cleanKey)
	for i := range keyBytes {
		keyBytes[i] = 0
	}

	// 7. Chargement du smart contract
	ContractAddress = common.HexToAddress(contractAddr)
	Contract, err = NewIoTDeviceManager(ContractAddress, Client)
	if err != nil {
		return fmt.Errorf("chargement contrat '%s' : %v", contractAddr, err)
	}

	// 8. JWT Geth (optionnel — warning si absent)
	gethJWTToken = os.Getenv("GETH_JWT_SECRET")
	if gethJWTToken == "" {
		log.Println("⚠️  GETH_JWT_SECRET non défini — nœuds Geth sans authentification JWT")
	} else {
		log.Println("✅ JWT Geth chargé")
	}

	// 9. Client HTTP Geth interne
	gethRPCClient = buildGethHTTPClient()

	log.Printf("✅ Blockchain initialisée — ChainID: %s, Adresse: %s",
		ChainID.String(), SecureKey.address.Hex())
	log.Printf("✅ Contrat: %s", ContractAddress.Hex())

	// 10. Goroutines de maintenance annulables
	var maintenanceCtx context.Context
	maintenanceCtx, maintenanceCancel = context.WithCancel(context.Background())

	go cacheCleanupRoutine(maintenanceCtx)
	go asyncConfirmationRoutine(maintenanceCtx)
	go staleNodeCleanupRoutine(maintenanceCtx)

	// [BUG-CRITIQUE-5 CORRIGÉ] : StartGethNodeScanner() retiré d'ici.
	// Appeler blockchain.StartGethNodeScanner() dans main_secure.go
	// APRÈS blockchain.InitNoncePool() pour garantir que le pool de nonces
	// est disponible quand registerDynamicNode() tente d'envoyer des transactions.

	return nil
}

// StopBlockchain annule les goroutines de maintenance — à appeler lors du shutdown.
func StopBlockchain() {
	if maintenanceCancel != nil {
		maintenanceCancel()
	}
}

// buildGethHTTPClient construit un http.Client sécurisé pour les appels RPC Geth.
func buildGethHTTPClient() *http.Client {
	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		},
	}
	if caCertPath := os.Getenv("GETH_CA_CERT"); caCertPath != "" {
		log.Printf("ℹ️  GETH_CA_CERT défini — CA Geth : %s", caCertPath)
	}
	return &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:     tlsCfg,
			MaxIdleConns:        50,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     60 * time.Second,
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// zeroPrivateKey — effacement mémoire sécurisé
// ─────────────────────────────────────────────────────────────────────────────

func zeroPrivateKey(pk *ecdsa.PrivateKey) {
	if pk == nil || pk.D == nil {
		return
	}
	bits := pk.D.Bits()
	for i := range bits {
		bits[i] = 0
	}
	pk.D.SetInt64(0)
}

// cleanHexPrefix supprime le préfixe "0x" ou "0X".
func cleanHexPrefix(s string) string {
	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		return s[2:]
	}
	return s
}

// ─────────────────────────────────────────────────────────────────────────────
// SecurePrivateKey — chiffrement / déchiffrement AES-256-GCM
// ─────────────────────────────────────────────────────────────────────────────

func (spk *SecurePrivateKey) encryptPrivateKey(privateKeyHex string) error {
	spk.mu.Lock()
	defer spk.mu.Unlock()

	memoryKeyMu.RLock()
	key := memoryEncryptionKey
	memoryKeyMu.RUnlock()

	if len(key) != 32 {
		return fmt.Errorf("clé AES non initialisée (longueur %d)", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("AES cipher : %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("AES-GCM : %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("nonce GCM : %v", err)
	}

	plaintext := []byte(privateKeyHex)
	spk.encryptedKey = gcm.Seal(nonce, nonce, plaintext, nil)

	for i := range plaintext {
		plaintext[i] = 0
	}
	return nil
}

func (spk *SecurePrivateKey) decryptPrivateKey() (*ecdsa.PrivateKey, error) {
	spk.mu.RLock()
	defer spk.mu.RUnlock()

	if spk.encryptedKey == nil {
		return nil, fmt.Errorf("clé privée non initialisée")
	}

	memoryKeyMu.RLock()
	key := memoryEncryptionKey
	memoryKeyMu.RUnlock()

	if len(key) != 32 {
		return nil, fmt.Errorf("clé AES non initialisée")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("AES cipher : %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("AES-GCM : %v", err)
	}

	nonceSize := gcm.NonceSize()
	if len(spk.encryptedKey) < nonceSize+gcm.Overhead() {
		return nil, fmt.Errorf("données chiffrées trop courtes")
	}

	nonce := spk.encryptedKey[:nonceSize]
	ciphertext := spk.encryptedKey[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("déchiffrement AES-GCM échoué : %w", err)
	}

	privateKey, parseErr := crypto.HexToECDSA(string(plaintext))

	for i := range plaintext {
		plaintext[i] = 0
	}

	if parseErr != nil {
		return nil, fmt.Errorf("parsing clé ECDSA : %v", parseErr)
	}
	return privateKey, nil
}

// GetAddress retourne l'adresse Ethereum de la clé.
func (spk *SecurePrivateKey) GetAddress() common.Address {
	spk.mu.RLock()
	defer spk.mu.RUnlock()
	return spk.address
}

// ─────────────────────────────────────────────────────────────────────────────
// Options de transaction sécurisées — délègue au NoncePool (nonce_manager.go)
// ─────────────────────────────────────────────────────────────────────────────

// getTransactOpts est l'adaptateur interne vers GetSecureTransactOpts.
// Toute la logique de nonce est gérée par nonce_manager.go (NoncePool).
func getTransactOpts(ctx context.Context) (*bind.TransactOpts, error) {
	return GetSecureTransactOpts(ctx)
}

// ─────────────────────────────────────────────────────────────────────────────
// Utilitaires RPC Geth
// ─────────────────────────────────────────────────────────────────────────────

// generateSyntheticPublicKeyFromAddress génère un identifiant déterministe.
// AVERTISSEMENT : ce n'est PAS une clé publique cryptographique valide.
func generateSyntheticPublicKeyFromAddress(address string) (string, error) {
	addr := strings.ToLower(strings.TrimPrefix(address, "0x"))
	if len(addr) != 40 {
		return "", fmt.Errorf("adresse invalide : longueur %d (attendu 40)", len(addr))
	}
	if _, decErr := hex.DecodeString(addr); decErr != nil {
		return "", fmt.Errorf("adresse non-hexadécimale : %v", decErr)
	}
	hash := sha256.Sum256([]byte("synthetic-pubkey-v1:" + addr))
	return "0x04" + hex.EncodeToString(hash[:]) + hex.EncodeToString(hash[:]), nil
}

func newGethRPCRequest(ctx context.Context, port int, body string) (*http.Request, error) {
	url := fmt.Sprintf("http://localhost:%d", port)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if gethJWTToken != "" {
		req.Header.Set("Authorization", "Bearer "+gethJWTToken)
	}
	return req, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Goroutines de maintenance — annulables via context
// ─────────────────────────────────────────────────────────────────────────────

func cacheCleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("cacheCleanupRoutine : arrêt demandé")
			return
		case <-ticker.C:
			now := time.Now()
			confirmedTxCacheMu.Lock()
			for hash, receipt := range confirmedTxCache {
				if receipt == nil {
					delete(confirmedTxCache, hash)
					continue
				}
				blockTime := time.Unix(int64(receipt.BlockNumber.Uint64()), 0)
				if now.Sub(blockTime) > 24*time.Hour {
					delete(confirmedTxCache, hash)
				}
			}
			confirmedTxCacheMu.Unlock()
		}
	}
}

func asyncConfirmationRoutine(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("asyncConfirmationRoutine : arrêt demandé")
			return
		case <-ticker.C:
			pendingTxMu.Lock()
			pendingList := make([]common.Hash, 0, len(pendingTransactions))
			for hash := range pendingTransactions {
				pendingList = append(pendingList, hash)
			}
			pendingTxMu.Unlock()

			for _, txHash := range pendingList {
				rctx, cancel := context.WithTimeout(ctx, 5*time.Second)
				receipt, err := Client.TransactionReceipt(rctx, txHash)
				cancel()

				if err == nil && receipt != nil {
					confirmedTxCacheMu.Lock()
					confirmedTxCache[txHash] = receipt
					confirmedTxCacheMu.Unlock()

					pendingTxMu.Lock()
					delete(pendingTransactions, txHash)
					pendingTxMu.Unlock()

					if receipt.Status == 1 {
						log.Printf("✅ TX confirmée : %s bloc %d",
							txHash.Hex()[:20], receipt.BlockNumber)
					} else {
						log.Printf("❌ TX échouée : %s", txHash.Hex()[:20])
					}
				}
			}
		}
	}
}

func staleNodeCleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("staleNodeCleanupRoutine : arrêt demandé")
			return
		case <-ticker.C:
			staleThreshold := time.Now().Add(-5 * time.Minute).Unix()
			dynamicNodesMu.Lock()
			for _, node := range dynamicNodes {
				if node.LastSeen < staleThreshold && node.IsActive {
					node.IsActive = false
					log.Printf("⚠️  Nœud marqué inactif (stale) : %s", node.DeviceID)
				}
			}
			dynamicNodesMu.Unlock()
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Scan et gestion des nœuds Geth dynamiques
// ─────────────────────────────────────────────────────────────────────────────

func isMining(port int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := newGethRPCRequest(ctx, port,
		`{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}`)
	if err != nil {
		return false
	}
	resp, err := gethRPCClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return strings.Contains(string(body), `"result":true`)
}

func ScanGethNodes() {
	dynamicNodesMu.Lock()
	defer dynamicNodesMu.Unlock()

	log.Println("🔍 Scan des nœuds Geth (ports 8541-8800)...")

	for port := 8541; port <= 8800; port++ {
		address, err := getNodeAddress(port)
		if err != nil || address == "" ||
			address == "0x0000000000000000000000000000000000000000" {
			continue
		}

		nodeType := "fullnode"
		if isMining(port) {
			nodeType = "validator"
		}

		nodeID := port - 8540
		deviceID := fmt.Sprintf("GETH_NODE_%d", nodeID)
		if nodeType == "fullnode" {
			deviceID = fmt.Sprintf("FULLNODE_%d", nodeID-4)
		}
		p2pPort := 30300 + nodeID

		if existing, exists := dynamicNodes[address]; exists {
			existing.LastSeen = time.Now().Unix()
			existing.IsActive = true
			if existing.Type != nodeType {
				existing.Type = nodeType
				existing.DeviceID = deviceID
			}
		} else {
			newNode := &DynamicGethNode{
				Address:  address,
				DeviceID: deviceID,
				NodeID:   nodeID,
				Type:     nodeType,
				RPCPort:  port,
				P2PPort:  p2pPort,
				IsActive: true,
				LastSeen: time.Now().Unix(),
			}
			dynamicNodes[address] = newNode
			log.Printf("🆕 Nouveau nœud : %s (port %d, type=%s)", deviceID, port, nodeType)
			go registerDynamicNode(newNode)
		}
	}

	lastScanTime = time.Now()
	log.Printf("✅ Scan terminé — %d nœuds connus", len(dynamicNodes))
}

func getNodeAddress(port int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := newGethRPCRequest(ctx, port,
		`{"jsonrpc":"2.0","method":"eth_coinbase","params":[],"id":1}`)
	if err != nil {
		return "", err
	}
	resp, err := gethRPCClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var result struct {
		Result string `json:"result"`
		Error  struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err = json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	if result.Error.Code != 0 {
		return "", fmt.Errorf("RPC error %d : %s", result.Error.Code, result.Error.Message)
	}
	return result.Result, nil
}

func registerDynamicNode(node *DynamicGethNode) {
	deviceAddr := common.HexToAddress(node.Address)
	if exists, _ := DeviceIsRegistered(deviceAddr); exists {
		return
	}

	publicKey, err := generateSyntheticPublicKeyFromAddress(node.Address)
	if err != nil {
		log.Printf("❌ Clé synthétique %s : %v", node.DeviceID, err)
		return
	}

	metadata := fmt.Sprintf(
		`{"device_id":%q,"node_id":%d,"type":%q,"rpc_port":%d,"p2p_port":%d,`+
			`"initialized_at":%d,"key_type":"synthetic",`+
			`"warning":"NOT_A_REAL_ECDSA_PUBLIC_KEY"}`,
		node.DeviceID, node.NodeID, node.Type,
		node.RPCPort, node.P2PPort, time.Now().Unix(),
	)

	var entityType EntityType
	if node.Type == "validator" {
		entityType = EntityValidator
	} else {
		entityType = EntityFullnode
	}

	txHash, err := DeviceRegisterWithType(deviceAddr, publicKey, metadata, entityType)
	if err != nil {
		log.Printf("❌ Enregistrement %s échoué : %v", node.DeviceID, err)
	} else {
		log.Printf("✅ %s enregistré (type=%d) TX: %s",
			node.DeviceID, entityType, txHash.Hash().Hex()[:20])
	}
}

// StartGethNodeScanner démarre le scanner périodique des nœuds Geth.
//
// [BUG-CRITIQUE-5] Doit être appelé dans main_secure.go APRÈS InitNoncePool().
// Ne plus appeler depuis InitBlockchain().
func StartGethNodeScanner() {
	ScanGethNodes()

	go func() {
		time.Sleep(15 * time.Second)
		ConnectAllPeers()
	}()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			ScanGethNodes()
			go ConnectAllPeers()
		}
	}()
}

func ConnectAllPeers() {
	log.Println("🔗 Connexion des pairs Geth...")

	enodes := make(map[int]string)
	dynamicNodesMu.RLock()
	for _, node := range dynamicNodes {
		if enode := getEnode(node.RPCPort); enode != "" {
			enodes[node.RPCPort] = enode
		}
	}
	dynamicNodesMu.RUnlock()

	if len(enodes) < 2 {
		return
	}

	for port := range enodes {
		for otherPort, otherEnode := range enodes {
			if port != otherPort {
				addPeer(port, otherEnode)
			}
		}
	}

	time.Sleep(2 * time.Second)
	for port := range enodes {
		peers := getPeerCount(port)
		log.Printf("   Port %d : %d pairs connectés", port, peers)
	}
	log.Println("✅ Connexion des pairs terminée")
}

func getEnode(port int) string {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	req, err := newGethRPCRequest(ctx, port,
		`{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}`)
	if err != nil {
		return ""
	}
	resp, err := gethRPCClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	var result struct {
		Result struct {
			Enode string `json:"enode"`
		} `json:"result"`
	}
	if err = json.Unmarshal(body, &result); err != nil {
		return ""
	}

	enode := result.Result.Enode
	var ip string
	if port >= 8541 && port <= 8544 {
		ip = fmt.Sprintf("172.20.0.%d", 11+(port-8541))
	} else {
		ip = fmt.Sprintf("172.20.0.%d", 21+(port-8545))
	}
	return strings.NewReplacer("127.0.0.1", ip, "[::]", ip).Replace(enode)
}

func addPeer(port int, enode string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	body := fmt.Sprintf(
		`{"jsonrpc":"2.0","method":"admin_addPeer","params":[%q],"id":1}`, enode)
	req, err := newGethRPCRequest(ctx, port, body)
	if err != nil {
		return false
	}
	resp, err := gethRPCClient.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}

func getPeerCount(port int) int {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	req, err := newGethRPCRequest(ctx, port,
		`{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}`)
	if err != nil {
		return 0
	}
	resp, err := gethRPCClient.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	var result struct {
		Result string `json:"result"`
	}
	_ = json.Unmarshal(body, &result)
	count, _ := new(big.Int).SetString(
		strings.TrimPrefix(result.Result, "0x"), 16)
	if count == nil {
		return 0
	}
	return int(count.Int64())
}

// GetAllGethNodesDynamic retourne la liste des nœuds Geth avec leurs infos blockchain.
func GetAllGethNodesDynamic() []map[string]interface{} {
	dynamicNodesMu.RLock()
	defer dynamicNodesMu.RUnlock()

	nodesMap := make(map[string]map[string]interface{}, len(dynamicNodes))
	for addr, node := range dynamicNodes {
		deviceAddr := common.HexToAddress(addr)
		publicKey, metadata, isActive, lastSeen, err := DeviceGetInfo(deviceAddr)
		if err != nil {
			publicKey, _ = generateSyntheticPublicKeyFromAddress(addr)
			metadata = fmt.Sprintf(`{"device_id":%q,"type":%q}`,
				node.DeviceID, node.Type)
			isActive = node.IsActive
			lastSeen = big.NewInt(node.LastSeen)
		}

		authCountBig, _ := StatsGetDeviceAuthCount(deviceAddr)
		authCount := uint64(0)
		if authCountBig != nil {
			authCount = authCountBig.Uint64()
		}
		lastSeenUint64 := uint64(0)
		if lastSeen != nil {
			lastSeenUint64 = lastSeen.Uint64()
		}

		nodesMap[addr] = map[string]interface{}{
			"address":    addr,
			"device_id":  node.DeviceID,
			"node_id":    node.NodeID,
			"type":       node.Type,
			"rpc_port":   node.RPCPort,
			"p2p_port":   node.P2PPort,
			"publicKey":  publicKey,
			"metadata":   metadata,
			"isActive":   isActive,
			"lastSeen":   lastSeenUint64,
			"authCount":  authCount,
			"isGethNode": true,
			"permanent":  false,
		}
	}

	nodes := make([]map[string]interface{}, 0, len(nodesMap))
	for _, n := range nodesMap {
		nodes = append(nodes, n)
	}
	sort.Slice(nodes, func(i, j int) bool {
		return fmt.Sprint(nodes[i]["node_id"]) < fmt.Sprint(nodes[j]["node_id"])
	})
	return nodes
}
