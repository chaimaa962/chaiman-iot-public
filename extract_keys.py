import json
from eth_keyfile import decode_keyfile

def extract_private_key(keystore_file, password):
    with open(keystore_file, 'r') as f:
        keyfile = json.load(f)
    private_key = decode_keyfile(keyfile, password.encode())
    return private_key.hex()

# Pour chaque nœud
for i in 1 2 3 4; do
    KEYSTORE=$(ls node${i}_data/keystore/ 2>/dev/null)
    if [ -n "$KEYSTORE" ]; then
        echo "Node $i:"
        python3 -c "
import json
from eth_keyfile import decode_keyfile

with open('node${i}_data/keystore/$KEYSTORE', 'r') as f:
    keyfile = json.load(f)
pk = decode_keyfile(keyfile, '123'.encode())
print(pk.hex())
"
    fi
done
