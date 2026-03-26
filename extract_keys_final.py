import json
import os
from eth_keyfile import decode_keyfile_json

password = "123"

print("\n" + "="*70)
print("🔑 EXTRACTION DES CLÉS PRIVÉES")
print("="*70)

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
        
        print(f"\n✅ NŒUD {i}")
        print(f"   Adresse: 0x{address}")
        print(f"   Clé privée: {private_key_hex}")
        
        # Sauvegarder
        with open(f"node{i}_private_key.txt", "w") as f:
            f.write(private_key_hex)
        print(f"   ✅ Sauvegardée dans node{i}_private_key.txt")
        
    except Exception as e:
        print(f"❌ Node {i}: {e}")

print("\n" + "="*70)
print("📋 RÉCAPITULATIF DES CLÉS")
print("="*70)

for i in range(1, 5):
    key_file = f"node{i}_private_key.txt"
    if os.path.exists(key_file):
        with open(key_file, 'r') as f:
            key = f.read().strip()
        print(f"Node {i}: 0x{key}")

print("\n" + "="*70)
