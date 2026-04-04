const { ethers } = require("hardhat");
const fs = require("fs");

async function main() {
  console.log("\n📝 DÉPLOIEMENT DU CONTRAT");
  console.log("=".repeat(50));

  const IoTDeviceManager = await ethers.getContractFactory("IoTDeviceManager");
  const deviceManager = await IoTDeviceManager.deploy();
  await deviceManager.waitForDeployment();
  const address = await deviceManager.getAddress();
  
  console.log(`✅ Contrat déployé à: ${address}`);
  console.log("=".repeat(50) + "\n");
  
  fs.writeFileSync("deployed-address.txt", address);
  console.log("📁 Adresse sauvegardée dans deployed-address.txt");
}

main().catch(console.error);
