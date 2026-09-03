// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title SweepDelegate
/// @notice Minimal EIP-7702 delegate used by deposit addresses to sweep native currency
///         and ERC-20 tokens.
/// @dev This contract is intentionally not upgradeable and exposes no arbitrary-call primitive.
///      When its code is executed through an EIP-7702 delegation, address(this) is the
///      authorizing deposit address, so the ERC-20 transfer is made by that address.
contract SweepDelegate {
    error UnauthorizedCaller(address caller);
    error ZeroAddress();
    error InvalidToken(address token);
    error TokenTransferFailed(address token);
    error NativeTransferFailed(address recipient, uint256 amount);

    bytes4 public constant SWEEP_SUCCESS = bytes4(keccak256("SWEEP_SUCCESS"));

    /// @notice The only contract allowed to invoke sweep functions.
    address public immutable EXECUTOR;

    constructor(address executor) {
        if (executor == address(0)) revert ZeroAddress();
        EXECUTOR = executor;
    }

    modifier onlyExecutor() {
        if (msg.sender != EXECUTOR) revert UnauthorizedCaller(msg.sender);
        _;
    }

    /// @notice Transfer ERC-20 tokens from the authorizing EOA to a collection address.
    /// @param token ERC-20 token contract.
    /// @param recipient Approved collection address, checked by BatchSweepExecutor.
    /// @param amount Amount in the token's smallest unit.
    /// @return magic A fixed value that lets the executor distinguish real execution from
    ///               a successful empty call to an address without delegated code.
    function sweepERC20(
        address token,
        address recipient,
        uint256 amount
    ) external onlyExecutor returns (bytes4 magic) {
        if (token == address(0) || token.code.length == 0) revert InvalidToken(token);
        if (recipient == address(0)) revert ZeroAddress();

        _safeTransfer(token, recipient, amount);
        return SWEEP_SUCCESS;
    }

    /// @notice Transfer native currency (ETH/BNB) from the authorizing EOA.
    /// @dev Gas is paid by the outer sponsor transaction, so `amount` may equal the EOA's
    ///      entire native balance. The recipient is approved by BatchSweepExecutor.
    /// @param recipient Approved collection address.
    /// @param amount Amount in wei.
    /// @return magic The same fixed success marker used by ERC-20 sweeps.
    function sweepNative(
        address payable recipient,
        uint256 amount
    ) external onlyExecutor returns (bytes4 magic) {
        if (recipient == address(0)) revert ZeroAddress();

        bool success;
        assembly ("memory-safe") {
            // Do not copy return data. A whitelisted contract recipient cannot force an
            // unbounded memory copy even if it returns a very large payload.
            success := call(gas(), recipient, amount, 0, 0, 0, 0)
        }
        if (!success) revert NativeTransferFailed(recipient, amount);
        return SWEEP_SUCCESS;
    }

    /// @dev Supports ERC-20 implementations that return true or no return data.
    ///      The assembly call copies at most 32 bytes, preventing a malicious token from
    ///      forcing this contract to copy an unbounded return/revert payload.
    function _safeTransfer(address token, address recipient, uint256 amount) private {
        bool success;
        bool returnValue;

        assembly ("memory-safe") {
            let ptr := mload(0x40)
            mstore(ptr, shl(224, 0xa9059cbb)) // transfer(address,uint256)
            mstore(add(ptr, 0x04), recipient)
            mstore(add(ptr, 0x24), amount)

            success := call(gas(), token, 0, ptr, 0x44, ptr, 0x20)

            switch returndatasize()
            case 0 {
                returnValue := 1
            }
            default {
                // A standard bool return value occupies exactly one ABI word. Requiring at
                // least 32 bytes rejects malformed short responses.
                if iszero(lt(returndatasize(), 0x20)) {
                    returnValue := iszero(iszero(mload(ptr)))
                }
            }
        }

        if (!success || !returnValue) revert TokenTransferFailed(token);
    }
}
