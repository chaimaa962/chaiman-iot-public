#!/bin/bash

echo "════════════════════════════════════════════════════════════════"
echo "🚀 DÉMARRAGE AUTOMATIQUE DU SYSTÈME IoT + BLOCKCHAIN + IA + ZKP"
echo "════════════════════════════════════════════════════════════════"

cd ~/iot-docker-project
mkdir -p logs

# ============================================
# 1. NETTOYAGE
# ============================================
echo "🧹 Nettoyage des conteneurs..."
docker stop $(docker ps -aq --filter "name=geth") 2>/dev/null
docker rm $(docker ps -aq --filter "name=geth") 2>/dev/null
docker stop iot-ml-analyzer 2>/dev/null && docker rm iot-ml-analyzer 2>/dev/null

# ============================================
# 2. BLOCKCHAIN
# ============================================
echo "📦 1/5 - Démarrage Blockchain PoA..."
docker-compose up -d 2>/dev/null || docker compose up -d 2>/dev/null
echo "⏳ Attente 30s..."
sleep 30

# ============================================
# 3. IA (Ancienne version Docker - commentée mais conservée)
# ============================================
echo "🤖 2/5 - Démarrage Service IA (version Docker)..."
docker stop iot-ml-analyzer 2>/dev/null && docker rm iot-ml-analyzer 2>/dev/null
docker-compose up -d ml-analyzer 2>/dev/null || docker compose up -d ml-analyzer 2>/dev/null
echo "⏳ Attente 15s..."
sleep 15
curl -s http://localhost:5001/health >/dev/null 2>&1 && echo "   ✅ IA Docker prête" || echo "   ⚠️ IA Docker en cours..."

# ============================================
# 3b. NOUVEAU: IA V4 AVEC APPRENTISSAGE CONTINU (Python direct)
# ============================================
echo "🤖 2b/5 - Démarrage IA V4 avec apprentissage continu..."
pkill -f "detection.py" 2>/dev/null
sleep 2
cd ~/iot-docker-project/iot-backend/ai

# Activer l'environnement virtuel et lancer l'IA
source venv/bin/activate
nohup python3 detection.py > ~/iot-docker-project/logs/ia_v4.log 2>&1 &
IA_PID=$!
echo "   ✅ IA V4 démarrée (PID: $IA_PID)"

# Attendre le chargement du modèle
echo "   ⏳ Chargement du modèle V4..."
sleep 30

# Vérifier que l'IA répond
if curl -s http://localhost:5001/v4/health > /dev/null 2>&1; then
    echo "   ✅ IA V4 prête (apprentissage continu activé)"
else
    echo "   ⚠️ IA V4 en cours de chargement..."
fi

cd ~/iot-docker-project

# ============================================
# 4. BACKEND (SCAN + CONNECT AUTO)
# ============================================
echo "⚙️  3/5 - Démarrage Backend Go (scan + connexion auto)..."
pkill -f "iot-backend" 2>/dev/null
sleep 2
cd ~/iot-docker-project/iot-backend
go build -o iot-backend ./cmd/api/main_secure.go 2>/dev/null
if [ -f "./iot-backend" ]; then
    nohup ./iot-backend > ../logs/backend.log 2>&1 &
    echo "   ✅ Backend démarré (PID: $!)"
else
    echo "   ❌ Erreur compilation"
fi
cd ~/iot-docker-project

# ============================================
# 5. SIMULATEUR IoT
# ============================================
echo "📡 4/5 - Démarrage Simulateur IoT..."
cd ~/iot-docker-project/iot-backend
if [ -f "./iot-simulator" ]; then
    nohup ./iot-simulator > ../logs/iot-simulator.log 2>&1 &
    echo "   ✅ Simulateur IoT démarré"
else
    go build -o iot-simulator ./simulators/iot_simulator.go 2>/dev/null
    if [ -f "./iot-simulator" ]; then
        nohup ./iot-simulator > ../logs/iot-simulator.log 2>&1 &
        echo "   ✅ Simulateur IoT démarré"
    fi
fi
cd ~/iot-docker-project

# ============================================
# RÉSUMÉ
# ============================================
echo ""
echo "════════════════════════════════════════════════════════════════"
echo "✅ SYSTÈME DÉMARRÉ !"
echo "════════════════════════════════════════════════════════════════"
echo "📊 Services: Blockchain:8541 | Backend:8080 | IA:5001 | IA V4:5001 | Frontend:3000"
echo "📋 Nœuds Geth: $(docker ps --filter 'name=geth' -q | wc -l) actifs"
echo "🤖 IA V4: Apprentissage continu activé (ré-entraînement toutes les heures)"
echo "🔧 Logs: tail -f logs/backend.log"
echo "🔧 Logs IA V4: tail -f logs/ia_v4.log"
echo "🛑 Arrêt: ./stop-all.sh"
