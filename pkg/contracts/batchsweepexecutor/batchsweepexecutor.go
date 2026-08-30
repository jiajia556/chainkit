// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package batchsweepexecutor

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// BatchSweepExecutorCollectItem is an auto generated low-level Go binding around an user-defined struct.
type BatchSweepExecutorCollectItem struct {
	TaskId       *big.Int
	Account      common.Address
	Token        common.Address
	Recipient    common.Address
	Amount       *big.Int
	CallGasLimit *big.Int
}

// BatchSweepExecutorMetaData contains all meta data concerning the BatchSweepExecutor contract.
var BatchSweepExecutorMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"initialOwner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"initialOperator\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"supplied\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maximum\",\"type\":\"uint256\"}],\"name\":\"BatchTooLarge\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ContractPaused\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateAlreadyInitialized\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateNotInitialized\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"EmptyBatch\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegate\",\"type\":\"address\"}],\"name\":\"InvalidDelegate\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"name\":\"NotPendingOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ReentrantCall\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"name\":\"UnauthorizedOperator\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"name\":\"UnauthorizedOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroAddress\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nextItemIndex\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"gasRemaining\",\"type\":\"uint256\"}],\"name\":\"BatchStopped\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"taskId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"requestedAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"enumBatchSweepExecutor.ResultCode\",\"name\":\"result\",\"type\":\"uint8\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"returnDataHash\",\"type\":\"bytes32\"}],\"name\":\"CollectResult\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOperator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOperator\",\"type\":\"address\"}],\"name\":\"OperatorUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferStarted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"paused\",\"type\":\"bool\"}],\"name\":\"PauseUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"allowed\",\"type\":\"bool\"}],\"name\":\"RecipientPermissionUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegate\",\"type\":\"address\"}],\"name\":\"SweepDelegateInitialized\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"MAX_BATCH_ITEMS\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MAX_ITEM_GAS\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MIN_ITEM_GAS\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"acceptOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"}],\"name\":\"allowedRecipient\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"allowed\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"taskId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"callGasLimit\",\"type\":\"uint256\"}],\"internalType\":\"structBatchSweepExecutor.CollectItem[]\",\"name\":\"items\",\"type\":\"tuple[]\"}],\"name\":\"collect\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegate\",\"type\":\"address\"}],\"name\":\"initializeSweepDelegate\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"isDelegatedAccount\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"operator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"paused\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pendingOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOperator\",\"type\":\"address\"}],\"name\":\"setOperator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"newPaused\",\"type\":\"bool\"}],\"name\":\"setPaused\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"allowed\",\"type\":\"bool\"}],\"name\":\"setRecipientAllowed\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"sweepDelegate\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// BatchSweepExecutorABI is the input ABI used to generate the binding from.
// Deprecated: Use BatchSweepExecutorMetaData.ABI instead.
var BatchSweepExecutorABI = BatchSweepExecutorMetaData.ABI

// BatchSweepExecutor is an auto generated Go binding around an Ethereum contract.
type BatchSweepExecutor struct {
	BatchSweepExecutorCaller     // Read-only binding to the contract
	BatchSweepExecutorTransactor // Write-only binding to the contract
	BatchSweepExecutorFilterer   // Log filterer for contract events
}

// BatchSweepExecutorCaller is an auto generated read-only Go binding around an Ethereum contract.
type BatchSweepExecutorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BatchSweepExecutorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BatchSweepExecutorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BatchSweepExecutorFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BatchSweepExecutorFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BatchSweepExecutorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BatchSweepExecutorSession struct {
	Contract     *BatchSweepExecutor // Generic contract binding to set the session for
	CallOpts     bind.CallOpts       // Call options to use throughout this session
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// BatchSweepExecutorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BatchSweepExecutorCallerSession struct {
	Contract *BatchSweepExecutorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts             // Call options to use throughout this session
}

// BatchSweepExecutorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BatchSweepExecutorTransactorSession struct {
	Contract     *BatchSweepExecutorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// BatchSweepExecutorRaw is an auto generated low-level Go binding around an Ethereum contract.
type BatchSweepExecutorRaw struct {
	Contract *BatchSweepExecutor // Generic contract binding to access the raw methods on
}

// BatchSweepExecutorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BatchSweepExecutorCallerRaw struct {
	Contract *BatchSweepExecutorCaller // Generic read-only contract binding to access the raw methods on
}

// BatchSweepExecutorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BatchSweepExecutorTransactorRaw struct {
	Contract *BatchSweepExecutorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBatchSweepExecutor creates a new instance of BatchSweepExecutor, bound to a specific deployed contract.
func NewBatchSweepExecutor(address common.Address, backend bind.ContractBackend) (*BatchSweepExecutor, error) {
	contract, err := bindBatchSweepExecutor(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BatchSweepExecutor{BatchSweepExecutorCaller: BatchSweepExecutorCaller{contract: contract}, BatchSweepExecutorTransactor: BatchSweepExecutorTransactor{contract: contract}, BatchSweepExecutorFilterer: BatchSweepExecutorFilterer{contract: contract}}, nil
}

// NewBatchSweepExecutorCaller creates a new read-only instance of BatchSweepExecutor, bound to a specific deployed contract.
func NewBatchSweepExecutorCaller(address common.Address, caller bind.ContractCaller) (*BatchSweepExecutorCaller, error) {
	contract, err := bindBatchSweepExecutor(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BatchSweepExecutorCaller{contract: contract}, nil
}

// NewBatchSweepExecutorTransactor creates a new write-only instance of BatchSweepExecutor, bound to a specific deployed contract.
func NewBatchSweepExecutorTransactor(address common.Address, transactor bind.ContractTransactor) (*BatchSweepExecutorTransactor, error) {
	contract, err := bindBatchSweepExecutor(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BatchSweepExecutorTransactor{contract: contract}, nil
}

// NewBatchSweepExecutorFilterer creates a new log filterer instance of BatchSweepExecutor, bound to a specific deployed contract.
func NewBatchSweepExecutorFilterer(address common.Address, filterer bind.ContractFilterer) (*BatchSweepExecutorFilterer, error) {
	contract, err := bindBatchSweepExecutor(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BatchSweepExecutorFilterer{contract: contract}, nil
}

// bindBatchSweepExecutor binds a generic wrapper to an already deployed contract.
func bindBatchSweepExecutor(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BatchSweepExecutorMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BatchSweepExecutor *BatchSweepExecutorRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BatchSweepExecutor.Contract.BatchSweepExecutorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BatchSweepExecutor *BatchSweepExecutorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BatchSweepExecutor.Contract.BatchSweepExecutorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BatchSweepExecutor *BatchSweepExecutorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BatchSweepExecutor.Contract.BatchSweepExecutorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BatchSweepExecutor *BatchSweepExecutorCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BatchSweepExecutor.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BatchSweepExecutor *BatchSweepExecutorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BatchSweepExecutor.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BatchSweepExecutor *BatchSweepExecutorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BatchSweepExecutor.Contract.contract.Transact(opts, method, params...)
}

// MAXBATCHITEMS is a free data retrieval call binding the contract method 0x2b95390f.
//
// Solidity: function MAX_BATCH_ITEMS() view returns(uint256)
func (_BatchSweepExecutor *BatchSweepExecutorCaller) MAXBATCHITEMS(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BatchSweepExecutor.contract.Call(opts, &out, "MAX_BATCH_ITEMS")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXBATCHITEMS is a free data retrieval call binding the contract method 0x2b95390f.
//
// Solidity: function MAX_BATCH_ITEMS() view returns(uint256)
func (_BatchSweepExecutor *BatchSweepExecutorSession) MAXBATCHITEMS() (*big.Int, error) {
	return _BatchSweepExecutor.Contract.MAXBATCHITEMS(&_BatchSweepExecutor.CallOpts)
}

// MAXBATCHITEMS is a free data retrieval call binding the contract method 0x2b95390f.
//
// Solidity: function MAX_BATCH_ITEMS() view returns(uint256)
func (_BatchSweepExecutor *BatchSweepExecutorCallerSession) MAXBATCHITEMS() (*big.Int, error) {
	return _BatchSweepExecutor.Contract.MAXBATCHITEMS(&_BatchSweepExecutor.CallOpts)
}

// MAXITEMGAS is a free data retrieval call binding the contract method 0x6d7242f5.
//
// Solidity: function MAX_ITEM_GAS() view returns(uint256)
func (_BatchSweepExecutor *BatchSweepExecutorCaller) MAXITEMGAS(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BatchSweepExecutor.contract.Call(opts, &out, "MAX_ITEM_GAS")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXITEMGAS is a free data retrieval call binding the contract method 0x6d7242f5.
//
// Solidity: function MAX_ITEM_GAS() view returns(uint256)
func (_BatchSweepExecutor *BatchSweepExecutorSession) MAXITEMGAS() (*big.Int, error) {
	return _BatchSweepExecutor.Contract.MAXITEMGAS(&_BatchSweepExecutor.CallOpts)
}

// MAXITEMGAS is a free data retrieval call binding the contract method 0x6d7242f5.
//
// Solidity: function MAX_ITEM_GAS() view returns(uint256)
func (_BatchSweepExecutor *BatchSweepExecutorCallerSession) MAXITEMGAS() (*big.Int, error) {
	return _BatchSweepExecutor.Contract.MAXITEMGAS(&_BatchSweepExecutor.CallOpts)
}

// MINITEMGAS is a free data retrieval call binding the contract method 0x8340ead1.
//
// Solidity: function MIN_ITEM_GAS() view returns(uint256)
func (_BatchSweepExecutor *BatchSweepExecutorCaller) MINITEMGAS(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BatchSweepExecutor.contract.Call(opts, &out, "MIN_ITEM_GAS")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MINITEMGAS is a free data retrieval call binding the contract method 0x8340ead1.
//
// Solidity: function MIN_ITEM_GAS() view returns(uint256)
func (_BatchSweepExecutor *BatchSweepExecutorSession) MINITEMGAS() (*big.Int, error) {
	return _BatchSweepExecutor.Contract.MINITEMGAS(&_BatchSweepExecutor.CallOpts)
}

// MINITEMGAS is a free data retrieval call binding the contract method 0x8340ead1.
//
// Solidity: function MIN_ITEM_GAS() view returns(uint256)
func (_BatchSweepExecutor *BatchSweepExecutorCallerSession) MINITEMGAS() (*big.Int, error) {
	return _BatchSweepExecutor.Contract.MINITEMGAS(&_BatchSweepExecutor.CallOpts)
}

// AllowedRecipient is a free data retrieval call binding the contract method 0xdb846026.
//
// Solidity: function allowedRecipient(address recipient) view returns(bool allowed)
func (_BatchSweepExecutor *BatchSweepExecutorCaller) AllowedRecipient(opts *bind.CallOpts, recipient common.Address) (bool, error) {
	var out []interface{}
	err := _BatchSweepExecutor.contract.Call(opts, &out, "allowedRecipient", recipient)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// AllowedRecipient is a free data retrieval call binding the contract method 0xdb846026.
//
// Solidity: function allowedRecipient(address recipient) view returns(bool allowed)
func (_BatchSweepExecutor *BatchSweepExecutorSession) AllowedRecipient(recipient common.Address) (bool, error) {
	return _BatchSweepExecutor.Contract.AllowedRecipient(&_BatchSweepExecutor.CallOpts, recipient)
}

// AllowedRecipient is a free data retrieval call binding the contract method 0xdb846026.
//
// Solidity: function allowedRecipient(address recipient) view returns(bool allowed)
func (_BatchSweepExecutor *BatchSweepExecutorCallerSession) AllowedRecipient(recipient common.Address) (bool, error) {
	return _BatchSweepExecutor.Contract.AllowedRecipient(&_BatchSweepExecutor.CallOpts, recipient)
}

// IsDelegatedAccount is a free data retrieval call binding the contract method 0xb80a852c.
//
// Solidity: function isDelegatedAccount(address account) view returns(bool)
func (_BatchSweepExecutor *BatchSweepExecutorCaller) IsDelegatedAccount(opts *bind.CallOpts, account common.Address) (bool, error) {
	var out []interface{}
	err := _BatchSweepExecutor.contract.Call(opts, &out, "isDelegatedAccount", account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsDelegatedAccount is a free data retrieval call binding the contract method 0xb80a852c.
//
// Solidity: function isDelegatedAccount(address account) view returns(bool)
func (_BatchSweepExecutor *BatchSweepExecutorSession) IsDelegatedAccount(account common.Address) (bool, error) {
	return _BatchSweepExecutor.Contract.IsDelegatedAccount(&_BatchSweepExecutor.CallOpts, account)
}

// IsDelegatedAccount is a free data retrieval call binding the contract method 0xb80a852c.
//
// Solidity: function isDelegatedAccount(address account) view returns(bool)
func (_BatchSweepExecutor *BatchSweepExecutorCallerSession) IsDelegatedAccount(account common.Address) (bool, error) {
	return _BatchSweepExecutor.Contract.IsDelegatedAccount(&_BatchSweepExecutor.CallOpts, account)
}

// Operator is a free data retrieval call binding the contract method 0x570ca735.
//
// Solidity: function operator() view returns(address)
func (_BatchSweepExecutor *BatchSweepExecutorCaller) Operator(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BatchSweepExecutor.contract.Call(opts, &out, "operator")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Operator is a free data retrieval call binding the contract method 0x570ca735.
//
// Solidity: function operator() view returns(address)
func (_BatchSweepExecutor *BatchSweepExecutorSession) Operator() (common.Address, error) {
	return _BatchSweepExecutor.Contract.Operator(&_BatchSweepExecutor.CallOpts)
}

// Operator is a free data retrieval call binding the contract method 0x570ca735.
//
// Solidity: function operator() view returns(address)
func (_BatchSweepExecutor *BatchSweepExecutorCallerSession) Operator() (common.Address, error) {
	return _BatchSweepExecutor.Contract.Operator(&_BatchSweepExecutor.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BatchSweepExecutor *BatchSweepExecutorCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BatchSweepExecutor.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BatchSweepExecutor *BatchSweepExecutorSession) Owner() (common.Address, error) {
	return _BatchSweepExecutor.Contract.Owner(&_BatchSweepExecutor.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BatchSweepExecutor *BatchSweepExecutorCallerSession) Owner() (common.Address, error) {
	return _BatchSweepExecutor.Contract.Owner(&_BatchSweepExecutor.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_BatchSweepExecutor *BatchSweepExecutorCaller) Paused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _BatchSweepExecutor.contract.Call(opts, &out, "paused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_BatchSweepExecutor *BatchSweepExecutorSession) Paused() (bool, error) {
	return _BatchSweepExecutor.Contract.Paused(&_BatchSweepExecutor.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_BatchSweepExecutor *BatchSweepExecutorCallerSession) Paused() (bool, error) {
	return _BatchSweepExecutor.Contract.Paused(&_BatchSweepExecutor.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BatchSweepExecutor *BatchSweepExecutorCaller) PendingOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BatchSweepExecutor.contract.Call(opts, &out, "pendingOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BatchSweepExecutor *BatchSweepExecutorSession) PendingOwner() (common.Address, error) {
	return _BatchSweepExecutor.Contract.PendingOwner(&_BatchSweepExecutor.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BatchSweepExecutor *BatchSweepExecutorCallerSession) PendingOwner() (common.Address, error) {
	return _BatchSweepExecutor.Contract.PendingOwner(&_BatchSweepExecutor.CallOpts)
}

// SweepDelegate is a free data retrieval call binding the contract method 0x2b5e2e8d.
//
// Solidity: function sweepDelegate() view returns(address)
func (_BatchSweepExecutor *BatchSweepExecutorCaller) SweepDelegate(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BatchSweepExecutor.contract.Call(opts, &out, "sweepDelegate")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// SweepDelegate is a free data retrieval call binding the contract method 0x2b5e2e8d.
//
// Solidity: function sweepDelegate() view returns(address)
func (_BatchSweepExecutor *BatchSweepExecutorSession) SweepDelegate() (common.Address, error) {
	return _BatchSweepExecutor.Contract.SweepDelegate(&_BatchSweepExecutor.CallOpts)
}

// SweepDelegate is a free data retrieval call binding the contract method 0x2b5e2e8d.
//
// Solidity: function sweepDelegate() view returns(address)
func (_BatchSweepExecutor *BatchSweepExecutorCallerSession) SweepDelegate() (common.Address, error) {
	return _BatchSweepExecutor.Contract.SweepDelegate(&_BatchSweepExecutor.CallOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BatchSweepExecutor *BatchSweepExecutorTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BatchSweepExecutor.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BatchSweepExecutor *BatchSweepExecutorSession) AcceptOwnership() (*types.Transaction, error) {
	return _BatchSweepExecutor.Contract.AcceptOwnership(&_BatchSweepExecutor.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BatchSweepExecutor *BatchSweepExecutorTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _BatchSweepExecutor.Contract.AcceptOwnership(&_BatchSweepExecutor.TransactOpts)
}

// Collect is a paid mutator transaction binding the contract method 0x46e4f1a5.
//
// Solidity: function collect((uint256,address,address,address,uint256,uint256)[] items) returns()
func (_BatchSweepExecutor *BatchSweepExecutorTransactor) Collect(opts *bind.TransactOpts, items []BatchSweepExecutorCollectItem) (*types.Transaction, error) {
	return _BatchSweepExecutor.contract.Transact(opts, "collect", items)
}

// Collect is a paid mutator transaction binding the contract method 0x46e4f1a5.
//
// Solidity: function collect((uint256,address,address,address,uint256,uint256)[] items) returns()
func (_BatchSweepExecutor *BatchSweepExecutorSession) Collect(items []BatchSweepExecutorCollectItem) (*types.Transaction, error) {
	return _BatchSweepExecutor.Contract.Collect(&_BatchSweepExecutor.TransactOpts, items)
}

// Collect is a paid mutator transaction binding the contract method 0x46e4f1a5.
//
// Solidity: function collect((uint256,address,address,address,uint256,uint256)[] items) returns()
func (_BatchSweepExecutor *BatchSweepExecutorTransactorSession) Collect(items []BatchSweepExecutorCollectItem) (*types.Transaction, error) {
	return _BatchSweepExecutor.Contract.Collect(&_BatchSweepExecutor.TransactOpts, items)
}

// InitializeSweepDelegate is a paid mutator transaction binding the contract method 0x4666748b.
//
// Solidity: function initializeSweepDelegate(address delegate) returns()
func (_BatchSweepExecutor *BatchSweepExecutorTransactor) InitializeSweepDelegate(opts *bind.TransactOpts, delegate common.Address) (*types.Transaction, error) {
	return _BatchSweepExecutor.contract.Transact(opts, "initializeSweepDelegate", delegate)
}

// InitializeSweepDelegate is a paid mutator transaction binding the contract method 0x4666748b.
//
// Solidity: function initializeSweepDelegate(address delegate) returns()
func (_BatchSweepExecutor *BatchSweepExecutorSession) InitializeSweepDelegate(delegate common.Address) (*types.Transaction, error) {
	return _BatchSweepExecutor.Contract.InitializeSweepDelegate(&_BatchSweepExecutor.TransactOpts, delegate)
}

// InitializeSweepDelegate is a paid mutator transaction binding the contract method 0x4666748b.
//
// Solidity: function initializeSweepDelegate(address delegate) returns()
func (_BatchSweepExecutor *BatchSweepExecutorTransactorSession) InitializeSweepDelegate(delegate common.Address) (*types.Transaction, error) {
	return _BatchSweepExecutor.Contract.InitializeSweepDelegate(&_BatchSweepExecutor.TransactOpts, delegate)
}

// SetOperator is a paid mutator transaction binding the contract method 0xb3ab15fb.
//
// Solidity: function setOperator(address newOperator) returns()
func (_BatchSweepExecutor *BatchSweepExecutorTransactor) SetOperator(opts *bind.TransactOpts, newOperator common.Address) (*types.Transaction, error) {
	return _BatchSweepExecutor.contract.Transact(opts, "setOperator", newOperator)
}

// SetOperator is a paid mutator transaction binding the contract method 0xb3ab15fb.
//
// Solidity: function setOperator(address newOperator) returns()
func (_BatchSweepExecutor *BatchSweepExecutorSession) SetOperator(newOperator common.Address) (*types.Transaction, error) {
	return _BatchSweepExecutor.Contract.SetOperator(&_BatchSweepExecutor.TransactOpts, newOperator)
}

// SetOperator is a paid mutator transaction binding the contract method 0xb3ab15fb.
//
// Solidity: function setOperator(address newOperator) returns()
func (_BatchSweepExecutor *BatchSweepExecutorTransactorSession) SetOperator(newOperator common.Address) (*types.Transaction, error) {
	return _BatchSweepExecutor.Contract.SetOperator(&_BatchSweepExecutor.TransactOpts, newOperator)
}

// SetPaused is a paid mutator transaction binding the contract method 0x16c38b3c.
//
// Solidity: function setPaused(bool newPaused) returns()
func (_BatchSweepExecutor *BatchSweepExecutorTransactor) SetPaused(opts *bind.TransactOpts, newPaused bool) (*types.Transaction, error) {
	return _BatchSweepExecutor.contract.Transact(opts, "setPaused", newPaused)
}

// SetPaused is a paid mutator transaction binding the contract method 0x16c38b3c.
//
// Solidity: function setPaused(bool newPaused) returns()
func (_BatchSweepExecutor *BatchSweepExecutorSession) SetPaused(newPaused bool) (*types.Transaction, error) {
	return _BatchSweepExecutor.Contract.SetPaused(&_BatchSweepExecutor.TransactOpts, newPaused)
}

// SetPaused is a paid mutator transaction binding the contract method 0x16c38b3c.
//
// Solidity: function setPaused(bool newPaused) returns()
func (_BatchSweepExecutor *BatchSweepExecutorTransactorSession) SetPaused(newPaused bool) (*types.Transaction, error) {
	return _BatchSweepExecutor.Contract.SetPaused(&_BatchSweepExecutor.TransactOpts, newPaused)
}

// SetRecipientAllowed is a paid mutator transaction binding the contract method 0xb0adb83c.
//
// Solidity: function setRecipientAllowed(address recipient, bool allowed) returns()
func (_BatchSweepExecutor *BatchSweepExecutorTransactor) SetRecipientAllowed(opts *bind.TransactOpts, recipient common.Address, allowed bool) (*types.Transaction, error) {
	return _BatchSweepExecutor.contract.Transact(opts, "setRecipientAllowed", recipient, allowed)
}

// SetRecipientAllowed is a paid mutator transaction binding the contract method 0xb0adb83c.
//
// Solidity: function setRecipientAllowed(address recipient, bool allowed) returns()
func (_BatchSweepExecutor *BatchSweepExecutorSession) SetRecipientAllowed(recipient common.Address, allowed bool) (*types.Transaction, error) {
	return _BatchSweepExecutor.Contract.SetRecipientAllowed(&_BatchSweepExecutor.TransactOpts, recipient, allowed)
}

// SetRecipientAllowed is a paid mutator transaction binding the contract method 0xb0adb83c.
//
// Solidity: function setRecipientAllowed(address recipient, bool allowed) returns()
func (_BatchSweepExecutor *BatchSweepExecutorTransactorSession) SetRecipientAllowed(recipient common.Address, allowed bool) (*types.Transaction, error) {
	return _BatchSweepExecutor.Contract.SetRecipientAllowed(&_BatchSweepExecutor.TransactOpts, recipient, allowed)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BatchSweepExecutor *BatchSweepExecutorTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _BatchSweepExecutor.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BatchSweepExecutor *BatchSweepExecutorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BatchSweepExecutor.Contract.TransferOwnership(&_BatchSweepExecutor.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BatchSweepExecutor *BatchSweepExecutorTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BatchSweepExecutor.Contract.TransferOwnership(&_BatchSweepExecutor.TransactOpts, newOwner)
}

// BatchSweepExecutorBatchStoppedIterator is returned from FilterBatchStopped and is used to iterate over the raw logs and unpacked data for BatchStopped events raised by the BatchSweepExecutor contract.
type BatchSweepExecutorBatchStoppedIterator struct {
	Event *BatchSweepExecutorBatchStopped // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BatchSweepExecutorBatchStoppedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchSweepExecutorBatchStopped)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BatchSweepExecutorBatchStopped)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BatchSweepExecutorBatchStoppedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchSweepExecutorBatchStoppedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchSweepExecutorBatchStopped represents a BatchStopped event raised by the BatchSweepExecutor contract.
type BatchSweepExecutorBatchStopped struct {
	NextItemIndex *big.Int
	GasRemaining  *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterBatchStopped is a free log retrieval operation binding the contract event 0x8f1403bec0a17defea3a808da2ff7a763f2d127d0f930b27fcc8ff9ef9e216ce.
//
// Solidity: event BatchStopped(uint256 indexed nextItemIndex, uint256 gasRemaining)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) FilterBatchStopped(opts *bind.FilterOpts, nextItemIndex []*big.Int) (*BatchSweepExecutorBatchStoppedIterator, error) {

	var nextItemIndexRule []interface{}
	for _, nextItemIndexItem := range nextItemIndex {
		nextItemIndexRule = append(nextItemIndexRule, nextItemIndexItem)
	}

	logs, sub, err := _BatchSweepExecutor.contract.FilterLogs(opts, "BatchStopped", nextItemIndexRule)
	if err != nil {
		return nil, err
	}
	return &BatchSweepExecutorBatchStoppedIterator{contract: _BatchSweepExecutor.contract, event: "BatchStopped", logs: logs, sub: sub}, nil
}

// WatchBatchStopped is a free log subscription operation binding the contract event 0x8f1403bec0a17defea3a808da2ff7a763f2d127d0f930b27fcc8ff9ef9e216ce.
//
// Solidity: event BatchStopped(uint256 indexed nextItemIndex, uint256 gasRemaining)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) WatchBatchStopped(opts *bind.WatchOpts, sink chan<- *BatchSweepExecutorBatchStopped, nextItemIndex []*big.Int) (event.Subscription, error) {

	var nextItemIndexRule []interface{}
	for _, nextItemIndexItem := range nextItemIndex {
		nextItemIndexRule = append(nextItemIndexRule, nextItemIndexItem)
	}

	logs, sub, err := _BatchSweepExecutor.contract.WatchLogs(opts, "BatchStopped", nextItemIndexRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchSweepExecutorBatchStopped)
				if err := _BatchSweepExecutor.contract.UnpackLog(event, "BatchStopped", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseBatchStopped is a log parse operation binding the contract event 0x8f1403bec0a17defea3a808da2ff7a763f2d127d0f930b27fcc8ff9ef9e216ce.
//
// Solidity: event BatchStopped(uint256 indexed nextItemIndex, uint256 gasRemaining)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) ParseBatchStopped(log types.Log) (*BatchSweepExecutorBatchStopped, error) {
	event := new(BatchSweepExecutorBatchStopped)
	if err := _BatchSweepExecutor.contract.UnpackLog(event, "BatchStopped", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BatchSweepExecutorCollectResultIterator is returned from FilterCollectResult and is used to iterate over the raw logs and unpacked data for CollectResult events raised by the BatchSweepExecutor contract.
type BatchSweepExecutorCollectResultIterator struct {
	Event *BatchSweepExecutorCollectResult // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BatchSweepExecutorCollectResultIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchSweepExecutorCollectResult)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BatchSweepExecutorCollectResult)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BatchSweepExecutorCollectResultIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchSweepExecutorCollectResultIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchSweepExecutorCollectResult represents a CollectResult event raised by the BatchSweepExecutor contract.
type BatchSweepExecutorCollectResult struct {
	TaskId          *big.Int
	Account         common.Address
	Token           common.Address
	Recipient       common.Address
	RequestedAmount *big.Int
	Result          uint8
	ReturnDataHash  [32]byte
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterCollectResult is a free log retrieval operation binding the contract event 0x87b90f229a62351bfcf57edd6a6ef56caca9eedd001aacac5f83d760fa9b6266.
//
// Solidity: event CollectResult(uint256 indexed taskId, address indexed account, address indexed token, address recipient, uint256 requestedAmount, uint8 result, bytes32 returnDataHash)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) FilterCollectResult(opts *bind.FilterOpts, taskId []*big.Int, account []common.Address, token []common.Address) (*BatchSweepExecutorCollectResultIterator, error) {

	var taskIdRule []interface{}
	for _, taskIdItem := range taskId {
		taskIdRule = append(taskIdRule, taskIdItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _BatchSweepExecutor.contract.FilterLogs(opts, "CollectResult", taskIdRule, accountRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return &BatchSweepExecutorCollectResultIterator{contract: _BatchSweepExecutor.contract, event: "CollectResult", logs: logs, sub: sub}, nil
}

// WatchCollectResult is a free log subscription operation binding the contract event 0x87b90f229a62351bfcf57edd6a6ef56caca9eedd001aacac5f83d760fa9b6266.
//
// Solidity: event CollectResult(uint256 indexed taskId, address indexed account, address indexed token, address recipient, uint256 requestedAmount, uint8 result, bytes32 returnDataHash)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) WatchCollectResult(opts *bind.WatchOpts, sink chan<- *BatchSweepExecutorCollectResult, taskId []*big.Int, account []common.Address, token []common.Address) (event.Subscription, error) {

	var taskIdRule []interface{}
	for _, taskIdItem := range taskId {
		taskIdRule = append(taskIdRule, taskIdItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _BatchSweepExecutor.contract.WatchLogs(opts, "CollectResult", taskIdRule, accountRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchSweepExecutorCollectResult)
				if err := _BatchSweepExecutor.contract.UnpackLog(event, "CollectResult", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCollectResult is a log parse operation binding the contract event 0x87b90f229a62351bfcf57edd6a6ef56caca9eedd001aacac5f83d760fa9b6266.
//
// Solidity: event CollectResult(uint256 indexed taskId, address indexed account, address indexed token, address recipient, uint256 requestedAmount, uint8 result, bytes32 returnDataHash)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) ParseCollectResult(log types.Log) (*BatchSweepExecutorCollectResult, error) {
	event := new(BatchSweepExecutorCollectResult)
	if err := _BatchSweepExecutor.contract.UnpackLog(event, "CollectResult", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BatchSweepExecutorOperatorUpdatedIterator is returned from FilterOperatorUpdated and is used to iterate over the raw logs and unpacked data for OperatorUpdated events raised by the BatchSweepExecutor contract.
type BatchSweepExecutorOperatorUpdatedIterator struct {
	Event *BatchSweepExecutorOperatorUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BatchSweepExecutorOperatorUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchSweepExecutorOperatorUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BatchSweepExecutorOperatorUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BatchSweepExecutorOperatorUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchSweepExecutorOperatorUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchSweepExecutorOperatorUpdated represents a OperatorUpdated event raised by the BatchSweepExecutor contract.
type BatchSweepExecutorOperatorUpdated struct {
	PreviousOperator common.Address
	NewOperator      common.Address
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterOperatorUpdated is a free log retrieval operation binding the contract event 0xfbe5b6cbafb274f445d7fed869dc77a838d8243a22c460de156560e8857cad03.
//
// Solidity: event OperatorUpdated(address indexed previousOperator, address indexed newOperator)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) FilterOperatorUpdated(opts *bind.FilterOpts, previousOperator []common.Address, newOperator []common.Address) (*BatchSweepExecutorOperatorUpdatedIterator, error) {

	var previousOperatorRule []interface{}
	for _, previousOperatorItem := range previousOperator {
		previousOperatorRule = append(previousOperatorRule, previousOperatorItem)
	}
	var newOperatorRule []interface{}
	for _, newOperatorItem := range newOperator {
		newOperatorRule = append(newOperatorRule, newOperatorItem)
	}

	logs, sub, err := _BatchSweepExecutor.contract.FilterLogs(opts, "OperatorUpdated", previousOperatorRule, newOperatorRule)
	if err != nil {
		return nil, err
	}
	return &BatchSweepExecutorOperatorUpdatedIterator{contract: _BatchSweepExecutor.contract, event: "OperatorUpdated", logs: logs, sub: sub}, nil
}

// WatchOperatorUpdated is a free log subscription operation binding the contract event 0xfbe5b6cbafb274f445d7fed869dc77a838d8243a22c460de156560e8857cad03.
//
// Solidity: event OperatorUpdated(address indexed previousOperator, address indexed newOperator)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) WatchOperatorUpdated(opts *bind.WatchOpts, sink chan<- *BatchSweepExecutorOperatorUpdated, previousOperator []common.Address, newOperator []common.Address) (event.Subscription, error) {

	var previousOperatorRule []interface{}
	for _, previousOperatorItem := range previousOperator {
		previousOperatorRule = append(previousOperatorRule, previousOperatorItem)
	}
	var newOperatorRule []interface{}
	for _, newOperatorItem := range newOperator {
		newOperatorRule = append(newOperatorRule, newOperatorItem)
	}

	logs, sub, err := _BatchSweepExecutor.contract.WatchLogs(opts, "OperatorUpdated", previousOperatorRule, newOperatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchSweepExecutorOperatorUpdated)
				if err := _BatchSweepExecutor.contract.UnpackLog(event, "OperatorUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOperatorUpdated is a log parse operation binding the contract event 0xfbe5b6cbafb274f445d7fed869dc77a838d8243a22c460de156560e8857cad03.
//
// Solidity: event OperatorUpdated(address indexed previousOperator, address indexed newOperator)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) ParseOperatorUpdated(log types.Log) (*BatchSweepExecutorOperatorUpdated, error) {
	event := new(BatchSweepExecutorOperatorUpdated)
	if err := _BatchSweepExecutor.contract.UnpackLog(event, "OperatorUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BatchSweepExecutorOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the BatchSweepExecutor contract.
type BatchSweepExecutorOwnershipTransferStartedIterator struct {
	Event *BatchSweepExecutorOwnershipTransferStarted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BatchSweepExecutorOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchSweepExecutorOwnershipTransferStarted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BatchSweepExecutorOwnershipTransferStarted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BatchSweepExecutorOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchSweepExecutorOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchSweepExecutorOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the BatchSweepExecutor contract.
type BatchSweepExecutorOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BatchSweepExecutorOwnershipTransferStartedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BatchSweepExecutor.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BatchSweepExecutorOwnershipTransferStartedIterator{contract: _BatchSweepExecutor.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *BatchSweepExecutorOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BatchSweepExecutor.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchSweepExecutorOwnershipTransferStarted)
				if err := _BatchSweepExecutor.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferStarted is a log parse operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) ParseOwnershipTransferStarted(log types.Log) (*BatchSweepExecutorOwnershipTransferStarted, error) {
	event := new(BatchSweepExecutorOwnershipTransferStarted)
	if err := _BatchSweepExecutor.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BatchSweepExecutorOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the BatchSweepExecutor contract.
type BatchSweepExecutorOwnershipTransferredIterator struct {
	Event *BatchSweepExecutorOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BatchSweepExecutorOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchSweepExecutorOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BatchSweepExecutorOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BatchSweepExecutorOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchSweepExecutorOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchSweepExecutorOwnershipTransferred represents a OwnershipTransferred event raised by the BatchSweepExecutor contract.
type BatchSweepExecutorOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BatchSweepExecutorOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BatchSweepExecutor.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BatchSweepExecutorOwnershipTransferredIterator{contract: _BatchSweepExecutor.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BatchSweepExecutorOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BatchSweepExecutor.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchSweepExecutorOwnershipTransferred)
				if err := _BatchSweepExecutor.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) ParseOwnershipTransferred(log types.Log) (*BatchSweepExecutorOwnershipTransferred, error) {
	event := new(BatchSweepExecutorOwnershipTransferred)
	if err := _BatchSweepExecutor.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BatchSweepExecutorPauseUpdatedIterator is returned from FilterPauseUpdated and is used to iterate over the raw logs and unpacked data for PauseUpdated events raised by the BatchSweepExecutor contract.
type BatchSweepExecutorPauseUpdatedIterator struct {
	Event *BatchSweepExecutorPauseUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BatchSweepExecutorPauseUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchSweepExecutorPauseUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BatchSweepExecutorPauseUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BatchSweepExecutorPauseUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchSweepExecutorPauseUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchSweepExecutorPauseUpdated represents a PauseUpdated event raised by the BatchSweepExecutor contract.
type BatchSweepExecutorPauseUpdated struct {
	Paused bool
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterPauseUpdated is a free log retrieval operation binding the contract event 0x77860e247ab9186dbe64e5bd0e0b93273cc4273e01818420e788f500078886f5.
//
// Solidity: event PauseUpdated(bool paused)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) FilterPauseUpdated(opts *bind.FilterOpts) (*BatchSweepExecutorPauseUpdatedIterator, error) {

	logs, sub, err := _BatchSweepExecutor.contract.FilterLogs(opts, "PauseUpdated")
	if err != nil {
		return nil, err
	}
	return &BatchSweepExecutorPauseUpdatedIterator{contract: _BatchSweepExecutor.contract, event: "PauseUpdated", logs: logs, sub: sub}, nil
}

// WatchPauseUpdated is a free log subscription operation binding the contract event 0x77860e247ab9186dbe64e5bd0e0b93273cc4273e01818420e788f500078886f5.
//
// Solidity: event PauseUpdated(bool paused)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) WatchPauseUpdated(opts *bind.WatchOpts, sink chan<- *BatchSweepExecutorPauseUpdated) (event.Subscription, error) {

	logs, sub, err := _BatchSweepExecutor.contract.WatchLogs(opts, "PauseUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchSweepExecutorPauseUpdated)
				if err := _BatchSweepExecutor.contract.UnpackLog(event, "PauseUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePauseUpdated is a log parse operation binding the contract event 0x77860e247ab9186dbe64e5bd0e0b93273cc4273e01818420e788f500078886f5.
//
// Solidity: event PauseUpdated(bool paused)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) ParsePauseUpdated(log types.Log) (*BatchSweepExecutorPauseUpdated, error) {
	event := new(BatchSweepExecutorPauseUpdated)
	if err := _BatchSweepExecutor.contract.UnpackLog(event, "PauseUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BatchSweepExecutorRecipientPermissionUpdatedIterator is returned from FilterRecipientPermissionUpdated and is used to iterate over the raw logs and unpacked data for RecipientPermissionUpdated events raised by the BatchSweepExecutor contract.
type BatchSweepExecutorRecipientPermissionUpdatedIterator struct {
	Event *BatchSweepExecutorRecipientPermissionUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BatchSweepExecutorRecipientPermissionUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchSweepExecutorRecipientPermissionUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BatchSweepExecutorRecipientPermissionUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BatchSweepExecutorRecipientPermissionUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchSweepExecutorRecipientPermissionUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchSweepExecutorRecipientPermissionUpdated represents a RecipientPermissionUpdated event raised by the BatchSweepExecutor contract.
type BatchSweepExecutorRecipientPermissionUpdated struct {
	Recipient common.Address
	Allowed   bool
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterRecipientPermissionUpdated is a free log retrieval operation binding the contract event 0x0f89d0212fb3085400640538b4ffbf16f0e0fee26864da229ac4d3a00a4ba62d.
//
// Solidity: event RecipientPermissionUpdated(address indexed recipient, bool allowed)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) FilterRecipientPermissionUpdated(opts *bind.FilterOpts, recipient []common.Address) (*BatchSweepExecutorRecipientPermissionUpdatedIterator, error) {

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _BatchSweepExecutor.contract.FilterLogs(opts, "RecipientPermissionUpdated", recipientRule)
	if err != nil {
		return nil, err
	}
	return &BatchSweepExecutorRecipientPermissionUpdatedIterator{contract: _BatchSweepExecutor.contract, event: "RecipientPermissionUpdated", logs: logs, sub: sub}, nil
}

// WatchRecipientPermissionUpdated is a free log subscription operation binding the contract event 0x0f89d0212fb3085400640538b4ffbf16f0e0fee26864da229ac4d3a00a4ba62d.
//
// Solidity: event RecipientPermissionUpdated(address indexed recipient, bool allowed)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) WatchRecipientPermissionUpdated(opts *bind.WatchOpts, sink chan<- *BatchSweepExecutorRecipientPermissionUpdated, recipient []common.Address) (event.Subscription, error) {

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _BatchSweepExecutor.contract.WatchLogs(opts, "RecipientPermissionUpdated", recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchSweepExecutorRecipientPermissionUpdated)
				if err := _BatchSweepExecutor.contract.UnpackLog(event, "RecipientPermissionUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRecipientPermissionUpdated is a log parse operation binding the contract event 0x0f89d0212fb3085400640538b4ffbf16f0e0fee26864da229ac4d3a00a4ba62d.
//
// Solidity: event RecipientPermissionUpdated(address indexed recipient, bool allowed)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) ParseRecipientPermissionUpdated(log types.Log) (*BatchSweepExecutorRecipientPermissionUpdated, error) {
	event := new(BatchSweepExecutorRecipientPermissionUpdated)
	if err := _BatchSweepExecutor.contract.UnpackLog(event, "RecipientPermissionUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BatchSweepExecutorSweepDelegateInitializedIterator is returned from FilterSweepDelegateInitialized and is used to iterate over the raw logs and unpacked data for SweepDelegateInitialized events raised by the BatchSweepExecutor contract.
type BatchSweepExecutorSweepDelegateInitializedIterator struct {
	Event *BatchSweepExecutorSweepDelegateInitialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BatchSweepExecutorSweepDelegateInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchSweepExecutorSweepDelegateInitialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BatchSweepExecutorSweepDelegateInitialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BatchSweepExecutorSweepDelegateInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchSweepExecutorSweepDelegateInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchSweepExecutorSweepDelegateInitialized represents a SweepDelegateInitialized event raised by the BatchSweepExecutor contract.
type BatchSweepExecutorSweepDelegateInitialized struct {
	Delegate common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterSweepDelegateInitialized is a free log retrieval operation binding the contract event 0xd0a9f96ddd4c7bbecc65a031d025a018dc3e8afcd95ea3cccdec6977c2bad9bd.
//
// Solidity: event SweepDelegateInitialized(address indexed delegate)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) FilterSweepDelegateInitialized(opts *bind.FilterOpts, delegate []common.Address) (*BatchSweepExecutorSweepDelegateInitializedIterator, error) {

	var delegateRule []interface{}
	for _, delegateItem := range delegate {
		delegateRule = append(delegateRule, delegateItem)
	}

	logs, sub, err := _BatchSweepExecutor.contract.FilterLogs(opts, "SweepDelegateInitialized", delegateRule)
	if err != nil {
		return nil, err
	}
	return &BatchSweepExecutorSweepDelegateInitializedIterator{contract: _BatchSweepExecutor.contract, event: "SweepDelegateInitialized", logs: logs, sub: sub}, nil
}

// WatchSweepDelegateInitialized is a free log subscription operation binding the contract event 0xd0a9f96ddd4c7bbecc65a031d025a018dc3e8afcd95ea3cccdec6977c2bad9bd.
//
// Solidity: event SweepDelegateInitialized(address indexed delegate)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) WatchSweepDelegateInitialized(opts *bind.WatchOpts, sink chan<- *BatchSweepExecutorSweepDelegateInitialized, delegate []common.Address) (event.Subscription, error) {

	var delegateRule []interface{}
	for _, delegateItem := range delegate {
		delegateRule = append(delegateRule, delegateItem)
	}

	logs, sub, err := _BatchSweepExecutor.contract.WatchLogs(opts, "SweepDelegateInitialized", delegateRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchSweepExecutorSweepDelegateInitialized)
				if err := _BatchSweepExecutor.contract.UnpackLog(event, "SweepDelegateInitialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSweepDelegateInitialized is a log parse operation binding the contract event 0xd0a9f96ddd4c7bbecc65a031d025a018dc3e8afcd95ea3cccdec6977c2bad9bd.
//
// Solidity: event SweepDelegateInitialized(address indexed delegate)
func (_BatchSweepExecutor *BatchSweepExecutorFilterer) ParseSweepDelegateInitialized(log types.Log) (*BatchSweepExecutorSweepDelegateInitialized, error) {
	event := new(BatchSweepExecutorSweepDelegateInitialized)
	if err := _BatchSweepExecutor.contract.UnpackLog(event, "SweepDelegateInitialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
