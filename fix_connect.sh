#!/bin/bash

# Connecter manuellement les nœuds
echo "🔗 Connexion des nœuds..."

# Récupérer les enodes
for i in 1 2 3 4 5; do
    echo "Récupération enode du nœud $i..."
    ENODE=$(docker exec geth-node$i sh -c "geth attach --datadir /data --exec 'admin.nodeInfo.enode'" 2>/dev/null | tr -d '"')
    if [ -n "$ENODE" ]; then
        echo "Node$i: ${ENODE:0:80}..."
        eval "ENODE$i=$ENODE"
    fi
done

# Connecter les nœuds entre eux
for i in 1 2 3 4 5; do
    for j in 1 2 3 4 5; do
        if [ $i -ne $j ]; then
            ENODE_VAR="ENODE$j"
            ENODE=${!ENODE_VAR}
            if [ -n "$ENODE" ]; then
                echo "Connecter node$i -> node$j"
                docker exec geth-node$i sh -c "geth attach --datadir /data --exec 'admin.addPeer(\"$ENODE\")'" 2>/dev/null
            fi
        fi
    done
done

echo "✅ Connexion terminée"

# Vérifier les pairs
echo ""
echo "📊 Vérification des connexions:"
for i in 1 2 3 4 5; do
    PEERS=$(docker exec geth-node$i sh -c "geth attach --datadir /data --exec 'net.peerCount'" 2>/dev/null)
    echo "Node$i: $PEERS pairs"
done
