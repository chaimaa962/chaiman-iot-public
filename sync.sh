#!/bin/bash
echo "🔄 Synchronisation avec GitHub..."

cd ~/iot-docker-project

# Ajouter les changements
git add .

# Vérifier s'il y a des modifications
if git diff --cached --quiet && git diff --quiet; then
    echo "✅ Aucun changement à synchroniser"
    exit 0
fi

# Faire un commit
git commit -m "Mise à jour automatique - $(date '+%Y-%m-%d %H:%M:%S')"

# Récupérer les changements distants et pousser
git pull origin main --rebase && git push

if [ $? -eq 0 ]; then
    echo "✅ Synchronisation terminée !"
else
    echo "❌ Erreur lors de la synchronisation"
fi
