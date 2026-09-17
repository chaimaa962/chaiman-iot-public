#!/usr/bin/env python3
import json
import subprocess
import sys

# Vérifier que le contrat existe
try:
    with open('artifacts/contracts/IoTDeviceManager.sol/IoTDeviceManager.json', 'r') as f:
        contract_json = json.load(f)
        bytecode = contract_json['bytecode']
        abi = contract_json['abi']
        print(f"Bytecode trouvé: {len(bytecode)} caractères")
except FileNotFoundError:
    print("Contrat non compilé. Exécutez 'npx hardhat compile' d'abord")
    sys.exit(1)

print("Pour déployer, utilisez: npx hardhat run scripts/deploy.js --network localhost")
