#!/bin/bash
docker stop geth-node1 geth-node2 2>/dev/null
docker rm geth-node1 geth-node2 2>/dev/null
echo "✅ Nœuds arrêtés"
