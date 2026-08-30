// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package sweepdelegate

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

// SweepDelegateMetaData contains all meta data concerning the SweepDelegate contract.
var SweepDelegateMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"executor\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"InvalidToken\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"NativeTransferFailed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"TokenTransferFailed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"name\":\"UnauthorizedCaller\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"EXECUTOR\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SWEEP_SUCCESS\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"\",\"type\":\"bytes4\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"sweepERC20\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"magic\",\"type\":\"bytes4\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"sweepNative\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"magic\",\"type\":\"bytes4\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// SweepDelegateABI is the input ABI used to generate the binding from.
// Deprecated: Use SweepDelegateMetaData.ABI instead.
var SweepDelegateABI = SweepDelegateMetaData.ABI

// SweepDelegate is an auto generated Go binding around an Ethereum contract.
type SweepDelegate struct {
	SweepDelegateCaller     // Read-only binding to the contract
	SweepDelegateTransactor // Write-only binding to the contract
	SweepDelegateFilterer   // Log filterer for contract events
}

// SweepDelegateCaller is an auto generated read-only Go binding around an Ethereum contract.
type SweepDelegateCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SweepDelegateTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SweepDelegateTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SweepDelegateFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SweepDelegateFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SweepDelegateSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SweepDelegateSession struct {
	Contract     *SweepDelegate    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SweepDelegateCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SweepDelegateCallerSession struct {
	Contract *SweepDelegateCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// SweepDelegateTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SweepDelegateTransactorSession struct {
	Contract     *SweepDelegateTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// SweepDelegateRaw is an auto generated low-level Go binding around an Ethereum contract.
type SweepDelegateRaw struct {
	Contract *SweepDelegate // Generic contract binding to access the raw methods on
}

// SweepDelegateCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SweepDelegateCallerRaw struct {
	Contract *SweepDelegateCaller // Generic read-only contract binding to access the raw methods on
}

// SweepDelegateTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SweepDelegateTransactorRaw struct {
	Contract *SweepDelegateTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSweepDelegate creates a new instance of SweepDelegate, bound to a specific deployed contract.
func NewSweepDelegate(address common.Address, backend bind.ContractBackend) (*SweepDelegate, error) {
	contract, err := bindSweepDelegate(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SweepDelegate{SweepDelegateCaller: SweepDelegateCaller{contract: contract}, SweepDelegateTransactor: SweepDelegateTransactor{contract: contract}, SweepDelegateFilterer: SweepDelegateFilterer{contract: contract}}, nil
}

// NewSweepDelegateCaller creates a new read-only instance of SweepDelegate, bound to a specific deployed contract.
func NewSweepDelegateCaller(address common.Address, caller bind.ContractCaller) (*SweepDelegateCaller, error) {
	contract, err := bindSweepDelegate(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SweepDelegateCaller{contract: contract}, nil
}

// NewSweepDelegateTransactor creates a new write-only instance of SweepDelegate, bound to a specific deployed contract.
func NewSweepDelegateTransactor(address common.Address, transactor bind.ContractTransactor) (*SweepDelegateTransactor, error) {
	contract, err := bindSweepDelegate(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SweepDelegateTransactor{contract: contract}, nil
}

// NewSweepDelegateFilterer creates a new log filterer instance of SweepDelegate, bound to a specific deployed contract.
func NewSweepDelegateFilterer(address common.Address, filterer bind.ContractFilterer) (*SweepDelegateFilterer, error) {
	contract, err := bindSweepDelegate(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SweepDelegateFilterer{contract: contract}, nil
}

// bindSweepDelegate binds a generic wrapper to an already deployed contract.
func bindSweepDelegate(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := SweepDelegateMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SweepDelegate *SweepDelegateRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SweepDelegate.Contract.SweepDelegateCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SweepDelegate *SweepDelegateRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SweepDelegate.Contract.SweepDelegateTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SweepDelegate *SweepDelegateRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SweepDelegate.Contract.SweepDelegateTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SweepDelegate *SweepDelegateCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SweepDelegate.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SweepDelegate *SweepDelegateTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SweepDelegate.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SweepDelegate *SweepDelegateTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SweepDelegate.Contract.contract.Transact(opts, method, params...)
}

// EXECUTOR is a free data retrieval call binding the contract method 0x630dc7cb.
//
// Solidity: function EXECUTOR() view returns(address)
func (_SweepDelegate *SweepDelegateCaller) EXECUTOR(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SweepDelegate.contract.Call(opts, &out, "EXECUTOR")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// EXECUTOR is a free data retrieval call binding the contract method 0x630dc7cb.
//
// Solidity: function EXECUTOR() view returns(address)
func (_SweepDelegate *SweepDelegateSession) EXECUTOR() (common.Address, error) {
	return _SweepDelegate.Contract.EXECUTOR(&_SweepDelegate.CallOpts)
}

// EXECUTOR is a free data retrieval call binding the contract method 0x630dc7cb.
//
// Solidity: function EXECUTOR() view returns(address)
func (_SweepDelegate *SweepDelegateCallerSession) EXECUTOR() (common.Address, error) {
	return _SweepDelegate.Contract.EXECUTOR(&_SweepDelegate.CallOpts)
}

// SWEEPSUCCESS is a free data retrieval call binding the contract method 0x3554972c.
//
// Solidity: function SWEEP_SUCCESS() view returns(bytes4)
func (_SweepDelegate *SweepDelegateCaller) SWEEPSUCCESS(opts *bind.CallOpts) ([4]byte, error) {
	var out []interface{}
	err := _SweepDelegate.contract.Call(opts, &out, "SWEEP_SUCCESS")

	if err != nil {
		return *new([4]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([4]byte)).(*[4]byte)

	return out0, err

}

// SWEEPSUCCESS is a free data retrieval call binding the contract method 0x3554972c.
//
// Solidity: function SWEEP_SUCCESS() view returns(bytes4)
func (_SweepDelegate *SweepDelegateSession) SWEEPSUCCESS() ([4]byte, error) {
	return _SweepDelegate.Contract.SWEEPSUCCESS(&_SweepDelegate.CallOpts)
}

// SWEEPSUCCESS is a free data retrieval call binding the contract method 0x3554972c.
//
// Solidity: function SWEEP_SUCCESS() view returns(bytes4)
func (_SweepDelegate *SweepDelegateCallerSession) SWEEPSUCCESS() ([4]byte, error) {
	return _SweepDelegate.Contract.SWEEPSUCCESS(&_SweepDelegate.CallOpts)
}

// SweepERC20 is a paid mutator transaction binding the contract method 0x503690d1.
//
// Solidity: function sweepERC20(address token, address recipient, uint256 amount) returns(bytes4 magic)
func (_SweepDelegate *SweepDelegateTransactor) SweepERC20(opts *bind.TransactOpts, token common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _SweepDelegate.contract.Transact(opts, "sweepERC20", token, recipient, amount)
}

// SweepERC20 is a paid mutator transaction binding the contract method 0x503690d1.
//
// Solidity: function sweepERC20(address token, address recipient, uint256 amount) returns(bytes4 magic)
func (_SweepDelegate *SweepDelegateSession) SweepERC20(token common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _SweepDelegate.Contract.SweepERC20(&_SweepDelegate.TransactOpts, token, recipient, amount)
}

// SweepERC20 is a paid mutator transaction binding the contract method 0x503690d1.
//
// Solidity: function sweepERC20(address token, address recipient, uint256 amount) returns(bytes4 magic)
func (_SweepDelegate *SweepDelegateTransactorSession) SweepERC20(token common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _SweepDelegate.Contract.SweepERC20(&_SweepDelegate.TransactOpts, token, recipient, amount)
}

// SweepNative is a paid mutator transaction binding the contract method 0x9646f3ea.
//
// Solidity: function sweepNative(address recipient, uint256 amount) returns(bytes4 magic)
func (_SweepDelegate *SweepDelegateTransactor) SweepNative(opts *bind.TransactOpts, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _SweepDelegate.contract.Transact(opts, "sweepNative", recipient, amount)
}

// SweepNative is a paid mutator transaction binding the contract method 0x9646f3ea.
//
// Solidity: function sweepNative(address recipient, uint256 amount) returns(bytes4 magic)
func (_SweepDelegate *SweepDelegateSession) SweepNative(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _SweepDelegate.Contract.SweepNative(&_SweepDelegate.TransactOpts, recipient, amount)
}

// SweepNative is a paid mutator transaction binding the contract method 0x9646f3ea.
//
// Solidity: function sweepNative(address recipient, uint256 amount) returns(bytes4 magic)
func (_SweepDelegate *SweepDelegateTransactorSession) SweepNative(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _SweepDelegate.Contract.SweepNative(&_SweepDelegate.TransactOpts, recipient, amount)
}
