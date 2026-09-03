// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface ISweepDelegate {
    function EXECUTOR() external view returns (address);

    function SWEEP_SUCCESS() external view returns (bytes4);

    function sweepERC20(
        address token,
        address recipient,
        uint256 amount
    ) external returns (bytes4 magic);

    function sweepNative(
        address payable recipient,
        uint256 amount
    ) external returns (bytes4 magic);
}

/// @title BatchSweepExecutor
/// @notice Pays for and executes batched native/ERC-20 sweeps from EIP-7702 deposit addresses.
/// @dev The transaction sender must be OPERATOR. Individual item failures do not revert
///      the whole batch; off-chain task state must be reconciled from CollectResult events.
contract BatchSweepExecutor {
    error UnauthorizedOwner(address caller);
    error UnauthorizedOperator(address caller);
    error ZeroAddress();
    error DelegateAlreadyInitialized();
    error DelegateNotInitialized();
    error InvalidDelegate(address delegate);
    error ContractPaused();
    error EmptyBatch();
    error BatchTooLarge(uint256 supplied, uint256 maximum);
    error ReentrantCall();
    error NotPendingOwner(address caller);

    enum ResultCode {
        Success,
        InvalidAccount,
        InvalidToken,
        RecipientNotAllowed,
        WrongDelegation,
        DelegateCallFailed,
        InvalidReturnValue,
        InsufficientExecutionGas
    }

    struct CollectItem {
        uint256 taskId;
        address account;
        /// @dev address(0) represents the chain's native currency (ETH/BNB).
        address token;
        address recipient;
        uint256 amount;
        uint256 callGasLimit;
    }

    uint256 public constant MAX_BATCH_ITEMS = 100;
    uint256 public constant MIN_ITEM_GAS = 40_000;
    uint256 public constant MAX_ITEM_GAS = 500_000;
    uint256 private constant POST_CALL_GAS_RESERVE = 30_000;
    bytes3 private constant DELEGATION_PREFIX = 0xef0100;

    address public owner;
    address public pendingOwner;
    address public operator;
    address public sweepDelegate;
    bool public paused;

    mapping(address recipient => bool allowed) public allowedRecipient;

    uint256 private _unlocked = 1;

    event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner);
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);
    event OperatorUpdated(address indexed previousOperator, address indexed newOperator);
    event RecipientPermissionUpdated(address indexed recipient, bool allowed);
    event SweepDelegateInitialized(address indexed delegate);
    event PauseUpdated(bool paused);
    event CollectResult(
        uint256 indexed taskId,
        address indexed account,
        address indexed token,
        address recipient,
        uint256 requestedAmount,
        ResultCode result,
        bytes32 returnDataHash
    );
    event BatchStopped(uint256 indexed nextItemIndex, uint256 gasRemaining);

    constructor(address initialOwner, address initialOperator) {
        if (initialOwner == address(0) || initialOperator == address(0)) revert ZeroAddress();
        owner = initialOwner;
        operator = initialOperator;
        emit OwnershipTransferred(address(0), initialOwner);
        emit OperatorUpdated(address(0), initialOperator);
    }

    modifier onlyOwner() {
        if (msg.sender != owner) revert UnauthorizedOwner(msg.sender);
        _;
    }

    modifier onlyOperator() {
        if (msg.sender != operator) revert UnauthorizedOperator(msg.sender);
        _;
    }

    modifier nonReentrant() {
        if (_unlocked != 1) revert ReentrantCall();
        _unlocked = 2;
        _;
        _unlocked = 1;
    }

    /// @notice Permanently pairs this executor with a freshly deployed SweepDelegate.
    /// @dev Two-step deployment avoids a circular constructor dependency:
    ///      1. deploy this executor; 2. deploy SweepDelegate(address(this));
    ///      3. call initializeSweepDelegate(delegate).
    function initializeSweepDelegate(address delegate) external onlyOwner {
        if (sweepDelegate != address(0)) revert DelegateAlreadyInitialized();
        if (delegate == address(0) || delegate.code.length == 0) {
            revert InvalidDelegate(delegate);
        }

        try ISweepDelegate(delegate).EXECUTOR() returns (address configuredExecutor) {
            if (configuredExecutor != address(this)) revert InvalidDelegate(delegate);
        } catch {
            revert InvalidDelegate(delegate);
        }

        sweepDelegate = delegate;
        emit SweepDelegateInitialized(delegate);
    }

    function setOperator(address newOperator) external onlyOwner {
        if (newOperator == address(0)) revert ZeroAddress();
        address previousOperator = operator;
        operator = newOperator;
        emit OperatorUpdated(previousOperator, newOperator);
    }

    function setRecipientAllowed(address recipient, bool allowed) external onlyOwner {
        if (recipient == address(0)) revert ZeroAddress();
        allowedRecipient[recipient] = allowed;
        emit RecipientPermissionUpdated(recipient, allowed);
    }

    function setPaused(bool newPaused) external onlyOwner {
        paused = newPaused;
        emit PauseUpdated(newPaused);
    }

    function transferOwnership(address newOwner) external onlyOwner {
        if (newOwner == address(0)) revert ZeroAddress();
        pendingOwner = newOwner;
        emit OwnershipTransferStarted(owner, newOwner);
    }

    function acceptOwnership() external {
        if (msg.sender != pendingOwner) revert NotPendingOwner(msg.sender);
        address previousOwner = owner;
        owner = msg.sender;
        pendingOwner = address(0);
        emit OwnershipTransferred(previousOwner, msg.sender);
    }

    /// @notice Sweep a batch of ERC-20 or native balances from delegated deposit addresses.
    /// @dev For a first-time address, its EIP-7702 authorization must be included in the
    ///      outer type-4 transaction. Authorization processing occurs before this call,
    ///      so _isDelegatedToExpectedImplementation succeeds during execution.
    function collect(CollectItem[] calldata items) external onlyOperator nonReentrant {
        if (paused) revert ContractPaused();
        address delegate = sweepDelegate;
        if (delegate == address(0)) revert DelegateNotInitialized();

        uint256 length = items.length;
        if (length == 0) revert EmptyBatch();
        if (length > MAX_BATCH_ITEMS) revert BatchTooLarge(length, MAX_BATCH_ITEMS);

        bytes4 expectedMagic = ISweepDelegate(delegate).SWEEP_SUCCESS();

        for (uint256 i; i < length; ) {
            CollectItem calldata item = items[i];
            (ResultCode result, bytes32 returnDataHash) = _collectOne(
                item,
                delegate,
                expectedMagic
            );

            emit CollectResult(
                item.taskId,
                item.account,
                item.token,
                item.recipient,
                item.amount,
                result,
                returnDataHash
            );

            // Avoid iterating over all remaining items after the transaction no longer has
            // enough gas for the requested per-item cap. Items without a CollectResult were
            // not attempted and can safely remain queued off-chain.
            if (result == ResultCode.InsufficientExecutionGas) {
                emit BatchStopped(i + 1, gasleft());
                break;
            }

            unchecked {
                ++i;
            }
        }
    }

    function _collectOne(
        CollectItem calldata item,
        address delegate,
        bytes4 expectedMagic
    ) private returns (ResultCode result, bytes32 returnDataHash) {
        if (item.account == address(0)) return (ResultCode.InvalidAccount, bytes32(0));
        if (item.token != address(0) && item.token.code.length == 0) {
            return (ResultCode.InvalidToken, bytes32(0));
        }
        if (!allowedRecipient[item.recipient]) {
            return (ResultCode.RecipientNotAllowed, bytes32(0));
        }
        if (!_isDelegatedTo(item.account, delegate)) {
            return (ResultCode.WrongDelegation, bytes32(0));
        }

        uint256 callGas = item.callGasLimit;
        if (callGas < MIN_ITEM_GAS) callGas = MIN_ITEM_GAS;
        if (callGas > MAX_ITEM_GAS) callGas = MAX_ITEM_GAS;
        if (gasleft() <= callGas + POST_CALL_GAS_RESERVE) {
            return (ResultCode.InsufficientExecutionGas, bytes32(0));
        }

        bytes memory callData;
        if (item.token == address(0)) {
            callData = abi.encodeCall(
                ISweepDelegate.sweepNative,
                (payable(item.recipient), item.amount)
            );
        } else {
            callData = abi.encodeCall(
                ISweepDelegate.sweepERC20,
                (item.token, item.recipient, item.amount)
            );
        }

        (bool success, bytes memory returnData) = item.account.call{gas: callGas}(callData);
        returnDataHash = keccak256(returnData);

        if (!success) return (ResultCode.DelegateCallFailed, returnDataHash);
        if (returnData.length < 32 || abi.decode(returnData, (bytes4)) != expectedMagic) {
            return (ResultCode.InvalidReturnValue, returnDataHash);
        }
        return (ResultCode.Success, returnDataHash);
    }

    /// @notice Returns whether account currently contains the exact EIP-7702 delegation
    ///         indicator `0xef0100 || expectedDelegate`.
    function isDelegatedAccount(address account) external view returns (bool) {
        address delegate = sweepDelegate;
        return delegate != address(0) && _isDelegatedTo(account, delegate);
    }

    function _isDelegatedTo(address account, address expectedDelegate) private view returns (bool) {
        bytes memory code = account.code;
        if (code.length != 23) return false;
        return keccak256(code) == keccak256(abi.encodePacked(DELEGATION_PREFIX, expectedDelegate));
    }
}
