# transactions/signature.py
import hashlib
import json
import time
import secrets
import asyncio
from ecdsa_keys import ECDSAManager
from zkp_integration import ZKPGenerator

class TransactionBuilder:
    def __init__(self, device_keys, use_zkp=True):
        self.device_keys = device_keys
        self.sequence = 0
        self.use_zkp = use_zkp
        self.zkp_gen = ZKPGenerator()
        
    def build_message(self, sensor_data):
        """Construit un message avec anti-rejeu"""
        self.sequence += 1
        
        # Génération du nonce
        nonce = secrets.token_hex(16)
        
        # Timestamp actuel
        timestamp = int(time.time())
        
        # Construction du message complet
        message = {
            "device_id": self.device_keys["device_id"],
            "sequence": self.sequence,
            "timestamp": timestamp,
            "nonce": nonce,
            "data": sensor_data
        }
        
        # Créer le payload à signer
        payload_str = json.dumps(message, sort_keys=True)
        
        return message, payload_str
    
    def sign_transaction(self, payload_str):
        """Signe la transaction avec ECDSA"""
        signature = ECDSAManager.sign_message(
            self.device_keys["private_key"],
            payload_str
        )
        
        return signature
    
    async def prepare_transaction(self, sensor_data):
        """Prépare une transaction complète avec preuve ZKP"""
        message, payload_str = self.build_message(sensor_data)
        signature = self.sign_transaction(payload_str)
        
        # Calculer le hash du message
        message_hash = hashlib.sha256(payload_str.encode()).hexdigest()
        
        # Générer la preuve ZKP
        if self.use_zkp:
            # Soit avec le serveur ZKP
            proof = await self.zkp_gen.generate_proof(
                signature["signature_hex"],
                message_hash,
                self.device_keys["address"],
                self.sequence,
                message["timestamp"],
                self.device_keys["private_key"]
            )
            
            # Fallback: simulation si serveur indisponible
            if not proof:
                proof = ZKPGenerator.simulate_proof(
                    signature["signature_hex"],
                    message_hash
                )
        else:
            proof = None
        
        transaction = {
            "message": message,
            "message_hash": message_hash,
            "signature": signature,
            "address": self.device_keys["address"],
            "public_key": self.device_keys["public_key"],
            "proof": proof
        }
        
        return transaction
