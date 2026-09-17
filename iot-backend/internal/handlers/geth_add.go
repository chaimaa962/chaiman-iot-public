package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ScriptState struct {
	Running bool
	Output  string
	Mutex   sync.RWMutex
}

var scriptState = &ScriptState{
	Running: false,
	Output:  "",
}

type AddNodesRequest struct {
	Validators int `json:"validators"`
	Fullnodes  int `json:"fullnodes"`
}

type AddNodesResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ✅ FIX MED-2: Suppression de requireAdminAuth de ce fichier
// La version sécurisée avec hmac.Equal est dans main_secure.go
// NE PAS RAJOUTER DE FONCTION requireAdminAuth ICI

// ✅ FIX MED-1: Validation de chemin sécurisée avec HasPrefix
var allowedScriptDirs = []string{
	"iot-docker-project",
	"iot-project",
}

// ✅ FIX MED-1: validateScriptPath utilise HasPrefix au lieu de Contains
func validateScriptPath(scriptPath string) (string, error) {
	cleanPath := filepath.Clean(scriptPath)

	if strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("chemin invalide: path traversal détecté")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("impossible de déterminer le répertoire home: %v", err)
	}

	// ✅ FIX MED-1: Vérifier que le chemin est DANS le répertoire autorisé
	found := false
	for _, dir := range allowedScriptDirs {
		allowedBase := filepath.Join(homeDir, dir)
		// Vérifier que cleanPath commence par allowedBase + séparateur
		prefix := allowedBase + string(filepath.Separator)
		if strings.HasPrefix(cleanPath, prefix) || cleanPath == allowedBase {
			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("chemin hors répertoire autorisé: %s", cleanPath)
	}

	return cleanPath, nil
}

func verifyScriptIntegrity(scriptPath string) error {
	expectedHash := os.Getenv("SCRIPT_SHA256")
	if expectedHash == "" {
		return fmt.Errorf("SCRIPT_SHA256 non défini — intégrité non vérifiable")
	}

	data, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("lecture script impossible: %v", err)
	}

	hash := sha256.Sum256(data)
	actualHash := hex.EncodeToString(hash[:])

	if !strings.EqualFold(actualHash, expectedHash) {
		return fmt.Errorf("intégrité script compromise: hash attendu %s, reçu %s", expectedHash, actualHash)
	}

	return nil
}

func AddGethNodes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req AddNodesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AddNodesResponse{
			Success: false,
			Message: "Requête invalide",
			Error:   err.Error(),
		})
		return
	}

	if req.Validators < 0 || req.Fullnodes < 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AddNodesResponse{
			Success: false,
			Message: "Les valeurs ne peuvent pas être négatives",
		})
		return
	}

	if req.Validators > 10 || req.Fullnodes > 10 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AddNodesResponse{
			Success: false,
			Message: "Maximum 10 validateurs et 10 fullnodes autorisés",
		})
		return
	}

	if req.Validators == 0 && req.Fullnodes == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AddNodesResponse{
			Success: false,
			Message: "Au moins un validateur ou full node requis",
		})
		return
	}

	scriptState.Mutex.RLock()
	if scriptState.Running {
		scriptState.Mutex.RUnlock()
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(AddNodesResponse{
			Success: false,
			Message: "Un script est déjà en cours d'exécution",
		})
		return
	}
	scriptState.Mutex.RUnlock()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AddNodesResponse{
			Success: false,
			Message: "Impossible de déterminer le répertoire home",
		})
		return
	}

	scriptPath := filepath.Join(homeDir, "iot-docker-project", "nv_node.sh")

	validatedPath, err := validateScriptPath(scriptPath)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(AddNodesResponse{
			Success: false,
			Message: "Chemin script non autorisé",
			Error:   err.Error(),
		})
		return
	}

	info, err := os.Stat(validatedPath)
	if err != nil || info.IsDir() {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(AddNodesResponse{
			Success: false,
			Message: "Script nv_node.sh non trouvé",
			Error:   "Vérifiez que le script existe dans " + filepath.Dir(validatedPath),
		})
		return
	}

	if err := verifyScriptIntegrity(validatedPath); err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(AddNodesResponse{
			Success: false,
			Message: "Vérification d'intégrité échouée",
			Error:   err.Error(),
		})
		return
	}

	if err := os.Chmod(validatedPath, 0755); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AddNodesResponse{
			Success: false,
			Message: "Impossible de rendre le script exécutable",
			Error:   err.Error(),
		})
		return
	}

	scriptState.Mutex.Lock()
	scriptState.Running = true
	scriptState.Output = ""
	scriptState.Mutex.Unlock()

	json.NewEncoder(w).Encode(AddNodesResponse{
		Success: true,
		Message: "Démarrage de l'ajout des nœuds...",
	})

	// FIX: Passer validators et fullnodes comme ARGUMENTS du script
	// Plus de stdin injection — les valeurs ne peuvent pas contenir de code shell
	go executeAddScript(validatedPath, req.Validators, req.Fullnodes)
}

// FIX CRITIQUE: Les valeurs sont passées comme arguments positionnels du script.
// Le script doit lire $1 (validators) et $2 (fullnodes) au lieu de stdin.
// Exemple script: #!/bin/bash
//   VALIDATORS=$1
//   FULLNODES=$2
// Cela élimine complètement l'injection via stdin car exec.Command
// ne passe PAS ces arguments à travers un shell.
func executeAddScript(scriptPath string, validators, fullnodes int) {
	defer func() {
		scriptState.Mutex.Lock()
		scriptState.Running = false
		scriptState.Mutex.Unlock()
	}()

	// FIX: Convertir les entiers en strings pour les args — AUCUNE interpolation shell
	validatorsArg := strconv.Itoa(validators)
	fullnodesArg := strconv.Itoa(fullnodes)

	// FIX: exec.Command avec args directs — pas de bash -c, pas de shell intermédiaire
	// Les arguments sont passés directement au processus sans interprétation shell
	cmd := exec.Command(scriptPath, validatorsArg, fullnodesArg)

	// FIX: Environnement minimal et contrôlé — pas os.Environ() en entier
	cmd.Env = []string{
		"PATH=/usr/local/bin:/usr/bin:/bin",
		"HOME=" + os.Getenv("HOME"),
		"USER=" + os.Getenv("USER"),
		fmt.Sprintf("VALIDATORS=%s", validatorsArg),
		fmt.Sprintf("FULLNODES=%s", fullnodesArg),
	}

	// FIX: Timeout sur l'exécution du script (10 minutes max)
	done := make(chan error, 1)
	timeout := time.NewTimer(10 * time.Minute)
	defer timeout.Stop()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		appendAddOutput("Erreur: impossible de capturer stdout - " + err.Error())
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		appendAddOutput("Erreur: impossible de capturer stderr - " + err.Error())
		return
	}

	if err := cmd.Start(); err != nil {
		appendAddOutput("Erreur: impossible de démarrer le script - " + err.Error())
		return
	}

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				appendAddOutput(string(buf[:n]))
			}
			if err != nil {
				break
			}
		}
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if n > 0 {
				appendAddOutput("[STDERR] " + string(buf[:n]))
			}
			if err != nil {
				break
			}
		}
	}()

	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			appendAddOutput("Script terminé avec erreur: " + err.Error())
		} else {
			appendAddOutput("✅ Script terminé avec succès!")
		}
        case <-timeout.C:
    if cmd.Process != nil {
        cmd.Process.Kill()
    }
    appendAddOutput("❌ Script interrompu: timeout de 10 minutes dépassé")
	}
}

func appendAddOutput(text string) {
	scriptState.Mutex.Lock()
	defer scriptState.Mutex.Unlock()
	if len(scriptState.Output)+len(text) > 1024*1024 {
		scriptState.Output = scriptState.Output[len(scriptState.Output)-500000:]
	}
	scriptState.Output += text + "\n"
}

func GetAddGethNodesStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	scriptState.Mutex.RLock()
	defer scriptState.Mutex.RUnlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"running": scriptState.Running,
		"output":  scriptState.Output,
	})
}
