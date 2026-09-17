// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package blockchain

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

// IoTDeviceManagerAuthRecord is an auto generated low-level Go binding around an user-defined struct.
type IoTDeviceManagerAuthRecord struct {
	Timestamp *big.Int
	Success   bool
	ProofHash [32]byte
}

// IoTDeviceManagerTransaction is an auto generated low-level Go binding around an user-defined struct.
type IoTDeviceManagerTransaction struct {
	DeviceId      string
	Data          []byte
	MessageHash   string
	Sequence      *big.Int
	Timestamp     *big.Int
	DeviceAddress common.Address
}

// IoTDeviceManagerMetaData contains all meta data concerning the IoTDeviceManager contract.
var IoTDeviceManagerMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"}],\"name\":\"AuthenticationRecorded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"}],\"name\":\"DeviceActivated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"DeviceAuthenticated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"}],\"name\":\"DeviceDeactivated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"publicKey\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"enumIoTDeviceManager.EntityType\",\"name\":\"entityType\",\"type\":\"uint8\"}],\"name\":\"DeviceRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"deviceId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"sequence\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"}],\"name\":\"TransactionRecorded\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"}],\"name\":\"activateDevice\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"activeDevicesCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"authCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"authHistory\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes32\",\"name\":\"proofHash\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"messageHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"authenticateDevice\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"}],\"name\":\"deactivateDevice\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"deviceList\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"deviceTransactions\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"deviceId\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"string\",\"name\":\"messageHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"sequence\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"devices\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"publicKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"lastSeen\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isActive\",\"type\":\"bool\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"enumIoTDeviceManager.EntityType\",\"name\":\"entityType\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAllDevices\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"}],\"name\":\"getAuthHistory\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes32\",\"name\":\"proofHash\",\"type\":\"bytes32\"}],\"internalType\":\"structIoTDeviceManager.AuthRecord[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"}],\"name\":\"getAuthHistorySize\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"}],\"name\":\"getDeviceAuthCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"}],\"name\":\"getDeviceInfo\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"publicKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"},{\"internalType\":\"bool\",\"name\":\"isActive\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"lastSeen\",\"type\":\"uint256\"},{\"internalType\":\"enumIoTDeviceManager.EntityType\",\"name\":\"entityType\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"deviceId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"start\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit\",\"type\":\"uint256\"}],\"name\":\"getDeviceTransactions\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"deviceId\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"string\",\"name\":\"messageHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"sequence\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"}],\"internalType\":\"structIoTDeviceManager.Transaction[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getDevicesCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"start\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit\",\"type\":\"uint256\"}],\"name\":\"getDevicesPaginated\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"}],\"name\":\"getEntityType\",\"outputs\":[{\"internalType\":\"enumIoTDeviceManager.EntityType\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getGlobalStats\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"_totalDevices\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_totalAuthentications\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_activeDevices\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_inactiveDevices\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"deviceId\",\"type\":\"string\"}],\"name\":\"getLatestTransaction\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"deviceId\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"string\",\"name\":\"messageHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"sequence\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"}],\"internalType\":\"structIoTDeviceManager.Transaction\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"limit\",\"type\":\"uint256\"}],\"name\":\"getRecentAuthHistory\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes32\",\"name\":\"proofHash\",\"type\":\"bytes32\"}],\"internalType\":\"structIoTDeviceManager.AuthRecord[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"deviceId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getTransaction\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"deviceId\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"string\",\"name\":\"messageHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"sequence\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"}],\"internalType\":\"structIoTDeviceManager.Transaction\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"deviceId\",\"type\":\"string\"}],\"name\":\"getTransactionCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"isDeviceListed\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"}],\"name\":\"isDeviceRegistered\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"lastSequence\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes32\",\"name\":\"proofHash\",\"type\":\"bytes32\"}],\"name\":\"recordAuthentication\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"deviceId\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"string\",\"name\":\"messageHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"sequence\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"recordTransaction\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_hash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"_signature\",\"type\":\"bytes\"}],\"name\":\"recoverSigner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"_publicKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_metadata\",\"type\":\"string\"},{\"internalType\":\"enumIoTDeviceManager.EntityType\",\"name\":\"_entityType\",\"type\":\"uint8\"}],\"name\":\"registerDevice\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalAuthentications\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalDevices\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"usedSignatures\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// IoTDeviceManagerABI is the input ABI used to generate the binding from.
// Deprecated: Use IoTDeviceManagerMetaData.ABI instead.
var IoTDeviceManagerABI = IoTDeviceManagerMetaData.ABI

// IoTDeviceManager is an auto generated Go binding around an Ethereum contract.
type IoTDeviceManager struct {
	IoTDeviceManagerCaller     // Read-only binding to the contract
	IoTDeviceManagerTransactor // Write-only binding to the contract
	IoTDeviceManagerFilterer   // Log filterer for contract events
}

// IoTDeviceManagerCaller is an auto generated read-only Go binding around an Ethereum contract.
type IoTDeviceManagerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IoTDeviceManagerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IoTDeviceManagerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IoTDeviceManagerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IoTDeviceManagerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IoTDeviceManagerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IoTDeviceManagerSession struct {
	Contract     *IoTDeviceManager // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IoTDeviceManagerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IoTDeviceManagerCallerSession struct {
	Contract *IoTDeviceManagerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// IoTDeviceManagerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IoTDeviceManagerTransactorSession struct {
	Contract     *IoTDeviceManagerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// IoTDeviceManagerRaw is an auto generated low-level Go binding around an Ethereum contract.
type IoTDeviceManagerRaw struct {
	Contract *IoTDeviceManager // Generic contract binding to access the raw methods on
}

// IoTDeviceManagerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IoTDeviceManagerCallerRaw struct {
	Contract *IoTDeviceManagerCaller // Generic read-only contract binding to access the raw methods on
}

// IoTDeviceManagerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IoTDeviceManagerTransactorRaw struct {
	Contract *IoTDeviceManagerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIoTDeviceManager creates a new instance of IoTDeviceManager, bound to a specific deployed contract.
func NewIoTDeviceManager(address common.Address, backend bind.ContractBackend) (*IoTDeviceManager, error) {
	contract, err := bindIoTDeviceManager(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IoTDeviceManager{IoTDeviceManagerCaller: IoTDeviceManagerCaller{contract: contract}, IoTDeviceManagerTransactor: IoTDeviceManagerTransactor{contract: contract}, IoTDeviceManagerFilterer: IoTDeviceManagerFilterer{contract: contract}}, nil
}

// NewIoTDeviceManagerCaller creates a new read-only instance of IoTDeviceManager, bound to a specific deployed contract.
func NewIoTDeviceManagerCaller(address common.Address, caller bind.ContractCaller) (*IoTDeviceManagerCaller, error) {
	contract, err := bindIoTDeviceManager(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IoTDeviceManagerCaller{contract: contract}, nil
}

// NewIoTDeviceManagerTransactor creates a new write-only instance of IoTDeviceManager, bound to a specific deployed contract.
func NewIoTDeviceManagerTransactor(address common.Address, transactor bind.ContractTransactor) (*IoTDeviceManagerTransactor, error) {
	contract, err := bindIoTDeviceManager(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IoTDeviceManagerTransactor{contract: contract}, nil
}

// NewIoTDeviceManagerFilterer creates a new log filterer instance of IoTDeviceManager, bound to a specific deployed contract.
func NewIoTDeviceManagerFilterer(address common.Address, filterer bind.ContractFilterer) (*IoTDeviceManagerFilterer, error) {
	contract, err := bindIoTDeviceManager(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IoTDeviceManagerFilterer{contract: contract}, nil
}

// bindIoTDeviceManager binds a generic wrapper to an already deployed contract.
func bindIoTDeviceManager(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := IoTDeviceManagerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IoTDeviceManager *IoTDeviceManagerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IoTDeviceManager.Contract.IoTDeviceManagerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IoTDeviceManager *IoTDeviceManagerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IoTDeviceManager.Contract.IoTDeviceManagerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IoTDeviceManager *IoTDeviceManagerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IoTDeviceManager.Contract.IoTDeviceManagerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IoTDeviceManager *IoTDeviceManagerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IoTDeviceManager.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IoTDeviceManager *IoTDeviceManagerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IoTDeviceManager.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IoTDeviceManager *IoTDeviceManagerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IoTDeviceManager.Contract.contract.Transact(opts, method, params...)
}

// ActiveDevicesCount is a free data retrieval call binding the contract method 0x4df97928.
//
// Solidity: function activeDevicesCount() view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerCaller) ActiveDevicesCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "activeDevicesCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ActiveDevicesCount is a free data retrieval call binding the contract method 0x4df97928.
//
// Solidity: function activeDevicesCount() view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerSession) ActiveDevicesCount() (*big.Int, error) {
	return _IoTDeviceManager.Contract.ActiveDevicesCount(&_IoTDeviceManager.CallOpts)
}

// ActiveDevicesCount is a free data retrieval call binding the contract method 0x4df97928.
//
// Solidity: function activeDevicesCount() view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) ActiveDevicesCount() (*big.Int, error) {
	return _IoTDeviceManager.Contract.ActiveDevicesCount(&_IoTDeviceManager.CallOpts)
}

// AuthCount is a free data retrieval call binding the contract method 0xa17eec9f.
//
// Solidity: function authCount(address ) view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerCaller) AuthCount(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "authCount", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// AuthCount is a free data retrieval call binding the contract method 0xa17eec9f.
//
// Solidity: function authCount(address ) view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerSession) AuthCount(arg0 common.Address) (*big.Int, error) {
	return _IoTDeviceManager.Contract.AuthCount(&_IoTDeviceManager.CallOpts, arg0)
}

// AuthCount is a free data retrieval call binding the contract method 0xa17eec9f.
//
// Solidity: function authCount(address ) view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) AuthCount(arg0 common.Address) (*big.Int, error) {
	return _IoTDeviceManager.Contract.AuthCount(&_IoTDeviceManager.CallOpts, arg0)
}

// AuthHistory is a free data retrieval call binding the contract method 0xb879fee7.
//
// Solidity: function authHistory(address , uint256 ) view returns(uint256 timestamp, bool success, bytes32 proofHash)
func (_IoTDeviceManager *IoTDeviceManagerCaller) AuthHistory(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (struct {
	Timestamp *big.Int
	Success   bool
	ProofHash [32]byte
}, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "authHistory", arg0, arg1)

	outstruct := new(struct {
		Timestamp *big.Int
		Success   bool
		ProofHash [32]byte
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Timestamp = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Success = *abi.ConvertType(out[1], new(bool)).(*bool)
	outstruct.ProofHash = *abi.ConvertType(out[2], new([32]byte)).(*[32]byte)

	return *outstruct, err

}

// AuthHistory is a free data retrieval call binding the contract method 0xb879fee7.
//
// Solidity: function authHistory(address , uint256 ) view returns(uint256 timestamp, bool success, bytes32 proofHash)
func (_IoTDeviceManager *IoTDeviceManagerSession) AuthHistory(arg0 common.Address, arg1 *big.Int) (struct {
	Timestamp *big.Int
	Success   bool
	ProofHash [32]byte
}, error) {
	return _IoTDeviceManager.Contract.AuthHistory(&_IoTDeviceManager.CallOpts, arg0, arg1)
}

// AuthHistory is a free data retrieval call binding the contract method 0xb879fee7.
//
// Solidity: function authHistory(address , uint256 ) view returns(uint256 timestamp, bool success, bytes32 proofHash)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) AuthHistory(arg0 common.Address, arg1 *big.Int) (struct {
	Timestamp *big.Int
	Success   bool
	ProofHash [32]byte
}, error) {
	return _IoTDeviceManager.Contract.AuthHistory(&_IoTDeviceManager.CallOpts, arg0, arg1)
}

// DeviceList is a free data retrieval call binding the contract method 0x979ebb42.
//
// Solidity: function deviceList(uint256 ) view returns(address)
func (_IoTDeviceManager *IoTDeviceManagerCaller) DeviceList(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "deviceList", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// DeviceList is a free data retrieval call binding the contract method 0x979ebb42.
//
// Solidity: function deviceList(uint256 ) view returns(address)
func (_IoTDeviceManager *IoTDeviceManagerSession) DeviceList(arg0 *big.Int) (common.Address, error) {
	return _IoTDeviceManager.Contract.DeviceList(&_IoTDeviceManager.CallOpts, arg0)
}

// DeviceList is a free data retrieval call binding the contract method 0x979ebb42.
//
// Solidity: function deviceList(uint256 ) view returns(address)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) DeviceList(arg0 *big.Int) (common.Address, error) {
	return _IoTDeviceManager.Contract.DeviceList(&_IoTDeviceManager.CallOpts, arg0)
}

// DeviceTransactions is a free data retrieval call binding the contract method 0x98962ad6.
//
// Solidity: function deviceTransactions(string , uint256 ) view returns(string deviceId, bytes data, string messageHash, uint256 sequence, uint256 timestamp, address deviceAddress)
func (_IoTDeviceManager *IoTDeviceManagerCaller) DeviceTransactions(opts *bind.CallOpts, arg0 string, arg1 *big.Int) (struct {
	DeviceId      string
	Data          []byte
	MessageHash   string
	Sequence      *big.Int
	Timestamp     *big.Int
	DeviceAddress common.Address
}, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "deviceTransactions", arg0, arg1)

	outstruct := new(struct {
		DeviceId      string
		Data          []byte
		MessageHash   string
		Sequence      *big.Int
		Timestamp     *big.Int
		DeviceAddress common.Address
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.DeviceId = *abi.ConvertType(out[0], new(string)).(*string)
	outstruct.Data = *abi.ConvertType(out[1], new([]byte)).(*[]byte)
	outstruct.MessageHash = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.Sequence = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.Timestamp = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.DeviceAddress = *abi.ConvertType(out[5], new(common.Address)).(*common.Address)

	return *outstruct, err

}

// DeviceTransactions is a free data retrieval call binding the contract method 0x98962ad6.
//
// Solidity: function deviceTransactions(string , uint256 ) view returns(string deviceId, bytes data, string messageHash, uint256 sequence, uint256 timestamp, address deviceAddress)
func (_IoTDeviceManager *IoTDeviceManagerSession) DeviceTransactions(arg0 string, arg1 *big.Int) (struct {
	DeviceId      string
	Data          []byte
	MessageHash   string
	Sequence      *big.Int
	Timestamp     *big.Int
	DeviceAddress common.Address
}, error) {
	return _IoTDeviceManager.Contract.DeviceTransactions(&_IoTDeviceManager.CallOpts, arg0, arg1)
}

// DeviceTransactions is a free data retrieval call binding the contract method 0x98962ad6.
//
// Solidity: function deviceTransactions(string , uint256 ) view returns(string deviceId, bytes data, string messageHash, uint256 sequence, uint256 timestamp, address deviceAddress)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) DeviceTransactions(arg0 string, arg1 *big.Int) (struct {
	DeviceId      string
	Data          []byte
	MessageHash   string
	Sequence      *big.Int
	Timestamp     *big.Int
	DeviceAddress common.Address
}, error) {
	return _IoTDeviceManager.Contract.DeviceTransactions(&_IoTDeviceManager.CallOpts, arg0, arg1)
}

// Devices is a free data retrieval call binding the contract method 0xe7b4cac6.
//
// Solidity: function devices(address ) view returns(string publicKey, string metadata, uint256 lastSeen, bool isActive, address owner, uint8 entityType)
func (_IoTDeviceManager *IoTDeviceManagerCaller) Devices(opts *bind.CallOpts, arg0 common.Address) (struct {
	PublicKey  string
	Metadata   string
	LastSeen   *big.Int
	IsActive   bool
	Owner      common.Address
	EntityType uint8
}, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "devices", arg0)

	outstruct := new(struct {
		PublicKey  string
		Metadata   string
		LastSeen   *big.Int
		IsActive   bool
		Owner      common.Address
		EntityType uint8
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.PublicKey = *abi.ConvertType(out[0], new(string)).(*string)
	outstruct.Metadata = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.LastSeen = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.IsActive = *abi.ConvertType(out[3], new(bool)).(*bool)
	outstruct.Owner = *abi.ConvertType(out[4], new(common.Address)).(*common.Address)
	outstruct.EntityType = *abi.ConvertType(out[5], new(uint8)).(*uint8)

	return *outstruct, err

}

// Devices is a free data retrieval call binding the contract method 0xe7b4cac6.
//
// Solidity: function devices(address ) view returns(string publicKey, string metadata, uint256 lastSeen, bool isActive, address owner, uint8 entityType)
func (_IoTDeviceManager *IoTDeviceManagerSession) Devices(arg0 common.Address) (struct {
	PublicKey  string
	Metadata   string
	LastSeen   *big.Int
	IsActive   bool
	Owner      common.Address
	EntityType uint8
}, error) {
	return _IoTDeviceManager.Contract.Devices(&_IoTDeviceManager.CallOpts, arg0)
}

// Devices is a free data retrieval call binding the contract method 0xe7b4cac6.
//
// Solidity: function devices(address ) view returns(string publicKey, string metadata, uint256 lastSeen, bool isActive, address owner, uint8 entityType)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) Devices(arg0 common.Address) (struct {
	PublicKey  string
	Metadata   string
	LastSeen   *big.Int
	IsActive   bool
	Owner      common.Address
	EntityType uint8
}, error) {
	return _IoTDeviceManager.Contract.Devices(&_IoTDeviceManager.CallOpts, arg0)
}

// GetAllDevices is a free data retrieval call binding the contract method 0x7a2a530d.
//
// Solidity: function getAllDevices() view returns(address[])
func (_IoTDeviceManager *IoTDeviceManagerCaller) GetAllDevices(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "getAllDevices")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetAllDevices is a free data retrieval call binding the contract method 0x7a2a530d.
//
// Solidity: function getAllDevices() view returns(address[])
func (_IoTDeviceManager *IoTDeviceManagerSession) GetAllDevices() ([]common.Address, error) {
	return _IoTDeviceManager.Contract.GetAllDevices(&_IoTDeviceManager.CallOpts)
}

// GetAllDevices is a free data retrieval call binding the contract method 0x7a2a530d.
//
// Solidity: function getAllDevices() view returns(address[])
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) GetAllDevices() ([]common.Address, error) {
	return _IoTDeviceManager.Contract.GetAllDevices(&_IoTDeviceManager.CallOpts)
}

// GetAuthHistory is a free data retrieval call binding the contract method 0xf320a66a.
//
// Solidity: function getAuthHistory(address deviceAddress) view returns((uint256,bool,bytes32)[])
func (_IoTDeviceManager *IoTDeviceManagerCaller) GetAuthHistory(opts *bind.CallOpts, deviceAddress common.Address) ([]IoTDeviceManagerAuthRecord, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "getAuthHistory", deviceAddress)

	if err != nil {
		return *new([]IoTDeviceManagerAuthRecord), err
	}

	out0 := *abi.ConvertType(out[0], new([]IoTDeviceManagerAuthRecord)).(*[]IoTDeviceManagerAuthRecord)

	return out0, err

}

// GetAuthHistory is a free data retrieval call binding the contract method 0xf320a66a.
//
// Solidity: function getAuthHistory(address deviceAddress) view returns((uint256,bool,bytes32)[])
func (_IoTDeviceManager *IoTDeviceManagerSession) GetAuthHistory(deviceAddress common.Address) ([]IoTDeviceManagerAuthRecord, error) {
	return _IoTDeviceManager.Contract.GetAuthHistory(&_IoTDeviceManager.CallOpts, deviceAddress)
}

// GetAuthHistory is a free data retrieval call binding the contract method 0xf320a66a.
//
// Solidity: function getAuthHistory(address deviceAddress) view returns((uint256,bool,bytes32)[])
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) GetAuthHistory(deviceAddress common.Address) ([]IoTDeviceManagerAuthRecord, error) {
	return _IoTDeviceManager.Contract.GetAuthHistory(&_IoTDeviceManager.CallOpts, deviceAddress)
}

// GetAuthHistorySize is a free data retrieval call binding the contract method 0x0564c34a.
//
// Solidity: function getAuthHistorySize(address deviceAddress) view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerCaller) GetAuthHistorySize(opts *bind.CallOpts, deviceAddress common.Address) (*big.Int, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "getAuthHistorySize", deviceAddress)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAuthHistorySize is a free data retrieval call binding the contract method 0x0564c34a.
//
// Solidity: function getAuthHistorySize(address deviceAddress) view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerSession) GetAuthHistorySize(deviceAddress common.Address) (*big.Int, error) {
	return _IoTDeviceManager.Contract.GetAuthHistorySize(&_IoTDeviceManager.CallOpts, deviceAddress)
}

// GetAuthHistorySize is a free data retrieval call binding the contract method 0x0564c34a.
//
// Solidity: function getAuthHistorySize(address deviceAddress) view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) GetAuthHistorySize(deviceAddress common.Address) (*big.Int, error) {
	return _IoTDeviceManager.Contract.GetAuthHistorySize(&_IoTDeviceManager.CallOpts, deviceAddress)
}

// GetDeviceAuthCount is a free data retrieval call binding the contract method 0x9f3241f3.
//
// Solidity: function getDeviceAuthCount(address deviceAddress) view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerCaller) GetDeviceAuthCount(opts *bind.CallOpts, deviceAddress common.Address) (*big.Int, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "getDeviceAuthCount", deviceAddress)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetDeviceAuthCount is a free data retrieval call binding the contract method 0x9f3241f3.
//
// Solidity: function getDeviceAuthCount(address deviceAddress) view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerSession) GetDeviceAuthCount(deviceAddress common.Address) (*big.Int, error) {
	return _IoTDeviceManager.Contract.GetDeviceAuthCount(&_IoTDeviceManager.CallOpts, deviceAddress)
}

// GetDeviceAuthCount is a free data retrieval call binding the contract method 0x9f3241f3.
//
// Solidity: function getDeviceAuthCount(address deviceAddress) view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) GetDeviceAuthCount(deviceAddress common.Address) (*big.Int, error) {
	return _IoTDeviceManager.Contract.GetDeviceAuthCount(&_IoTDeviceManager.CallOpts, deviceAddress)
}

// GetDeviceInfo is a free data retrieval call binding the contract method 0x4869fe26.
//
// Solidity: function getDeviceInfo(address deviceAddress) view returns(string publicKey, string metadata, bool isActive, uint256 lastSeen, uint8 entityType)
func (_IoTDeviceManager *IoTDeviceManagerCaller) GetDeviceInfo(opts *bind.CallOpts, deviceAddress common.Address) (struct {
	PublicKey  string
	Metadata   string
	IsActive   bool
	LastSeen   *big.Int
	EntityType uint8
}, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "getDeviceInfo", deviceAddress)

	outstruct := new(struct {
		PublicKey  string
		Metadata   string
		IsActive   bool
		LastSeen   *big.Int
		EntityType uint8
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.PublicKey = *abi.ConvertType(out[0], new(string)).(*string)
	outstruct.Metadata = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.IsActive = *abi.ConvertType(out[2], new(bool)).(*bool)
	outstruct.LastSeen = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.EntityType = *abi.ConvertType(out[4], new(uint8)).(*uint8)

	return *outstruct, err

}

// GetDeviceInfo is a free data retrieval call binding the contract method 0x4869fe26.
//
// Solidity: function getDeviceInfo(address deviceAddress) view returns(string publicKey, string metadata, bool isActive, uint256 lastSeen, uint8 entityType)
func (_IoTDeviceManager *IoTDeviceManagerSession) GetDeviceInfo(deviceAddress common.Address) (struct {
	PublicKey  string
	Metadata   string
	IsActive   bool
	LastSeen   *big.Int
	EntityType uint8
}, error) {
	return _IoTDeviceManager.Contract.GetDeviceInfo(&_IoTDeviceManager.CallOpts, deviceAddress)
}

// GetDeviceInfo is a free data retrieval call binding the contract method 0x4869fe26.
//
// Solidity: function getDeviceInfo(address deviceAddress) view returns(string publicKey, string metadata, bool isActive, uint256 lastSeen, uint8 entityType)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) GetDeviceInfo(deviceAddress common.Address) (struct {
	PublicKey  string
	Metadata   string
	IsActive   bool
	LastSeen   *big.Int
	EntityType uint8
}, error) {
	return _IoTDeviceManager.Contract.GetDeviceInfo(&_IoTDeviceManager.CallOpts, deviceAddress)
}

// GetDeviceTransactions is a free data retrieval call binding the contract method 0xc1078635.
//
// Solidity: function getDeviceTransactions(string deviceId, uint256 start, uint256 limit) view returns((string,bytes,string,uint256,uint256,address)[])
func (_IoTDeviceManager *IoTDeviceManagerCaller) GetDeviceTransactions(opts *bind.CallOpts, deviceId string, start *big.Int, limit *big.Int) ([]IoTDeviceManagerTransaction, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "getDeviceTransactions", deviceId, start, limit)

	if err != nil {
		return *new([]IoTDeviceManagerTransaction), err
	}

	out0 := *abi.ConvertType(out[0], new([]IoTDeviceManagerTransaction)).(*[]IoTDeviceManagerTransaction)

	return out0, err

}

// GetDeviceTransactions is a free data retrieval call binding the contract method 0xc1078635.
//
// Solidity: function getDeviceTransactions(string deviceId, uint256 start, uint256 limit) view returns((string,bytes,string,uint256,uint256,address)[])
func (_IoTDeviceManager *IoTDeviceManagerSession) GetDeviceTransactions(deviceId string, start *big.Int, limit *big.Int) ([]IoTDeviceManagerTransaction, error) {
	return _IoTDeviceManager.Contract.GetDeviceTransactions(&_IoTDeviceManager.CallOpts, deviceId, start, limit)
}

// GetDeviceTransactions is a free data retrieval call binding the contract method 0xc1078635.
//
// Solidity: function getDeviceTransactions(string deviceId, uint256 start, uint256 limit) view returns((string,bytes,string,uint256,uint256,address)[])
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) GetDeviceTransactions(deviceId string, start *big.Int, limit *big.Int) ([]IoTDeviceManagerTransaction, error) {
	return _IoTDeviceManager.Contract.GetDeviceTransactions(&_IoTDeviceManager.CallOpts, deviceId, start, limit)
}

// GetDevicesCount is a free data retrieval call binding the contract method 0x6c3a0355.
//
// Solidity: function getDevicesCount() view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerCaller) GetDevicesCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "getDevicesCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetDevicesCount is a free data retrieval call binding the contract method 0x6c3a0355.
//
// Solidity: function getDevicesCount() view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerSession) GetDevicesCount() (*big.Int, error) {
	return _IoTDeviceManager.Contract.GetDevicesCount(&_IoTDeviceManager.CallOpts)
}

// GetDevicesCount is a free data retrieval call binding the contract method 0x6c3a0355.
//
// Solidity: function getDevicesCount() view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) GetDevicesCount() (*big.Int, error) {
	return _IoTDeviceManager.Contract.GetDevicesCount(&_IoTDeviceManager.CallOpts)
}

// GetDevicesPaginated is a free data retrieval call binding the contract method 0xd3bde1ad.
//
// Solidity: function getDevicesPaginated(uint256 start, uint256 limit) view returns(address[])
func (_IoTDeviceManager *IoTDeviceManagerCaller) GetDevicesPaginated(opts *bind.CallOpts, start *big.Int, limit *big.Int) ([]common.Address, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "getDevicesPaginated", start, limit)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetDevicesPaginated is a free data retrieval call binding the contract method 0xd3bde1ad.
//
// Solidity: function getDevicesPaginated(uint256 start, uint256 limit) view returns(address[])
func (_IoTDeviceManager *IoTDeviceManagerSession) GetDevicesPaginated(start *big.Int, limit *big.Int) ([]common.Address, error) {
	return _IoTDeviceManager.Contract.GetDevicesPaginated(&_IoTDeviceManager.CallOpts, start, limit)
}

// GetDevicesPaginated is a free data retrieval call binding the contract method 0xd3bde1ad.
//
// Solidity: function getDevicesPaginated(uint256 start, uint256 limit) view returns(address[])
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) GetDevicesPaginated(start *big.Int, limit *big.Int) ([]common.Address, error) {
	return _IoTDeviceManager.Contract.GetDevicesPaginated(&_IoTDeviceManager.CallOpts, start, limit)
}

// GetEntityType is a free data retrieval call binding the contract method 0x402bfbec.
//
// Solidity: function getEntityType(address deviceAddress) view returns(uint8)
func (_IoTDeviceManager *IoTDeviceManagerCaller) GetEntityType(opts *bind.CallOpts, deviceAddress common.Address) (uint8, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "getEntityType", deviceAddress)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetEntityType is a free data retrieval call binding the contract method 0x402bfbec.
//
// Solidity: function getEntityType(address deviceAddress) view returns(uint8)
func (_IoTDeviceManager *IoTDeviceManagerSession) GetEntityType(deviceAddress common.Address) (uint8, error) {
	return _IoTDeviceManager.Contract.GetEntityType(&_IoTDeviceManager.CallOpts, deviceAddress)
}

// GetEntityType is a free data retrieval call binding the contract method 0x402bfbec.
//
// Solidity: function getEntityType(address deviceAddress) view returns(uint8)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) GetEntityType(deviceAddress common.Address) (uint8, error) {
	return _IoTDeviceManager.Contract.GetEntityType(&_IoTDeviceManager.CallOpts, deviceAddress)
}

// GetGlobalStats is a free data retrieval call binding the contract method 0x6b4169c3.
//
// Solidity: function getGlobalStats() view returns(uint256 _totalDevices, uint256 _totalAuthentications, uint256 _activeDevices, uint256 _inactiveDevices)
func (_IoTDeviceManager *IoTDeviceManagerCaller) GetGlobalStats(opts *bind.CallOpts) (struct {
	TotalDevices         *big.Int
	TotalAuthentications *big.Int
	ActiveDevices        *big.Int
	InactiveDevices      *big.Int
}, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "getGlobalStats")

	outstruct := new(struct {
		TotalDevices         *big.Int
		TotalAuthentications *big.Int
		ActiveDevices        *big.Int
		InactiveDevices      *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.TotalDevices = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.TotalAuthentications = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.ActiveDevices = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.InactiveDevices = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetGlobalStats is a free data retrieval call binding the contract method 0x6b4169c3.
//
// Solidity: function getGlobalStats() view returns(uint256 _totalDevices, uint256 _totalAuthentications, uint256 _activeDevices, uint256 _inactiveDevices)
func (_IoTDeviceManager *IoTDeviceManagerSession) GetGlobalStats() (struct {
	TotalDevices         *big.Int
	TotalAuthentications *big.Int
	ActiveDevices        *big.Int
	InactiveDevices      *big.Int
}, error) {
	return _IoTDeviceManager.Contract.GetGlobalStats(&_IoTDeviceManager.CallOpts)
}

// GetGlobalStats is a free data retrieval call binding the contract method 0x6b4169c3.
//
// Solidity: function getGlobalStats() view returns(uint256 _totalDevices, uint256 _totalAuthentications, uint256 _activeDevices, uint256 _inactiveDevices)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) GetGlobalStats() (struct {
	TotalDevices         *big.Int
	TotalAuthentications *big.Int
	ActiveDevices        *big.Int
	InactiveDevices      *big.Int
}, error) {
	return _IoTDeviceManager.Contract.GetGlobalStats(&_IoTDeviceManager.CallOpts)
}

// GetLatestTransaction is a free data retrieval call binding the contract method 0x168d68fa.
//
// Solidity: function getLatestTransaction(string deviceId) view returns((string,bytes,string,uint256,uint256,address))
func (_IoTDeviceManager *IoTDeviceManagerCaller) GetLatestTransaction(opts *bind.CallOpts, deviceId string) (IoTDeviceManagerTransaction, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "getLatestTransaction", deviceId)

	if err != nil {
		return *new(IoTDeviceManagerTransaction), err
	}

	out0 := *abi.ConvertType(out[0], new(IoTDeviceManagerTransaction)).(*IoTDeviceManagerTransaction)

	return out0, err

}

// GetLatestTransaction is a free data retrieval call binding the contract method 0x168d68fa.
//
// Solidity: function getLatestTransaction(string deviceId) view returns((string,bytes,string,uint256,uint256,address))
func (_IoTDeviceManager *IoTDeviceManagerSession) GetLatestTransaction(deviceId string) (IoTDeviceManagerTransaction, error) {
	return _IoTDeviceManager.Contract.GetLatestTransaction(&_IoTDeviceManager.CallOpts, deviceId)
}

// GetLatestTransaction is a free data retrieval call binding the contract method 0x168d68fa.
//
// Solidity: function getLatestTransaction(string deviceId) view returns((string,bytes,string,uint256,uint256,address))
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) GetLatestTransaction(deviceId string) (IoTDeviceManagerTransaction, error) {
	return _IoTDeviceManager.Contract.GetLatestTransaction(&_IoTDeviceManager.CallOpts, deviceId)
}

// GetRecentAuthHistory is a free data retrieval call binding the contract method 0xd00e9fbd.
//
// Solidity: function getRecentAuthHistory(address deviceAddress, uint256 limit) view returns((uint256,bool,bytes32)[])
func (_IoTDeviceManager *IoTDeviceManagerCaller) GetRecentAuthHistory(opts *bind.CallOpts, deviceAddress common.Address, limit *big.Int) ([]IoTDeviceManagerAuthRecord, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "getRecentAuthHistory", deviceAddress, limit)

	if err != nil {
		return *new([]IoTDeviceManagerAuthRecord), err
	}

	out0 := *abi.ConvertType(out[0], new([]IoTDeviceManagerAuthRecord)).(*[]IoTDeviceManagerAuthRecord)

	return out0, err

}

// GetRecentAuthHistory is a free data retrieval call binding the contract method 0xd00e9fbd.
//
// Solidity: function getRecentAuthHistory(address deviceAddress, uint256 limit) view returns((uint256,bool,bytes32)[])
func (_IoTDeviceManager *IoTDeviceManagerSession) GetRecentAuthHistory(deviceAddress common.Address, limit *big.Int) ([]IoTDeviceManagerAuthRecord, error) {
	return _IoTDeviceManager.Contract.GetRecentAuthHistory(&_IoTDeviceManager.CallOpts, deviceAddress, limit)
}

// GetRecentAuthHistory is a free data retrieval call binding the contract method 0xd00e9fbd.
//
// Solidity: function getRecentAuthHistory(address deviceAddress, uint256 limit) view returns((uint256,bool,bytes32)[])
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) GetRecentAuthHistory(deviceAddress common.Address, limit *big.Int) ([]IoTDeviceManagerAuthRecord, error) {
	return _IoTDeviceManager.Contract.GetRecentAuthHistory(&_IoTDeviceManager.CallOpts, deviceAddress, limit)
}

// GetTransaction is a free data retrieval call binding the contract method 0x12b35c1c.
//
// Solidity: function getTransaction(string deviceId, uint256 index) view returns((string,bytes,string,uint256,uint256,address))
func (_IoTDeviceManager *IoTDeviceManagerCaller) GetTransaction(opts *bind.CallOpts, deviceId string, index *big.Int) (IoTDeviceManagerTransaction, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "getTransaction", deviceId, index)

	if err != nil {
		return *new(IoTDeviceManagerTransaction), err
	}

	out0 := *abi.ConvertType(out[0], new(IoTDeviceManagerTransaction)).(*IoTDeviceManagerTransaction)

	return out0, err

}

// GetTransaction is a free data retrieval call binding the contract method 0x12b35c1c.
//
// Solidity: function getTransaction(string deviceId, uint256 index) view returns((string,bytes,string,uint256,uint256,address))
func (_IoTDeviceManager *IoTDeviceManagerSession) GetTransaction(deviceId string, index *big.Int) (IoTDeviceManagerTransaction, error) {
	return _IoTDeviceManager.Contract.GetTransaction(&_IoTDeviceManager.CallOpts, deviceId, index)
}

// GetTransaction is a free data retrieval call binding the contract method 0x12b35c1c.
//
// Solidity: function getTransaction(string deviceId, uint256 index) view returns((string,bytes,string,uint256,uint256,address))
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) GetTransaction(deviceId string, index *big.Int) (IoTDeviceManagerTransaction, error) {
	return _IoTDeviceManager.Contract.GetTransaction(&_IoTDeviceManager.CallOpts, deviceId, index)
}

// GetTransactionCount is a free data retrieval call binding the contract method 0x5e233b55.
//
// Solidity: function getTransactionCount(string deviceId) view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerCaller) GetTransactionCount(opts *bind.CallOpts, deviceId string) (*big.Int, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "getTransactionCount", deviceId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTransactionCount is a free data retrieval call binding the contract method 0x5e233b55.
//
// Solidity: function getTransactionCount(string deviceId) view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerSession) GetTransactionCount(deviceId string) (*big.Int, error) {
	return _IoTDeviceManager.Contract.GetTransactionCount(&_IoTDeviceManager.CallOpts, deviceId)
}

// GetTransactionCount is a free data retrieval call binding the contract method 0x5e233b55.
//
// Solidity: function getTransactionCount(string deviceId) view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) GetTransactionCount(deviceId string) (*big.Int, error) {
	return _IoTDeviceManager.Contract.GetTransactionCount(&_IoTDeviceManager.CallOpts, deviceId)
}

// IsDeviceListed is a free data retrieval call binding the contract method 0xfb3627a9.
//
// Solidity: function isDeviceListed(address ) view returns(bool)
func (_IoTDeviceManager *IoTDeviceManagerCaller) IsDeviceListed(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "isDeviceListed", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsDeviceListed is a free data retrieval call binding the contract method 0xfb3627a9.
//
// Solidity: function isDeviceListed(address ) view returns(bool)
func (_IoTDeviceManager *IoTDeviceManagerSession) IsDeviceListed(arg0 common.Address) (bool, error) {
	return _IoTDeviceManager.Contract.IsDeviceListed(&_IoTDeviceManager.CallOpts, arg0)
}

// IsDeviceListed is a free data retrieval call binding the contract method 0xfb3627a9.
//
// Solidity: function isDeviceListed(address ) view returns(bool)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) IsDeviceListed(arg0 common.Address) (bool, error) {
	return _IoTDeviceManager.Contract.IsDeviceListed(&_IoTDeviceManager.CallOpts, arg0)
}

// IsDeviceRegistered is a free data retrieval call binding the contract method 0x659228c9.
//
// Solidity: function isDeviceRegistered(address deviceAddress) view returns(bool)
func (_IoTDeviceManager *IoTDeviceManagerCaller) IsDeviceRegistered(opts *bind.CallOpts, deviceAddress common.Address) (bool, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "isDeviceRegistered", deviceAddress)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsDeviceRegistered is a free data retrieval call binding the contract method 0x659228c9.
//
// Solidity: function isDeviceRegistered(address deviceAddress) view returns(bool)
func (_IoTDeviceManager *IoTDeviceManagerSession) IsDeviceRegistered(deviceAddress common.Address) (bool, error) {
	return _IoTDeviceManager.Contract.IsDeviceRegistered(&_IoTDeviceManager.CallOpts, deviceAddress)
}

// IsDeviceRegistered is a free data retrieval call binding the contract method 0x659228c9.
//
// Solidity: function isDeviceRegistered(address deviceAddress) view returns(bool)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) IsDeviceRegistered(deviceAddress common.Address) (bool, error) {
	return _IoTDeviceManager.Contract.IsDeviceRegistered(&_IoTDeviceManager.CallOpts, deviceAddress)
}

// LastSequence is a free data retrieval call binding the contract method 0x4387324a.
//
// Solidity: function lastSequence(string ) view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerCaller) LastSequence(opts *bind.CallOpts, arg0 string) (*big.Int, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "lastSequence", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastSequence is a free data retrieval call binding the contract method 0x4387324a.
//
// Solidity: function lastSequence(string ) view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerSession) LastSequence(arg0 string) (*big.Int, error) {
	return _IoTDeviceManager.Contract.LastSequence(&_IoTDeviceManager.CallOpts, arg0)
}

// LastSequence is a free data retrieval call binding the contract method 0x4387324a.
//
// Solidity: function lastSequence(string ) view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) LastSequence(arg0 string) (*big.Int, error) {
	return _IoTDeviceManager.Contract.LastSequence(&_IoTDeviceManager.CallOpts, arg0)
}

// RecoverSigner is a free data retrieval call binding the contract method 0x97aba7f9.
//
// Solidity: function recoverSigner(bytes32 _hash, bytes _signature) pure returns(address)
func (_IoTDeviceManager *IoTDeviceManagerCaller) RecoverSigner(opts *bind.CallOpts, _hash [32]byte, _signature []byte) (common.Address, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "recoverSigner", _hash, _signature)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// RecoverSigner is a free data retrieval call binding the contract method 0x97aba7f9.
//
// Solidity: function recoverSigner(bytes32 _hash, bytes _signature) pure returns(address)
func (_IoTDeviceManager *IoTDeviceManagerSession) RecoverSigner(_hash [32]byte, _signature []byte) (common.Address, error) {
	return _IoTDeviceManager.Contract.RecoverSigner(&_IoTDeviceManager.CallOpts, _hash, _signature)
}

// RecoverSigner is a free data retrieval call binding the contract method 0x97aba7f9.
//
// Solidity: function recoverSigner(bytes32 _hash, bytes _signature) pure returns(address)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) RecoverSigner(_hash [32]byte, _signature []byte) (common.Address, error) {
	return _IoTDeviceManager.Contract.RecoverSigner(&_IoTDeviceManager.CallOpts, _hash, _signature)
}

// TotalAuthentications is a free data retrieval call binding the contract method 0x1aea3e39.
//
// Solidity: function totalAuthentications() view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerCaller) TotalAuthentications(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "totalAuthentications")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalAuthentications is a free data retrieval call binding the contract method 0x1aea3e39.
//
// Solidity: function totalAuthentications() view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerSession) TotalAuthentications() (*big.Int, error) {
	return _IoTDeviceManager.Contract.TotalAuthentications(&_IoTDeviceManager.CallOpts)
}

// TotalAuthentications is a free data retrieval call binding the contract method 0x1aea3e39.
//
// Solidity: function totalAuthentications() view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) TotalAuthentications() (*big.Int, error) {
	return _IoTDeviceManager.Contract.TotalAuthentications(&_IoTDeviceManager.CallOpts)
}

// TotalDevices is a free data retrieval call binding the contract method 0x6121898e.
//
// Solidity: function totalDevices() view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerCaller) TotalDevices(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "totalDevices")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalDevices is a free data retrieval call binding the contract method 0x6121898e.
//
// Solidity: function totalDevices() view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerSession) TotalDevices() (*big.Int, error) {
	return _IoTDeviceManager.Contract.TotalDevices(&_IoTDeviceManager.CallOpts)
}

// TotalDevices is a free data retrieval call binding the contract method 0x6121898e.
//
// Solidity: function totalDevices() view returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) TotalDevices() (*big.Int, error) {
	return _IoTDeviceManager.Contract.TotalDevices(&_IoTDeviceManager.CallOpts)
}

// UsedSignatures is a free data retrieval call binding the contract method 0xf978fd61.
//
// Solidity: function usedSignatures(bytes32 ) view returns(bool)
func (_IoTDeviceManager *IoTDeviceManagerCaller) UsedSignatures(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "usedSignatures", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// UsedSignatures is a free data retrieval call binding the contract method 0xf978fd61.
//
// Solidity: function usedSignatures(bytes32 ) view returns(bool)
func (_IoTDeviceManager *IoTDeviceManagerSession) UsedSignatures(arg0 [32]byte) (bool, error) {
	return _IoTDeviceManager.Contract.UsedSignatures(&_IoTDeviceManager.CallOpts, arg0)
}

// UsedSignatures is a free data retrieval call binding the contract method 0xf978fd61.
//
// Solidity: function usedSignatures(bytes32 ) view returns(bool)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) UsedSignatures(arg0 [32]byte) (bool, error) {
	return _IoTDeviceManager.Contract.UsedSignatures(&_IoTDeviceManager.CallOpts, arg0)
}

// ActivateDevice is a paid mutator transaction binding the contract method 0x127157c3.
//
// Solidity: function activateDevice(address deviceAddress) returns()
func (_IoTDeviceManager *IoTDeviceManagerTransactor) ActivateDevice(opts *bind.TransactOpts, deviceAddress common.Address) (*types.Transaction, error) {
	return _IoTDeviceManager.contract.Transact(opts, "activateDevice", deviceAddress)
}

// ActivateDevice is a paid mutator transaction binding the contract method 0x127157c3.
//
// Solidity: function activateDevice(address deviceAddress) returns()
func (_IoTDeviceManager *IoTDeviceManagerSession) ActivateDevice(deviceAddress common.Address) (*types.Transaction, error) {
	return _IoTDeviceManager.Contract.ActivateDevice(&_IoTDeviceManager.TransactOpts, deviceAddress)
}

// ActivateDevice is a paid mutator transaction binding the contract method 0x127157c3.
//
// Solidity: function activateDevice(address deviceAddress) returns()
func (_IoTDeviceManager *IoTDeviceManagerTransactorSession) ActivateDevice(deviceAddress common.Address) (*types.Transaction, error) {
	return _IoTDeviceManager.Contract.ActivateDevice(&_IoTDeviceManager.TransactOpts, deviceAddress)
}

// AuthenticateDevice is a paid mutator transaction binding the contract method 0xbba1fd2d.
//
// Solidity: function authenticateDevice(address deviceAddress, bytes32 messageHash, bytes signature) returns(bool)
func (_IoTDeviceManager *IoTDeviceManagerTransactor) AuthenticateDevice(opts *bind.TransactOpts, deviceAddress common.Address, messageHash [32]byte, signature []byte) (*types.Transaction, error) {
	return _IoTDeviceManager.contract.Transact(opts, "authenticateDevice", deviceAddress, messageHash, signature)
}

// AuthenticateDevice is a paid mutator transaction binding the contract method 0xbba1fd2d.
//
// Solidity: function authenticateDevice(address deviceAddress, bytes32 messageHash, bytes signature) returns(bool)
func (_IoTDeviceManager *IoTDeviceManagerSession) AuthenticateDevice(deviceAddress common.Address, messageHash [32]byte, signature []byte) (*types.Transaction, error) {
	return _IoTDeviceManager.Contract.AuthenticateDevice(&_IoTDeviceManager.TransactOpts, deviceAddress, messageHash, signature)
}

// AuthenticateDevice is a paid mutator transaction binding the contract method 0xbba1fd2d.
//
// Solidity: function authenticateDevice(address deviceAddress, bytes32 messageHash, bytes signature) returns(bool)
func (_IoTDeviceManager *IoTDeviceManagerTransactorSession) AuthenticateDevice(deviceAddress common.Address, messageHash [32]byte, signature []byte) (*types.Transaction, error) {
	return _IoTDeviceManager.Contract.AuthenticateDevice(&_IoTDeviceManager.TransactOpts, deviceAddress, messageHash, signature)
}

// DeactivateDevice is a paid mutator transaction binding the contract method 0xcf65952c.
//
// Solidity: function deactivateDevice(address deviceAddress) returns()
func (_IoTDeviceManager *IoTDeviceManagerTransactor) DeactivateDevice(opts *bind.TransactOpts, deviceAddress common.Address) (*types.Transaction, error) {
	return _IoTDeviceManager.contract.Transact(opts, "deactivateDevice", deviceAddress)
}

// DeactivateDevice is a paid mutator transaction binding the contract method 0xcf65952c.
//
// Solidity: function deactivateDevice(address deviceAddress) returns()
func (_IoTDeviceManager *IoTDeviceManagerSession) DeactivateDevice(deviceAddress common.Address) (*types.Transaction, error) {
	return _IoTDeviceManager.Contract.DeactivateDevice(&_IoTDeviceManager.TransactOpts, deviceAddress)
}

// DeactivateDevice is a paid mutator transaction binding the contract method 0xcf65952c.
//
// Solidity: function deactivateDevice(address deviceAddress) returns()
func (_IoTDeviceManager *IoTDeviceManagerTransactorSession) DeactivateDevice(deviceAddress common.Address) (*types.Transaction, error) {
	return _IoTDeviceManager.Contract.DeactivateDevice(&_IoTDeviceManager.TransactOpts, deviceAddress)
}

// RecordAuthentication is a paid mutator transaction binding the contract method 0xf8db1d98.
//
// Solidity: function recordAuthentication(address deviceAddress, bool success, bytes32 proofHash) returns()
func (_IoTDeviceManager *IoTDeviceManagerTransactor) RecordAuthentication(opts *bind.TransactOpts, deviceAddress common.Address, success bool, proofHash [32]byte) (*types.Transaction, error) {
	return _IoTDeviceManager.contract.Transact(opts, "recordAuthentication", deviceAddress, success, proofHash)
}

// RecordAuthentication is a paid mutator transaction binding the contract method 0xf8db1d98.
//
// Solidity: function recordAuthentication(address deviceAddress, bool success, bytes32 proofHash) returns()
func (_IoTDeviceManager *IoTDeviceManagerSession) RecordAuthentication(deviceAddress common.Address, success bool, proofHash [32]byte) (*types.Transaction, error) {
	return _IoTDeviceManager.Contract.RecordAuthentication(&_IoTDeviceManager.TransactOpts, deviceAddress, success, proofHash)
}

// RecordAuthentication is a paid mutator transaction binding the contract method 0xf8db1d98.
//
// Solidity: function recordAuthentication(address deviceAddress, bool success, bytes32 proofHash) returns()
func (_IoTDeviceManager *IoTDeviceManagerTransactorSession) RecordAuthentication(deviceAddress common.Address, success bool, proofHash [32]byte) (*types.Transaction, error) {
	return _IoTDeviceManager.Contract.RecordAuthentication(&_IoTDeviceManager.TransactOpts, deviceAddress, success, proofHash)
}

// RecordTransaction is a paid mutator transaction binding the contract method 0xe62dc669.
//
// Solidity: function recordTransaction(address deviceAddress, string deviceId, bytes data, string messageHash, uint256 sequence, uint256 timestamp) returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerTransactor) RecordTransaction(opts *bind.TransactOpts, deviceAddress common.Address, deviceId string, data []byte, messageHash string, sequence *big.Int, timestamp *big.Int) (*types.Transaction, error) {
	return _IoTDeviceManager.contract.Transact(opts, "recordTransaction", deviceAddress, deviceId, data, messageHash, sequence, timestamp)
}

// RecordTransaction is a paid mutator transaction binding the contract method 0xe62dc669.
//
// Solidity: function recordTransaction(address deviceAddress, string deviceId, bytes data, string messageHash, uint256 sequence, uint256 timestamp) returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerSession) RecordTransaction(deviceAddress common.Address, deviceId string, data []byte, messageHash string, sequence *big.Int, timestamp *big.Int) (*types.Transaction, error) {
	return _IoTDeviceManager.Contract.RecordTransaction(&_IoTDeviceManager.TransactOpts, deviceAddress, deviceId, data, messageHash, sequence, timestamp)
}

// RecordTransaction is a paid mutator transaction binding the contract method 0xe62dc669.
//
// Solidity: function recordTransaction(address deviceAddress, string deviceId, bytes data, string messageHash, uint256 sequence, uint256 timestamp) returns(uint256)
func (_IoTDeviceManager *IoTDeviceManagerTransactorSession) RecordTransaction(deviceAddress common.Address, deviceId string, data []byte, messageHash string, sequence *big.Int, timestamp *big.Int) (*types.Transaction, error) {
	return _IoTDeviceManager.Contract.RecordTransaction(&_IoTDeviceManager.TransactOpts, deviceAddress, deviceId, data, messageHash, sequence, timestamp)
}

// RegisterDevice is a paid mutator transaction binding the contract method 0xbc717f9a.
//
// Solidity: function registerDevice(address deviceAddress, string _publicKey, string _metadata, uint8 _entityType) returns()
func (_IoTDeviceManager *IoTDeviceManagerTransactor) RegisterDevice(opts *bind.TransactOpts, deviceAddress common.Address, _publicKey string, _metadata string, _entityType uint8) (*types.Transaction, error) {
	return _IoTDeviceManager.contract.Transact(opts, "registerDevice", deviceAddress, _publicKey, _metadata, _entityType)
}

// RegisterDevice is a paid mutator transaction binding the contract method 0xbc717f9a.
//
// Solidity: function registerDevice(address deviceAddress, string _publicKey, string _metadata, uint8 _entityType) returns()
func (_IoTDeviceManager *IoTDeviceManagerSession) RegisterDevice(deviceAddress common.Address, _publicKey string, _metadata string, _entityType uint8) (*types.Transaction, error) {
	return _IoTDeviceManager.Contract.RegisterDevice(&_IoTDeviceManager.TransactOpts, deviceAddress, _publicKey, _metadata, _entityType)
}

// RegisterDevice is a paid mutator transaction binding the contract method 0xbc717f9a.
//
// Solidity: function registerDevice(address deviceAddress, string _publicKey, string _metadata, uint8 _entityType) returns()
func (_IoTDeviceManager *IoTDeviceManagerTransactorSession) RegisterDevice(deviceAddress common.Address, _publicKey string, _metadata string, _entityType uint8) (*types.Transaction, error) {
	return _IoTDeviceManager.Contract.RegisterDevice(&_IoTDeviceManager.TransactOpts, deviceAddress, _publicKey, _metadata, _entityType)
}

// IoTDeviceManagerAuthenticationRecordedIterator is returned from FilterAuthenticationRecorded and is used to iterate over the raw logs and unpacked data for AuthenticationRecorded events raised by the IoTDeviceManager contract.
type IoTDeviceManagerAuthenticationRecordedIterator struct {
	Event *IoTDeviceManagerAuthenticationRecorded // Event containing the contract specifics and raw log

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
func (it *IoTDeviceManagerAuthenticationRecordedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IoTDeviceManagerAuthenticationRecorded)
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
		it.Event = new(IoTDeviceManagerAuthenticationRecorded)
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
func (it *IoTDeviceManagerAuthenticationRecordedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IoTDeviceManagerAuthenticationRecordedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IoTDeviceManagerAuthenticationRecorded represents a AuthenticationRecorded event raised by the IoTDeviceManager contract.
type IoTDeviceManagerAuthenticationRecorded struct {
	DeviceAddress common.Address
	Timestamp     *big.Int
	Success       bool
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterAuthenticationRecorded is a free log retrieval operation binding the contract event 0x479914d0568c4a3927420aa15b2f8838a29126342e32c563637c6a993bba692e.
//
// Solidity: event AuthenticationRecorded(address indexed deviceAddress, uint256 timestamp, bool success)
func (_IoTDeviceManager *IoTDeviceManagerFilterer) FilterAuthenticationRecorded(opts *bind.FilterOpts, deviceAddress []common.Address) (*IoTDeviceManagerAuthenticationRecordedIterator, error) {

	var deviceAddressRule []interface{}
	for _, deviceAddressItem := range deviceAddress {
		deviceAddressRule = append(deviceAddressRule, deviceAddressItem)
	}

	logs, sub, err := _IoTDeviceManager.contract.FilterLogs(opts, "AuthenticationRecorded", deviceAddressRule)
	if err != nil {
		return nil, err
	}
	return &IoTDeviceManagerAuthenticationRecordedIterator{contract: _IoTDeviceManager.contract, event: "AuthenticationRecorded", logs: logs, sub: sub}, nil
}

// WatchAuthenticationRecorded is a free log subscription operation binding the contract event 0x479914d0568c4a3927420aa15b2f8838a29126342e32c563637c6a993bba692e.
//
// Solidity: event AuthenticationRecorded(address indexed deviceAddress, uint256 timestamp, bool success)
func (_IoTDeviceManager *IoTDeviceManagerFilterer) WatchAuthenticationRecorded(opts *bind.WatchOpts, sink chan<- *IoTDeviceManagerAuthenticationRecorded, deviceAddress []common.Address) (event.Subscription, error) {

	var deviceAddressRule []interface{}
	for _, deviceAddressItem := range deviceAddress {
		deviceAddressRule = append(deviceAddressRule, deviceAddressItem)
	}

	logs, sub, err := _IoTDeviceManager.contract.WatchLogs(opts, "AuthenticationRecorded", deviceAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IoTDeviceManagerAuthenticationRecorded)
				if err := _IoTDeviceManager.contract.UnpackLog(event, "AuthenticationRecorded", log); err != nil {
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

// ParseAuthenticationRecorded is a log parse operation binding the contract event 0x479914d0568c4a3927420aa15b2f8838a29126342e32c563637c6a993bba692e.
//
// Solidity: event AuthenticationRecorded(address indexed deviceAddress, uint256 timestamp, bool success)
func (_IoTDeviceManager *IoTDeviceManagerFilterer) ParseAuthenticationRecorded(log types.Log) (*IoTDeviceManagerAuthenticationRecorded, error) {
	event := new(IoTDeviceManagerAuthenticationRecorded)
	if err := _IoTDeviceManager.contract.UnpackLog(event, "AuthenticationRecorded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IoTDeviceManagerDeviceActivatedIterator is returned from FilterDeviceActivated and is used to iterate over the raw logs and unpacked data for DeviceActivated events raised by the IoTDeviceManager contract.
type IoTDeviceManagerDeviceActivatedIterator struct {
	Event *IoTDeviceManagerDeviceActivated // Event containing the contract specifics and raw log

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
func (it *IoTDeviceManagerDeviceActivatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IoTDeviceManagerDeviceActivated)
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
		it.Event = new(IoTDeviceManagerDeviceActivated)
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
func (it *IoTDeviceManagerDeviceActivatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IoTDeviceManagerDeviceActivatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IoTDeviceManagerDeviceActivated represents a DeviceActivated event raised by the IoTDeviceManager contract.
type IoTDeviceManagerDeviceActivated struct {
	DeviceAddress common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterDeviceActivated is a free log retrieval operation binding the contract event 0x6e6ce1da269ab1f571a737cdc56e1d5ffad073ff99aa5925df6d9d972fb2b3d4.
//
// Solidity: event DeviceActivated(address indexed deviceAddress)
func (_IoTDeviceManager *IoTDeviceManagerFilterer) FilterDeviceActivated(opts *bind.FilterOpts, deviceAddress []common.Address) (*IoTDeviceManagerDeviceActivatedIterator, error) {

	var deviceAddressRule []interface{}
	for _, deviceAddressItem := range deviceAddress {
		deviceAddressRule = append(deviceAddressRule, deviceAddressItem)
	}

	logs, sub, err := _IoTDeviceManager.contract.FilterLogs(opts, "DeviceActivated", deviceAddressRule)
	if err != nil {
		return nil, err
	}
	return &IoTDeviceManagerDeviceActivatedIterator{contract: _IoTDeviceManager.contract, event: "DeviceActivated", logs: logs, sub: sub}, nil
}

// WatchDeviceActivated is a free log subscription operation binding the contract event 0x6e6ce1da269ab1f571a737cdc56e1d5ffad073ff99aa5925df6d9d972fb2b3d4.
//
// Solidity: event DeviceActivated(address indexed deviceAddress)
func (_IoTDeviceManager *IoTDeviceManagerFilterer) WatchDeviceActivated(opts *bind.WatchOpts, sink chan<- *IoTDeviceManagerDeviceActivated, deviceAddress []common.Address) (event.Subscription, error) {

	var deviceAddressRule []interface{}
	for _, deviceAddressItem := range deviceAddress {
		deviceAddressRule = append(deviceAddressRule, deviceAddressItem)
	}

	logs, sub, err := _IoTDeviceManager.contract.WatchLogs(opts, "DeviceActivated", deviceAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IoTDeviceManagerDeviceActivated)
				if err := _IoTDeviceManager.contract.UnpackLog(event, "DeviceActivated", log); err != nil {
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

// ParseDeviceActivated is a log parse operation binding the contract event 0x6e6ce1da269ab1f571a737cdc56e1d5ffad073ff99aa5925df6d9d972fb2b3d4.
//
// Solidity: event DeviceActivated(address indexed deviceAddress)
func (_IoTDeviceManager *IoTDeviceManagerFilterer) ParseDeviceActivated(log types.Log) (*IoTDeviceManagerDeviceActivated, error) {
	event := new(IoTDeviceManagerDeviceActivated)
	if err := _IoTDeviceManager.contract.UnpackLog(event, "DeviceActivated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IoTDeviceManagerDeviceAuthenticatedIterator is returned from FilterDeviceAuthenticated and is used to iterate over the raw logs and unpacked data for DeviceAuthenticated events raised by the IoTDeviceManager contract.
type IoTDeviceManagerDeviceAuthenticatedIterator struct {
	Event *IoTDeviceManagerDeviceAuthenticated // Event containing the contract specifics and raw log

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
func (it *IoTDeviceManagerDeviceAuthenticatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IoTDeviceManagerDeviceAuthenticated)
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
		it.Event = new(IoTDeviceManagerDeviceAuthenticated)
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
func (it *IoTDeviceManagerDeviceAuthenticatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IoTDeviceManagerDeviceAuthenticatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IoTDeviceManagerDeviceAuthenticated represents a DeviceAuthenticated event raised by the IoTDeviceManager contract.
type IoTDeviceManagerDeviceAuthenticated struct {
	DeviceAddress common.Address
	Timestamp     *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterDeviceAuthenticated is a free log retrieval operation binding the contract event 0x5caa76bbd236e4b7e7bd9a888c8b726858221bac425e0fa9e9f2a574c1860f86.
//
// Solidity: event DeviceAuthenticated(address indexed deviceAddress, uint256 timestamp)
func (_IoTDeviceManager *IoTDeviceManagerFilterer) FilterDeviceAuthenticated(opts *bind.FilterOpts, deviceAddress []common.Address) (*IoTDeviceManagerDeviceAuthenticatedIterator, error) {

	var deviceAddressRule []interface{}
	for _, deviceAddressItem := range deviceAddress {
		deviceAddressRule = append(deviceAddressRule, deviceAddressItem)
	}

	logs, sub, err := _IoTDeviceManager.contract.FilterLogs(opts, "DeviceAuthenticated", deviceAddressRule)
	if err != nil {
		return nil, err
	}
	return &IoTDeviceManagerDeviceAuthenticatedIterator{contract: _IoTDeviceManager.contract, event: "DeviceAuthenticated", logs: logs, sub: sub}, nil
}

// WatchDeviceAuthenticated is a free log subscription operation binding the contract event 0x5caa76bbd236e4b7e7bd9a888c8b726858221bac425e0fa9e9f2a574c1860f86.
//
// Solidity: event DeviceAuthenticated(address indexed deviceAddress, uint256 timestamp)
func (_IoTDeviceManager *IoTDeviceManagerFilterer) WatchDeviceAuthenticated(opts *bind.WatchOpts, sink chan<- *IoTDeviceManagerDeviceAuthenticated, deviceAddress []common.Address) (event.Subscription, error) {

	var deviceAddressRule []interface{}
	for _, deviceAddressItem := range deviceAddress {
		deviceAddressRule = append(deviceAddressRule, deviceAddressItem)
	}

	logs, sub, err := _IoTDeviceManager.contract.WatchLogs(opts, "DeviceAuthenticated", deviceAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IoTDeviceManagerDeviceAuthenticated)
				if err := _IoTDeviceManager.contract.UnpackLog(event, "DeviceAuthenticated", log); err != nil {
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

// ParseDeviceAuthenticated is a log parse operation binding the contract event 0x5caa76bbd236e4b7e7bd9a888c8b726858221bac425e0fa9e9f2a574c1860f86.
//
// Solidity: event DeviceAuthenticated(address indexed deviceAddress, uint256 timestamp)
func (_IoTDeviceManager *IoTDeviceManagerFilterer) ParseDeviceAuthenticated(log types.Log) (*IoTDeviceManagerDeviceAuthenticated, error) {
	event := new(IoTDeviceManagerDeviceAuthenticated)
	if err := _IoTDeviceManager.contract.UnpackLog(event, "DeviceAuthenticated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IoTDeviceManagerDeviceDeactivatedIterator is returned from FilterDeviceDeactivated and is used to iterate over the raw logs and unpacked data for DeviceDeactivated events raised by the IoTDeviceManager contract.
type IoTDeviceManagerDeviceDeactivatedIterator struct {
	Event *IoTDeviceManagerDeviceDeactivated // Event containing the contract specifics and raw log

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
func (it *IoTDeviceManagerDeviceDeactivatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IoTDeviceManagerDeviceDeactivated)
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
		it.Event = new(IoTDeviceManagerDeviceDeactivated)
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
func (it *IoTDeviceManagerDeviceDeactivatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IoTDeviceManagerDeviceDeactivatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IoTDeviceManagerDeviceDeactivated represents a DeviceDeactivated event raised by the IoTDeviceManager contract.
type IoTDeviceManagerDeviceDeactivated struct {
	DeviceAddress common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterDeviceDeactivated is a free log retrieval operation binding the contract event 0xca2220ee93ea32595f56c66fee84698b8ff679dad662cde58c0e54d54e5af983.
//
// Solidity: event DeviceDeactivated(address indexed deviceAddress)
func (_IoTDeviceManager *IoTDeviceManagerFilterer) FilterDeviceDeactivated(opts *bind.FilterOpts, deviceAddress []common.Address) (*IoTDeviceManagerDeviceDeactivatedIterator, error) {

	var deviceAddressRule []interface{}
	for _, deviceAddressItem := range deviceAddress {
		deviceAddressRule = append(deviceAddressRule, deviceAddressItem)
	}

	logs, sub, err := _IoTDeviceManager.contract.FilterLogs(opts, "DeviceDeactivated", deviceAddressRule)
	if err != nil {
		return nil, err
	}
	return &IoTDeviceManagerDeviceDeactivatedIterator{contract: _IoTDeviceManager.contract, event: "DeviceDeactivated", logs: logs, sub: sub}, nil
}

// WatchDeviceDeactivated is a free log subscription operation binding the contract event 0xca2220ee93ea32595f56c66fee84698b8ff679dad662cde58c0e54d54e5af983.
//
// Solidity: event DeviceDeactivated(address indexed deviceAddress)
func (_IoTDeviceManager *IoTDeviceManagerFilterer) WatchDeviceDeactivated(opts *bind.WatchOpts, sink chan<- *IoTDeviceManagerDeviceDeactivated, deviceAddress []common.Address) (event.Subscription, error) {

	var deviceAddressRule []interface{}
	for _, deviceAddressItem := range deviceAddress {
		deviceAddressRule = append(deviceAddressRule, deviceAddressItem)
	}

	logs, sub, err := _IoTDeviceManager.contract.WatchLogs(opts, "DeviceDeactivated", deviceAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IoTDeviceManagerDeviceDeactivated)
				if err := _IoTDeviceManager.contract.UnpackLog(event, "DeviceDeactivated", log); err != nil {
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

// ParseDeviceDeactivated is a log parse operation binding the contract event 0xca2220ee93ea32595f56c66fee84698b8ff679dad662cde58c0e54d54e5af983.
//
// Solidity: event DeviceDeactivated(address indexed deviceAddress)
func (_IoTDeviceManager *IoTDeviceManagerFilterer) ParseDeviceDeactivated(log types.Log) (*IoTDeviceManagerDeviceDeactivated, error) {
	event := new(IoTDeviceManagerDeviceDeactivated)
	if err := _IoTDeviceManager.contract.UnpackLog(event, "DeviceDeactivated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IoTDeviceManagerDeviceRegisteredIterator is returned from FilterDeviceRegistered and is used to iterate over the raw logs and unpacked data for DeviceRegistered events raised by the IoTDeviceManager contract.
type IoTDeviceManagerDeviceRegisteredIterator struct {
	Event *IoTDeviceManagerDeviceRegistered // Event containing the contract specifics and raw log

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
func (it *IoTDeviceManagerDeviceRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IoTDeviceManagerDeviceRegistered)
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
		it.Event = new(IoTDeviceManagerDeviceRegistered)
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
func (it *IoTDeviceManagerDeviceRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IoTDeviceManagerDeviceRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IoTDeviceManagerDeviceRegistered represents a DeviceRegistered event raised by the IoTDeviceManager contract.
type IoTDeviceManagerDeviceRegistered struct {
	DeviceAddress common.Address
	PublicKey     string
	EntityType    uint8
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterDeviceRegistered is a free log retrieval operation binding the contract event 0xadd6077226737886880132c2643b33a3812038937aafffa2b00f9ef0f3c0b29c.
//
// Solidity: event DeviceRegistered(address indexed deviceAddress, string publicKey, uint8 entityType)
func (_IoTDeviceManager *IoTDeviceManagerFilterer) FilterDeviceRegistered(opts *bind.FilterOpts, deviceAddress []common.Address) (*IoTDeviceManagerDeviceRegisteredIterator, error) {

	var deviceAddressRule []interface{}
	for _, deviceAddressItem := range deviceAddress {
		deviceAddressRule = append(deviceAddressRule, deviceAddressItem)
	}

	logs, sub, err := _IoTDeviceManager.contract.FilterLogs(opts, "DeviceRegistered", deviceAddressRule)
	if err != nil {
		return nil, err
	}
	return &IoTDeviceManagerDeviceRegisteredIterator{contract: _IoTDeviceManager.contract, event: "DeviceRegistered", logs: logs, sub: sub}, nil
}

// WatchDeviceRegistered is a free log subscription operation binding the contract event 0xadd6077226737886880132c2643b33a3812038937aafffa2b00f9ef0f3c0b29c.
//
// Solidity: event DeviceRegistered(address indexed deviceAddress, string publicKey, uint8 entityType)
func (_IoTDeviceManager *IoTDeviceManagerFilterer) WatchDeviceRegistered(opts *bind.WatchOpts, sink chan<- *IoTDeviceManagerDeviceRegistered, deviceAddress []common.Address) (event.Subscription, error) {

	var deviceAddressRule []interface{}
	for _, deviceAddressItem := range deviceAddress {
		deviceAddressRule = append(deviceAddressRule, deviceAddressItem)
	}

	logs, sub, err := _IoTDeviceManager.contract.WatchLogs(opts, "DeviceRegistered", deviceAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IoTDeviceManagerDeviceRegistered)
				if err := _IoTDeviceManager.contract.UnpackLog(event, "DeviceRegistered", log); err != nil {
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

// ParseDeviceRegistered is a log parse operation binding the contract event 0xadd6077226737886880132c2643b33a3812038937aafffa2b00f9ef0f3c0b29c.
//
// Solidity: event DeviceRegistered(address indexed deviceAddress, string publicKey, uint8 entityType)
func (_IoTDeviceManager *IoTDeviceManagerFilterer) ParseDeviceRegistered(log types.Log) (*IoTDeviceManagerDeviceRegistered, error) {
	event := new(IoTDeviceManagerDeviceRegistered)
	if err := _IoTDeviceManager.contract.UnpackLog(event, "DeviceRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IoTDeviceManagerTransactionRecordedIterator is returned from FilterTransactionRecorded and is used to iterate over the raw logs and unpacked data for TransactionRecorded events raised by the IoTDeviceManager contract.
type IoTDeviceManagerTransactionRecordedIterator struct {
	Event *IoTDeviceManagerTransactionRecorded // Event containing the contract specifics and raw log

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
func (it *IoTDeviceManagerTransactionRecordedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IoTDeviceManagerTransactionRecorded)
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
		it.Event = new(IoTDeviceManagerTransactionRecorded)
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
func (it *IoTDeviceManagerTransactionRecordedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IoTDeviceManagerTransactionRecordedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IoTDeviceManagerTransactionRecorded represents a TransactionRecorded event raised by the IoTDeviceManager contract.
type IoTDeviceManagerTransactionRecorded struct {
	DeviceAddress common.Address
	DeviceId      string
	Sequence      *big.Int
	Timestamp     *big.Int
	BlockNumber   *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterTransactionRecorded is a free log retrieval operation binding the contract event 0x8aebf80de26fd1cfce16f5e9a05341326492e77136df3565d8da753e57ce1f04.
//
// Solidity: event TransactionRecorded(address indexed deviceAddress, string deviceId, uint256 sequence, uint256 timestamp, uint256 blockNumber)
func (_IoTDeviceManager *IoTDeviceManagerFilterer) FilterTransactionRecorded(opts *bind.FilterOpts, deviceAddress []common.Address) (*IoTDeviceManagerTransactionRecordedIterator, error) {

	var deviceAddressRule []interface{}
	for _, deviceAddressItem := range deviceAddress {
		deviceAddressRule = append(deviceAddressRule, deviceAddressItem)
	}

	logs, sub, err := _IoTDeviceManager.contract.FilterLogs(opts, "TransactionRecorded", deviceAddressRule)
	if err != nil {
		return nil, err
	}
	return &IoTDeviceManagerTransactionRecordedIterator{contract: _IoTDeviceManager.contract, event: "TransactionRecorded", logs: logs, sub: sub}, nil
}

// WatchTransactionRecorded is a free log subscription operation binding the contract event 0x8aebf80de26fd1cfce16f5e9a05341326492e77136df3565d8da753e57ce1f04.
//
// Solidity: event TransactionRecorded(address indexed deviceAddress, string deviceId, uint256 sequence, uint256 timestamp, uint256 blockNumber)
func (_IoTDeviceManager *IoTDeviceManagerFilterer) WatchTransactionRecorded(opts *bind.WatchOpts, sink chan<- *IoTDeviceManagerTransactionRecorded, deviceAddress []common.Address) (event.Subscription, error) {

	var deviceAddressRule []interface{}
	for _, deviceAddressItem := range deviceAddress {
		deviceAddressRule = append(deviceAddressRule, deviceAddressItem)
	}

	logs, sub, err := _IoTDeviceManager.contract.WatchLogs(opts, "TransactionRecorded", deviceAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IoTDeviceManagerTransactionRecorded)
				if err := _IoTDeviceManager.contract.UnpackLog(event, "TransactionRecorded", log); err != nil {
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

// ParseTransactionRecorded is a log parse operation binding the contract event 0x8aebf80de26fd1cfce16f5e9a05341326492e77136df3565d8da753e57ce1f04.
//
// Solidity: event TransactionRecorded(address indexed deviceAddress, string deviceId, uint256 sequence, uint256 timestamp, uint256 blockNumber)
func (_IoTDeviceManager *IoTDeviceManagerFilterer) ParseTransactionRecorded(log types.Log) (*IoTDeviceManagerTransactionRecorded, error) {
	event := new(IoTDeviceManagerTransactionRecorded)
	if err := _IoTDeviceManager.contract.UnpackLog(event, "TransactionRecorded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
