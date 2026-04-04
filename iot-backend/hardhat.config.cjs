require("@nomicfoundation/hardhat-toolbox");

module.exports = {
  solidity: "0.8.19",
  networks: {
    localhost: {
      url: "http://172.20.0.11:8541",
      chainId: 2026,
    },
  },
};
