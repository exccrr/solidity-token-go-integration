// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./Token.sol";

contract Game {
    Token public token;
    address public owner;

    event BetPlaced(address indexed player, uint256 amount, uint8 guess, uint8 winning);
    event Win(address indexed player, uint256 prize);
    event Loss(address indexed player);

    constructor(address _token) {
        token = Token(_token);
        owner = msg.sender;
    }

    function play(uint8 guess) external {
        require(guess >= 1 && guess <= 10, "Guess out of range");

        uint256 betAmount = 10 * 10 ** token.decimals();
        require(token.transferFrom(msg.sender, address(this), betAmount), "Payment failed");

        uint8 winning = uint8(uint256(blockhash(block.number - 1)) % 10 + 1);

        emit BetPlaced(msg.sender, betAmount, guess, winning);

        if (guess == winning) {
            uint256 prize = 20 * 10 ** token.decimals();
            require(token.transfer(msg.sender, prize), "Prize transfer failed");
            emit Win(msg.sender, prize);
        } else {
            emit Loss(msg.sender);
        }
    }

    function withdraw() external {
        require(msg.sender == owner, "Only owner can withdraw");
        uint256 balance = token.balanceOf(address(this));
        token.transfer(owner, balance);
    }
}
