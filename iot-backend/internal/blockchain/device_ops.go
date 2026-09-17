// internal/blockchain/device_ops.go
// FICHIER CORRIGÉ v5.3
//
// CORRECTIONS v5.2 → v5.3 :
//
//   [FIX-NONCE-15] DeviceRegisterWithType : utilise ReleaseNonceSync() au lieu de
//                  ReleaseNonce() pour les erreurs de nonce détectées AVANT soumission.
//
//                  PROBLÈME dans v5.2 :
//                    Quand RegisterDevice reçoit "nonce too low" (erreur renvoyée par
//                    go-ethereum AVANT que la tx atteigne le mempool), ReleaseNonce()
//                    était appelé. Celui-ci lance une goroutine avec un sleep de 2s,
//                    ce qui est conçu pour laisser Geth purger une tx DÉJÀ SOUMISE.
//                    Mais ici la tx n'a JAMAIS été soumise → les 2s sont inutiles et
//                    créent une fenêtre pendant laquelle le retry suivant peut
//                    acquérir un nonce erroné (avant la fin du reset).
//
//                  SOLUTION : appeler ReleaseNonceSync() — synchrone, 700ms — pour
//                    les cas où la tx n'a pas été soumise (erreur go-ethereum locale
//                    comme "nonce too low"). Le TxSemaphore (déjà tenu par
//                    DeviceRegisterWithType) garantit qu'aucune autre goroutine write
//                    ne peut interférer pendant le reset.
//
//                  DISTINCTION :
//                    - "nonce too low" renvoyé par go-ethereum CLIENT → tx pas soumise
//                      → ReleaseNonceSync (700ms, pas de purge mempool nécessaire)
//                    - "replacement transaction underpriced" → tx peut être dans le
//                      mempool → ReleaseNonce async (2s, laisse Geth purger)
//
//   Toutes les corrections v5.2 sont conservées (FIX-NONCE-8/9/10/12).

package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ─────────────────────────────────────────────────────────────────────────────
// TxSemaphore — sérialise TOUTES les transactions blockchain write (capacity=1)
//
// [FIX-NONCE-12] : exporté pour être partagé avec handlers/device.go.
//
// Utilisé par :
//   - DeviceRegisterWithType (ce fichier)
//   - ActivateDevice handler (handlers/device.go)   [FIX-NONCE-13]
//   - DeactivateDevice handler (handlers/device.go)  [FIX-NONCE-14]
//
// Garantit qu'une seule tx write est en vol à la fois, éliminant les gaps
// de nonce causés par des opérations concurrentes qui échouent avant soumission.
//
// Capacité 1 = mutex non réentrant : 1 goroutine à la fois.
// Les autres bloquent sur l'envoi jusqu'à libération du slot.
//
// NOTE : AuthRecordWithMeta N'utilise PAS ce sémaphore volontairement.
// Les auth sont trop fréquentes pour être sérialisées (200 devices = 200 auth/15s).
// Leur retry logic interne gère les conflits de nonce.
// ─────────────────────────────────────────────────────────────────────────────

var TxSemaphore = make(chan struct{}, 1)

// ─────────────────────────────────────────────────────────────────────────────
// AuthenticationRecord — structure retournée par GetRecentAuthentications
// ─────────────────────────────────────────────────────────────────────────────

type AuthenticationRecord struct {
	TxHash        string
	DeviceID      string
	DeviceAddress string
	Success       bool
	Message       string
	Timestamp     int64
	BlockNumber   uint64
	Sequence      uint64
}

// ─────────────────────────────────────────────────────────────────────────────
// Cache local des authentifications (en mémoire)
// ─────────────────────────────────────────────────────────────────────────────

var (
	authCache    []AuthenticationRecord
	authCacheMu  sync.RWMutex
	maxCacheSize = 10000
)

// ─────────────────────────────────────────────────────────────────────────────
// DeviceIsRegistered — vérifie l'enregistrement on-chain
// ─────────────────────────────────────────────────────────────────────────────

func DeviceIsRegistered(address common.Address) (bool, error) {
	if Contract == nil {
		return false, fmt.Errorf("contrat non initialisé")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return Contract.IsDeviceRegistered(&bind.CallOpts{Context: ctx}, address)
}

// ─────────────────────────────────────────────────────────────────────────────
// IsDeviceActive — vérifie si le device est actif on-chain
// ─────────────────────────────────────────────────────────────────────────────

func IsDeviceActive(address common.Address) (bool, error) {
	if Contract == nil {
		return false, fmt.Errorf("contrat non initialisé")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	info, err := Contract.Devices(&bind.CallOpts{Context: ctx}, address)
	if err != nil {
		return false, fmt.Errorf("IsDeviceActive: %w", err)
	}
	return info.IsActive, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// DeviceGetInfo — lit toutes les infos d'un device
// ─────────────────────────────────────────────────────────────────────────────

func DeviceGetInfo(address common.Address) (publicKey, metadata string, isActive bool, lastSeen *big.Int, err error) {
	if Contract == nil {
		return "", "", false, nil, fmt.Errorf("contrat non initialisé")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	info, e := Contract.GetDeviceInfo(&bind.CallOpts{Context: ctx}, address)
	if e != nil {
		return "", "", false, nil, fmt.Errorf("DeviceGetInfo: %w", e)
	}
	return info.PublicKey, info.Metadata, info.IsActive, info.LastSeen, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// DeviceRegisterWithType — enregistre un device avec sérialisation et retry
//
// [FIX-NONCE-8]  TxSemaphore : 1 seule tx write active à la fois.
// [FIX-NONCE-9]  Backoff doublé : 4s, 8s + 500ms post-reset.
// [FIX-NONCE-10] Gas price bump +15% par retry.
// [FIX-NONCE-12] TxSemaphore partagé avec ActivateDevice/DeactivateDevice handlers.
// [FIX-NONCE-15] ReleaseNonceSync pour "nonce too low" (tx pas soumise),
//                ReleaseNonce async pour "replacement underpriced" (tx peut être
//                dans le mempool et doit être purgée par Geth).
// ─────────────────────────────────────────────────────────────────────────────

func DeviceRegisterWithType(
	address common.Address,
	publicKey string,
	metadata string,
	entityType EntityType,
) (*types.Transaction, error) {
	if Contract == nil {
		return nil, fmt.Errorf("contrat non initialisé")
	}

	// [FIX-NONCE-8/12] Acquérir le sémaphore global.
	// Bloque si ActivateDevice ou DeactivateDevice est en cours.
	TxSemaphore <- struct{}{}
	defer func() { <-TxSemaphore }()

	const maxRetries = 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// [FIX-NONCE-9] Backoff doublé : 4s puis 8s.
			waitTime := time.Duration(4<<uint(attempt-1)) * time.Second
			fmt.Printf("device_ops: DeviceRegisterWithType retry %d/%d dans %v (adresse: %s)\n",
				attempt+1, maxRetries, waitTime, address.Hex()[:12])
			time.Sleep(waitTime)

			if err := ResetNoncePool(); err != nil {
				fmt.Printf("device_ops: ResetNoncePool error: %v\n", err)
			}
			time.Sleep(500 * time.Millisecond)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		auth, err := GetSecureTransactOpts(ctx)
		if err != nil {
			cancel()
			lastErr = fmt.Errorf("GetSecureTransactOpts (attempt %d): %w", attempt+1, err)
			continue
		}

		// [FIX-NONCE-10] Bump du gas price pour les retries (évite "replacement underpriced").
		if attempt > 0 && auth.GasPrice != nil {
			bumpPercent := int64(15 * attempt)
			bump := new(big.Int).Mul(auth.GasPrice, big.NewInt(bumpPercent))
			bump.Div(bump, big.NewInt(100))
			auth.GasPrice = new(big.Int).Add(auth.GasPrice, bump)
			fmt.Printf("device_ops: Gas price bumped +%d%% → %s gwei (attempt %d)\n",
				bumpPercent, new(big.Int).Div(auth.GasPrice, big.NewInt(1e9)), attempt+1)
		}

		auth.GasLimit = uint64(0)
		estimatedGas, estimErr := estimateRegisterGas(ctx, address, publicKey, metadata, entityType)
		if estimErr != nil {
			auth.GasLimit = uint64(800_000)
			fmt.Printf("device_ops: estimateRegisterGas error (fallback 800k): %v\n", estimErr)
		} else {
			buffered := estimatedGas * 3 / 2
			if buffered < 600_000 {
				buffered = 600_000
			}
			if buffered > 3_000_000 {
				buffered = 3_000_000
			}
			auth.GasLimit = buffered
		}

		tx, err := Contract.RegisterDevice(auth, address, publicKey, metadata, uint8(entityType))
		cancel()

		if err != nil {
			errMsg := err.Error()

			if strings.Contains(errMsg, "nonce too low") ||
				strings.Contains(errMsg, "already known") {
				// [FIX-NONCE-15] "nonce too low" : erreur go-ethereum locale, la tx n'a
				// PAS été soumise au mempool Geth. ReleaseNonceSync resynchronise en 700ms
				// sans délai de purge inutile.
				if auth.Nonce != nil {
					ReleaseNonceSync(auth.Nonce.Uint64())
				}
				lastErr = fmt.Errorf("RegisterDevice (attempt %d, nonce too low): %w", attempt+1, err)
				fmt.Printf("device_ops: nonce too low détecté → ReleaseNonceSync, retry: %v\n", errMsg)
				continue
			}

			if strings.Contains(errMsg, "replacement transaction underpriced") {
				// [FIX-NONCE-15] "replacement underpriced" : la tx PRÉCÉDENTE est encore
				// dans le mempool. Geth doit la purger avant qu'un nouveau nonce identique
				// soit accepté. ReleaseNonce async (2s) laisse ce délai.
				if auth.Nonce != nil {
					ReleaseNonce(auth.Nonce.Uint64())
				}
				lastErr = fmt.Errorf("RegisterDevice (attempt %d, replacement underpriced): %w", attempt+1, err)
				fmt.Printf("device_ops: replacement underpriced → ReleaseNonce async, retry: %v\n", errMsg)
				continue
			}

			// Erreur non-retriable (revert contrat, etc.).
			if auth.Nonce != nil {
				ReleaseNonceSync(auth.Nonce.Uint64())
			}
			return nil, fmt.Errorf("RegisterDevice: %w", err)
		}

		ConfirmTransaction(auth.Nonce.Uint64())
		fmt.Printf("device_ops: tx soumise nonce=%d gas=%d hash=%s\n",
			auth.Nonce.Uint64(), auth.GasLimit, tx.Hash().Hex()[:20])
		return tx, nil
	}

	return nil, fmt.Errorf("DeviceRegisterWithType: échec après %d tentatives: %w", maxRetries, lastErr)
}

// estimateRegisterGas estime le gas via eth_estimateGas.
// [FIX-NONCE-3] Timeout augmenté à 15s.
func estimateRegisterGas(
	ctx context.Context,
	address common.Address,
	publicKey string,
	metadata string,
	entityType EntityType,
) (uint64, error) {
	if Client == nil || Contract == nil {
		return 0, fmt.Errorf("client ou contrat non initialisé")
	}

	abi, err := IoTDeviceManagerMetaData.GetAbi()
	if err != nil {
		return 0, fmt.Errorf("ABI: %w", err)
	}

	calldata, err := abi.Pack("registerDevice", address, publicKey, metadata, uint8(entityType))
	if err != nil {
		return 0, fmt.Errorf("pack ABI: %w", err)
	}

	from := SecureKey.GetAddress()
	contractAddr := ContractAddress

	estCtx, estCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer estCancel()

	msg := ethereum.CallMsg{
		From: from,
		To:   &contractAddr,
		Data: calldata,
	}

	gas, err := Client.EstimateGas(estCtx, msg)
	if err != nil {
		return 0, fmt.Errorf("EstimateGas: %w", err)
	}
	return gas, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// StatsGetDeviceAuthCount
// ─────────────────────────────────────────────────────────────────────────────

func StatsGetDeviceAuthCount(address common.Address) (*big.Int, error) {
	if Contract == nil {
		return big.NewInt(0), fmt.Errorf("contrat non initialisé")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := Contract.GetDeviceAuthCount(&bind.CallOpts{Context: ctx}, address)
	if err != nil {
		return big.NewInt(0), nil
	}
	return count, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// AuthRecord — enregistre une authentification on-chain (sans sémaphore)
// ─────────────────────────────────────────────────────────────────────────────

func AuthRecord(
	address common.Address,
	success bool,
	proofHash [32]byte,
) (common.Hash, error) {
	if Contract == nil {
		return common.Hash{}, fmt.Errorf("contrat non initialisé")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	auth, err := GetSecureTransactOpts(ctx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("GetSecureTransactOpts: %w", err)
	}

	tx, err := Contract.RecordAuthentication(auth, address, success, proofHash)
	if err != nil {
		if auth.Nonce != nil {
			ReleaseNonce(auth.Nonce.Uint64())
		}
		return common.Hash{}, fmt.Errorf("RecordAuthentication: %w", err)
	}

	ConfirmTransaction(auth.Nonce.Uint64())
	txHash := tx.Hash()

	record := AuthenticationRecord{
		TxHash:        txHash.Hex(),
		DeviceID:      address.Hex()[:12],
		DeviceAddress: strings.ToLower(address.Hex()),
		Success:       success,
		Message:       "",
		Timestamp:     time.Now().Unix(),
		BlockNumber:   0,
		Sequence:      0,
	}
	authCacheMu.Lock()
	authCache = append([]AuthenticationRecord{record}, authCache...)
	if len(authCache) > maxCacheSize {
		authCache = authCache[:maxCacheSize]
	}
	authCacheMu.Unlock()

	return txHash, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// AuthRecordWithMeta — version enrichie avec retry sur nonce error.
//
// NOTE : pas de TxSemaphore ici — les auth sont très fréquentes et le retry
// interne gère les conflits sans bloquer toutes les autres authentifications.
// ─────────────────────────────────────────────────────────────────────────────

func AuthRecordWithMeta(
	address common.Address,
	deviceID string,
	message string,
	sequence uint64,
	success bool,
	proofHash [32]byte,
) (common.Hash, error) {
	if Contract == nil {
		return common.Hash{}, fmt.Errorf("contrat non initialisé")
	}

	const maxRetries = 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			waitTime := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(waitTime)

			if attempt == 2 {
				_ = ResetNoncePool()
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		auth, err := GetSecureTransactOpts(ctx)
		if err != nil {
			cancel()
			lastErr = fmt.Errorf("GetSecureTransactOpts: %w", err)
			continue
		}

		tx, err := Contract.RecordAuthentication(auth, address, success, proofHash)
		if err != nil {
			cancel()
			if strings.Contains(err.Error(), "nonce too low") ||
				strings.Contains(err.Error(), "replacement transaction underpriced") {
				if auth.Nonce != nil {
					ReleaseNonce(auth.Nonce.Uint64())
				}
				lastErr = fmt.Errorf("RecordAuthentication: %w", err)
				continue
			}
			if auth.Nonce != nil {
				ReleaseNonce(auth.Nonce.Uint64())
			}
			cancel()
			return common.Hash{}, fmt.Errorf("RecordAuthentication: %w", err)
		}

		ConfirmTransaction(auth.Nonce.Uint64())
		txHash := tx.Hash()
		cancel()

		record := AuthenticationRecord{
			TxHash:        txHash.Hex(),
			DeviceID:      deviceID,
			DeviceAddress: strings.ToLower(address.Hex()),
			Success:       success,
			Message:       message,
			Timestamp:     time.Now().Unix(),
			BlockNumber:   0,
			Sequence:      sequence,
		}
		authCacheMu.Lock()
		authCache = append([]AuthenticationRecord{record}, authCache...)
		if len(authCache) > maxCacheSize {
			authCache = authCache[:maxCacheSize]
		}
		authCacheMu.Unlock()

		return txHash, nil
	}

	return common.Hash{}, fmt.Errorf("échec après %d tentatives: %w", maxRetries, lastErr)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetRecentAuthentications
// ─────────────────────────────────────────────────────────────────────────────

func GetRecentAuthentications(limit int) ([]AuthenticationRecord, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > maxCacheSize {
		limit = maxCacheSize
	}

	// Lire le cache in-memory (alimenté en temps réel par AuthRecordWithMeta)
	authCacheMu.RLock()
	cached := make([]AuthenticationRecord, len(authCache))
	copy(cached, authCache)
	authCacheMu.RUnlock()

	// [FIX-1] Ne PAS court-circuiter si cache >= limit :
	// après restart le cache est vide mais la blockchain a des données.
	// On scanne toujours la blockchain et on fusionne avec le cache.

	if Contract == nil || Client == nil {
		fmt.Println("GetRecentAuthentications: Contract ou Client nil, retour cache uniquement")
		return cached, nil
	}

	// [FIX-2] Timeout étendu 15s -> 60s pour les grandes fenêtres de blocs
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	currentBlock, err := Client.BlockNumber(ctx)
	if err != nil {
		fmt.Printf("GetRecentAuthentications: erreur BlockNumber: %v\n", err)
		return cached, nil
	}

	// [FIX-3] Scan depuis le bloc 0 pour récupérer tout l'historique.
	// Sur blockchain PoA locale le genesis est bas, pas de risque de timeout.
	fromBlock := uint64(0)
	fmt.Printf("GetRecentAuthentications: scan events bloc %d -> %d\n", fromBlock, currentBlock)

	filterOpts := &bind.FilterOpts{
		Start:   fromBlock,
		End:     &currentBlock,
		Context: ctx,
	}

	iter, err := Contract.FilterAuthenticationRecorded(filterOpts, nil)
	if err != nil {
		// [FIX-4] Logger l'erreur au lieu de l'ignorer silencieusement
		fmt.Printf("GetRecentAuthentications: FilterAuthenticationRecorded erreur: %v\n", err)
		return cached, nil
	}
	defer iter.Close()

	cachedHashes := make(map[string]bool, len(cached))
	for _, r := range cached {
		cachedHashes[r.TxHash] = true
	}

	var fromChain []AuthenticationRecord
	for iter.Next() {
		event := iter.Event
		txHash := event.Raw.TxHash.Hex()
		if cachedHashes[txHash] {
			continue
		}
		fromChain = append(fromChain, AuthenticationRecord{
			TxHash:        txHash,
			DeviceID:      event.DeviceAddress.Hex()[:12],
			DeviceAddress: strings.ToLower(event.DeviceAddress.Hex()),
			Success:       event.Success,
			Timestamp:     event.Timestamp.Int64(),
			BlockNumber:   event.Raw.BlockNumber,
		})
	}

	fmt.Printf("GetRecentAuthentications: cache=%d blockchain=%d\n", len(cached), len(fromChain))

	result := append(cached, fromChain...)
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// IsProtectedNode
// ─────────────────────────────────────────────────────────────────────────────

func IsProtectedNode(address common.Address) bool {
	dynamicNodesMu.RLock()
	defer dynamicNodesMu.RUnlock()
	_, exists := dynamicNodes[strings.ToLower(address.Hex())]
	return exists
}

// ─────────────────────────────────────────────────────────────────────────────
// GetAllGethNodes
// ─────────────────────────────────────────────────────────────────────────────

func GetAllGethNodes() []map[string]interface{} {
	return GetAllGethNodesDynamic()
}

// ─────────────────────────────────────────────────────────────────────────────
// GetTransactionAuth
// ─────────────────────────────────────────────────────────────────────────────

func GetTransactionAuth() (*bind.TransactOpts, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return GetSecureTransactOpts(ctx)
}
