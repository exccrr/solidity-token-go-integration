const hre = require("hardhat");

async function main() {
  const [deployer] = await hre.ethers.getSigners();

  const tokenAddress = "0xa33239e13303Fe9586C25b70ABd4D5d65E7B368f";
  console.log("Using token at:", tokenAddress);

  const Game = await hre.ethers.getContractFactory("Game");
  const game = await Game.deploy(tokenAddress);

  await game.deployed();
  console.log("Game deployed to:", game.address);
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
