#!/bin/bash

# Monitoring des certificats
DOMAIN="votre-domaine.com"
EXPIRY_DAYS_WARNING=30

expiry_date=$(openssl x509 -enddate -noout -in /etc/letsencrypt/live/$DOMAIN/fullchain.pem | cut -d= -f2)
expiry_epoch=$(date -d "$expiry_date" +%s)
current_epoch=$(date +%s)
days_left=$(( ($expiry_epoch - $current_epoch) / 86400 ))

if [ $days_left -lt $EXPIRY_DAYS_WARNING ]; then
    echo "⚠️ Certificat expire dans $days_left jours"
    # Envoyer alerte (email, webhook, etc.)
    curl -X POST https://hooks.slack.com/... \
         -d "{\"text\":\"Certificat $DOMAIN expire dans $days_left jours\"}"
else
    echo "✅ Certificat valide pour $days_left jours"
fi
