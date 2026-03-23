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

// IoTDeviceManagerMetaData contains all meta data concerning the IoTDeviceManager contract.
var IoTDeviceManagerMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"}],\"name\":\"AuthenticationRecorded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"DeviceAuthenticated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"}],\"name\":\"DeviceDeactivated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"publicKey\",\"type\":\"string\"}],\"name\":\"DeviceRegistered\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"}],\"name\":\"activateDevice\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"activeDevicesCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"authCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"authHistory\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes32\",\"name\":\"proofHash\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"messageHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"authenticateDevice\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"}],\"name\":\"deactivateDevice\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"deviceList\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"devices\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"publicKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"lastSeen\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isActive\",\"type\":\"bool\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAllDevices\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"}],\"name\":\"getAuthHistory\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes32\",\"name\":\"proofHash\",\"type\":\"bytes32\"}],\"internalType\":\"structIoTDeviceManager.AuthRecord[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"}],\"name\":\"getDeviceAuthCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"}],\"name\":\"getDeviceInfo\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getDevicesCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"start\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit\",\"type\":\"uint256\"}],\"name\":\"getDevicesPaginated\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getGlobalStats\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"_totalDevices\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_totalAuthentications\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_activeDevices\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_inactiveDevices\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"limit\",\"type\":\"uint256\"}],\"name\":\"getRecentAuthHistory\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes32\",\"name\":\"proofHash\",\"type\":\"bytes32\"}],\"internalType\":\"structIoTDeviceManager.AuthRecord[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"isDeviceListed\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"}],\"name\":\"isDeviceRegistered\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes32\",\"name\":\"proofHash\",\"type\":\"bytes32\"}],\"name\":\"recordAuthentication\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_hash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"_signature\",\"type\":\"bytes\"}],\"name\":\"recoverSigner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"deviceAddress\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"_publicKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_metadata\",\"type\":\"string\"}],\"name\":\"registerDevice\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalAuthentications\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalDevices\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"usedSignatures\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b50611bd7806100206000396000f3fe608060405234801561001057600080fd5b50600436106101585760003560e01c80639f3241f3116100c3578063d3bde1ad1161007c578063d3bde1ad146102ec578063e7b4cac6146102ff578063f320a66a14610323578063f8db1d9814610336578063f978fd6114610349578063fb3627a91461035c57610158565b80639f3241f31461025e578063a17eec9f14610271578063b879fee714610284578063bba1fd2d146102a6578063cf65952c146102b9578063d00e9fbd146102cc57610158565b80636b4169c3116101155780636b4169c3146101e35780636c3a0355146101fb5780636f6dfeb0146102035780637a2a530d14610216578063979ebb421461022b57806397aba7f91461024b57610158565b8063127157c31461015d5780631aea3e39146101725780634869fe26146101905780634df97928146101b35780636121898e146101bb578063659228c9146101c3575b600080fd5b61017061016b36600461153b565b61036f565b005b61017a610470565b6040516101879190611a70565b60405180910390f35b6101a361019e36600461153b565b610476565b6040516101879493929190611844565b61017a6105d7565b61017a6105dd565b6101d66101d136600461153b565b6105e3565b6040516101879190611808565b6101eb610607565b6040516101879493929190611a9f565b61017a610634565b6101706102113660046115f3565b61063a565b61021e610811565b6040516101879190611760565b61023e610239366004611683565b610873565b604051610187919061174c565b61023e61025936600461169b565b61089d565b61017a61026c36600461153b565b610943565b61017a61027f36600461153b565b61095e565b61029761029236600461165a565b610970565b60405161018793929190611a89565b6101d66102b436600461159e565b6109b6565b6101706102c736600461153b565b610c53565b6102df6102da36600461165a565b610d7b565b60405161018791906117ad565b61021e6102fa3660046116e0565b610f3c565b61031261030d36600461153b565b611088565b60405161018795949392919061187d565b6102df61033136600461153b565b6111d3565b61017061034436600461155c565b6112a1565b6101d6610357366004611683565b6113c0565b6101d661036a36600461153b565b6113d5565b6001600160a01b03811660009081526020819052604090206002015481906103b25760405162461bcd60e51b81526004016103a9906119e3565b60405180910390fd5b6001600160a01b0380831660009081526020819052604090206003015483913361010090920416146103f65760405162461bcd60e51b81526004016103a990611986565b6001600160a01b03831660009081526020819052604090206003015460ff16156104325760405162461bcd60e51b81526004016103a990611900565b6001600160a01b0383166000908152602081905260408120600301805460ff19166001179055600880549161046683611b5a565b9190505550505050565b60075481565b6001600160a01b038116600090815260208190526040812060038101546002820154825460609485949093849391928392600184019260ff169184906104bb90611b25565b80601f01602080910402602001604051908101604052809291908181526020018280546104e790611b25565b80156105345780601f1061050957610100808354040283529160200191610534565b820191906000526020600020905b81548152906001019060200180831161051757829003601f168201915b5050505050935082805461054790611b25565b80601f016020809104026020016040519081016040528092919081815260200182805461057390611b25565b80156105c05780601f10610595576101008083540402835291602001916105c0565b820191906000526020600020905b8154815290600101906020018083116105a357829003601f168201915b505050505092509450945094509450509193509193565b60085481565b60065481565b6001600160a01b03811660009081526020819052604090206002015415155b919050565b6000806000806006546007546008546008546006546106269190611af7565b935093509350935090919293565b60045490565b6001600160a01b038316600090815260208190526040902060020154156106735760405162461bcd60e51b81526004016103a9906119ac565b6040805160a08101825283815260208082018490524282840152600160608301523360808301526001600160a01b03861660009081528082529290922081518051929391926106c592849201906113ea565b5060208281015180516106de92600185019201906113ea565b50604082810151600283015560608301516003909201805460809094015160ff1990941692151592909217610100600160a81b0319166101006001600160a01b03948516021790915590841660009081526005602052205460ff166107a1576004805460018082019092557f8a35acfbc15ff81a39ae7d344fd709f28e8600b4aa8c65c6b64bfe7fe36bd19b0180546001600160a01b0319166001600160a01b0386169081179091556000908152600560205260409020805460ff191690911790555b600680549060006107b183611b5a565b9091555050600880549060006107c683611b5a565b9190505550826001600160a01b03167f40ee62d3a0cf66ef791a7a1f9d6d563ec877b394b8f5c097b6238ac82a68c0e2836040516108049190611831565b60405180910390a2505050565b6060600480548060200260200160405190810160405280929190818152602001828054801561086957602002820191906000526020600020905b81546001600160a01b0316815260019091019060200180831161084b575b5050505050905090565b6004818154811061088357600080fd5b6000918252602090912001546001600160a01b0316905081565b600081516041146108c05760405162461bcd60e51b81526004016103a990611a0d565b60208201516040830151606084015160001a601b8110156108e9576108e6601b82611ad2565b90505b6001868285856040516000815260200160405260405161090c9493929190611813565b6020604051602081039080840390855afa15801561092e573d6000803e3d6000fd5b5050604051601f190151979650505050505050565b6001600160a01b031660009081526003602052604090205490565b60036020526000908152604090205481565b6001602052816000526040600020818154811061098c57600080fd5b600091825260209091206003909102018054600182015460029092015490935060ff909116915083565b6001600160a01b03831660009081526020819052604081206002015484906109f05760405162461bcd60e51b81526004016103a9906119e3565b6001600160a01b0385166000908152602081905260409020600381015460ff16610a2c5760405162461bcd60e51b81526004016103a99061192d565b60008581526002602052604090205460ff1615610a5b5760405162461bcd60e51b81526004016103a9906118c9565b6000610a67868661089d565b9050866001600160a01b0316816001600160a01b031614610a9a5760405162461bcd60e51b81526004016103a990611a0d565b60016002600088815260200190815260200160002060006101000a81548160ff02191690831515021790555042826002018190555060016000886001600160a01b03166001600160a01b0316815260200190815260200160002060405180606001604052804281526020016001151581526020018881525090806001815401808255809150506001900390600052602060002090600302016000909190919091506000820151816000015560208201518160010160006101000a81548160ff02191690831515021790555060408201518160020155505060036000886001600160a01b03166001600160a01b031681526020019081526020016000206000815480929190610ba790611b5a565b909155505060078054906000610bbc83611b5a565b9190505550866001600160a01b03167f5caa76bbd236e4b7e7bd9a888c8b726858221bac425e0fa9e9f2a574c1860f8642604051610bfa9190611a70565b60405180910390a2866001600160a01b03167f479914d0568c4a3927420aa15b2f8838a29126342e32c563637c6a993bba692e426001604051610c3e929190611a79565b60405180910390a25060019695505050505050565b6001600160a01b0381166000908152602081905260409020600201548190610c8d5760405162461bcd60e51b81526004016103a9906119e3565b6001600160a01b038083166000908152602081905260409020600301548391336101009092041614610cd15760405162461bcd60e51b81526004016103a990611986565b6001600160a01b03831660009081526020819052604090206003015460ff16610d0c5760405162461bcd60e51b81526004016103a990611957565b6001600160a01b0383166000908152602081905260408120600301805460ff191690556008805491610d3d83611b0e565b90915550506040516001600160a01b038416907fca2220ee93ea32595f56c66fee84698b8ff679dad662cde58c0e54d54e5af98390600090a2505050565b6001600160a01b0382166000908152602081905260409020600201546060908390610db85760405162461bcd60e51b81526004016103a9906119e3565b6001600160a01b03841660009081526001602052604081205490848211610de0576000610dea565b610dea8583611af7565b90506000610df88284611af7565b905060008167ffffffffffffffff811115610e2357634e487b7160e01b600052604160045260246000fd5b604051908082528060200260200182016040528015610e5c57816020015b610e4961146e565b815260200190600190039081610e415790505b50905060005b82811015610f30576001600160a01b0389166000908152600160205260409020610e8c8286611aba565b81548110610eaa57634e487b7160e01b600052603260045260246000fd5b9060005260206000209060030201604051806060016040529081600082015481526020016001820160009054906101000a900460ff16151515158152602001600282015481525050828281518110610f1257634e487b7160e01b600052603260045260246000fd5b60200260200101819052508080610f2890611b5a565b915050610e62565b50979650505050505050565b6004546060908310610f605760405162461bcd60e51b81526004016103a990611a39565b6000610f6c8385611aba565b600454909150811115610f7e57506004545b6000610f8a8583611af7565b905060008167ffffffffffffffff811115610fb557634e487b7160e01b600052604160045260246000fd5b604051908082528060200260200182016040528015610fde578160200160208202803683370190505b50905060005b8281101561107e576004610ff88289611aba565b8154811061101657634e487b7160e01b600052603260045260246000fd5b9060005260206000200160009054906101000a90046001600160a01b031682828151811061105457634e487b7160e01b600052603260045260246000fd5b6001600160a01b03909216602092830291909101909101528061107681611b5a565b915050610fe4565b5095945050505050565b6000602081905290815260409020805481906110a390611b25565b80601f01602080910402602001604051908101604052809291908181526020018280546110cf90611b25565b801561111c5780601f106110f15761010080835404028352916020019161111c565b820191906000526020600020905b8154815290600101906020018083116110ff57829003601f168201915b50505050509080600101805461113190611b25565b80601f016020809104026020016040519081016040528092919081815260200182805461115d90611b25565b80156111aa5780601f1061117f576101008083540402835291602001916111aa565b820191906000526020600020905b81548152906001019060200180831161118d57829003601f168201915b50505050600283015460039093015491929160ff8116915061010090046001600160a01b031685565b6001600160a01b03811660009081526020819052604090206002015460609082906112105760405162461bcd60e51b81526004016103a9906119e3565b6001600160a01b038316600090815260016020908152604080832080548251818502810185019093528083529193909284015b8282101561129457600084815260209081902060408051606081018252600386029092018054835260018082015460ff16151584860152600290910154918301919091529083529092019101611243565b5050505091505b50919050565b6001600160a01b03831660009081526020819052604090206002015483906112db5760405162461bcd60e51b81526004016103a9906119e3565b6001600160a01b038416600081815260016020818152604080842081516060810183524281528915158185019081528184018a815283548088018555938852858820925160039485029093019283559051958201805460ff19169615159690961790955593516002909401939093559383525290812080549161135d83611b5a565b90915550506007805490600061137283611b5a565b9190505550836001600160a01b03167f479914d0568c4a3927420aa15b2f8838a29126342e32c563637c6a993bba692e42856040516113b2929190611a79565b60405180910390a250505050565b60026020526000908152604090205460ff1681565b60056020526000908152604090205460ff1681565b8280546113f690611b25565b90600052602060002090601f016020900481019282611418576000855561145e565b82601f1061143157805160ff191683800117855561145e565b8280016001018555821561145e579182015b8281111561145e578251825591602001919060010190611443565b5061146a92915061148e565b5090565b604080516060810182526000808252602082018190529181019190915290565b5b8082111561146a576000815560010161148f565b80356001600160a01b038116811461060257600080fd5b600082601f8301126114ca578081fd5b813567ffffffffffffffff808211156114e5576114e5611b8b565b604051601f8301601f19168101602001828111828210171561150957611509611b8b565b604052828152848301602001861015611520578384fd5b82602086016020830137918201602001929092529392505050565b60006020828403121561154c578081fd5b611555826114a3565b9392505050565b600080600060608486031215611570578182fd5b611579846114a3565b92506020840135801515811461158d578283fd5b929592945050506040919091013590565b6000806000606084860312156115b2578283fd5b6115bb846114a3565b925060208401359150604084013567ffffffffffffffff8111156115dd578182fd5b6115e9868287016114ba565b9150509250925092565b600080600060608486031215611607578283fd5b611610846114a3565b9250602084013567ffffffffffffffff8082111561162c578384fd5b611638878388016114ba565b9350604086013591508082111561164d578283fd5b506115e9868287016114ba565b6000806040838503121561166c578182fd5b611675836114a3565b946020939093013593505050565b600060208284031215611694578081fd5b5035919050565b600080604083850312156116ad578182fd5b82359150602083013567ffffffffffffffff8111156116ca578182fd5b6116d6858286016114ba565b9150509250929050565b600080604083850312156116f2578182fd5b50508035926020909101359150565b60008151808452815b818110156117265760208185018101518683018201520161170a565b818111156117375782602083870101525b50601f01601f19169290920160200192915050565b6001600160a01b0391909116815260200190565b6020808252825182820181905260009190848201906040850190845b818110156117a15783516001600160a01b03168352928401929184019160010161177c565b50909695505050505050565b602080825282518282018190526000919060409081850190868401855b828110156117fb578151805185528681015115158786015285015185850152606090930192908501906001016117ca565b5091979650505050505050565b901515815260200190565b93845260ff9290921660208401526040830152606082015260800190565b6000602082526115556020830184611701565b6000608082526118576080830187611701565b82810360208401526118698187611701565b941515604084015250506060015292915050565b600060a0825261189060a0830188611701565b82810360208401526118a28188611701565b6040840196909652505091151560608301526001600160a01b031660809091015292915050565b60208082526017908201527f5369676e61747572652064656a61207574696c69736565000000000000000000604082015260600190565b60208082526013908201527220b83830b932b4b6103232b5309030b1ba34b360691b604082015260600190565b60208082526010908201526f20b83830b932b4b61034b730b1ba34b360811b604082015260600190565b60208082526015908201527420b83830b932b4b6103232b5309034b730b1ba34b360591b604082015260600190565b6020808252600c908201526b4e6f6e206175746f7269736560a01b604082015260600190565b60208082526018908201527f417070617265696c2064656a6120656e72656769737472650000000000000000604082015260600190565b60208082526010908201526f417070617265696c20696e636f6e6e7560801b604082015260600190565b6020808252601290820152715369676e617475726520696e76616c69646560701b604082015260600190565b60208082526019908201527f537461727420696e646578206f7574206f6620626f756e647300000000000000604082015260600190565b90815260200190565b9182521515602082015260400190565b9283529015156020830152604082015260600190565b93845260208401929092526040830152606082015260800190565b60008219821115611acd57611acd611b75565b500190565b600060ff821660ff84168060ff03821115611aef57611aef611b75565b019392505050565b600082821015611b0957611b09611b75565b500390565b600081611b1d57611b1d611b75565b506000190190565b600281046001821680611b3957607f821691505b6020821081141561129b57634e487b7160e01b600052602260045260246000fd5b6000600019821415611b6e57611b6e611b75565b5060010190565b634e487b7160e01b600052601160045260246000fd5b634e487b7160e01b600052604160045260246000fdfea26469706673582212202693c55372ac7d39b92af75c857d8aaa4430e282668a31ba992b005e0e7b01a564736f6c63430008000033",
}

// IoTDeviceManagerABI is the input ABI used to generate the binding from.
// Deprecated: Use IoTDeviceManagerMetaData.ABI instead.
var IoTDeviceManagerABI = IoTDeviceManagerMetaData.ABI

// IoTDeviceManagerBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use IoTDeviceManagerMetaData.Bin instead.
var IoTDeviceManagerBin = IoTDeviceManagerMetaData.Bin

// DeployIoTDeviceManager deploys a new Ethereum contract, binding an instance of IoTDeviceManager to it.
func DeployIoTDeviceManager(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *IoTDeviceManager, error) {
	parsed, err := IoTDeviceManagerMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(IoTDeviceManagerBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &IoTDeviceManager{IoTDeviceManagerCaller: IoTDeviceManagerCaller{contract: contract}, IoTDeviceManagerTransactor: IoTDeviceManagerTransactor{contract: contract}, IoTDeviceManagerFilterer: IoTDeviceManagerFilterer{contract: contract}}, nil
}

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

// Devices is a free data retrieval call binding the contract method 0xe7b4cac6.
//
// Solidity: function devices(address ) view returns(string publicKey, string metadata, uint256 lastSeen, bool isActive, address owner)
func (_IoTDeviceManager *IoTDeviceManagerCaller) Devices(opts *bind.CallOpts, arg0 common.Address) (struct {
	PublicKey string
	Metadata  string
	LastSeen  *big.Int
	IsActive  bool
	Owner     common.Address
}, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "devices", arg0)

	outstruct := new(struct {
		PublicKey string
		Metadata  string
		LastSeen  *big.Int
		IsActive  bool
		Owner     common.Address
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.PublicKey = *abi.ConvertType(out[0], new(string)).(*string)
	outstruct.Metadata = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.LastSeen = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.IsActive = *abi.ConvertType(out[3], new(bool)).(*bool)
	outstruct.Owner = *abi.ConvertType(out[4], new(common.Address)).(*common.Address)

	return *outstruct, err

}

// Devices is a free data retrieval call binding the contract method 0xe7b4cac6.
//
// Solidity: function devices(address ) view returns(string publicKey, string metadata, uint256 lastSeen, bool isActive, address owner)
func (_IoTDeviceManager *IoTDeviceManagerSession) Devices(arg0 common.Address) (struct {
	PublicKey string
	Metadata  string
	LastSeen  *big.Int
	IsActive  bool
	Owner     common.Address
}, error) {
	return _IoTDeviceManager.Contract.Devices(&_IoTDeviceManager.CallOpts, arg0)
}

// Devices is a free data retrieval call binding the contract method 0xe7b4cac6.
//
// Solidity: function devices(address ) view returns(string publicKey, string metadata, uint256 lastSeen, bool isActive, address owner)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) Devices(arg0 common.Address) (struct {
	PublicKey string
	Metadata  string
	LastSeen  *big.Int
	IsActive  bool
	Owner     common.Address
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
// Solidity: function getDeviceInfo(address deviceAddress) view returns(string, string, bool, uint256)
func (_IoTDeviceManager *IoTDeviceManagerCaller) GetDeviceInfo(opts *bind.CallOpts, deviceAddress common.Address) (string, string, bool, *big.Int, error) {
	var out []interface{}
	err := _IoTDeviceManager.contract.Call(opts, &out, "getDeviceInfo", deviceAddress)

	if err != nil {
		return *new(string), *new(string), *new(bool), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)
	out1 := *abi.ConvertType(out[1], new(string)).(*string)
	out2 := *abi.ConvertType(out[2], new(bool)).(*bool)
	out3 := *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)

	return out0, out1, out2, out3, err

}

// GetDeviceInfo is a free data retrieval call binding the contract method 0x4869fe26.
//
// Solidity: function getDeviceInfo(address deviceAddress) view returns(string, string, bool, uint256)
func (_IoTDeviceManager *IoTDeviceManagerSession) GetDeviceInfo(deviceAddress common.Address) (string, string, bool, *big.Int, error) {
	return _IoTDeviceManager.Contract.GetDeviceInfo(&_IoTDeviceManager.CallOpts, deviceAddress)
}

// GetDeviceInfo is a free data retrieval call binding the contract method 0x4869fe26.
//
// Solidity: function getDeviceInfo(address deviceAddress) view returns(string, string, bool, uint256)
func (_IoTDeviceManager *IoTDeviceManagerCallerSession) GetDeviceInfo(deviceAddress common.Address) (string, string, bool, *big.Int, error) {
	return _IoTDeviceManager.Contract.GetDeviceInfo(&_IoTDeviceManager.CallOpts, deviceAddress)
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

// RegisterDevice is a paid mutator transaction binding the contract method 0x6f6dfeb0.
//
// Solidity: function registerDevice(address deviceAddress, string _publicKey, string _metadata) returns()
func (_IoTDeviceManager *IoTDeviceManagerTransactor) RegisterDevice(opts *bind.TransactOpts, deviceAddress common.Address, _publicKey string, _metadata string) (*types.Transaction, error) {
	return _IoTDeviceManager.contract.Transact(opts, "registerDevice", deviceAddress, _publicKey, _metadata)
}

// RegisterDevice is a paid mutator transaction binding the contract method 0x6f6dfeb0.
//
// Solidity: function registerDevice(address deviceAddress, string _publicKey, string _metadata) returns()
func (_IoTDeviceManager *IoTDeviceManagerSession) RegisterDevice(deviceAddress common.Address, _publicKey string, _metadata string) (*types.Transaction, error) {
	return _IoTDeviceManager.Contract.RegisterDevice(&_IoTDeviceManager.TransactOpts, deviceAddress, _publicKey, _metadata)
}

// RegisterDevice is a paid mutator transaction binding the contract method 0x6f6dfeb0.
//
// Solidity: function registerDevice(address deviceAddress, string _publicKey, string _metadata) returns()
func (_IoTDeviceManager *IoTDeviceManagerTransactorSession) RegisterDevice(deviceAddress common.Address, _publicKey string, _metadata string) (*types.Transaction, error) {
	return _IoTDeviceManager.Contract.RegisterDevice(&_IoTDeviceManager.TransactOpts, deviceAddress, _publicKey, _metadata)
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
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterDeviceRegistered is a free log retrieval operation binding the contract event 0x40ee62d3a0cf66ef791a7a1f9d6d563ec877b394b8f5c097b6238ac82a68c0e2.
//
// Solidity: event DeviceRegistered(address indexed deviceAddress, string publicKey)
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

// WatchDeviceRegistered is a free log subscription operation binding the contract event 0x40ee62d3a0cf66ef791a7a1f9d6d563ec877b394b8f5c097b6238ac82a68c0e2.
//
// Solidity: event DeviceRegistered(address indexed deviceAddress, string publicKey)
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

// ParseDeviceRegistered is a log parse operation binding the contract event 0x40ee62d3a0cf66ef791a7a1f9d6d563ec877b394b8f5c097b6238ac82a68c0e2.
//
// Solidity: event DeviceRegistered(address indexed deviceAddress, string publicKey)
func (_IoTDeviceManager *IoTDeviceManagerFilterer) ParseDeviceRegistered(log types.Log) (*IoTDeviceManagerDeviceRegistered, error) {
	event := new(IoTDeviceManagerDeviceRegistered)
	if err := _IoTDeviceManager.contract.UnpackLog(event, "DeviceRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
