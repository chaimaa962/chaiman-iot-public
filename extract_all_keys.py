import json
import os
from eth_keyfile import decode_keyfile_json

password = "123"

for i in range(1, 5):
    keystore_dir = f"node{i}_data/keystore"
    if not os.path.exists(keystore_dir):
        print(f"❌ Node {i}: dossier keystore non trouvé")
        continue
    
    keystore_files = os.listdir(keystore_dir)
    if not keystore_files:
        print(f"❌ Node {i}: aucun fichier keystore")
        continue
    
    keystore_file = os.path.join(keystore_dir, keystore_files[0])
    
    try:
        with open(keystore_file, 'r') as f:
            keyfile = json.load(f)
        
        private_key_bytes = decode_keyfile_json(keyfile, password)
        private_key_hex = private_key_bytes.hex()
        
        # Récupérer l'adresse depuis le nom du fichier
        address = keystore_files[0].split('--')[-1]
        
        print(f"\n{'='*60}")
        print(f"🔑 NŒUD {i}")
        print(f"{'='*60}")
        print(f"Adresse: 0x{address}")
        print(f"Clé privée: {private_key_hex}")
        print(f"Clé avec 0x: 0x{private_key_hex}")
        print(f"{'='*60}")
        
    except Exception as e:
        print(f"❌ Node {i}: erreur - {e}")

print("\n\n📋 COPIEZ CES CLÉS DANS VOTRE .env:")
for i in range(1, 5):
    keystore_dir = f"node{i}_data/keystore"
    if os.path.exists(keystore_dir):
        keystore_files = os.listdir(keystore_dir)
        if keystore_files:
            print(f"GETH_NODE_{i}_KEY=0x...")
