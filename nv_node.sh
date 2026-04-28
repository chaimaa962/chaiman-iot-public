#!/bin/bash

# ============================================
# Script pour Ajouter de Nouveaux Nœuds au Réseau POA
# Version COMPLÈTE - Création + Connexion automatique
# ============================================

set -e

BASE_DIR="$HOME/iot-docker-project"
BACKEND_URL="http://localhost:8080"
NETWORK_ID=2026
PASSWORD="password123"

echo "=========================================="
echo "➕ Ajouter de nouveaux nœuds au réseau POA"
echo "=========================================="
echo ""

# Détecter le dernier numéro de validateur
echo "🔍 Détection des nœuds existants..."

LAST_VAL=0
for node in $(docker ps -a --format "{{.Names}}" 2>/dev/null | grep "geth-validator" | sort -V); do
    num=$(echo $node | sed 's/geth-validator//')
    if [ "$num" -gt "$LAST_VAL" ]; then
        LAST_VAL=$num
    fi
done
# Si aucun trouvé, commencer à 4
[ "$LAST_VAL" -eq 0 ] && LAST_VAL=4

LAST_FN=0
for node in $(docker ps -a --format "{{.Names}}" 2>/dev/null | grep "geth-fullnode" | sort -V); do
    num=$(echo $node | sed 's/geth-fullnode//')
    if [ "$num" -gt "$LAST_FN" ]; then
        LAST_FN=$num
    fi
done
[ "$LAST_FN" -eq 0 ] && LAST_FN=4

echo ""
echo "📊 Nœuds existants :"
echo "  - Dernier validateur: validator$LAST_VAL"
echo "  - Dernier full node: fullnode$LAST_FN"
echo ""

read -p "Combien de VALIDATEURS supplémentaires ? (0-10) : " NEW_VALIDATORS
read -p "Combien de FULL NODES supplémentaires ? (0-10) : " NEW_FULLNODES

NEW_VALIDATORS=${NEW_VALIDATORS:-0}
NEW_FULLNODES=${NEW_FULLNODES:-0}

if [ "$NEW_VALIDATORS" -eq 0 ] && [ "$NEW_FULLNODES" -eq 0 ]; then
    echo "❌ Aucun nœud à ajouter. Sortie."
    exit 0
fi

echo ""
echo "📝 Ajout de $NEW_VALIDATORS validateur(s) et $NEW_FULLNODES full node(s)..."

# ============================================
# CRÉER LES NOUVEAUX VALIDATEURS
# ============================================
for i in $(seq 1 $NEW_VALIDATORS); do
    NODE_NUM=$((LAST_VAL + i))
    NODE_NAME="validator$NODE_NUM"
    NODE_DIR="$BASE_DIR/nodes/$NODE_NAME"
    
    echo ""
    echo "🔧 Création de $NODE_NAME (numéro $NODE_NUM)..."
    
    mkdir -p $NODE_DIR/keystore
    echo -n "$PASSWORD" > $NODE_DIR/password.txt
    
    echo "  📦 Création du compte Ethereum..."
    ADDR=$(docker run --rm \
        -v $NODE_DIR:/data \
        ethereum/client-go:v1.13.15 \
        account new --password /data/password.txt --keystore /data/keystore 2>&1 | \
        grep "Public address of the key" | awk '{print $NF}')
    echo "  ✅ Adresse: $ADDR"
    
    echo "  🔧 Initialisation avec genesis.json..."
    docker run --rm \
        -v $BASE_DIR/genesis.json:/genesis.json:ro \
        -v $NODE_DIR:/data \
        ethereum/client-go:v1.13.15 \
        --datadir /data init /genesis.json 2>&1 | grep -E "(Successfully|Genesis)" || true
    
    RPC_PORT=$((8540 + NODE_NUM))
    P2P_PORT=$((30300 + NODE_NUM))
    IP_ADDR="172.20.0.$((10 + NODE_NUM))"
    
    echo "  📝 Ports: RPC=$RPC_PORT, P2P=$P2P_PORT, IP=$IP_ADDR"
    
    cat >> $BASE_DIR/docker-compose.yml << EOF

  $NODE_NAME:
    image: ethereum/client-go:v1.13.15
    container_name: geth-$NODE_NAME
    hostname: $NODE_NAME
    networks:
      poa-network:
        ipv4_address: $IP_ADDR
    volumes:
      - $BASE_DIR/nodes/$NODE_NAME:/root/.ethereum
    command:
      - --datadir=/root/.ethereum
      - --networkid=$NETWORK_ID
      - --port=$P2P_PORT
      - --http
      - --http.port=$RPC_PORT
      - --http.addr=0.0.0.0
      - --http.api=eth,web3,net,admin,personal,miner,clique,txpool
      - --http.corsdomain=*
      - --authrpc.port=$((8550 + NODE_NUM))
      - --mine
      - --miner.etherbase=$ADDR
      - --unlock=$ADDR
      - --password=/root/.ethereum/password.txt
      - --allow-insecure-unlock
      - --syncmode=full
      - --snapshot=false
      - --netrestrict=172.20.0.0/16
      - --nodiscover
      - --verbosity=3
    ports:
      - "$RPC_PORT:$RPC_PORT"
      - "$P2P_PORT:$P2P_PORT/tcp"
      - "$P2P_PORT:$P2P_PORT/udp"
    restart: unless-stopped
EOF
    
    echo "  ✅ $NODE_NAME ajouté au docker-compose.yml"
done

# ============================================
# CRÉER LES NOUVEAUX FULL NODES
# ============================================
for i in $(seq 1 $NEW_FULLNODES); do
    NODE_NUM=$((LAST_FN + i))
    NODE_NAME="fullnode$NODE_NUM"
    NODE_DIR="$BASE_DIR/nodes/$NODE_NAME"
    
    echo ""
    echo "🔧 Création de $NODE_NAME (numéro $NODE_NUM)..."
    
    mkdir -p $NODE_DIR
    
    echo "  🔧 Initialisation avec genesis.json..."
    docker run --rm \
        -v $BASE_DIR/genesis.json:/genesis.json:ro \
        -v $NODE_DIR:/data \
        ethereum/client-go:v1.13.15 \
        --datadir /data init /genesis.json 2>&1 | grep -E "(Successfully|Genesis)" || true
    
    RPC_PORT=$((8540 + LAST_VAL + NEW_VALIDATORS + i))
    P2P_PORT=$((30300 + LAST_VAL + NEW_VALIDATORS + i))
    IP_ADDR="172.20.0.$((20 + NODE_NUM))"
    
    echo "  📝 Ports: RPC=$RPC_PORT, P2P=$P2P_PORT, IP=$IP_ADDR"
    
    cat >> $BASE_DIR/docker-compose.yml << EOF

  $NODE_NAME:
    image: ethereum/client-go:v1.13.15
    container_name: geth-$NODE_NAME
    hostname: $NODE_NAME
    networks:
      poa-network:
        ipv4_address: $IP_ADDR
    volumes:
      - $BASE_DIR/nodes/$NODE_NAME:/root/.ethereum
    command:
      - --datadir=/root/.ethereum
      - --networkid=$NETWORK_ID
      - --port=$P2P_PORT
      - --http
      - --http.port=$RPC_PORT
      - --http.addr=0.0.0.0
      - --http.api=eth,web3,net,admin,txpool
      - --http.corsdomain=*
      - --syncmode=full
      - --snapshot=false
      - --netrestrict=172.20.0.0/16
      - --nodiscover
      - --verbosity=3
    ports:
      - "$RPC_PORT:$RPC_PORT"
      - "$P2P_PORT:$P2P_PORT/tcp"
      - "$P2P_PORT:$P2P_PORT/udp"
    restart: unless-stopped
EOF
    
    echo "  ✅ $NODE_NAME ajouté au docker-compose.yml"
done

# ============================================
# DÉMARRAGE DES NOUVEAUX NŒUDS
# ============================================
echo ""
echo "=========================================="
echo "🚀 Démarrage des nouveaux nœuds..."
echo "=========================================="

cd $BASE_DIR

for i in $(seq 1 $NEW_VALIDATORS); do
    NODE_NUM=$((LAST_VAL + i))
    docker compose -f docker-compose.yml up -d validator$NODE_NUM 2>/dev/null || docker-compose -f docker-compose.yml up -d validator$NODE_NUM 2>/dev/null
    echo "  ✅ validator$NODE_NUM démarré"
done

for i in $(seq 1 $NEW_FULLNODES); do
    NODE_NUM=$((LAST_FN + i))
    docker compose -f docker-compose.yml up -d fullnode$NODE_NUM 2>/dev/null || docker-compose -f docker-compose.yml up -d fullnode$NODE_NUM 2>/dev/null
    echo "  ✅ fullnode$NODE_NUM démarré"
done

echo ""
echo "⏳ Attente de 30 secondes pour le démarrage..."
sleep 30

# ============================================
# CONNEXION AUTOMATIQUE DES NOUVEAUX NŒUDS
# ============================================
echo ""
echo "=========================================="
echo "🔗 Connexion des nouveaux nœuds au réseau..."
echo "=========================================="

# Récupérer TOUS les enodes (existants + nouveaux)
declare -A ENODES

echo "📡 Récupération des enodes..."

# Scanner tous les validateurs
for port in $(seq 8541 $((8540 + LAST_VAL + NEW_VALIDATORS))); do
    enode=$(curl -s -X POST http://localhost:$port \
        -H 'Content-Type: application/json' \
        -d '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' 2>/dev/null | jq -r '.result.enode' 2>/dev/null)
    
    if [ -n "$enode" ] && [ "$enode" != "null" ]; then
        # Calculer l'IP selon le port
        if [ $port -le 8544 ]; then
            IP="172.20.0.$((11 + port - 8541))"
        else
            IP="172.20.0.$((10 + port - 8540))"
        fi
        enode_fixed=$(echo "$enode" | sed "s/127.0.0.1/$IP/" | sed "s/\[::\]/$IP/")
        ENODES[$port]=$enode_fixed
        echo "  ✅ Port $port: enode récupéré (IP: $IP)"
    fi
done

# Scanner tous les fullnodes
for port in $(seq 8545 $((8540 + LAST_FN + NEW_FULLNODES + NEW_VALIDATORS))); do
    enode=$(curl -s -X POST http://localhost:$port \
        -H 'Content-Type: application/json' \
        -d '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' 2>/dev/null | jq -r '.result.enode' 2>/dev/null)
    
    if [ -n "$enode" ] && [ "$enode" != "null" ]; then
        IP="172.20.0.$((20 + port - 8544))"
        enode_fixed=$(echo "$enode" | sed "s/127.0.0.1/$IP/" | sed "s/\[::\]/$IP/")
        ENODES[$port]=$enode_fixed
        echo "  ✅ Port $port: enode récupéré (IP: $IP)"
    fi
done

echo ""
echo "🔗 Connexion de tous les nœuds entre eux (${#ENODES[@]} nœuds)..."

# Connecter chaque nœud à tous les autres
for port in "${!ENODES[@]}"; do
    for other_port in "${!ENODES[@]}"; do
        if [ "$port" != "$other_port" ]; then
            result=$(curl -s -X POST http://localhost:$port \
                -H 'Content-Type: application/json' \
                -d "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"${ENODES[$other_port]}\"],\"id\":1}" 2>/dev/null)
            if echo "$result" | grep -q "true"; then
                echo "  ✅ Port $port -> Port $other_port"
            fi
        fi
    done
done

echo ""
echo "📊 Vérification des pairs connectés..."
sleep 5

for port in "${!ENODES[@]}"; do
    peers=$(curl -s -X POST http://localhost:$port \
        -H 'Content-Type: application/json' \
        -d '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' 2>/dev/null | jq -r '.result' 2>/dev/null | xargs printf '%d' 2>/dev/null)
    block=$(curl -s -X POST http://localhost:$port \
        -H 'Content-Type: application/json' \
        -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' 2>/dev/null | jq -r '.result' 2>/dev/null | xargs printf '%d' 2>/dev/null)
    echo "  📊 Port $port: $peers pairs | Bloc #$block"
done

echo ""
echo "=========================================="
echo "✅ Nouveaux nœuds ajoutés et connectés !"
echo "=========================================="
echo ""
echo "📌 Résumé :"
echo "  - Total validateurs: $((LAST_VAL + NEW_VALIDATORS))"
echo "  - Total full nodes: $((LAST_FN + NEW_FULLNODES))"
echo "  - Total nœuds: ${#ENODES[@]}"
echo ""
echo "🔧 Commandes utiles :"
echo "  - Voir les logs: docker logs -f geth-validator1"
echo "  - Scanner les nœuds: curl http://localhost:8080/api/geth-nodes"
