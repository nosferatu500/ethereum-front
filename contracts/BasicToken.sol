pragma solidity ^0.4.18;

import "./ERC20Basic.sol";
import "./SafeMath.sol";
import "./Privileged.sol";

/**
 * @title Basic token1
 * @dev Basic version of StandardToken, with no allowances.
 * @dev added privileged functions
 * @dev added tokenStorage
 */
contract BasicToken is ERC20Basic {
  using SafeMath for uint256;

  Privileged public privileged;

  event Burn(address indexed burner, uint256 value);

  mapping(address => uint256) balances;

  address public tokenStorage;

  /**
  * @dev transfer token for a specified address
  * @param _to The address to transfer to.
  * @param _value The amount to be transferred.
  */
  function transfer(address _to, uint256 _value) public returns (bool) {
    require(_to != address(0));
    require(_value <= balances[msg.sender]);

    balances[msg.sender] = balances[msg.sender].sub(_value);
    balances[_to] = balances[_to].add(_value);
    Transfer(msg.sender, _to, _value);
    return true;
  }

  /**
  * @dev Gets the balance of the specified address.
  * @param _owner The address to query the the balance of.
  * @return An uint256 representing the amount owned by the passed address.
  */
  function balanceOf(address _owner) public view returns (uint256 balance) {
    return balances[_owner];
  }

  /**
  * @dev privileged token transfer from tokenStorage to a specified address
  * @param _to The address to transfer to.
  * @param _value The amount to be transferred.
  */
  function privilegedTransfer(address _to, uint256 _value) public returns (bool) {
    require(privileged.isPrivileged(msg.sender));
    require(_to != address(0));
    require(_value <= balances[tokenStorage]);
    require(_value > 0);

    balances[tokenStorage] = balances[tokenStorage].sub(_value);
    balances[_to] = balances[_to].add(_value);
    Transfer(tokenStorage, _to, _value);

    return true;
  }

  /**
  * @dev Burns a specific amount of tokens from a preferred address.
  * @param _value The amount of token to be burned.
  */
  function privilegedBurn(uint256 _value) public returns (bool) {
    require(privileged.isPrivileged(msg.sender));
    require(_value > 0);
    require(_value <= balances[msg.sender]);

    address burner = msg.sender;
    balances[burner] = balances[burner].sub(_value);
    totalSupply = totalSupply.sub(_value);
    Burn(burner, _value);
    return true;
  }
}