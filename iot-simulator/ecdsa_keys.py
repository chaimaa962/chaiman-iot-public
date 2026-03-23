# ecdsa_keys.py
from eth_keys import keys
import os
import json
import hashlib
import time

class ECDSAManager:
    """Gère les clés ECDSA pour les appareils IoT"""
    
    @staticmethod
    def generate_keys(device_id):
        """Génère une paire de clés ECDSA pour un appareil"""
        private_key_bytes = os.urandom(32)
        private_key = keys.PrivateKey(private_key_bytes)
        public_key = private_key.public_key
        address = public_key.to_checksum_address()
        
        keys_data = {
            "device_id": device_id,
            "private_key": private_key_bytes.hex(),
            "public_key": public_key.to_hex(),
            "address": address,
            "created_at": int(time.time())
        }
        
        # Sauvegarder dans un fichier sécurisé
        os.makedirs("keys", exist_ok=True)
        with open(f"keys/{device_id}.key", "w") as f:
            json.dump(keys_data, f)
        
        return keys_data
    
    @staticmethod
    def load_keys(device_id):
        """Charge les clés d'un appareil"""
        try:
            with open(f"keys/{device_id}.key", "r") as f:
                return json.load(f)
        except:
            return None
    
    @staticmethod
    def sign_message(private_key_hex, message):
        """Signe un message avec la clé privée ECDSA"""
        private_key_bytes = bytes.fromhex(private_key_hex)
        private_key = keys.PrivateKey(private_key_bytes)
        
        # Hacher le message
        if isinstance(message, str):
            message = message.encode()
        message_hash = hashlib.sha256(message).digest()
        
        # Signer
        signature = private_key.sign_msg(message_hash)
        
        return {
            "r": hex(signature.r),
            "s": hex(signature.s),
            "v": signature.v,
            "signature_hex": signature.to_hex()
        }
    
    @staticmethod
    def verify_signature(public_key_hex, message, signature_hex):
        """Vérifie une signature ECDSA"""
        public_key = keys.PublicKey.from_hex(public_key_hex)
        signature = keys.Signature.from_hex(signature_hex)
        
        if isinstance(message, str):
            message = message.encode()
        message_hash = hashlib.sha256(message).digest()
        
        return public_key.verify_msg(message_hash, signature)
