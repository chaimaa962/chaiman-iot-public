#!/bin/bash
# Tester si l'adresse peut appeler registerDevice

curl -X POST http://localhost:8541 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc":"2.0",
    "method":"eth_call",
    "params":[{
        "from":"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
        "to":"0x5FbDB2315678afecb367f032d93F642f64180aa3",
        "data":"0x3f0f0d8a000000000000000000000000f39fd6e51aad88f6f4ce6ab8827279cfffb92266"
    }],
    "id":1
}'
