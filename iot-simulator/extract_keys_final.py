import json
from eth_account import Account

# Mot de passe = 123
password = "123"

nodes = ['node1', 'node2', 'node3', 'node4']
results = {}

print("="*70)
print("🔐 EXTRACTION DES CLÉS PRIVÉES")
print("="*70)

for node in nodes:
    print(f"\n📁 {node.upper()}")
    print("-"*50)
    
    try:
        import glob
        keystore_files = glob.glob(f'/home/chaimaa/iot-blockchain/{node}/keystore/UTC--*')
        
        if not keystore_files:
            print(f"❌ Aucun keystore trouvé pour {node}")
            continue
            
        keystore_path = keystore_files[0]
        print(f"📄 Fichier: {keystore_path}")
        
        # Charge le keystore
        with open(keystore_path, 'r') as f:
            keystore = json.load(f)
        
        address = keystore['address']
        print(f"📍 Adresse: 0x{address}")
        
        # Déchiffre la clé privée
        private_key = Account.decrypt(keystore, password)
        private_key_hex = private_key.hex()
        
        print(f"✅ Clé privée trouvée!")
        print(f"🔑 0x{private_key_hex}")
        
        # Sauvegarde
        with open(f'{node}_private_key.txt', 'w') as f:
            f.write(f"0x{private_key_hex}")
        
        results[node] = {
            'address': f"0x{address}",
            'private_key': f"0x{private_key_hex}"
        }
            
    except Exception as e:
        print(f"❌ Erreur: {e}")

print("\n" + "="*70)
print("📋 RÉSUMÉ POUR .env")
print("="*70)

# Affiche le contenu pour .env
env_content = ""
for node, data in results.items():
    print(f"{node.upper()}: {data['address']}")
    env_line = f"{node.upper().replace('NODE', 'GETH_NODE')}_KEY={data['private_key']}"
    env_content += env_line + "\n"
    print(f"   {env_line}")
    print()

print("\n" + "="*70)
print("📝 COPIE CE BLOC POUR .env")
print("="*70)
print(env_content)

# Sauvegarde aussi dans un fichier .env
with open('.env', 'w') as f:
    f.write(env_content)

print("\n✅ Fichier .env créé avec les clés privées!")
