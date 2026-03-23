import json
from eth_account import Account
import getpass

# Mot de passe que tu as utilisé pour créer les comptes
password = "test123"  # Change si nécessaire

nodes = ['node1', 'node2', 'node3', 'node4']

for node in nodes:
    try:
        # Cherche le fichier keystore
        import glob
        keystore_files = glob.glob(f'/home/chaimaa/iot-blockchain/{node}/keystore/UTC--*')
        
        if not keystore_files:
            print(f"❌ Aucun keystore trouvé pour {node}")
            continue
            
        keystore_path = keystore_files[0]
        print(f"\n📁 {node}: {keystore_path}")
        
        # Charge le keystore
        with open(keystore_path, 'r') as f:
            keystore = json.load(f)
        
        # Déchiffre la clé privée
        private_key = Account.decrypt(keystore, password)
        
        # Convertit en hex
        private_key_hex = private_key.hex()
        
        print(f"✅ Adresse: {keystore['address']}")
        print(f"🔑 Clé privée: 0x{private_key_hex}")
        
        # Sauvegarde dans un fichier
        with open(f'{node}_private_key.txt', 'w') as f:
            f.write(f"0x{private_key_hex}")
            
    except Exception as e:
        print(f"❌ Erreur pour {node}: {e}")

print("\n✅ Clés sauvegardées dans nodeX_private_key.txt")
