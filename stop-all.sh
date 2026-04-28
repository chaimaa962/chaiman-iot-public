#!/bin/bash

echo "════════════════════════════════════════════════════════════════"
echo "🛑 ARRÊT DU SYSTÈME"
echo "════════════════════════════════════════════════════════════════"

cd ~/iot-docker-project

# ============================================
# 1. ARRÊT DE L'IA V4 (NOUVEAU)
# ============================================
echo "🤖 Arrêt de l'IA V4 avec apprentissage continu..."
pkill -f "detection.py" 2>/dev/null && echo "   ✅ IA V4 arrêtée" || echo "   ⚠️ Aucune IA V4 en cours"

# ============================================
# 2. ARRÊT DU SIMULATEUR IoT
# ============================================
echo "📡 Arrêt du simulateur IoT..."
if [ -f "logs/iot-simulator.pid" ]; then
    kill $(cat logs/iot-simulator.pid) 2>/dev/null && echo "   ✅ Simulateur IoT arrêté (PID)"
    rm -f logs/iot-simulator.pid
fi
pkill -f "iot-simulator" 2>/dev/null && echo "   ✅ Simulateur IoT arrêté"
pkill -f "iot_devices_simulator" 2>/dev/null
pkill -f "geth_nodes_simulator" 2>/dev/null

# ============================================
# 3. ARRÊT DU BACKEND GO
# ============================================
echo "🔐 Arrêt du backend Go..."
if [ -f "logs/backend.pid" ]; then
    kill $(cat logs/backend.pid) 2>/dev/null && echo "   ✅ Backend arrêté (PID)"
    rm -f logs/backend.pid
fi
pkill -f "main_secure" 2>/dev/null && echo "   ✅ Backend arrêté"
pkill -f "iot-backend" 2>/dev/null

# ============================================
# 4. ARRÊT DU FRONTEND
# ============================================
echo "🎨 Arrêt du frontend React..."
if [ -f "logs/frontend.pid" ]; then
    kill $(cat logs/frontend.pid) 2>/dev/null && echo "   ✅ Frontend arrêté (PID)"
    rm -f logs/frontend.pid
fi
pkill -f "react-scripts start" 2>/dev/null && echo "   ✅ Frontend arrêté"
pkill -f "node.*react-scripts" 2>/dev/null

# ============================================
# 5. ARRÊT DES CONTENEURS DOCKER
# ============================================
echo "🐳 Arrêt des conteneurs Docker..."
docker stop iot-ml-analyzer 2>/dev/null
docker stop iot-ml-monitor 2>/dev/null
docker stop iot-backend 2>/dev/null
docker stop $(docker ps -aq --filter "name=geth") 2>/dev/null
docker-compose down 2>/dev/null || docker compose down 2>/dev/null
echo "   ✅ Conteneurs arrêtés"

# ============================================
# 6. NETTOYAGE
# ============================================
echo "🧹 Nettoyage..."
rm -f logs/*.pid 2>/dev/null

# Optionnel : supprimer les conteneurs
read -p "Supprimer les conteneurs ? (o/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Oo]$ ]]; then
    docker rm iot-ml-analyzer 2>/dev/null
    docker rm iot-ml-monitor 2>/dev/null
    docker rm iot-backend 2>/dev/null
    docker rm $(docker ps -aq --filter "name=geth") 2>/dev/null
    echo "   ✅ Conteneurs supprimés"
fi

echo ""
echo "════════════════════════════════════════════════════════════════"
echo "✅ Système arrêté avec succès !"
echo "════════════════════════════════════════════════════════════════"
