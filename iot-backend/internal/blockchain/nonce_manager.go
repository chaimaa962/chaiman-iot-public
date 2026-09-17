package blockchain

// nonce_manager.go — VERSION CORRIGÉE v4.4
//
// CORRECTIONS v4.3 → v4.4 :
//
//   [FIX-NONCE-16] ReleaseNonceSync : log diagnostique amélioré — affiche le nonce
//                  libéré pour faciliter le traçage des gaps dans les logs de prod.
//
//                  RAPPEL du contrat d'utilisation (inchangé depuis v4.3) :
//                    Cette fonction DOIT être appelée en tenant blockchain.TxSemaphore.
//                    C'est désormais garanti par device.go v5.3 (ActivateDevice,
//                    DeactivateDevice) et device_ops.go v5.3 (DeviceRegisterWithType
//                    pour "nonce too low").
//
//   Toutes les corrections v4.3 sont conservées (FIX-NONCE-11).
//   Toutes les corrections v4.2 sont conservées (FIX-NONCE-5/6/7).

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// ─────────────────────────────────────────────────────────────────────────────
// Types internes
// ─────────────────────────────────────────────────────────────────────────────

type nonceRequest struct {
	ctx    context.Context
	respCh chan nonceResponse
}

type nonceResponse struct {
	opts *bind.TransactOpts
	err  error
}

type resetRequest struct {
	respCh chan error
}

// ─────────────────────────────────────────────────────────────────────────────
// NonceSerializer — propriétaire unique du compteur de nonce
// ─────────────────────────────────────────────────────────────────────────────

type NonceSerializer struct {
	account  common.Address
	reqCh    chan nonceRequest
	resetCh  chan resetRequest
	stopCh   chan struct{}
	stopOnce sync.Once
}

func NewNonceSerializer(account common.Address) *NonceSerializer {
	return &NonceSerializer{
		account: account,
		reqCh:   make(chan nonceRequest, 64),
		resetCh: make(chan resetRequest, 4),
		stopCh:  make(chan struct{}),
	}
}

// run est la goroutine owner — seul endroit où currentNonce est lu ou écrit.
func (ns *NonceSerializer) run(ctx context.Context) {
	var currentNonce uint64

	fetchNonce := func(opCtx context.Context) error {
		n, err := Client.PendingNonceAt(opCtx, ns.account)
		if err != nil {
			return fmt.Errorf("PendingNonceAt: %w", err)
		}
		currentNonce = n
		return nil
	}

	initCtx, initCancel := context.WithTimeout(ctx, 15*time.Second)
	if err := fetchNonce(initCtx); err != nil {
		fmt.Printf("nonce_manager: avertissement init: %v\n", err)
	}
	initCancel()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ns.stopCh:
			return

		case req := <-ns.resetCh:
			opCtx, opCancel := context.WithTimeout(ctx, 10*time.Second)
			req.respCh <- fetchNonce(opCtx)
			opCancel()

		case req := <-ns.reqCh:
			if req.ctx.Err() != nil {
				req.respCh <- nonceResponse{err: req.ctx.Err()}
				continue
			}

			opts, err := ns.buildOpts(req.ctx, currentNonce)
			if err != nil {
				req.respCh <- nonceResponse{err: err}
				continue
			}

			currentNonce++
			req.respCh <- nonceResponse{opts: opts}
		}
	}
}

// buildOpts construit des TransactOpts pour le nonce donné.
func (ns *NonceSerializer) buildOpts(ctx context.Context, nonce uint64) (*bind.TransactOpts, error) {
	gasPrice, err := Client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("SuggestGasPrice: %w", err)
	}

	maxGasPrice := new(big.Int).SetUint64(500_000_000_000)
	if gasPrice.Cmp(maxGasPrice) > 0 {
		gasPrice = maxGasPrice
	}

	privateKey, err := SecureKey.decryptPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("déchiffrement clé: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, ChainID)
	if err != nil {
		return nil, fmt.Errorf("NewKeyedTransactorWithChainID: %w", err)
	}

	auth.Nonce    = big.NewInt(int64(nonce))
	auth.Value    = big.NewInt(0)
	auth.GasLimit = uint64(0)
	auth.GasPrice = gasPrice

	return auth, nil
}

// buildOptsWithBump construit des TransactOpts avec un gas price majoré d'un pourcentage.
// Utilisé lors des retries pour éviter "replacement transaction underpriced".
func (ns *NonceSerializer) buildOptsWithBump(ctx context.Context, nonce uint64, bumpPercent int64) (*bind.TransactOpts, error) {
	gasPrice, err := Client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("SuggestGasPrice: %w", err)
	}

	if bumpPercent > 0 {
		bump := new(big.Int).Mul(gasPrice, big.NewInt(bumpPercent))
		bump.Div(bump, big.NewInt(100))
		gasPrice.Add(gasPrice, bump)
	}

	maxGasPrice := new(big.Int).SetUint64(500_000_000_000)
	if gasPrice.Cmp(maxGasPrice) > 0 {
		gasPrice = maxGasPrice
	}

	privateKey, err := SecureKey.decryptPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("déchiffrement clé: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, ChainID)
	if err != nil {
		return nil, fmt.Errorf("NewKeyedTransactorWithChainID: %w", err)
	}

	auth.Nonce    = big.NewInt(int64(nonce))
	auth.Value    = big.NewInt(0)
	auth.GasLimit = uint64(0)
	auth.GasPrice = gasPrice

	return auth, nil
}

func (ns *NonceSerializer) acquire(ctx context.Context) (*bind.TransactOpts, error) {
	respCh := make(chan nonceResponse, 1)
	req := nonceRequest{ctx: ctx, respCh: respCh}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case ns.reqCh <- req:
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-respCh:
		return resp.opts, resp.err
	}
}

func (ns *NonceSerializer) reset() error {
	respCh := make(chan error, 1)
	select {
	case ns.resetCh <- resetRequest{respCh: respCh}:
	case <-time.After(5 * time.Second):
		return fmt.Errorf("reset: timeout — sérialiseur trop occupé")
	}
	select {
	case err := <-respCh:
		return err
	case <-time.After(20 * time.Second):
		return fmt.Errorf("reset: timeout — pas de réponse de la blockchain")
	}
}

func (ns *NonceSerializer) stop() {
	ns.stopOnce.Do(func() { close(ns.stopCh) })
}

// ─────────────────────────────────────────────────────────────────────────────
// Singleton global
// ─────────────────────────────────────────────────────────────────────────────

var (
	globalSerializer   *NonceSerializer
	serializerOnce     sync.Once
	serializerCancelFn context.CancelFunc
	serializerMu       sync.Mutex
)

func InitNoncePoolWithContext(parentCtx context.Context) error {
	var initErr error
	serializerOnce.Do(func() {
		if SecureKey == nil {
			initErr = fmt.Errorf("SecureKey non initialisé — appeler InitBlockchain d'abord")
			return
		}
		if Client == nil {
			initErr = fmt.Errorf("Client Ethereum non initialisé — appeler InitBlockchain d'abord")
			return
		}

		account := SecureKey.GetAddress()
		globalSerializer = NewNonceSerializer(account)

		runCtx, cancel := context.WithCancel(parentCtx)
		serializerMu.Lock()
		serializerCancelFn = cancel
		serializerMu.Unlock()

		go globalSerializer.run(runCtx)

		time.Sleep(200 * time.Millisecond)
	})
	return initErr
}

func InitNoncePool() error {
	return InitNoncePoolWithContext(context.Background())
}

func StopNoncePool() {
	serializerMu.Lock()
	cancel := serializerCancelFn
	serializerMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// ResetNoncePool force une resynchronisation depuis la blockchain.
// [FIX-NONCE-4] Utilisé par device_ops.go avant chaque retry.
// [FIX-NONCE-7] Délai de stabilisation 200ms après le reset.
func ResetNoncePool() error {
	if globalSerializer == nil {
		return fmt.Errorf("nonce serializer non initialisé")
	}
	err := globalSerializer.reset()
	if err == nil {
		time.Sleep(200 * time.Millisecond)
	}
	return err
}

// ─────────────────────────────────────────────────────────────────────────────
// GetSecureTransactOpts — API publique principale
// ─────────────────────────────────────────────────────────────────────────────

func GetSecureTransactOpts(ctx context.Context) (*bind.TransactOpts, error) {
	if globalSerializer == nil {
		return nil, fmt.Errorf("nonce serializer non initialisé — appeler InitNoncePool d'abord")
	}

	const maxRetries = 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// [FIX-NONCE-6] Backoff doublé : 500ms, 1000ms (était 250ms, 500ms)
			wait := time.Duration(1<<uint(attempt)) * 500 * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
		}

		opts, err := globalSerializer.acquire(ctx)
		if err == nil {
			return opts, nil
		}

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if attempt == maxRetries-1 {
			return nil, fmt.Errorf("GetSecureTransactOpts: %w (après %d tentatives)", err, maxRetries)
		}
	}

	return nil, fmt.Errorf("GetSecureTransactOpts: échec inattendu")
}

// ─────────────────────────────────────────────────────────────────────────────
// ConfirmTransaction / ReleaseNonce / ReleaseNonceSync
// ─────────────────────────────────────────────────────────────────────────────

// ConfirmTransaction — no-op dans v4.
func ConfirmTransaction(nonce uint64) {
	_ = nonce
}

// ReleaseNonce — resynchronise le nonce depuis la blockchain après un rejet.
// VERSION ASYNCHRONE — pour les cas où une tx a été soumise puis rejetée par Geth
// (nonce too low, replacement underpriced). Le délai de 2s laisse Geth purger la tx.
//
// ATTENTION : N'utiliser que quand la tx a réellement été soumise au mempool.
// Pour les gas estimation failures (tx jamais soumise), utiliser ReleaseNonceSync.
//
// [FIX-NONCE-5] : délai augmenté 500ms → 2000ms (≈ 0.5 bloc PoA).
func ReleaseNonce(nonce uint64) {
	_ = nonce
	if globalSerializer != nil {
		go func() {
			time.Sleep(2 * time.Second)
			if err := globalSerializer.reset(); err != nil {
				fmt.Printf("nonce_manager: ReleaseNonce reset error: %v\n", err)
			}
		}()
	}
}

// ReleaseNonceSync — resynchronise le nonce de façon SYNCHRONE et BLOQUANTE.
//
// [FIX-NONCE-11] CORRECTION DU BUG PRINCIPAL :
//
// À utiliser quand une transaction a échoué AVANT soumission au mempool
// (typiquement : gas estimation revert = "already active", "already registered", etc.).
// Dans ce cas, Geth ne connaît pas la tx → PendingNonceAt retourne immédiatement
// le bon nonce → délai de 500ms suffit (pas besoin des 2s de purge txpool).
//
// IMPORTANT : Cette fonction DOIT être appelée en tenant le TxSemaphore
// (blockchain.TxSemaphore) pour garantir qu'aucune autre goroutine n'acquiert
// un nouveau nonce entre l'échec et le reset.
//
// Séquence correcte dans ActivateDevice/DeactivateDevice handlers :
//   blockchain.TxSemaphore <- struct{}{}     // acquérir
//   defer func() { <-blockchain.TxSemaphore }()
//   auth, _ := blockchain.GetTransactionAuth()  // nonce N alloué
//   tx, err := Contract.ActivateDevice(auth, ...)
//   if err != nil {
//       blockchain.ReleaseNonceSync(auth.Nonce.Uint64())  // reset synchrone
//       // Le TxSemaphore reste tenu jusqu'au defer → aucun gap possible
//   }
// ReleaseNonceSync resynchronise le nonce de façon SYNCHRONE et BLOQUANTE.
//
// [FIX-NONCE-11] CORRECTION DU BUG PRINCIPAL :
//
// À utiliser quand une transaction a échoué AVANT soumission au mempool
// (typiquement : gas estimation revert = "already active", "already registered",
// "nonce too low" local go-ethereum).
// Dans ce cas, Geth ne connaît pas la tx → PendingNonceAt retourne immédiatement
// le bon nonce → délai de 500ms suffit (pas besoin des 2s de purge txpool).
//
// CONTRAT : Cette fonction DOIT être appelée en tenant blockchain.TxSemaphore
// (acquis dans DeviceRegisterWithType, ActivateDevice, DeactivateDevice).
// Garantit qu'aucune autre goroutine write n'acquiert de nonce entre l'échec
// et la fin du reset.
//
// [FIX-NONCE-16] Log diagnostique : affiche le nonce libéré.
func ReleaseNonceSync(nonce uint64) {
	if globalSerializer == nil {
		return
	}
	fmt.Printf("nonce_manager: ReleaseNonceSync nonce=%d — resync synchrone\n", nonce)
	// Délai court : la tx n'a PAS été soumise, Geth n'a rien à purger.
	// 500ms suffit pour que le sérialiseur intègre le nouvel état.
	time.Sleep(500 * time.Millisecond)
	if err := globalSerializer.reset(); err != nil {
		fmt.Printf("nonce_manager: ReleaseNonceSync reset error (nonce=%d): %v\n", nonce, err)
		return
	}
	// Stabilisation post-reset (FIX-NONCE-7).
	time.Sleep(200 * time.Millisecond)
	fmt.Printf("nonce_manager: ReleaseNonceSync terminé (nonce=%d libéré)\n", nonce)
}

// ─────────────────────────────────────────────────────────────────────────────
// NoncePool — alias de type pour la rétrocompatibilité
// ─────────────────────────────────────────────────────────────────────────────

type NoncePool = NonceSerializer

var _ common.Address
