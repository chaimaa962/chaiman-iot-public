#!/bin/bash
# Node Manager V12 - Démarrage sans minage → connexion → activation minage
# Usage: ./node_manager.sh [command] [options]

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GENESIS_FILE="$PROJECT_DIR/genesis.json"
NETWORK_ID="1337"
BASE_HTTP_PORT=8545
BASE_AUTHRPC_PORT=8551
BASE_DISCOVERY_PORT=30303
GETH_BINARY="geth"
DATA_DIR_PREFIX="node"
PASSWORD_PREFIX="password"
ADDRESS_PREFIX="node"

RED='\\033[0;31m'
GREEN='\\033[0;32m'
YELLOW='\\033[1;33m'
BLUE='\\033[0;34m'
CYAN='\\033[0;36m'
NC='\\033[0m'

show_help() {
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║         GESTIONNAIRE DE NŒUDS BLOCKCHAIN POA V12          ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${GREEN}Commandes:${NC}"
    echo "  start            - Démarrage → connexion → minage"
    echo "  recover          - Récupération complète"
    echo "  stop             - Arrêt des nœuds"
    echo "  force-clean      - Nettoyage ULTRA agressif"
    echo "  check-all        - Vérifie les blocs"
    echo ""
}

check_geth() {
    if ! command -v $GETH_BINARY &> /dev/null; then
        echo -e "${RED}Erreur: geth non installé${NC}"
        exit 1
    fi
}

get_existing_nodes() {
    local nodes=""
    for dir in "$PROJECT_DIR"/node*_data; do
        if [ -d "$dir" ]; then
            local num=$(basename "$dir" | sed 's/node//' | sed 's/_data//' | grep -oE '^[0-9]+')
            if [ ! -z "$num" ]; then
                local addr_file="$PROJECT_DIR/${ADDRESS_PREFIX}${num}_address.txt"
                if [ -f "$addr_file" ] && [ -d "$dir/keystore" ]; then
                    if [ -z "$nodes" ]; then
                        nodes="$num"
                    else
                        nodes="$nodes $num"
                    fi
                fi
            fi
        fi
    done
    echo "$nodes"
}

kill_port() {
    local port=$1
    if command -v fuser &> /dev/null; then
        fuser -k ${port}/tcp 2>/dev/null || true
        fuser -k ${port}/udp 2>/dev/null || true
    fi
    local pids=$(ss -tlnp 2>/dev/null | grep ":$port " | grep -oE 'pid=[0-9]+' | sed 's/pid=//' || echo "")
    for pid in $pids; do
        if [ ! -z "$pid" ]; then
            kill -9 $pid 2>/dev/null || true
        fi
    done
    local lsof_pids=$(lsof -t -i :$port 2>/dev/null || echo "")
    for pid in $lsof_pids; do
        if [ ! -z "$pid" ]; then
            kill -9 $pid 2>/dev/null || true
        fi
    done
}

force_clean() {
    echo -e "${YELLOW}════════════════════════════════════════════════════════════${NC}"
    echo -e "${YELLOW}  NETTOYAGE ULTRA AGRESSIF${NC}"
    echo -e "${YELLOW}════════════════════════════════════════════════════════════${NC}"
    
    pkill -9 -f "geth" 2>/dev/null || true
    killall -9 geth 2>/dev/null || true
    sleep 2
    
    for i in 0 1 2 3 4 5 6 7 8 9; do
        kill_port $((8545 + i * 10))
        kill_port $((8551 + i * 10))
    done
    
    for port in 30303 30304 30305 30306 30307 30308 30309 30310; do
        kill_port $port
    done
    
    sleep 1
    
    find "$PROJECT_DIR" -name "LOCK" -type f -delete 2>/dev/null || true
    find "$PROJECT_DIR" -name "*.ipc" -type f -delete 2>/dev/null || true
    rm -f "$PROJECT_DIR"/node*.pid 2>/dev/null || true
    
    sudo chown -R $USER:$USER "$PROJECT_DIR" 2>/dev/null || true
    chmod -R u+rwx "$PROJECT_DIR"/node*_data 2>/dev/null || true
    
    echo -e "${GREEN}✓ Nettoyage terminé${NC}"
}

init_all_nodes() {
    if [ ! -f "$GENESIS_FILE" ]; then
        echo -e "${RED}Erreur: genesis.json manquant!${NC}"
        return 1
    fi
    
    echo -e "${BLUE}Initialisation...${NC}"
    
    for dir in "$PROJECT_DIR"/node*_data; do
        if [ -d "$dir" ]; then
            local num=$(basename "$dir" | sed 's/node//' | sed 's/_data//' | grep -oE '^[0-9]+')
            if [ ! -z "$num" ]; then
                echo -n "  Nœud $num... "
                rm -rf "$dir/geth" "$dir/geth.ipc" 2>/dev/null || true
                $GETH_BINARY init --datadir "$dir" "$GENESIS_FILE" 2>&1 | grep "Successfully" | head -1 || echo "OK"
            fi
        fi
    done
    
    echo -e "${GREEN}✓ Initialisés${NC}"
}

# V12: Démarrage SANS minage → connexion → activation minage
start_nodes_v12() {
    local nodes=$(get_existing_nodes)
    
    if [ -z "$nodes" ]; then
        echo -e "${RED}Aucun nœud existant!${NC}"
        return 1
    fi
    
    local nodes_array=($nodes)
    
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║  Démarrage V12 - Sans minage → Connexion → Minage        ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    
    # ÉTAPE 1: Démarrer tous les nœuds SANS minage (pas de --mine)
    echo -e "${CYAN}ÉTAPE 1: Démarrage SANS minage...${NC}"
    
    for num in $nodes; do
        local data_dir="$PROJECT_DIR/${DATA_DIR_PREFIX}${num}_data"
        local address_file="$PROJECT_DIR/${ADDRESS_PREFIX}${num}_address.txt"
        local password_file="$PROJECT_DIR/${PASSWORD_PREFIX}${num}.txt"
        
        if [ ! -d "$data_dir" ] || [ ! -f "$address_file" ] || [ ! -f "$password_file" ]; then
            echo -e "${RED}Nœud $num incomplet${NC}"
            continue
        fi
        
        local http_port=$((BASE_HTTP_PORT + (num - 1) * 10))
        local authrpc_port=$((BASE_AUTHRPC_PORT + (num - 1) * 10))
        local discovery_port=$((BASE_DISCOVERY_PORT + num - 1))
        local address=$(cat "$address_file")
        
        rm -f "$data_dir/geth/LOCK" "$data_dir/geth.ipc" 2>/dev/null || true
        
        if [ ! -d "$data_dir/geth/chaindata" ]; then
            $GETH_BINARY init --datadir "$data_dir" "$GENESIS_FILE" > /dev/null 2>&1
        fi
        
        echo -e "${BLUE}  Démarrage nœud $num (HTTP:$http_port, SANS minage)...${NC}"
        
        # Démarrer SANS --mine mais AVEC --unlock pour que le compte soit prêt
        nohup $GETH_BINARY --datadir "$data_dir" --networkid $NETWORK_ID --http --http.addr "0.0.0.0" --http.port $http_port --http.api "eth,net,web3,personal,clique,admin,miner" --http.corsdomain "*" --http.vhosts "*" --authrpc.addr "localhost" --authrpc.port $authrpc_port --authrpc.vhosts "localhost" --port $discovery_port --allow-insecure-unlock --unlock "$address" --password "$password_file" --maxpeers 25 --nat "any" --verbosity 2 > "$PROJECT_DIR/node${num}.log" 2>&1 &
        
        echo $! > "$PROJECT_DIR/node${num}.pid"
    done
    
    # Attente minimale pour que les RPC soient prêts
    echo ""
    echo -e "${CYAN}Attente RPC (2s)...${NC}"
    sleep 2
    
    # Vérifier que tous les nœuds sont actifs
    local active=0
    for num in $nodes; do
        local http_port=$((BASE_HTTP_PORT + (num - 1) * 10))
        local pid=$(cat "$PROJECT_DIR/node${num}.pid" 2>/dev/null)
        
        if [ ! -z "$pid" ] && ps -p $pid > /dev/null 2>&1; then
            if curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:$http_port > /dev/null 2>&1; then
                ((active++))
            fi
        fi
    done
    
    echo -e "${GREEN}✓ $active/${#nodes_array[@]} nœuds RPC prêts${NC}"
    
    # ÉTAPE 2: Connecter tous les nœuds entre eux
    echo ""
    echo -e "${CYAN}ÉTAPE 2: Connexion des nœuds...${NC}"
    
    local enodes=""
    for num in $nodes; do
        local http_port=$((BASE_HTTP_PORT + (num - 1) * 10))
        local enode_info=$(curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' http://localhost:$http_port 2>/dev/null)
        local enode=$(echo "$enode_info" | grep -oE '"enode":"[^"]+' | sed 's/"enode":"//' || echo "")
        if [ ! -z "$enode" ]; then
            if [ -z "$enodes" ]; then
                enodes="$enode"
            else
                enodes="$enodes $enode"
            fi
        fi
    done
    
    for num in $nodes; do
        local http_port=$((BASE_HTTP_PORT + (num - 1) * 10))
        for enode in $enodes; do
            if [[ ! "$enode" =~ "@127.0.0.1:$((BASE_DISCOVERY_PORT + num - 1))" ]]; then
                curl -s -X POST -H "Content-Type: application/json" --data "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode\\"],\\"id\\":1}" http://localhost:$http_port > /dev/null 2>&1
            fi
        done
        echo -e "${GREEN}  ✓ Nœud $num connecté${NC}"
    done
    
    # ÉTAPE 3: Activer le minage sur tous les nœuds
    echo ""
    echo -e "${CYAN}ÉTAPE 3: Activation du minage...${NC}"
    
    for num in $nodes; do
        local http_port=$((BASE_HTTP_PORT + (num - 1) * 10))
        
        # Activer le minage via RPC
        curl -s -X POST -H "Content-Type: application/json" \
            --data '{"jsonrpc":"2.0","method":"miner_start","params":[],"id":1}' \
            http://localhost:$http_port > /dev/null 2>&1
        
        echo -e "${GREEN}  ✓ Minage activé sur nœud $num${NC}"
    done
    
    echo ""
    echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}  ✓ ${#nodes_array[@]} nœuds connectés et minant ensemble${NC}"
    echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
    echo ""
    echo -e "${CYAN}La blockchain produit des blocs!${NC}"
    echo ""
    echo "Vérifiez dans 30 secondes:"
    echo "  sleep 30 && ./node_manager.sh check-all"
}

check_all_blocks() {
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║  VÉRIFICATION DES BLOCS                                    ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    
    local nodes=$(get_existing_nodes)
    if [ -z "$nodes" ]; then
        echo -e "${RED}Aucun nœud${NC}"
        return 1
    fi
    
    local active=0
    local total_blocks=0
    
    for num in $nodes; do
        local http_port=$((BASE_HTTP_PORT + (num - 1) * 10))
        
        local response=$(curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:$http_port 2>/dev/null)
        
        if [ -z "$response" ] || [[ "$response" == *"error"* ]]; then
            echo -e "${RED}Nœud $num: Inaccessible${NC}"
            continue
        fi
        
        local block_hex=$(echo "$response" | grep -oE '"result":"[0-9a-fx]+' | sed 's/"result":"//' || echo "0x0")
        local block_num=$(printf "%d" "$block_hex" 2>/dev/null || echo "0")
        
        local block_info=$(curl -s -X POST -H "Content-Type: application/json" --data "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"eth_getBlockByNumber\\",\\"params\\":[\\"latest\\",false],\\"id\\":1}" http://localhost:$http_port 2>/dev/null)
        
        local hash=$(echo "$block_info" | grep -oE '"hash":"[^"]+' | head -1 | sed 's/"hash":"//' || echo "N/A")
        local miner=$(echo "$block_info" | grep -oE '"miner":"[^"]+' | sed 's/"miner":"//' || echo "N/A")
        
        local peers_resp=$(curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' http://localhost:$http_port 2>/dev/null)
        local peers_hex=$(echo "$peers_resp" | grep -oE '"result":"[0-9a-fx]+' | sed 's/"result":"//' || echo "0x0")
        local peers=$(printf "%d" "$peers_hex" 2>/dev/null || echo "0")
        
        echo -e "${GREEN}Nœud $num:${NC} Bloc ${CYAN}#$block_num${NC}, ${hash:0:20}..., Mineur: ${miner:0:20}..., Pairs: $peers"
        ((active++))
        total_blocks=$((total_blocks + block_num))
    done
    
    local nodes_array=($nodes)
    
    echo ""
    echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}  Actifs: $active/${#nodes_array[@]} | Blocs totaux: $total_blocks${NC}"
    echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
}

stop_nodes() {
    echo -e "${BLUE}Arrêt des nœuds...${NC}"
    local stopped=0
    
    for pid_file in "$PROJECT_DIR"/node*.pid; do
        if [ -f "$pid_file" ]; then
            local pid=$(cat "$pid_file" 2>/dev/null)
            if [ ! -z "$pid" ]; then
                kill $pid 2>/dev/null && ((stopped++))
            fi
            rm -f "$pid_file"
        fi
    done
    
    echo -e "${GREEN}✓ $stopped arrêté(s)${NC}"
}

check_geth

case "${1:-help}" in
    start)
        start_nodes_v12
        ;;
    recover)
        force_clean
        init_all_nodes
        start_nodes_v12
        ;;
    stop)
        stop_nodes
        ;;
    force-clean)
        force_clean
        ;;
    init-all)
        init_all_nodes
        ;;
    check-all)
        check_all_blocks
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo -e "${RED}Commande inconnue: $1${NC}"
        show_help
        exit 1
        ;;
esac
