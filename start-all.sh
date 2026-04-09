#!/bin/bash

echo "════════════════════════════════════════════════════════════════"
echo "🚀 DÉMARRAGE AUTOMATIQUE DU SYSTÈME IoT + BLOCKCHAIN + IA + ZKP"
echo "════════════════════════════════════════════════════════════════"

# Couleurs
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# ============================================
# 1. DÉMARRER LA BLOCKCHAIN POA
# ============================================
echo -e "\n${YELLOW}📦 1/5 - Démarrage de la Blockchain PoA...${NC}"
docker-compose up -d validator1 validator2 validator3 validator4 fullnode1 fullnode2 fullnode3 fullnode4

echo -e "${BLUE}⏳ Attente synchronisation des nœuds (30s)...${NC}"
sleep 30

# Connecter les peers
echo -e "${YELLOW}🔗 Connexion des validateurs...${NC}"
./demarer.sh 2>/dev/null | grep -E "peers|Bloc" || true

# ============================================
# 2. DÉMARRER LE SERVICE IA
# ============================================
echo -e "\n${YELLOW}🤖 2/5 - Démarrage du Service IA (LSTM)...${NC}"
docker-compose up -d ml-analyzer

echo -e "${BLUE}⏳ Attente chargement du modèle (20s)...${NC}"
sleep 20

# Vérifier IA
if curl -s http://localhost:5001/health > /dev/null; then
    echo -e "${GREEN}✅ Service IA opérationnel${NC}"
else
    echo -e "${RED}⚠️  Service IA non disponible${NC}"
fi

# ============================================
# 3. DÉMARRER LE BACKEND GO (ECDSA + ZKP)
# ============================================
echo -e "\n${YELLOW}🔐 3/5 - Démarrage du Backend Go (ZKP)...${NC}"

# Trouver un terminal disponible et lancer en arrière-plan
pkill -f "main_secure" 2>/dev/null || true

cd iot-backend
nohup go run ./cmd/api/main_secure.go > ../logs/backend.log 2>&1 &
BACKEND_PID=$!
cd ..

echo -e "${BLUE}⏳ Attente démarrage backend (15s)...${NC}"
sleep 15

if curl -s http://localhost:8080/api/health > /dev/null; then
    echo -e "${GREEN}✅ Backend Go opérationnel${NC}"
else
    echo -e "${RED}⚠️  Backend Go non disponible${NC}"
fi

# ============================================
# 4. LANCER LE SIMULATEUR
# ============================================
echo -e "\n${YELLOW}📡 4/5 - Lancement du Simulateur 4 nœuds...${NC}"

# Charger les variables d'environnement
export $(cat .env | xargs 2>/dev/null)

# Lancer le simulateur en arrière-plan
pkill -f "geth_nodes_simulator" 2>/dev/null || true
nohup python3 geth_nodes_simulator.py > logs/simulator.log 2>&1 &
SIMULATOR_PID=$!

echo -e "${GREEN}✅ Simulateur lancé (PID: $SIMULATOR_PID)${NC}"

# ============================================
# 5. LANCER LE DASHBOARD (optionnel)
# ============================================
echo -e "\n${YELLOW}📊 5/5 - Dashboard disponible...${NC}"
echo -e "${BLUE}Pour lancer le dashboard :${NC}"
echo -e "   cd iot-dashboard && npm start"
echo -e "   Puis ouvrir ${GREEN}http://localhost:3000${NC}"

# ============================================
# RÉSUMÉ
# ============================================
echo -e "\n${GREEN}════════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}✅ SYSTÈME DÉMARRÉ AVEC SUCCÈS !${NC}"
echo -e "${GREEN}════════════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "📋 ${BLUE}Services disponibles :${NC}"
echo -e "   🔗 Blockchain RPC:     ${GREEN}http://localhost:8541${NC}"
echo -e "   🤖 Service IA:         ${GREEN}http://localhost:5001${NC}"
echo -e "   🔐 Backend ZKP:        ${GREEN}http://localhost:8080${NC}"
echo -e "   📊 Dashboard:          ${GREEN}http://localhost:3000${NC} (npm start dans iot-dashboard)"
echo ""
echo -e "📋 ${BLUE}Logs :${NC}"
echo -e "   Backend:  ${YELLOW}tail -f logs/backend.log${NC}"
echo -e "   Simulateur: ${YELLOW}tail -f logs/simulator.log${NC}"
echo -e "   Blockchain: ${YELLOW}docker-compose logs -f validator1${NC}"
echo ""
echo -e "📋 ${BLUE}Arrêter tout :${NC}"
echo -e "   ${YELLOW}./stop-all.sh${NC}"

# Sauvegarder les PIDs pour l'arrêt
echo $BACKEND_PID > logs/backend.pid
echo $SIMULATOR_PID > logs/simulator.pid
