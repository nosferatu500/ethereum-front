pragma solidity ^0.4.18;

import "./MintableToken.sol";
import "./Privileged.sol";

/**
 * @title SingleTokenCoin
 */
contract SingleTokenCoin is MintableToken {
    
    string public constant name = "Example Token";
    
    string public constant symbol = "EXP";
    
    uint32 public constant decimals = 8;

    /**
     * @dev Constructor set Privileged contract address and set tokenStorage.
     */
    function SingleTokenCoin(address privilegedContractAddress) public {
        privileged = Privileged(privilegedContractAddress);
        tokenStorage = msg.sender;
    }
    
}