import json
import os
from eth_keyfile import decode_keyfile_json

password = "123"

print("\n" + "="*70)
print("🔑 EXTRACTION DES CLÉS PRIVÉES POUR LE SIMULATEUR")
print("="*70)

keys = {}

for i in range(1, 5):
    keystore_dir = f"node{i}_data/keystore"
    if not os.path.exists(keystore_dir):
        print(f"❌ Node {i}: dossier keystore non trouvé")
        continue
    
    files = os.listdir(keystore_dir)
    if not files:
        print(f"❌ Node {i}: aucun fichier keystore")
        continue
    
    keystore_file = os.path.join(keystore_dir, files[0])
    
    try:
        with open(keystore_file, 'r') as f:
            keyfile = json.load(f)
        
        # Déchiffrer la clé privée
        private_key_bytes = decode_keyfile_json(keyfile, password)
        private_key_hex = private_key_bytes.hex()
        
        # Récupérer l'adresse
        address = files[0].split('--')[-1]
        
        keys[f"GETH_NODE_{i}_KEY"] = private_key_hex
        keys[f"GETH_NODE_{i}_ADDRESS"] = address
        
        print(f"\n✅ NŒUD {i}")
        print(f"   Adresse: 0x{address}")
        print(f"   Clé privée: {private_key_hex[:20]}...{private_key_hex[-20:]}")
        
    except Exception as e:
        print(f"❌ Node {i}: {e}")

print("\n" + "="*70)
print("📋 FICHIER .env POUR LE SIMULATEUR")
print("="*70)
print("Copiez ceci dans ~/iot-simulator/.env :")
print("\n# ============================================")
print("# CLÉS PRIVÉES DES 4 NŒUDS GETH")
print("# ============================================")

for key, value in keys.items():
    if "KEY" in key:
        print(f"{key}=0x{value}")

print("\n# ============================================")
print("# ADRESSES DES 4 NŒUDS (pour vérification)")
print("# ============================================")
for key, value in keys.items():
    if "ADDRESS" in key:
        print(f"{key}=0x{value}")

# Sauvegarder les clés dans des fichiers
for i in range(1, 5):
    key_name = f"GETH_NODE_{i}_KEY"
    if key_name in keys:
        with open(f"node{i}_private_key.txt", "w") as f:
            f.write(keys[key_name])
        print(f"\n✅ Clé node{i} sauvegardée dans node{i}_private_key.txt")

print("\n" + "="*70)
