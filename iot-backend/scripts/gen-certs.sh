#!/bin/bash
set -e

CERT_DIR="./certs"
mkdir -p "$CERT_DIR"

echo "=== 1. Génération de la CA racine ==="
openssl genrsa -out "$CERT_DIR/ca.key" 4096
openssl req -new -x509 -days 3650 -key "$CERT_DIR/ca.key" \
  -out "$CERT_DIR/ca.crt" \
  -subj "/C=DZ/ST=Oran/L=Oran/O=IoT-CA/CN=IoT-Root-CA"

echo "=== 2. Génération de la clé serveur ==="
openssl genrsa -out "$CERT_DIR/server.key" 2048

echo "=== 3. CSR avec SAN ==="
cat > "$CERT_DIR/server.ext" <<EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
prompt = no

[req_distinguished_name]
C  = DZ
ST = Oran
L  = Oran
O  = IoT-Backend
CN = localhost

[v3_req]
subjectAltName = @alt_names
keyUsage = keyEncipherment, digitalSignature
extendedKeyUsage = serverAuth

[alt_names]
DNS.1 = localhost
DNS.2 = *.localhost
IP.1  = 127.0.0.1
IP.2  = ::1
EOF

openssl req -new -key "$CERT_DIR/server.key" \
  -out "$CERT_DIR/server.csr" \
  -config "$CERT_DIR/server.ext"

echo "=== 4. Signature par la CA ==="
openssl x509 -req -days 825 \
  -in "$CERT_DIR/server.csr" \
  -CA "$CERT_DIR/ca.crt" \
  -CAkey "$CERT_DIR/ca.key" \
  -CAcreateserial \
  -out "$CERT_DIR/server.crt" \
  -extensions v3_req \
  -extfile "$CERT_DIR/server.ext"

echo "=== 5. Vérification ==="
openssl verify -CAfile "$CERT_DIR/ca.crt" "$CERT_DIR/server.crt"
openssl x509 -in "$CERT_DIR/server.crt" -noout -text | grep -A4 "Subject Alternative Name"

echo "✅ Certificats générés dans $CERT_DIR/"
ls -la "$CERT_DIR/"*.crt "$CERT_DIR/"*.key
