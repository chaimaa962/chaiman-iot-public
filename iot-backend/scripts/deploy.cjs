const hre = require("hardhat");

async function main() {
  const [deployer] = await hre.ethers.getSigners();
  console.log("Déploiement avec:", deployer.address);

  const IoTDeviceManager = await hre.ethers.getContractFactory("IoTDeviceManager");
  const contract = await IoTDeviceManager.deploy();
  
  await contract.waitForDeployment();
  const address = await contract.getAddress();
  
  console.log("Contrat déployé à:", address);
  
  const fs = require("fs");
  fs.writeFileSync("deployed-address.txt", address);
}

main().catch(console.error);
