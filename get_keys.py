import json
import os
import sys

# Le mot de passe est "123"
password = "123"

for i in range(1, 5):
    keystore_dir = f"node{i}_data/keystore"
    if not os.path.exists(keystore_dir):
        print(f"❌ Node {i}: pas de keystore")
        continue
    
    files = os.listdir(keystore_dir)
    if not files:
        print(f"❌ Node {i}: keystore vide")
        continue
    
    keystore_file = os.path.join(keystore_dir, files[0])
    
    try:
        with open(keystore_file, 'r') as f:
            data = json.load(f)
        
        print(f"\n✅ Node {i} - Fichier: {files[0]}")
        print(f"   Adresse: {files[0].split('--')[-1]}")
        print(f"   Le fichier keystore contient la clé chiffrée")
        print(f"   Pour déchiffrer, il faut le mot de passe: {password}")
        
    except Exception as e:
        print(f"❌ Node {i}: {e}")

print("\n⚠️  Les clés privées sont chiffrées dans le keystore.")
print("💡 Alternative: Créer de nouvelles clés pour le simulateur")
