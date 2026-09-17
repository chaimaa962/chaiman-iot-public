package zkp

import (
	"math/big"
	"testing"
)

func TestSchnorr_FullFlow_ValidProof(t *testing.T) {
	// 1. Le DEVICE génère sa paire de clés localement (une seule fois, au démarrage)
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("génération clé: %v", err)
	}

	deviceAddress := "IOT_0001"
	timestamp := int64(1735600000)
	sequence := uint64(42)
	nonce := "test-nonce-abc123"

	// 2. Le DEVICE génère une preuve (ne révèle jamais kp.PrivateKey)
	proof, err := kp.GenerateProof(deviceAddress, timestamp, sequence, nonce)
	if err != nil {
		t.Fatalf("génération preuve: %v", err)
	}

	// 3. Le SERVEUR — qui ne connaît QUE kp.PublicKey — vérifie la preuve
	valid, err := VerifyProof(kp.PublicKey, proof, deviceAddress, timestamp, sequence, nonce)
	if err != nil {
		t.Fatalf("erreur vérification: %v", err)
	}
	if !valid {
		t.Fatal("la preuve légitime a été REJETÉE — bug dans le protocole")
	}
	t.Log("✅ Preuve légitime acceptée")
}

func TestSchnorr_RejectsWrongPrivateKey(t *testing.T) {
	// Un attaquant qui n'a PAS la vraie clé privée essaie de forger une preuve
	realKP, _ := GenerateKeyPair()
	attackerKP, _ := GenerateKeyPair() // clé différente

	deviceAddress := "IOT_0001"
	timestamp := int64(1735600000)
	sequence := uint64(1)
	nonce := "nonce-1"

	// L'attaquant génère une preuve avec SA PROPRE clé privée
	forgedProof, err := attackerKP.GenerateProof(deviceAddress, timestamp, sequence, nonce)
	if err != nil {
		t.Fatalf("génération preuve attaquant: %v", err)
	}

	// Mais essaie de la faire vérifier contre la clé PUBLIQUE du vrai device
	valid, _ := VerifyProof(realKP.PublicKey, forgedProof, deviceAddress, timestamp, sequence, nonce)
	if valid {
		t.Fatal("SÉCURITÉ CASSÉE : une preuve forgée avec la mauvaise clé a été ACCEPTÉE")
	}
	t.Log("✅ Preuve forgée correctement rejetée")
}

func TestSchnorr_RejectsReplay(t *testing.T) {
	// Un attaquant capture une preuve valide et essaie de la rejouer avec un
	// nouveau sequence/nonce (simulateur d'attaque par rejeu)
	kp, _ := GenerateKeyPair()
	deviceAddress := "IOT_0001"

	originalProof, _ := kp.GenerateProof(deviceAddress, 1735600000, 1, "nonce-A")

	// Rejeu avec une séquence différente (comme UpdateSequence le ferait détecter)
	// mais on vérifie ici QUE la couche cryptographique elle-même rejette le rejeu
	// (défense en profondeur, indépendante de la vérification de séquence applicative)
	valid, _ := VerifyProof(kp.PublicKey, originalProof, deviceAddress, 1735600000, 2, "nonce-A")
	if valid {
		t.Fatal("SÉCURITÉ CASSÉE : preuve rejouée avec sequence différente acceptée")
	}
	t.Log("✅ Rejeu avec sequence modifiée correctement rejeté")
}

func TestSchnorr_RejectsTamperedProof(t *testing.T) {
	kp, _ := GenerateKeyPair()
	deviceAddress := "IOT_0001"
	proof, _ := kp.GenerateProof(deviceAddress, 1735600000, 1, "nonce-A")

	// Un attaquant intercepte la preuve et modifie légèrement S
	tampered := &SchnorrProof{
		TX: proof.TX,
		TY: proof.TY,
		S:  new(big.Int).Add(proof.S, big.NewInt(1)), // +1 — modification triviale
	}

	valid, _ := VerifyProof(kp.PublicKey, tampered, deviceAddress, 1735600000, 1, "nonce-A")
	if valid {
		t.Fatal("SÉCURITÉ CASSÉE : preuve modifiée acceptée")
	}
	t.Log("✅ Preuve modifiée (tampering) correctement rejetée")
}

func TestSchnorr_ServerNeverSeesPrivateKey(t *testing.T) {
	// Test de conception : vérifie que le store serveur ne peut physiquement
	// stocker qu'une SchnorrPublicKey, jamais SchnorrKeyPair complet.
	store := NewInMemoryKeyStore()
	kp, _ := GenerateKeyPair()

	// Ceci ne compilerait PAS si on essayait de faire store.Set("addr", kp)
	// car le type attendu est *SchnorrPublicKey, pas *SchnorrKeyPair —
	// c'est une garantie au niveau du système de types, pas juste une convention.
	store.Set("IOT_0001", kp.PublicKey)

	retrieved, ok := store.Get("IOT_0001")
	if !ok {
		t.Fatal("clé publique non trouvée après stockage")
	}
	if !bigIntsEqual(retrieved.X, kp.PublicKey.X) || !bigIntsEqual(retrieved.Y, kp.PublicKey.Y) {
		t.Fatal("clé publique récupérée ne correspond pas à l'originale")
	}
	t.Log("✅ Le serveur ne manipule que des clés publiques (garanti par le typage)")
}

func TestSchnorr_HexSerializationRoundTrip(t *testing.T) {
	kp, _ := GenerateKeyPair()
	hexStr := kp.PublicKey.Hex()

	recovered, err := PublicKeyFromHex(hexStr)
	if err != nil {
		t.Fatalf("désérialisation: %v", err)
	}
	if !bigIntsEqual(recovered.X, kp.PublicKey.X) || !bigIntsEqual(recovered.Y, kp.PublicKey.Y) {
		t.Fatal("round-trip hex a corrompu la clé publique")
	}

	// Vérifie qu'une preuve fonctionne toujours après le round-trip
	proof, _ := kp.GenerateProof("IOT_0001", 1735600000, 1, "n1")
	valid, err := VerifyProof(recovered, proof, "IOT_0001", 1735600000, 1, "n1")
	if err != nil || !valid {
		t.Fatalf("vérification après round-trip hex échouée: valid=%v err=%v", valid, err)
	}
	t.Log("✅ Sérialisation hex round-trip fonctionnelle")
}

func BenchmarkSchnorr_GenerateProof(b *testing.B) {
	kp, _ := GenerateKeyPair()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = kp.GenerateProof("IOT_0001", 1735600000, uint64(i), "bench-nonce")
	}
}

func BenchmarkSchnorr_VerifyProof(b *testing.B) {
	kp, _ := GenerateKeyPair()
	proof, _ := kp.GenerateProof("IOT_0001", 1735600000, 1, "bench-nonce")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = VerifyProof(kp.PublicKey, proof, "IOT_0001", 1735600000, 1, "bench-nonce")
	}
}
