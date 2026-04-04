#!/bin/bash

# ============================================
# Script pour Ajouter de Nouveaux Nœuds au Réseau POA
# ============================================

set -e

BASE_DIR="$HOME/iot-docker-project"
NETWORK_ID=2026
PASSWORD="password123"

# Demander combien de nœuds ajouter
echo "=========================================="
echo "➕ Ajouter de nouveaux nœuds au réseau POA"
echo "=========================================="
echo ""
echo "Nœuds actuels :"
echo "  - 4 validateurs (validator1-4)"
echo "  - 4 full nodes (fullnode1-4)"
echo ""
read -p "Combien de VALIDATEURS supplémentaires voulez-vous ajouter ? (0-10) : " NEW_VALIDATORS
read -p "Combien de FULL NODES supplémentaires voulez-vous ajouter ? (0-10) : " NEW_FULLNODES

NEW_VALIDATORS=${NEW_VALIDATORS:-0}
NEW_FULLNODES=${NEW_FULLNODES:-0}

if [ "$NEW_VALIDATORS" -eq 0 ] && [ "$NEW_FULLNODES" -eq 0 ]; then
    echo "❌ Aucun nœud à ajouter. Sortie."
    exit 0
fi

echo ""
echo "📝 Ajout de $NEW_VALIDATORS validateur(s) et $NEW_FULLNODES full node(s)..."

# Trouver le dernier numéro de validateur et fullnode
LAST_VAL=4
LAST_FN=4

# Créer les nouveaux validateurs
for i in $(seq 1 $NEW_VALIDATORS); do
    NODE_NUM=$((LAST_VAL + i))
    NODE_NAME="validator$NODE_NUM"
    NODE_DIR="$BASE_DIR/nodes/$NODE_NAME"
    
    echo ""
    echo "🔧 Création de $NODE_NAME..."
    
    # Créer le répertoire
    mkdir -p $NODE_DIR/keystore
    
    # Créer le fichier password
    echo -n "$PASSWORD" > $NODE_DIR/password.txt
    
    # Créer le compte Ethereum
    echo "  📦 Création du compte Ethereum..."
    docker run --rm \
        -v $NODE_DIR:/data \
        ethereum/client-go:v1.13.15 \
        account new --password /data/password.txt --keystore /data/keystore 2>&1 | \
        grep "Public address of the key" | awk '{print "  ✅ Adresse: " $NF}'
    
    # Initialiser avec genesis.json
    echo "  🔧 Initialisation avec genesis.json..."
    docker run --rm \
        -v $BASE_DIR/genesis.json:/genesis.json:ro \
        -v $NODE_DIR:/data \
        ethereum/client-go:v1.13.15 \
        --datadir /data init /genesis.json 2>&1 | grep -E "(Successfully|Genesis)" || true
    
    # Calculer les ports
    RPC_PORT=$((8540 + NODE_NUM))
    P2P_PORT=$((30300 + NODE_NUM))
    IP_ADDR="172.20.0.$((10 + NODE_NUM))"
    
    echo "  📝 Ports: RPC=$RPC_PORT, P2P=$P2P_PORT, IP=$IP_ADDR"
    
    # Ajouter au docker-compose.yml
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
      - --http.api=eth,web3,net,admin,personal,miner,clique
      - --http.corsdomain=*
      - --authrpc.port=$((8550 + NODE_NUM))
      - --mine
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

# Créer les nouveaux full nodes
for i in $(seq 1 $NEW_FULLNODES); do
    NODE_NUM=$((LAST_FN + i))
    NODE_NAME="fullnode$NODE_NUM"
    NODE_DIR="$BASE_DIR/nodes/$NODE_NAME"
    
    echo ""
    echo "🔧 Création de $NODE_NAME..."
    
    # Créer le répertoire
    mkdir -p $NODE_DIR
    
    # Initialiser avec genesis.json
    echo "  🔧 Initialisation avec genesis.json..."
    docker run --rm \
        -v $BASE_DIR/genesis.json:/genesis.json:ro \
        -v $NODE_DIR:/data \
        ethereum/client-go:v1.13.15 \
        --datadir /data init /genesis.json 2>&1 | grep -E "(Successfully|Genesis)" || true
    
    # Calculer les ports (après les validateurs)
    RPC_PORT=$((8540 + 4 + NEW_VALIDATORS + i))
    P2P_PORT=$((30300 + 4 + NEW_VALIDATORS + i))
    IP_ADDR="172.20.0.$((20 + NODE_NUM))"
    
    echo "  📝 Ports: RPC=$RPC_PORT, P2P=$P2P_PORT, IP=$IP_ADDR"
    
    # Ajouter au docker-compose.yml
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
      - --http.api=eth,web3,net
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

echo ""
echo "=========================================="
echo "🚀 Démarrage des nouveaux nœuds..."
echo "=========================================="

# Démarrer seulement les nouveaux nœuds
for i in $(seq 1 $NEW_VALIDATORS); do
    NODE_NUM=$((LAST_VAL + i))
    docker-compose up -d validator$NODE_NUM
done

for i in $(seq 1 $NEW_FULLNODES); do
    NODE_NUM=$((LAST_FN + i))
    docker-compose up -d fullnode$NODE_NUM
done

echo ""
echo "⏳ Attente de 20 secondes pour le démarrage..."
sleep 20

echo ""
echo "=========================================="
echo "🔗 Connexion des nouveaux nœuds au réseau..."
echo "=========================================="

# Récupérer les ENODEs des validateurs existants
echo "Récupération des ENODEs des validateurs existants..."
ENODES=""
for port in 8541 8542 8543 8544; do
    enode=$(curl -s -X POST http://localhost:$port -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' 2>/dev/null | jq -r '.result.enode' 2>/dev/null)
    if [ -n "$enode" ] && [ "$enode" != "null" ]; then
        ENODES="$ENODES $enode"
        echo "  ✅ Validator port $port: $(echo $enode | cut -d'@' -f1)@..."
    fi
done

# Connecter les nouveaux validateurs aux validateurs existants
for i in $(seq 1 $NEW_VALIDATORS); do
    NODE_NUM=$((LAST_VAL + i))
    RPC_PORT=$((8540 + NODE_NUM))
    
    echo ""
    echo "Connexion de validator$NODE_NUM (port $RPC_PORT)..."
    
    # Récupérer son propre enode
    new_enode=$(curl -s -X POST http://localhost:$RPC_PORT -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' 2>/dev/null | jq -r '.result.enode' 2>/dev/null)
    
    # Connecter aux validateurs existants
    for existing_enode in $ENODES; do
        curl -s -X POST http://localhost:$RPC_PORT -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$existing_enode\\"],\\"id\\":1}" > /dev/null 2>&1 || true
    done
    
    # Connecter les validateurs existants à ce nouveau
    for port in 8541 8542 8543 8544; do
        curl -s -X POST http://localhost:$port -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$new_enode\\"],\\"id\\":1}" > /dev/null 2>&1 || true
    done
    
    echo "  ✅ Connecté au réseau"
done

# Connecter les nouveaux full nodes à tous les validateurs
for i in $(seq 1 $NEW_FULLNODES); do
    NODE_NUM=$((LAST_FN + i))
    RPC_PORT=$((8540 + 4 + NEW_VALIDATORS + i))
    
    echo ""
    echo "Connexion de fullnode$NODE_NUM (port $RPC_PORT)..."
    
    for existing_enode in $ENODES; do
        curl -s -X POST http://localhost:$RPC_PORT -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$existing_enode\\"],\\"id\\":1}" > /dev/null 2>&1 || true
    done
    
    echo "  ✅ Connecté au réseau"
done

echo ""
echo "⏳ Attente de 10 secondes pour la stabilisation..."
sleep 10

echo ""
echo "=========================================="
echo "📊 Vérification finale"
echo "=========================================="

# Vérifier tous les validateurs (anciens + nouveaux)
for port in $(seq 8541 $((8540 + 4 + NEW_VALIDATORS))); do
    peers=$(curl -s -X POST http://localhost:$port -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"admin_peers","params":[],"id":1}' 2>/dev/null | jq '.result | length' 2>/dev/null || echo "0")
    block=$(curl -s -X POST http://localhost:$port -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' 2>/dev/null | jq -r '.result' 2>/dev/null | xargs printf "%d\\n" 2>/dev/null || echo "0")
    echo "Validateur (port $port): $peers peers | Bloc #$block"
done

echo ""
echo "=========================================="
echo "✅ Nouveaux nœuds ajoutés avec succès !"
echo "=========================================="
echo ""
echo "📌 Résumé :"
echo "  - Validateurs totaux: $((4 + NEW_VALIDATORS))"
echo "  - Full nodes totaux: $((4 + NEW_FULLNODES))"
echo ""
echo "🚀 Pour redémarrer tout le réseau plus tard :"
echo "  ./restart-poa-network.sh"
'''

# Sauvegarder le script
with open('/mnt/kimi/output/add-new-nodes.sh', 'w') as f:
    f.write(add_nodes_script)

print("✅ Script pour ajouter des nœuds créé : add-new-nodes.sh")
print("\n" + "="*60)
print(add_nodes_script[:2000] + "...")
