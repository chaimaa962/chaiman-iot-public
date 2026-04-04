#!/bin/bash

echo "========================================="
echo "🚀 SETUP COMPLET DES NŒUDS GETH"
echo "========================================="

# Fonctions
print_step() { echo ""; echo "📌 $1"; echo "========================================="; }

# Vérifier si des nœuds existent déjà
NODES_EXIST=$(ls -d node*_data 2>/dev/null | wc -l)

if [ $NODES_EXIST -eq 0 ]; then
    print_step "Aucun nœud existant - Création des 4 premiers nœuds"
    
    # Créer les 4 premiers nœuds
    for i in 1 2 3 4; do
        echo "  📦 Création du nœud $i..."
        mkdir -p "node${i}_data"
        echo "123" > "password${i}.txt"
        
        ADDRESS=$(docker run --rm \
            -v $(pwd)/node${i}_data:/data \
            -v $(pwd)/password${i}.txt:/data/password.txt \
            ethereum/client-go:latest \
            account new --datadir=/data --password=/data/password.txt 2>&1 | \
            grep "Public address" | awk '{print $NF}')
        
        echo $ADDRESS > node${i}_address.txt
        echo "    ✅ Node$i: $ADDRESS"
    done
    echo "✅ 4 nœuds créés avec succès"
else
    print_step "Des nœuds existent déjà - Récupération des adresses"
fi

# Récupérer les adresses des nœuds existants
print_step "Étape 1: Récupération des adresses des nœuds"

ADDR_LIST=()
for i in $(seq 1 20); do
    if [ -f "node${i}_address.txt" ]; then
        ADDR=$(cat node${i}_address.txt | tr -d '"' | xargs)
        if [ -n "$ADDR" ]; then
            ADDR_LIST+=("$ADDR")
            echo "  ✅ Node$i: $ADDR"
        fi
    elif [ -d "node${i}_data/keystore" ]; then
        # Chercher dans le keystore
        for ks in node${i}_data/keystore/*; do
            if [ -f "$ks" ]; then
                ADDR=$(sudo cat "$ks" 2>/dev/null | grep -o '"address":"[^"]*"' | cut -d'"' -f4)
                if [ -n "$ADDR" ]; then
                    ADDR_LIST+=("$ADDR")
                    echo $ADDR > node${i}_address.txt
                    echo "  ✅ Node$i: $ADDR"
                    break
                fi
            fi
        done
    fi
done

if [ ${#ADDR_LIST[@]} -eq 0 ]; then
    echo "❌ Aucun nœud trouvé après création"
    exit 1
fi

echo "✅ ${#ADDR_LIST[@]} nœuds trouvés"

# Générer le genesis
print_step "Étape 2: Génération du genesis.json avec les adresses"

# Construire l'extradata pour Clique
EXTRA_PREFIX="0x0000000000000000000000000000000000000000000000000000000000000000"
EXTRA_SUFFIX="00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"

EXTRA_DATA="$EXTRA_PREFIX"
for addr in "${ADDR_LIST[@]}"; do
    EXTRA_DATA="${EXTRA_DATA}${addr#0x}"
done
EXTRA_DATA="${EXTRA_DATA}${EXTRA_SUFFIX}"

# Générer le genesis.json
cat > genesis.json << GENEOF
{
  "config": {
    "chainId": 1234,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "berlinBlock": 0,
    "londonBlock": 0,
    "clique": {
      "period": 5,
      "epoch": 30000
    }
  },
  "difficulty": "1",
  "gasLimit": "8000000",
  "extradata": "${EXTRA_DATA}",
  "alloc": {
GENEOF

for i in "${!ADDR_LIST[@]}"; do
    if [ $i -eq $((${#ADDR_LIST[@]} - 1)) ]; then
        echo "    \"${ADDR_LIST[$i]}\": {\"balance\": \"1000000000000000000000000\"}" >> genesis.json
    else
        echo "    \"${ADDR_LIST[$i]}\": {\"balance\": \"1000000000000000000000000\"}," >> genesis.json
    fi
done

cat >> genesis.json << GENEOF
  }
}
GENEOF

echo "✅ Genesis généré avec ${#ADDR_LIST[@]} signeurs"

# Arrêter les conteneurs existants
print_step "Étape 3: Arrêt des conteneurs existants"
for i in $(seq 1 ${#ADDR_LIST[@]}); do
    docker stop geth-node$i 2>/dev/null
    docker rm geth-node$i 2>/dev/null
done
echo "✅ Conteneurs arrêtés"

# Initialiser les nœuds avec le genesis
print_step "Étape 4: Initialisation des nœuds avec le genesis"
for i in $(seq 1 ${#ADDR_LIST[@]}); do
    if [ -d "node${i}_data" ]; then
        echo "  Initialisation node$i..."
        docker run --rm \
            -v $(pwd)/node${i}_data:/data \
            -v $(pwd)/genesis.json:/genesis.json \
            ethereum/client-go:latest \
            init /genesis.json 2>/dev/null
    fi
done
echo "✅ Initialisation terminée"

# Démarrer les nœuds
print_step "Étape 5: Démarrage des nœuds"
for i in $(seq 1 ${#ADDR_LIST[@]}); do
    ADDRESS=$(cat node${i}_address.txt)
    HTTP_PORT=$((8544 + $i))
    P2P_PORT=$((30300 + $i))
    
    echo "  Démarrage node$i sur port $HTTP_PORT..."
    
    docker run -d \
        --name geth-node$i \
        -v $(pwd)/node${i}_data:/data \
        -v $(pwd)/password${i}.txt:/data/password.txt \
        -p ${HTTP_PORT}:8545 \
        -p ${P2P_PORT}:30301 \
        ethereum/client-go:latest \
        --datadir=/data \
        --networkid=1234 \
        --port=30301 \
        --http \
        --http.port=8545 \
        --http.addr=0.0.0.0 \
        --http.api=eth,web3,net,admin,personal,miner,clique,txpool,debug \
        --http.corsdomain=* \
        --http.vhosts=* \
        --mine \
        --miner.etherbase="$ADDRESS" \
        --unlock="$ADDRESS" \
        --password=/data/password.txt \
        --allow-insecure-unlock \
        --verbosity=2 > /dev/null
    
    echo "    ✅ Node$i démarré"
done

# Attendre que les nœuds soient prêts
print_step "Étape 6: Attente du démarrage des nœuds"
sleep 15

# Connecter les nœuds
print_step "Étape 7: Connexion des nœuds entre eux"

# Récupérer les enodes
declare -a ENODES
for i in $(seq 1 ${#ADDR_LIST[@]}); do
    ENODE=$(curl -s -X POST http://localhost:$((8544 + $i)) \
        -H "Content-Type: application/json" \
        --data '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' 2>/dev/null | \
        grep -o '"enode":"[^"]*"' | cut -d'"' -f4)
    ENODES[$i]=$ENODE
    echo "  Node$i enode récupéré"
done

# Connecter chaque nœud aux autres
for i in $(seq 1 ${#ADDR_LIST[@]}); do
    for j in $(seq 1 ${#ADDR_LIST[@]}); do
        if [ $i -ne $j ] && [ -n "${ENODES[$j]}" ]; then
            curl -s -X POST http://localhost:$((8544 + $i)) \
                -H "Content-Type: application/json" \
                --data "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"${ENODES[$j]}\"],\"id\":1}" > /dev/null
        fi
    done
done
echo "✅ Nœuds connectés"

# Attendre la création des blocs
print_step "Étape 8: Attente 30 secondes pour la création des blocs"
sleep 30

# Vérification finale
print_step "Étape 9: Vérification finale"
echo ""
echo "📊 ÉTAT DES NŒUDS"
echo "========================================="

for i in $(seq 1 ${#ADDR_LIST[@]}); do
    PORT=$((8544 + $i))
    
    # Vérifier le bloc
    BLOCK=$(curl -s -X POST http://localhost:$PORT \
        -H "Content-Type: application/json" \
        --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' 2>/dev/null | \
        grep -o '"result":"0x[^"]*"' | cut -d'"' -f4)
    
    # Vérifier les pairs
    PEERS=$(curl -s -X POST http://localhost:$PORT \
        -H "Content-Type: application/json" \
        --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' 2>/dev/null | \
        grep -o '"result":"0x[^"]*"' | cut -d'"' -f4)
    
    if [ -n "$BLOCK" ]; then
        echo "✅ Node$i: $((BLOCK)) blocs, $((PEERS)) pairs"
    else
        echo "❌ Node$i: Non accessible"
    fi
done

echo ""
echo "========================================="
echo "✅ SETUP TERMINÉ !"
echo "========================================="
echo ""
echo "📝 Commandes utiles:"
echo "   ./node_manager.sh status    - Voir l'état des nœuds"
echo "   ./node_manager.sh add       - Ajouter un nouveau nœud"
echo "   ./node_manager.sh blocks    - Voir les blocs créés"
echo ""
echo "⛓️ Pour surveiller les blocs en temps réel:"
echo "   watch -n 2 './node_manager.sh blocks'"
echo "========================================="
