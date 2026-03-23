require("@nomicfoundation/hardhat-toolbox");

module.exports = {
  solidity: {
    version: "0.8.19",
    settings: {
      optimizer: {
        enabled: true,
        runs: 200
      }
    }
  },
  networks: {
    local: {
      url: "http://localhost:8545",
      accounts: ["e75b62988914a005dd33916558676b80c6e38838d951ea37824f6f98aca84e03"]
    }
  }
};
