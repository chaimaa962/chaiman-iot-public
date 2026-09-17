const { ethers } = require("hardhat");

async function main() {
    const contractAddress = "0x5FbDB2315678afecb367f032d93F642f64180aa3";
    const abi = JSON.parse(require('fs').readFileSync('iot_contract.abi', 'utf8'));
    
    const signer = await ethers.getSigner("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266");
    const contract = new ethers.Contract(contractAddress, abi, signer);
    
    // Vérifier si l'adresse a le rôle d'admin
    const isRegistered = await contract.isDeviceRegistered(signer.address);
    console.log("Adresse enregistrée?", isRegistered);
}
main().catch(console.error);
