#!/bin/bash

# Renouvellement automatique des certificats Let's Encrypt
echo "🔄 Vérification renouvellement certificats..."

# Renouveler si nécessaire
certbot renew --quiet --deploy-hook "systemctl restart iot-backend"

# Vérifier la date d'expiration
for cert in /etc/letsencrypt/live/*/fullchain.pem; do
    expiry=$(openssl x509 -enddate -noout -in "$cert" | cut -d= -f2)
    echo "📅 Certificat: $cert expire le: $expiry"
done

echo "✅ Vérification terminée"
