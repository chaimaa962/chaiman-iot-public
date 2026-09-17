#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}═══════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}🔍 CHAIMAN IoT - VÉRIFICATION FINALE COMPLÈTE${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════${NC}"
echo ""

# 1. BACKEND GO
echo -e "${YELLOW}📡 1. BACKEND GO (port 8080)${NC}"
if curl -s --max-time 3 http://localhost:8080/api/health 2>/dev/null | grep -q "OK"; then
    echo -e "   ${GREEN}✅ Backend OK${NC}"
    BACKEND_STATUS="OK"
else
    echo -e "   ${RED}❌ Backend indisponible${NC}"
    BACKEND_STATUS="KO"
fi

# 2. BLOCKCHAIN GETH
echo -e "\n${YELLOW}⛓️ 2. BLOCKCHAIN GETH (port 8541)${NC}"
if curl -s --max-time 3 -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:8541 2>/dev/null | grep -q '"result"'; then
    BLOCK=$(curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:8541 2>/dev/null | grep -o '"result":"0x[^"]*"' | cut -d'"' -f4)
    BLOCK_DEC=$((BLOCK))
    echo -e "   ${GREEN}✅ Geth OK (block #${BLOCK_DEC})${NC}"
    GETH_STATUS="OK"
else
    echo -e "   ${RED}❌ Geth indisponible${NC}"
    GETH_STATUS="KO"
fi

# 3. AI ENGINE
echo -e "\n${YELLOW}🤖 3. AI ENGINE (port 5001)${NC}"
if curl -s --max-time 3 http://localhost:5001/health 2>/dev/null | grep -q "healthy"; then
    echo -e "   ${GREEN}✅ AI Engine OK${NC}"
    AI_STATUS="OK"
else
    echo -e "   ${RED}❌ AI Engine indisponible${NC}"
    AI_STATUS="KO"
fi

# 4. TLS (si configuré)
echo -e "\n${YELLOW}🔐 4. TLS/SSL${NC}"
if [ -f "cert.pem" ] && [ -f "key.pem" ]; then
    echo -e "   ${GREEN}✅ Certificats TLS trouvés${NC}"
    # Vérifier expiration
    EXPIRY=$(openssl x509 -in cert.pem -noout -enddate 2>/dev/null | cut -d= -f2)
    if [ -n "$EXPIRY" ]; then
        echo -e "   ${GREEN}   Expiration: $EXPIRY${NC}"
    fi
    TLS_STATUS="OK"
else
    echo -e "   ${YELLOW}⚠️ TLS non configuré (mode développement)${NC}"
    TLS_STATUS="DEV"
fi

# 5. CONTRAT BLOCKCHAIN
echo -e "\n${YELLOW}📄 5. SMART CONTRACT${NC}"
CONTRACT_ADDR=$(grep CONTRACT_ADDRESS iot-backend/.env 2>/dev/null | cut -d'=' -f2)
if [ -n "$CONTRACT_ADDR" ]; then
    echo -e "   ${GREEN}✅ Contrat: ${CONTRACT_ADDR}${NC}"
    # Vérifier code du contrat
    CODE=$(curl -s -X POST -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getCode\",\"params\":[\"$CONTRACT_ADDR\",\"latest\"],\"id\":1}" http://localhost:8541 2>/dev/null | grep -o '"result":"0x[^"]*"' | cut -d'"' -f4)
    if [ ${#CODE} -gt 10 ]; then
        echo -e "   ${GREEN}   Code contrat présent (${#CODE} bytes)${NC}"
    fi
fi

# 6. STATS BACKEND
echo -e "\n${YELLOW}📊 6. STATISTIQUES${NC}"
if [ "$BACKEND_STATUS" = "OK" ]; then
    TOTAL_DEVICES=$(curl -s http://localhost:8080/api/devices 2>/dev/null | python3 -c "import sys,json; d=json.load(sys.stdin); print(len(d))" 2>/dev/null)
    TOTAL_AUTHS=$(curl -s http://localhost:8080/api/stats/global 2>/dev/null | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('totalAuths',0))" 2>/dev/null)
    echo -e "   ${GREEN}✅ Devices enregistrés: ${TOTAL_DEVICES:-0}${NC}"
    echo -e "   ${GREEN}✅ Authentifications: ${TOTAL_AUTHS:-0}${NC}"
fi

# 7. TEST SIMULATEUR
echo -e "\n${YELLOW}🚗 7. SIMULATEUR IoT${NC}"
if [ -f "iot_simulator" ]; then
    echo -e "   ${GREEN}✅ Binaire simulateur présent${NC}"
else
    echo -e "   ${YELLOW}⚠️ Simulateur non compilé (go build)${NC}"
fi

# 8. RÉSUMÉ
echo -e "\n${BLUE}═══════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}📋 RÉSUMÉ${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════${NC}"
echo -e "   Backend Go:      $([ "$BACKEND_STATUS" = "OK" ] && echo "✅" || echo "❌")"
echo -e "   Blockchain Geth: $([ "$GETH_STATUS" = "OK" ] && echo "✅" || echo "❌")"
echo -e "   AI Engine:       $([ "$AI_STATUS" = "OK" ] && echo "✅" || echo "❌")"
echo -e "   TLS:             $([ "$TLS_STATUS" = "OK" ] && echo "✅" || ( [ "$TLS_STATUS" = "DEV" ] && echo "⚠️ (dev)" || echo "❌"))"

echo -e "\n${BLUE}═══════════════════════════════════════════════════════════════════${NC}"
if [ "$BACKEND_STATUS" = "OK" ] && [ "$GETH_STATUS" = "OK" ]; then
    echo -e "${GREEN}✅ TOUT EST OPÉRATIONNEL !${NC}"
    echo -e "\nPour lancer le simulateur:"
    echo -e "   ./iot_simulator -backend=http://localhost:8080 -node=http://localhost:8541"
else
    echo -e "${RED}❌ Certains services ne sont pas disponibles${NC}"
    echo -e "\nPour démanger les services manquants:"
    if [ "$BACKEND_STATUS" != "OK" ]; then
        echo -e "   cd iot-backend && go build && ./iot-backend &"
    fi
    if [ "$GETH_STATUS" != "OK" ]; then
        echo -e "   docker start v1  # ou un autre conteneur Geth"
    fi
    if [ "$AI_STATUS" != "OK" ]; then
        echo -e "   cd iot-backend/ai && source venv/bin/activate && python3 detection_nosig.py &"
    fi
fi
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════${NC}"
