require("@nomicfoundation/hardhat-toolbox");

// Clé privée du validateur 1
const PRIVATE_KEY = "0x84ae84f3d0f1a7962b5175ed0a584e07053f9d2b569d86d447a52cb29af984d6";

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: "0.8.20",
  networks: {
    // Utiliser le réseau "localhost" qui est reconnu par défaut
    localhost: {
      url: "http://127.0.0.1:8541",
      chainId: 2026,
      accounts: [PRIVATE_KEY]
    }
  }
};
