pragma solidity ^0.4.18;

import "./StandardToken.sol";
import "./Ownable.sol";
import "./SafeMath.sol";

/**
 * @title Mintable token
 * @dev ERC20 Token, with mintable token creation
 */

contract MintableToken is StandardToken, Ownable {

  using SafeMath for uint256;

  event Mint(address indexed to, uint256 amount);
  event MintFinished();

  bool public mintingFinished = false;

  /**
   * @dev The MintableToken constructor.
   */
  function MintableToken() public {
    mint(msg.sender, 38000000 * 1E8);
    finishMinting();
  }

  modifier canMint() {
    require(!mintingFinished);
    _;
  }

  /**
   * @dev Function to stop minting new tokens.
   * @return True if the operation was successful.
   */
  function finishMinting() public onlyOwner canMint returns (bool) {
    mintingFinished = true;
    MintFinished();
    return true;
  }

  /**
   * @dev Function to mint tokens
   * @param _to The address that will receive the minted tokens.
   * @param _amount The amount of tokens to mint.
   * @return A boolean that indicates if the operation was successful.
   */
  function mint(address _to, uint256 _amount) onlyOwner canMint  public returns (bool) {

    totalSupply = totalSupply.add(_amount);
    balances[_to] = balances[_to].add(_amount);
    Mint(_to, _amount);
    Transfer(address(0), _to, _amount);
    return true;

  }
}
