package zkp

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/ecc"
	bn254mimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var bn254Modulus, _ = new(big.Int).SetString(
	"21888242871839275222246405745257275088548364400416034343698204186575808495617", 
	10,
)

type SecureZKPManager struct {
	provingKey       groth16.ProvingKey
	verifyingKey     groth16.VerifyingKey
	constraintSystem constraint.ConstraintSystem
	initialized      bool
	initMu           sync.Mutex
}

var (
	secureInstance *SecureZKPManager
	secureOnce     sync.Once
)

func GetSecureZKPManager() (*SecureZKPManager, error) {
	var initErr error
	secureOnce.Do(func() {
		secureInstance = &SecureZKPManager{}
		if err := secureInstance.Setup(); err != nil {
			initErr = err
			secureInstance = nil
		}
	})
	if initErr != nil {
		return nil, initErr
	}
	return secureInstance, nil
}

func (zm *SecureZKPManager) Setup() error {
	zm.initMu.Lock()
	defer zm.initMu.Unlock()
	if zm.initialized {
		return nil
	}

	var circuit SecureAuthCircuit
	cs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		return fmt.Errorf("erreur compilation: %v", err)
	}

	pk, vk, err := groth16.Setup(cs)
	if err != nil {
		return fmt.Errorf("erreur setup: %v", err)
	}

	zm.provingKey = pk
	zm.verifyingKey = vk
	zm.constraintSystem = cs
	zm.initialized = true
	return nil
}

func GenerateSecureSecret(signature []byte) *big.Int {
	hash := crypto.Keccak256Hash(signature)
	secret := new(big.Int).SetBytes(hash.Bytes())
	secret.Mod(secret, bn254Modulus)
	if secret.Sign() == 0 {
		secret.SetInt64(1)
	}
	return secret
}

func GenerateChallenge(timestamp int64, sequence uint64, randomNonce string) *big.Int {
	data := fmt.Sprintf("%d:%d:%s", timestamp, sequence, randomNonce)
	hash := crypto.Keccak256Hash([]byte(data))
	challenge := new(big.Int).SetBytes(hash.Bytes())
	challenge.Mod(challenge, bn254Modulus)
	return challenge
}

func ComputePublicHash(secret *big.Int, address common.Address, challenge *big.Int) *big.Int {
	secretMod := new(big.Int).Mod(secret, bn254Modulus)
	addrInt := new(big.Int).SetBytes(address.Bytes())
	addrMod := new(big.Int).Mod(addrInt, bn254Modulus)
	challengeMod := new(big.Int).Mod(challenge, bn254Modulus)

	mimcHasher := bn254mimc.NewMiMC()
	mimcHasher.Write(secretMod.Bytes())
	mimcHasher.Write(addrMod.Bytes())
	mimcHasher.Write(challengeMod.Bytes())
	
	hash := mimcHasher.Sum(nil)
	result := new(big.Int).SetBytes(hash)
	result.Mod(result, bn254Modulus)
	return result
}

func (zm *SecureZKPManager) GenerateProof(secret *big.Int, address common.Address, challenge *big.Int) ([]byte, *big.Int, error) {
	zm.initMu.Lock()
	defer zm.initMu.Unlock()
	if !zm.initialized {
		return nil, nil, fmt.Errorf("service non initialisé")
	}

	secretMod := new(big.Int).Mod(secret, bn254Modulus)
	addrInt := new(big.Int).SetBytes(address.Bytes())
	addrMod := new(big.Int).Mod(addrInt, bn254Modulus)
	challengeMod := new(big.Int).Mod(challenge, bn254Modulus)

	expectedHash := ComputePublicHash(secretMod, address, challengeMod)

	assignment := &SecureAuthCircuit{
		Secret:       secretMod,
		Address:      addrMod,
		Challenge:    challengeMod,
		ExpectedHash: expectedHash,
	}

	witness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField())
	if err != nil {
		return nil, nil, fmt.Errorf("erreur witness: %v", err)
	}

	proof, err := groth16.Prove(zm.constraintSystem, zm.provingKey, witness)
	if err != nil {
		return nil, nil, fmt.Errorf("erreur preuve: %v", err)
	}

	var buf bytes.Buffer
	_, err = proof.WriteTo(&buf)
	if err != nil {
		return nil, nil, fmt.Errorf("erreur sérialisation: %v", err)
	}

	return buf.Bytes(), expectedHash, nil
}

func (zm *SecureZKPManager) VerifyProof(proofBytes []byte, address common.Address, challenge *big.Int, expectedHash *big.Int) (bool, error) {
	zm.initMu.Lock()
	defer zm.initMu.Unlock()
	if !zm.initialized {
		return false, fmt.Errorf("service non initialisé")
	}
	if len(proofBytes) == 0 {
		return false, fmt.Errorf("preuve vide")
	}

	proof := groth16.NewProof(ecc.BN254)
	buf := bytes.NewReader(proofBytes)
	_, err := proof.ReadFrom(buf)
	if err != nil {
		return false, fmt.Errorf("désérialisation: %v", err)
	}

	addrInt := new(big.Int).SetBytes(address.Bytes())
	addrMod := new(big.Int).Mod(addrInt, bn254Modulus)
	challengeMod := new(big.Int).Mod(challenge, bn254Modulus)

	publicAssignment := &SecureAuthCircuit{
		Address:      addrMod,
		Challenge:    challengeMod,
		ExpectedHash: expectedHash,
	}

	witness, err := frontend.NewWitness(publicAssignment, ecc.BN254.ScalarField(), frontend.PublicOnly())
	if err != nil {
		return false, fmt.Errorf("witness public: %v", err)
	}

	err = groth16.Verify(proof, zm.verifyingKey, witness)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// VerifyECDSASignature vérifie une signature ECDSA Ethereum
// Accepte v = 0/1 (standard) ou v = 27/28 (legacy)
func VerifyECDSASignature(address common.Address, messageHash string, signatureHex string) (bool, error) {
	sigBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false, fmt.Errorf("hex invalide: %v", err)
	}
	if len(sigBytes) != 65 {
		return false, fmt.Errorf("longueur invalide: %d (attendu 65)", len(sigBytes))
	}

	msgHashBytes, err := hex.DecodeString(messageHash)
	if err != nil {
		return false, fmt.Errorf("hash invalide: %v", err)
	}
	if len(msgHashBytes) != 32 {
		return false, fmt.Errorf("hash doit être 32 bytes, reçu %d", len(msgHashBytes))
	}

	var msgHash [32]byte
	copy(msgHash[:], msgHashBytes)

	// 🔴 CORRECTION CRITIQUE: Normaliser v de 27/28 à 0/1 si nécessaire
	// Ethereum legacy: v = 27/28, Standard: v = 0/1
	v := sigBytes[64]
	if v >= 27 {
		v -= 27
		sigBytes[64] = v
	}

	// Vérifier que v est maintenant 0 ou 1
	if v > 1 {
		return false, fmt.Errorf("recovery id invalide après normalisation: %d", v)
	}

	pubKey, err := crypto.SigToPub(msgHash[:], sigBytes)
	if err != nil {
		return false, fmt.Errorf("récupération clé: %v", err)
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	if recoveredAddr != address {
		return false, fmt.Errorf("adresse ne correspond pas: attendu %s, obtenu %s", 
			address.Hex(), recoveredAddr.Hex())
	}

	return true, nil
}
// pkg/zkp/service_secure.go
// Ajoutez cette fonction à la fin du fichier, après VerifyECDSASignature

// ============================================
// FONCTIONS DE COMPATIBILITÉ
// ============================================

// GetZKPService - Alias pour GetSecureZKPManager (compatibilité avec l'ancien code)
func GetZKPService() (*SecureZKPManager, error) {
    return GetSecureZKPManager()
}

// GetZKPService est un alias pour GetSecureZKPManager
// Utilisé par les anciens handlers
