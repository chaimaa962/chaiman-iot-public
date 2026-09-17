#!/bin/bash

echo "═══════════════════════════════════════════════════════════════"
echo "🚀 DÉMARRAGE : 4 Validators + 4 Full Nodes (PoA)"
echo "   Mode: PRÉSERVE les données blockchain"
echo "═══════════════════════════════════════════════════════════════"

cd ~/iot-docker-project
GETH="ethereum/client-go:v1.10.26"

# ============================================
# [FIX-KEYSTORE-GLOB] get_address()
#
# PROBLÈME OBSERVÉ : "cat nodes/<node>/keystore/UTC--* | python3 -c
# '...json.load(sys.stdin)...'" plantait avec
# "json.decoder.JSONDecodeError: Extra data" sur les 8 nœuds.
#
# CAUSE : le wildcard UTC--* correspondait à PLUSIEURS fichiers keystore
# dans certains dossiers (restes d'anciennes exécutions / comptes créés
# plusieurs fois). "cat" concatène alors le contenu de tous ces fichiers
# JSON à la suite sans séparateur, et json.load() ne sait lire qu'UN SEUL
# objet JSON — d'où "Extra data" dès le début du second fichier.
#
# CONSÉQUENCE EN CASCADE : la commande python échouait (exception, pas de
# sortie), donc $V1..$V4 / $FN1..$FN4 / $ADDR devenaient des chaînes VIDES.
# genesis.json se retrouvait avec des adresses vides dans "extradata" et
# "alloc", et les conteneurs Docker démarraient avec
# --miner.etherbase="0x" et --unlock="0x" (adresses invalides) — d'où les
# nœuds "- 0x" (adresse vide) affichés dans les logs, et le blocage
# probable du script en attente de réponses de nœuds mal configurés.
#
# CORRECTION : sélectionner UN SEUL fichier keystore de façon déterministe
# (le premier par ordre alphabétique, généralement le plus ancien = le
# compte d'origine du nœud), et avertir si plusieurs fichiers existent
# pour permettre un nettoyage manuel ultérieur si besoin.
# ============================================
get_address() {
    local node_dir="$1"
    local files=("${node_dir}/keystore/"UTC--*)

    if [ ! -e "${files[0]}" ]; then
        echo "   ❌ Aucun keystore trouvé dans ${node_dir}/keystore/" >&2
        echo ""
        return 1
    fi

    if [ "${#files[@]}" -gt 1 ]; then
        echo "   ⚠️  ${node_dir} : ${#files[@]} keystores trouvés — utilisation de $(basename "${files[0]}") (envisager un nettoyage)" >&2
    fi

    python3 -c "
import json, sys
try:
    with open(sys.argv[1]) as f:
        print(json.load(f)['address'])
except Exception as e:
    print('', file=sys.stdout)
    print(f'Erreur lecture keystore {sys.argv[1]}: {e}', file=sys.stderr)
" "${files[0]}"
}

# ============================================
# 1. NETTOYAGE LÉGER (sans supprimer les blocs)
# ============================================
echo "🧹 Nettoyage léger (préservation des données)..."

# Arrêter les conteneurs proprement
docker stop $(docker ps -aq --filter "name=^v") 2>/dev/null
docker stop $(docker ps -aq --filter "name=^fn") 2>/dev/null
docker rm $(docker ps -aq --filter "name=^v") 2>/dev/null
docker rm $(docker ps -aq --filter "name=^fn") 2>/dev/null

# [NOUVEAU] Ne PAS supprimer les données blockchain
# sudo rm -rf ./nodes/validator*/geth   ← SUPPRIMÉ
# sudo rm -rf ./nodes/fullnode*/geth    ← SUPPRIMÉ

# Vérifier si les datadirs sont initialisés
NEED_INIT=false
for i in 1 2 3 4; do
    if [ ! -d "./nodes/validator${i}/geth" ] || [ ! -f "./nodes/validator${i}/geth/LOCK" ]; then
        NEED_INIT=true
        echo "   ⚠️  validator${i} nécessite initialisation"
    fi
    if [ ! -d "./nodes/fullnode${i}/geth" ] || [ ! -f "./nodes/fullnode${i}/geth/LOCK" ]; then
        NEED_INIT=true
        echo "   ⚠️  fullnode${i} nécessite initialisation"
    fi
done

# ============================================
# 2. INITIALISATION CONDITIONNELLE (uniquement si nécessaire)
# ============================================
if [ "$NEED_INIT" = true ]; then
    echo ""
    echo "📝 Initialisation conditionnelle des datadirs..."
    
    # [FIX-KEYSTORE-GLOB] Récupérer les adresses via get_address() —
    # sélectionne un seul fichier keystore, ne plante plus si plusieurs
    # fichiers UTC--* existent dans un même dossier.
    FN1=$(get_address nodes/fullnode1)
    FN2=$(get_address nodes/fullnode2)
    FN3=$(get_address nodes/fullnode3)
    FN4=$(get_address nodes/fullnode4)

    V1=$(get_address nodes/validator1)
    V2=$(get_address nodes/validator2)
    V3=$(get_address nodes/validator3)
    V4=$(get_address nodes/validator4)

    # [FIX-KEYSTORE-GLOB] Vérification : arrêter proprement si une adresse
    # est vide plutôt que de générer un genesis.json corrompu en silence
    # (c'est exactement ce qui s'est produit avant ce correctif : les
    # adresses vides ont produit un genesis invalide et des nœuds
    # --unlock="0x" cassés).
    for name_val in "V1:$V1" "V2:$V2" "V3:$V3" "V4:$V4" "FN1:$FN1" "FN2:$FN2" "FN3:$FN3" "FN4:$FN4"; do
        val="${name_val#*:}"
        key="${name_val%%:*}"
        if [ -z "$val" ]; then
            echo "❌ Adresse ${key} vide — arrêt du script pour éviter un genesis.json corrompu."
            echo "   Vérifiez le contenu de nodes/*/keystore/ (voir avertissements ci-dessus)."
            exit 1
        fi
    done

    cat > genesis.json << EOF
{
  "config": {
    "chainId": 1337,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "clique": {
      "period": 5,
      "epoch": 30000
    }
  },
  "difficulty": "1",
  "gasLimit": "8000000",
  "extradata": "0x0000000000000000000000000000000000000000000000000000000000000000${V1}${V2}${V3}${V4}0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
  "alloc": {
    "0x${V1}": { "balance": "1000000000000000000000" },
    "0x${V2}": { "balance": "1000000000000000000000" },
    "0x${V3}": { "balance": "1000000000000000000000" },
    "0x${V4}": { "balance": "1000000000000000000000" },
    "0x${FN1}": { "balance": "500000000000000000000" },
    "0x${FN2}": { "balance": "500000000000000000000" },
    "0x${FN3}": { "balance": "500000000000000000000" },
    "0x${FN4}": { "balance": "500000000000000000000" }
  }
}
EOF

    # Initialiser UNIQUEMENT les nœuds qui n'ont pas de données
    for i in 1 2 3 4; do
        if [ ! -d "./nodes/validator${i}/geth" ]; then
            docker run --rm \
              -v "$PWD/nodes/validator${i}:/root/.ethereum" \
              -v "$PWD/genesis.json:/genesis.json" \
              $GETH init --datadir=/root/.ethereum /genesis.json 2>/dev/null
            echo "   ✓ validator${i} initialisé"
        fi
    done
    for i in 1 2 3 4; do
        if [ ! -d "./nodes/fullnode${i}/geth" ]; then
            docker run --rm \
              -v "$PWD/nodes/fullnode${i}:/root/.ethereum" \
              -v "$PWD/genesis.json:/genesis.json" \
              $GETH init --datadir=/root/.ethereum /genesis.json 2>/dev/null
            echo "   ✓ fullnode${i} initialisé"
        fi
    done
else
    echo "   ✓ Tous les datadirs existent — pas d'initialisation nécessaire"
fi

# ============================================
# 3. LANCER LES 4 VALIDATEURS
# ============================================
echo ""
echo "⛏️  Lancement des 4 validateurs..."

for i in 1 2 3 4; do
    PORT=$((8540+i)); P2P=$((30300+i)); IP="172.17.0.$((i+1))"
    ADDR=$(get_address "nodes/validator${i}")
    if [ -z "$ADDR" ]; then
        echo "❌ Adresse validator${i} vide — arrêt (voir avertissements ci-dessus)."
        exit 1
    fi

    docker run -d --name "v${i}" --restart unless-stopped \
      -p ${PORT}:8545 -p ${P2P}:30303 \
      -v "$PWD/nodes/validator${i}:/root/.ethereum" $GETH \
      --datadir=/root/.ethereum --networkid=1337 \
      --http --http.addr=0.0.0.0 --http.port=8545 \
      --http.api=eth,web3,net,admin,personal,miner,txpool,clique,debug \
      --http.corsdomain="*" --http.vhosts="*" \
      --port=30303 --nat=extip:${IP} --discovery.port=30303 \
      --mine --miner.threads=1 \
      --miner.gaslimit=8000000 \
      --miner.etherbase="0x${ADDR}" --unlock="0x${ADDR}" \
      --password=/dev/null --allow-insecure-unlock --verbosity=3 2>/dev/null
    echo "   ✓ v${i} : $IP (port $PORT) - 0x${ADDR}"
done

# ============================================
# 4. LANCER LES 4 FULL NODES
# ============================================
echo "📡 Lancement des 4 full nodes..."

for i in 1 2 3 4; do
    PORT=$((8552+i)); P2P=$((30312+i)); IP="172.17.0.$((13+i))"
    ADDR=$(get_address "nodes/fullnode${i}")
    if [ -z "$ADDR" ]; then
        echo "❌ Adresse fullnode${i} vide — arrêt (voir avertissements ci-dessus)."
        exit 1
    fi

    docker run -d --name "fn${i}" --restart unless-stopped \
      -p ${PORT}:8545 -p ${P2P}:30303 \
      -v "$PWD/nodes/fullnode${i}:/root/.ethereum" $GETH \
      --datadir=/root/.ethereum --networkid=1337 \
      --http --http.addr=0.0.0.0 --http.port=8545 \
      --http.api=eth,web3,net,admin,personal,txpool,clique,debug \
      --http.corsdomain="*" --http.vhosts="*" \
      --port=30303 --nat=extip:${IP} --discovery.port=30303 \
      --unlock="0x${ADDR}" --password=/dev/null --allow-insecure-unlock --verbosity=3 2>/dev/null
    echo "   ✓ fn${i} : $IP (port $PORT) - 0x${ADDR}"
done

echo ""
echo "⏳ Attente démarrage (20s)..."
sleep 20

# ============================================
# 5. RÉCUPÉRER LES ENODES
# ============================================
echo ""
echo "🔗 Récupération des enodes..."

declare -A ENODES
declare -A NODE_PORTS

for i in 1 2 3 4; do
    PORT=$((8540+i))
    NODE_PORTS["v${i}"]=$PORT
    ENODE=$(curl -s -X POST http://localhost:$PORT -H "Content-Type: application/json" \
      -d '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' | jq -r '.result.enode')
    
    IP="172.17.0.$((i+1))"
    ENODE_FIXED=$(echo "$ENODE" | sed "s/\\[::\\]/${IP}/g" | sed 's/?discport=[0-9]*//g')
    ENODES["v${i}"]="$ENODE_FIXED"
    echo "   v${i}: ${ENODE_FIXED:0:70}..."
done

for i in 1 2 3 4; do
    PORT=$((8552+i))
    NODE_PORTS["fn${i}"]=$PORT
    ENODE=$(curl -s -X POST http://localhost:$PORT -H "Content-Type: application/json" \
      -d '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' | jq -r '.result.enode')
    
    IP="172.17.0.$((13+i))"
    ENODE_FIXED=$(echo "$ENODE" | sed "s/\\[::\\]/${IP}/g" | sed 's/?discport=[0-9]*//g')
    ENODES["fn${i}"]="$ENODE_FIXED"
    echo "   fn${i}: ${ENODE_FIXED:0:70}..."
done

# ============================================
# 6. CONNEXION MANUELLE DES PEERS
# ============================================
echo ""
echo "🔗 Connexion manuelle des peers..."

for name1 in v1 v2 v3 v4 fn1 fn2 fn3 fn4; do
    PORT1=${NODE_PORTS[$name1]}
    echo "   Connecting $name1 (port $PORT1)..."
    for name2 in v1 v2 v3 v4 fn1 fn2 fn3 fn4; do
        if [ "$name1" != "$name2" ]; then
            ENODE2=${ENODES[$name2]}
            curl -s -X POST http://localhost:$PORT1 -H "Content-Type: application/json" \
              -d "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"$ENODE2\"],\"id\":1}" > /dev/null 2>&1
        fi
    done
done

echo ""
echo "⏳ Attente synchronisation (15s)..."
sleep 15

# ============================================
# 7. VÉRIFICATION
# ============================================
echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "📊 VÉRIFICATION"
echo "═══════════════════════════════════════════════════════════════"

echo ""
echo "--- Peers connectés ---"
for name in v1 v2 v3 v4 fn1 fn2 fn3 fn4; do
    PORT=${NODE_PORTS[$name]}
    PEERS=$(curl -s -X POST http://localhost:$PORT -H "Content-Type: application/json" \
      -d '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' | jq -r '.result')
    PEERS_DEC=$(printf "%d" "$PEERS" 2>/dev/null || echo "$PEERS")
    echo "   $name (port $PORT) -> $PEERS_DEC peers"
done

echo ""
echo "--- Bloc actuel ---"
for name in v1 v2 v3 v4 fn1 fn2 fn3 fn4; do
    PORT=${NODE_PORTS[$name]}
    BLOCK=$(curl -s -X POST http://localhost:$PORT -H "Content-Type: application/json" \
      -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
    BLOCK_DEC=$(printf "%d" "$BLOCK" 2>/dev/null || echo "$BLOCK")
    echo "   $name (port $PORT) -> Bloc: $BLOCK_DEC"
done

echo ""
echo "--- Minage (validateurs) ---"
for i in 1 2 3 4; do
    PORT=$((8540+i))
    MINING=$(curl -s -X POST http://localhost:$PORT -H "Content-Type: application/json" \
      -d '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' | jq -r '.result')
    HASHRATE=$(curl -s -X POST http://localhost:$PORT -H "Content-Type: application/json" \
      -d '{"jsonrpc":"2.0","method":"eth_hashrate","params":[],"id":1}' | jq -r '.result')
    echo "   v${i} -> Mining: $MINING, Hashrate: $HASHRATE"
done

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "✅ TERMINÉ !"
echo "   Blockchain préservée — pas de réinitialisation"
echo "═══════════════════════════════════════════════════════════════"
