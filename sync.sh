#!/bin/bash
echo "🔄 Synchronisation avec GitHub..."

# Aller dans le dossier du projet
cd ~/iot-docker-project

# Ajouter tous les changements
git add .

# Vérifier s'il y a des modifications
if git diff --cached --quiet; then
    echo "✅ Aucun changement à synchroniser"
else
    # Faire un commit avec la date
    git commit -m "Mise à jour automatique - $(date '+%Y-%m-%d %H:%M:%S')"
    
    # Pousser sur GitHub
    git push
    
    echo "✅ Synchronisation terminée !"
fi
