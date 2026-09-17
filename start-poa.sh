#!/bin/bash

cd ~/iot-docker-project

echo "========================================="
echo "🚀 Démarrage Réseau PoA Complet"
echo "========================================="

# Nettoyage
docker rm -f v1 v2 v3 v4 fn1 fn2 fn3 fn4 2>/dev/null

# Initialiser et démarrer les validateurs
for i in 1 2 3 4; do
    PORT=$((8540+i))
    P2P=$((30300+i))
    IP="172.17.0.$((i+1))"
    ADDR=$(cat nodes/validator${i}/address.txt | tr -d '\n')
    
    echo "🔧 Initialisation v${i} (${ADDR})"
    
    # Initialiser avec genesis
    docker run --rm -v "$PWD/nodes/validator${i}:/root/.ethereum" \
      -v "$PWD/genesis.json:/root/genesis.json" \
      ethereum/client-go:v1.10.26 init /root/genesis.json 2>/dev/null
    
    # Démarrer le validateur
    docker run -d --name "v${i}" --restart unless-stopped \
      -p ${PORT}:8545 -p ${P2P}:30303 \
      -v "$PWD/nodes/validator${i}:/root/.ethereum" \
      ethereum/client-go:v1.10.26 \
      --datadir=/root/.ethereum \
      --networkid=2026 \
      --http --http.addr=0.0.0.0 --http.port=8545 \
      --http.api=eth,web3,net,admin,personal,miner,txpool,clique,debug \
      --http.corsdomain="*" \
      --port=30303 \
      --nat=extip:${IP} \
      --mine --miner.threads=1 \
      --miner.etherbase=${ADDR} \
      --unlock=${ADDR} \
      --password=/dev/null \
      --allow-insecure-unlock \
      --verbosity=3
    
    echo "   ✅ v${i} démarré sur port ${PORT}"
done

echo ""
echo "🔗 Connexion des validateurs entre eux..."

# Attendre que les nœuds démarrent
sleep 10

# Récupérer les enodes et connecter
declare -A ENODES
for i in 1 2 3 4; do
    PORT=$((8540+i))
    IP="172.17.0.$((i+1))"
    ENODE=$(curl -s -X POST http://localhost:${PORT} \
      -H "Content-Type: application/json" \
      --data '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' \
      | jq -r '.result.enode' | sed "s/\[::\]/${IP}/" | sed 's/?discport=.*//')
    ENODES["v${i}"]=$ENODE
done

# Connecter v1 aux autres
for i in 2 3 4; do
    curl -s -X POST http://localhost:8541 \
      -H "Content-Type: application/json" \
      --data "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"${ENODES[v$i]}\"],\"id\":1}" > /dev/null
    echo "   Connecté v1 -> v${i}"
done

echo ""
echo "⛏️  Démarrage du minage..."
for port in 8541 8542 8543 8544; do
    curl -s -X POST http://localhost:$port \
      -H "Content-Type: application/json" \
      --data '{"jsonrpc":"2.0","method":"miner_start","params":[1],"id":1}' > /dev/null
    echo "   Minage sur port $port"
done

echo ""
echo "📊 Vérification des blocs..."
sleep 5
BLOCK=$(curl -s -X POST http://localhost:8541 \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
echo "   Bloc actuel: $BLOCK"

echo ""
echo "========================================="
echo "✅ Réseau PoA prêt !"
echo "   🔗 http://localhost:8541"
echo "   📊 Surveiller: watch -n 2 'curl -s ...'"
echo "========================================="

echo ""
echo "🔗 Connexion des FULLNODES aux validateurs..."

# Démarrer les fullnodes
for i in 1 2 3 4; do
    PORT=$((8552+i))
    P2P=$((30312+i))
    IP="172.17.0.$((13+i))"
    ADDR=$(cat nodes/fullnode${i}/address.txt | tr -d '\n')
    
    echo "🔧 Initialisation fn${i} (${ADDR})"
    
    # Initialiser avec genesis
    docker run --rm -v "$PWD/nodes/fullnode${i}:/root/.ethereum" \
      -v "$PWD/genesis.json:/root/genesis.json" \
      ethereum/client-go:v1.10.26 init /root/genesis.json 2>/dev/null
    
    # Démarrer le fullnode
    docker run -d --name "fn${i}" --restart unless-stopped \
      -p ${PORT}:8545 -p ${P2P}:30303 \
      -v "$PWD/nodes/fullnode${i}:/root/.ethereum" \
      ethereum/client-go:v1.10.26 \
      --datadir=/root/.ethereum \
      --networkid=2026 \
      --http --http.addr=0.0.0.0 --http.port=8545 \
      --http.api=eth,web3,net,admin,personal,txpool,clique,debug \
      --http.corsdomain="*" \
      --port=30303 \
      --nat=extip:${IP} \
      --verbosity=3
    
    echo "   ✅ fn${i} démarré sur port ${PORT}"
done

echo ""
echo "🔗 Connexion de TOUS les nœuds (validateurs + fullnodes)..."

sleep 10

# Récupérer les enodes des validateurs
declare -A ENODES_VALIDATORS
for i in 1 2 3 4; do
    PORT=$((8540+i))
    IP="172.17.0.$((i+1))"
    ENODE=$(curl -s -X POST http://localhost:${PORT} \
      -H "Content-Type: application/json" \
      --data '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' \
      | jq -r '.result.enode' | sed "s/\[::\]/${IP}/" | sed 's/?discport=.*//')
    ENODES_VALIDATORS["v${i}"]=$ENODE
done

# Récupérer les enodes des fullnodes
declare -A ENODES_FULLNODES
for i in 1 2 3 4; do
    PORT=$((8552+i))
    IP="172.17.0.$((13+i))"
    ENODE=$(curl -s -X POST http://localhost:${PORT} \
      -H "Content-Type: application/json" \
      --data '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' \
      | jq -r '.result.enode' | sed "s/\[::\]/${IP}/" | sed 's/?discport=.*//')
    ENODES_FULLNODES["fn${i}"]=$ENODE
done

# Connecter chaque validateur à tous les autres nœuds
for v in 1 2 3 4; do
    PORT=$((8540+$v))
    echo "   Connecteur v${v}..."
    
    # Connecter aux autres validateurs
    for other in 1 2 3 4; do
        if [ "$v" != "$other" ]; then
            curl -s -X POST http://localhost:${PORT} \
              -H "Content-Type: application/json" \
              --data "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"${ENODES_VALIDATORS[v$other]}\"],\"id\":1}" > /dev/null
        fi
    done
    
    # Connecter aux fullnodes
    for fn in 1 2 3 4; do
        curl -s -X POST http://localhost:${PORT} \
          -H "Content-Type: application/json" \
          --data "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"${ENODES_FULLNODES[fn$fn]}\"],\"id\":1}" > /dev/null
    done
done

# Connecter chaque fullnode à tous les validateurs
for fn in 1 2 3 4; do
    PORT=$((8552+$fn))
    echo "   Connecteur fn${fn}..."
    
    for v in 1 2 3 4; do
        curl -s -X POST http://localhost:${PORT} \
          -H "Content-Type: application/json" \
          --data "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"${ENODES_VALIDATORS[v$v]}\"],\"id\":1}" > /dev/null
    done
done

echo ""
echo "📊 Vérification des connexions..."

sleep 3

# Afficher le nombre de pairs pour chaque nœud
for i in 1 2 3 4; do
    PEERS=$(curl -s -X POST http://localhost:854${i} \
      -H "Content-Type: application/json" \
      --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' | jq -r '.result')
    echo "   v${i}: $PEERS pairs"
done

for i in 1 2 3 4; do
    PEERS=$(curl -s -X POST http://localhost:855$((2+i)) \
      -H "Content-Type: application/json" \
      --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' | jq -r '.result')
    echo "   fn${i}: $PEERS pairs"
done

