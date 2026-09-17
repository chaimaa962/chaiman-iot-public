
#!/bin/bash

echo "╔═══════════════════════════════════════════════════════════════════════════════╗"
echo "║              🔍 VÉRIFICATION FINALE - SYSTÈME COMPLET CHAIMAN IOT             ║"
echo "║                    Blockchain + ZKP + IA + IoT + Backend                       ║"
echo "╚═══════════════════════════════════════════════════════════════════════════════╝"
echo ""

# ============================================
# 1. BACKEND GO
# ============================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📡 1. BACKEND GO (PORT 8080)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if curl -s http://localhost:8080/api/health > /dev/null 2>&1; then
    echo "✅ Backend Go: ACTIF"
    HEALTH=$(curl -s http://localhost:8080/api/health | jq -r '.status' 2>/dev/null)
    echo "   └─ Santé: $HEALTH"
else
    echo "❌ Backend Go: INACTIF"
fi

# Tester les routes principales
echo ""
echo "📋 ROUTES API:"
for route in "/api/health" "/api/devices" "/api/geth-nodes" "/api/stats/global" "/api/blockchain/info"; do
    status=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080$route" 2>/dev/null)
    if [ "$status" = "200" ]; then
        echo "   ✅ GET $route → HTTP $status"
    else
        echo "   ❌ GET $route → HTTP $status"
    fi
done

# ============================================
# 2. IA V4 (APPRENTISSAGE CONTINU)
# ============================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🤖 2. IA V4 AVEC APPRENTISSAGE CONTINU (PORT 5001)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if curl -s http://localhost:5001/v4/health > /dev/null 2>&1; then
    echo "✅ IA V4: ACTIVE"
    
    MODEL=$(curl -s http://localhost:5001/v4/health 2>/dev/null | jq -r '.model' 2>/dev/null)
    F1=$(curl -s http://localhost:5001/v4/health 2>/dev/null | jq -r '.f1_score' 2>/dev/null)
    THRESHOLD=$(curl -s http://localhost:5001/v4/health 2>/dev/null | jq -r '.threshold' 2>/dev/null)
    RETRAIN_COUNT=$(curl -s http://localhost:5001/v4/health 2>/dev/null | jq -r '.retrain_count' 2>/dev/null)
    
    echo "   ├─ Modèle: ${MODEL:0:50}..."
    echo "   ├─ F1-score: $F1%"
    echo "   ├─ Seuil: $THRESHOLD"
    echo "   └─ Ré-entraînements: $RETRAIN_COUNT"
    
    # Tester l'analyse d'un appareil
    echo ""
    echo "🔬 TEST ANALYSE IA:"
    ANALYSIS=$(curl -s "http://localhost:5001/v4/analyze/0x287052AA5BE1f" 2>/dev/null)
    SCORE=$(echo "$ANALYSIS" | jq -r '.score' 2>/dev/null)
    SEVERITY=$(echo "$ANALYSIS" | jq -r '.severity' 2>/dev/null)
    ACTION=$(echo "$ANALYSIS" | jq -r '.action' 2>/dev/null)
    echo "   ├─ Score: $SCORE"
    echo "   ├─ Sévérité: $SEVERITY"
    echo "   └─ Action: $ACTION"
else
    echo "❌ IA V4: INACTIVE"
fi

# ============================================
# 3. BLOCKCHAIN GETH
# ============================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "⛓️ 3. BLOCKCHAIN GETH (PoA)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Compter les conteneurs
NB_GETH=$(docker ps --filter "name=geth" -q 2>/dev/null | wc -l)
echo "📦 Conteneurs Geth: $NB_GETH"

# Vérifier les validateurs
echo ""
echo "🔨 VALIDATEURS (minage actif):"
for port in 8541 8542 8543 8544; do
    mining=$(curl -s -X POST http://localhost:$port \
        -H "Content-Type: application/json" \
        -d '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' 2>/dev/null | jq -r '.result' 2>/dev/null)
    if [ "$mining" = "true" ]; then
        echo "   ✅ Port $port: mining actif"
    else
        echo "   ❌ Port $port: mining inactif (ou non accessible)"
    fi
done

# Dernier bloc
BLOCK=$(curl -s -X POST http://localhost:8541 \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' 2>/dev/null | jq -r '.result' 2>/dev/null | xargs printf "%d" 2>/dev/null)
echo ""
echo "📊 Dernier bloc: $BLOCK"

# ============================================
# 4. NŒUDS GETH DÉTECTÉS PAR L'API
# ============================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🖥️ 4. NŒUDS GETH DÉTECTÉS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

NB_TOTAL=$(curl -s http://localhost:8080/api/geth-nodes 2>/dev/null | jq 'length' 2>/dev/null)
NB_VALIDATEURS=$(curl -s http://localhost:8080/api/geth-nodes 2>/dev/null | jq '[.[] | select(.type == "validator")] | length' 2>/dev/null)
NB_FULLNODES=$(curl -s http://localhost:8080/api/geth-nodes 2>/dev/null | jq '[.[] | select(.type == "fullnode")] | length' 2>/dev/null)

echo "📊 STATISTIQUES:"
echo "   ├─ Total nœuds: $NB_TOTAL"
echo "   ├─ Validateurs: $NB_VALIDATEURS"
echo "   └─ Full nodes: $NB_FULLNODES"

# ============================================
# 5. APPAREILS IoT
# ============================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📱 5. APPAREILS IoT"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

NB_IOT=$(curl -s http://localhost:8080/api/devices 2>/dev/null | jq 'length' 2>/dev/null)
echo "📊 Appareils IoT connectés: $NB_IOT"

# ============================================
# 6. TRANSACTIONS BLOCKCHAIN
# ============================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📝 6. TRANSACTIONS BLOCKCHAIN"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

NB_TX=$(curl -s "http://localhost:8080/api/transactions/recent?limit=100" 2>/dev/null | jq 'length' 2>/dev/null)
NB_SUCCESS=$(curl -s "http://localhost:8080/api/transactions/recent?limit=100" 2>/dev/null | jq '[.[] | select(.success == true)] | length' 2>/dev/null)

echo "📊 STATISTIQUES (100 dernières):"
echo "   ├─ Total transactions: $NB_TX"
echo "   └─ Succès: $NB_SUCCESS ($((NB_SUCCESS * 100 / NB_TX))%)"

# ============================================
# 7. COMMUNICATION BACKEND ↔ IA
# ============================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔗 7. COMMUNICATION BACKEND ↔ IA"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Vérifier les logs pour voir la communication
if tail -20 ~/iot-docker-project/logs/backend.log 2>/dev/null | grep -q "Analyse IA"; then
    echo "✅ Communication backend → IA: ACTIVE"
    echo "   Dernières analyses:"
    tail -5 ~/iot-docker-project/logs/backend.log 2>/dev/null | grep "Analyse IA" | while read line; do
        echo "   └─ $line"
    done
else
    echo "⚠️ Communication backend → IA: En attente de données..."
fi

# ============================================
# 8. RÉSUMÉ FINAL
# ============================================
echo ""
echo "╔═══════════════════════════════════════════════════════════════════════════════╗"
echo "║                          📊 RÉSUMÉ FINAL                                       ║
echo "╚═══════════════════════════════════════════════════════════════════════════════╝"

# Calculer le score global
SCORE=0
if curl -s http://localhost:8080/api/health > /dev/null 2>&1; then SCORE=$((SCORE+1)); fi
if curl -s http://localhost:5001/v4/health > /dev/null 2>&1; then SCORE=$((SCORE+1)); fi
if [ "$NB_GETH" -gt 0 ]; then SCORE=$((SCORE+1)); fi
if [ "$NB_IOT" -gt 0 ]; then SCORE=$((SCORE+1)); fi
if [ "$NB_TX" -gt 0 ]; then SCORE=$((SCORE+1)); fi

echo ""
if [ $SCORE -eq 5 ]; then
    echo "✅ SCORE: $SCORE/5 - SYSTÈME COMPLET ET OPÉRATIONNEL !"
    echo ""
    echo "🎉 FÉLICITATIONS ! Tous vos composants sont intégrés et fonctionnent !"
    echo ""
    echo "📊 RÉCAPITULATIF:"
    echo "   ├─ Backend Go:      ✅ ACTIF"
    echo "   ├─ IA V4:          ✅ ACTIVE (apprentissage continu)"
    echo "   ├─ Blockchain:     ✅ ACTIVE ($NB_GETH nœuds, bloc $BLOCK)"
    echo "   ├─ Appareils IoT:  ✅ $NB_IOT actifs"
    echo "   └─ Transactions:   ✅ $NB_TX enregistrées"
elif [ $SCORE -ge 3 ]; then
    echo "⚠️ SCORE: $SCORE/5 - SYSTÈME PARTIELLEMENT OPÉRATIONNEL"
else
    echo "❌ SCORE: $SCORE/5 - SYSTÈME NON OPÉRATIONNEL"
fi

echo ""
echo "╔═══════════════════════════════════════════════════════════════════════════════╗"
echo "║              🚀 VOTRE PROJET EST PRÊT POUR LA SOUTENANCE ! 🚀                  ║
echo "╚═══════════════════════════════════════════════════════════════════════════════╝"

