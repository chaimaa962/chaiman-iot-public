#!/bin/bash
# DIAGNOSTIC ET CORRECTION BLOCKCHAIN

echo "════════════════════════════════════════════════════════════"
echo "  DIAGNOSTIC BLOCKCHAIN"
echo "════════════════════════════════════════════════════════════"

# 1. Vérifier les données existantes
echo "1. Vérification des données blockchain..."
for dir in node*_data; do
    if [ -d "$dir/geth/chaindata" ]; then
        count=$(ls -1 "$dir/geth/chaindata" 2>/dev/null | wc -l)
        echo "  $dir/geth/chaindata: $count fichiers"
    else
        echo "  $dir/geth/chaindata: N'EXISTE PAS"
    fi
done

echo ""
echo "2. Arrêt de tous les geth..."
sudo pkill -9 geth 2>/dev/null || true
sleep 2

echo ""
echo "3. Suppression COMPLÈTE des données..."
for dir in node*_data; do
    if [ -d "$dir" ]; then
        echo "  Suppression $dir/geth..."
        sudo rm -rf "$dir/geth"
        sudo rm -f "$dir/geth.ipc"
    fi
done

echo ""
echo "4. Nettoyage fichiers..."
sudo rm -f node*.pid node*.log
find . -name "LOCK" -delete 2>/dev/null || true

echo ""
echo "5. Vérification post-nettoyage..."
for dir in node*_data; do
    if [ -d "$dir/geth" ]; then
        echo "  ⚠️  $dir/geth existe encore!"
    else
        echo "  ✅ $dir/geth supprimé"
    fi
done

echo ""
echo "════════════════════════════════════════════════════════════"
echo "  NETTOYAGE TERMINÉ"
echo "════════════════════════════════════════════════════════════"
echo ""
echo "Prochaines étapes:"
echo "  ./node_manager.sh init-all"
echo "  ./node_manager.sh start"
echo "  sleep 30 && ./node_manager.sh check-all"
