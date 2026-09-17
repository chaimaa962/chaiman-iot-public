package zkp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestKeystore_LoadOrCreate_FirstBoot(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "device_key.json")

	kp1, isNew, err := LoadOrCreateKeyPair(path)
	if err != nil {
		t.Fatalf("premier démarrage: %v", err)
	}
	if !isNew {
		t.Fatal("attendu isNew=true au premier démarrage")
	}

	// Vérifie que le fichier existe bien sur disque maintenant.
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("fichier clé non créé: %v", err)
	}

	// Simule un REDÉMARRAGE du device : recharge depuis le même fichier.
	kp2, isNew2, err := LoadOrCreateKeyPair(path)
	if err != nil {
		t.Fatalf("second démarrage: %v", err)
	}
	if isNew2 {
		t.Fatal("attendu isNew=false au second démarrage (clé déjà existante)")
	}

	if !bigIntsEqual(kp1.PrivateKey, kp2.PrivateKey) {
		t.Fatal("SÉCURITÉ CASSÉE : la clé privée a changé entre deux démarrages — le device perdrait son identité")
	}
	if kp1.PublicKey.Hex() != kp2.PublicKey.Hex() {
		t.Fatal("la clé publique a changé entre deux démarrages")
	}
	t.Log("✅ Identité du device stable à travers les redémarrages")
}

func TestKeystore_ProofStillWorksAfterReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "device_key.json")

	kp1, _, _ := LoadOrCreateKeyPair(path)
	kp2, _, _ := LoadOrCreateKeyPair(path) // simule un redémarrage

	// Une preuve générée après rechargement doit être vérifiable avec la
	// même clé publique qu'avant redémarrage (identité préservée de bout en bout).
	proof, err := kp2.GenerateProof("IOT_0001", 1735600000, 1, "n1")
	if err != nil {
		t.Fatalf("génération preuve après rechargement: %v", err)
	}
	valid, err := VerifyProof(kp1.PublicKey, proof, "IOT_0001", 1735600000, 1, "n1")
	if err != nil || !valid {
		t.Fatalf("preuve invalide après rechargement de clé: valid=%v err=%v", valid, err)
	}
	t.Log("✅ Authentification fonctionne correctement après redémarrage du device")
}

func TestKeystore_DetectsTampering(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "device_key.json")

	kp, _, _ := LoadOrCreateKeyPair(path)

	// Un attaquant modifie le fichier sur disque (remplace la clé publique
	// par une autre, tout en gardant l'ancienne clé privée) — simule une
	// tentative de substitution d'identité par accès physique/filesystem.
	otherKP, _ := GenerateKeyPair()
	tampered := persistedKeyPair{
		PrivateKeyHex: kp.PrivateKey.Text(16),
		PublicKeyHex:  otherKP.PublicKey.Hex(), // ne correspond plus à la clé privée !
	}
	raw, _ := json.MarshalIndent(tampered, "", "  ")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatalf("écriture fichier trafiqué: %v", err)
	}

	_, err := LoadKeyPair(path)
	if err == nil {
		t.Fatal("SÉCURITÉ CASSÉE : un fichier de clé trafiqué (incohérent) a été accepté")
	}
	t.Logf("✅ Fichier trafiqué correctement rejeté : %v", err)
}

func TestKeystore_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "device_key.json")

	_, _, err := LoadOrCreateKeyPair(path)
	if err != nil {
		t.Fatalf("création clé: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat fichier clé: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Fatalf("permissions incorrectes sur le fichier de clé: attendu 0600, obtenu %o — la clé privée pourrait être lisible par d'autres utilisateurs", perm)
	}
	t.Log("✅ Permissions fichier clé correctes (0600)")
}
