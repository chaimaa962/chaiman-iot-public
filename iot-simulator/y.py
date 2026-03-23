#!/usr/bin/env python3
"""
Script pour désactiver TOUS les anciens nœuds et enregistrer les 4 nouveaux avec bonnes adresses
"""

import os
import sys
import time
from web3 import Web3
from dotenv import load_dotenv

load_dotenv()

# Configuration
NODE_URL = os.getenv("NODE_URL", "http://localhost:8545")
PRIVATE_KEY = os.getenv("PRIVATE_KEY")  # Clé privée du propriétaire
CONTRACT_ADDRESS = os.getenv("CONTRACT_ADDRESS", "0x6466354d5708b118BC5561fEBC0C5Ba611a697B5")

# ABI complet du contrat
CONTRACT_ABI = [
    {
        "inputs": [{"internalType": "address", "name": "deviceAddress", "type": "address"}],
        "name": "deactivateDevice",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [{"internalType": "address", "name": "deviceAddress", "type": "address"},
                   {"internalType": "string", "name": "_publicKey", "type": "string"},
                   {"internalType": "string", "name": "_metadata", "type": "string"}],
        "name": "registerDevice",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [{"internalType": "address", "name": "deviceAddress", "type": "address"}],
        "name": "isDeviceRegistered",
        "outputs": [{"internalType": "bool", "name": "", "type": "bool"}],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [{"internalType": "address", "name": "deviceAddress", "type": "address"}],
        "name": "getDeviceInfo",
        "outputs": [
            {"internalType": "string", "name": "", "type": "string"},
            {"internalType": "string", "name": "", "type": "string"},
            {"internalType": "bool", "name": "", "type": "bool"},
            {"internalType": "uint256", "name": "", "type": "uint256"}
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [],
        "name": "getAllDevices",
        "outputs": [{"internalType": "address[]", "name": "", "type": "address[]"}],
        "stateMutability": "view",
        "type": "function"
    }
]

# Les 4 anciennes adresses à désactiver (toutes fausses sauf GETH_NODE_3 qui était fausse aussi)
ANCIENNES_ADRESSES = [
    "0x6abd7860df4a432a6e0748825d4c34e2eb3d2ba4",  # GETH_NODE_1
    "0xbe07ddd72c497dc88b41838dedb637d1cce17481",  # GETH_NODE_2
    "0x3d5e31a86612234e8b24e405896df5a95c521dc2",  # GETH_NODE_3 (ancienne fausse)
    "0xf5d36ce3d47850b0635a1f2f8de485ad2dc66986",  # GETH_NODE_4
]

# Les 4 nouvelles adresses correctes avec leurs clés privées
NOUVEAUX_NODES = [
    {
        "device_id": "GETH_NODE_1",
        "address": "0x6abd7860df4a432a6e0748825d4c34e2eb3d2ba4",
        "private_key": os.getenv("GETH_NODE_1_KEY", "0xa0cff945b340cfdc4ac7dc4e097d9c7161dd06ac84819b6d3058daab1b4ac4ae"),
        "interval": 5,
        "node_id": 1,
    },
    {
        "device_id": "GETH_NODE_2",
        "address": "0xbe07ddd72c497dc88b41838dedb637d1cce17481",
        "private_key": os.getenv("GETH_NODE_2_KEY", "0xbbc6b62a960333455f919c1eb821b647f56440b937c64d827df2275fc6cb2ae2"),
        "interval": 10,
        "node_id": 2,
    },
    {
        "device_id": "GETH_NODE_3",
        "address": "0x2235cb45508479ca2a70a99b4238ecf16c8ed36f",  # NOUVELLE ADRESSE CORRECTE
        "private_key": os.getenv("GETH_NODE_3_KEY", "0x7f81d8750ad924395e0e71b4366f5f158f398b4b371c4d737855088ce1840d68"),
        "interval": 15,
        "node_id": 3,
    },
    {
        "device_id": "GETH_NODE_4",
        "address": "0xf5d36ce3d47850b0635a1f2f8de485ad2dc66986",
        "private_key": os.getenv("GETH_NODE_4_KEY", "0x2d3a41dccb3d7bcbbad443b8ff7e7a14e7071233bc1f3470204802c9e91cc072"),
        "interval": 25,
        "node_id": 4,
    }
]

def send_transaction(w3, contract, account, private_key, function_name, *args):
    """Envoie une transaction et attend la confirmation"""
    try:
        tx = getattr(contract.functions, function_name)(*args).build_transaction({
            'from': account.address,
            'nonce': w3.eth.get_transaction_count(account.address),
            'gas': 500000,
            'gasPrice': w3.eth.gas_price
        })

        signed_tx = w3.eth.account.sign_transaction(tx, private_key)
        tx_hash = w3.eth.send_raw_transaction(signed_tx.rawTransaction)

        print(f"   ⏳ Attente confirmation {tx_hash.hex()[:20]}...")
        receipt = w3.eth.wait_for_transaction_receipt(tx_hash, timeout=60)

        if receipt.status == 1:
            print(f"   ✅ Confirmé bloc {receipt.blockNumber}")
            return True
        else:
            print(f"   ❌ Échoué (status {receipt.status})")
            return False
    except Exception as e:
        print(f"   ❌ Erreur: {e}")
        return False

def main():
    print("="*70)
    print("🔧 RÉINITIALISATION COMPLÈTE DES NŒUDS GETH")
    print("="*70)

    if not PRIVATE_KEY:
        print("❌ Clé privée manquante! Définissez PRIVATE_KEY dans .env")
        sys.exit(1)

    # Connexion
    w3 = Web3(Web3.HTTPProvider(NODE_URL))
    if not w3.is_connected():
        print(f"❌ Impossible de se connecter à {NODE_URL}")
        sys.exit(1)

    print(f"✅ Connecté à {NODE_URL} (Chain ID: {w3.eth.chain_id})")

    # Compte
    account = w3.eth.account.from_key(PRIVATE_KEY)
    print(f"✅ Compte propriétaire: {account.address}")

    # Contrat
    contract = w3.eth.contract(address=CONTRACT_ADDRESS, abi=CONTRACT_ABI)
    print(f"✅ Contrat: {CONTRACT_ADDRESS}")

    # ÉTAPE 1: Désactiver tous les anciens nœuds
    print(f"\n{'='*70}")
    print("📛 ÉTAPE 1: DÉSACTIVATION DE TOUS LES ANCIENS NŒUDS")
    print("="*70)

    for addr in ANCIENNES_ADRESSES:
        try:
            checksum_addr = Web3.to_checksum_address(addr)
            is_reg = contract.functions.isDeviceRegistered(checksum_addr).call()

            if is_reg:
                print(f"\n🔴 Désactivation de {addr}...")
                send_transaction(w3, contract, account, PRIVATE_KEY, 'deactivateDevice', checksum_addr)
            else:
                print(f"\nℹ️  {addr} non enregistré, ignoré")
        except Exception as e:
            print(f"\n❌ Erreur avec {addr}: {e}")

    # ÉTAPE 2: Enregistrer les 4 nouveaux nœuds
    print(f"\n{'='*70}")
    print("📝 ÉTAPE 2: ENREGISTREMENT DES 4 NOUVEAUX NŒUDS")
    print("="*70)

    for node in NOUVEAUX_NODES:
        try:
            addr = Web3.to_checksum_address(node['address'])

            # Vérifier si déjà enregistré
            is_reg = contract.functions.isDeviceRegistered(addr).call()
            if is_reg:
                print(f"\nℹ️  {node['device_id']} déjà enregistré, ignoré")
                continue

            print(f"\n📝 Enregistrement {node['device_id']} ({addr})...")

            # Générer clé publique simulée
            from eth_utils import keccak
            public_key = "0x" + keccak(hexstr=node['address']).hex()[:64]

            # Métadonnées
            metadata = f'{{"device_id":"{node["device_id"]}","node_id":{node["node_id"]},"type":"geth_node","interval":{node["interval"]},"permanent":true}}'

            # Enregistrer
            success = send_transaction(w3, contract, account, PRIVATE_KEY, 'registerDevice', addr, public_key, metadata)

            if success:
                print(f"   ✅ {node['device_id']} enregistré avec succès!")
            else:
                print(f"   ❌ Échec enregistrement {node['device_id']}")

            time.sleep(2)  # Délai entre les transactions

        except Exception as e:
            print(f"\n❌ Erreur {node['device_id']}: {e}")
            import traceback
            traceback.print_exc()

    # Vérification finale
    print(f"\n{'='*70}")
    print("📋 VÉRIFICATION FINALE")
    print("="*70)

    try:
        all_devices = contract.functions.getAllDevices().call()
        print(f"\nTotal appareils enregistrés: {len(all_devices)}")

        for addr in all_devices:
            try:
                pk, meta, is_active, last_seen = contract.functions.getDeviceInfo(addr).call()
                status = "🟢 Actif" if is_active else "🔴 Inactif"
                print(f"   {addr[:25]}... - {status}")
            except:
                print(f"   {addr[:25]}... - Erreur lecture")
    except Exception as e:
        print(f"❌ Erreur récupération liste: {e}")

    print(f"\n{'='*70}")
    print("✅ OPÉRATION TERMINÉE")
    print("="*70)
    print("\nVous pouvez maintenant relancer:")
    print("  1. Le backend: go run cmd/api/main.go")
    print("  2. Le simulateur: python3 geth_nodes_simulator.py")

if __name__ == "__main__":
    main()
