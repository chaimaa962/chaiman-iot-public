// scripts/deploy.js
const hre = require("hardhat");

async function main() {
  console.log("\n📝 DÉPLOIEMENT DU CONTRAT UNIQUE");
  console.log("=".repeat(50));

  console.log("\n1️⃣ Déploiement de IoTDeviceManager...");
  const IoTDeviceManager = await hre.ethers.getContractFactory("IoTDeviceManager");
  const deviceManager = await IoTDeviceManager.deploy();
  await deviceManager.waitForDeployment();
  const deviceManagerAddress = await deviceManager.getAddress();
  console.log(`   ✅ Contrat déployé: ${deviceManagerAddress}`);

  console.log("\n" + "=".repeat(50));
  console.log("🎉 DÉPLOIEMENT TERMINÉ !");
  console.log("=".repeat(50));
  console.log(`📌 Adresse du contrat: ${deviceManagerAddress}`);
  console.log("=".repeat(50) + "\n");
  
  // Sauvegarder l'adresse
  const fs = require("fs");
  const addresses = {
    IoTDeviceManager: deviceManagerAddress,
    timestamp: new Date().toISOString()
  };
  fs.writeFileSync("deployed-address.json", JSON.stringify(addresses, null, 2));
  console.log("📁 Adresse sauvegardée dans deployed-address.json");
}

main().catch(console.error);
