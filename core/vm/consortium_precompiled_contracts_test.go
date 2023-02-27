package vm

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

/*
*
testSortCodes is generated based on the following code
```
// SPDX-License-Identifier: GPL-3.0

pragma solidity >=0.8.0 <0.9.0;

	contract Validators {
	    address[] _validators = [
	    0x0000000000000000000000000000000000000010,
	    0x0000000000000000000000000000000000000011,
	    0x0000000000000000000000000000000000000012,
	    0x0000000000000000000000000000000000000013,
	    0x0000000000000000000000000000000000000014,
	    0x0000000000000000000000000000000000000015,
	    0x0000000000000000000000000000000000000016,
	    0x0000000000000000000000000000000000000017,
	    0x0000000000000000000000000000000000000018,
	    0x0000000000000000000000000000000000000019,
	    0x0000000000000000000000000000000000000020,
	    0x0000000000000000000000000000000000000021,
	    0x0000000000000000000000000000000000000022,
	    0x0000000000000000000000000000000000000023,
	    0x0000000000000000000000000000000000000024,
	    0x0000000000000000000000000000000000000025,
	    0x0000000000000000000000000000000000000026,
	    0x0000000000000000000000000000000000000027,
	    0x0000000000000000000000000000000000000028,
	    0x0000000000000000000000000000000000000029,
	    0x0000000000000000000000000000000000000030
	    ];
	    uint256[] _totalBalances = [1,4,6,2,8,9,10,3,16,20,100,12,22,30,50,60,5,18,16,22,21];
		constructor() {}
	    function getValidatorCandidates() public view returns (address[] memory _validatorList) {
	        _validatorList = _validators;
	    }
	    function totalBalances(address[] calldata _poolList) public view returns (uint256[] memory _balances) {
	        _balances = _totalBalances;
	    }
	}

```
*/
const testSortCode = `6080604052604051806102a00160405280601073ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001601173ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001601273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001601373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001601473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001601573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001601673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001601773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001601873ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001601973ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001602073ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001602173ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001602273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001602373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001602473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001602573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001602673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001602773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001602873ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001602973ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001603073ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815250600090601561044d92919061054d565b50604051806102a00160405280600160ff168152602001600460ff168152602001600660ff168152602001600260ff168152602001600860ff168152602001600960ff168152602001600a60ff168152602001600360ff168152602001601060ff168152602001601460ff168152602001606460ff168152602001600c60ff168152602001601660ff168152602001601e60ff168152602001603260ff168152602001603c60ff168152602001600560ff168152602001601260ff168152602001601060ff168152602001601660ff168152602001601560ff16815250600190601561053a9291906105d7565b5034801561054757600080fd5b50610646565b8280548282559060005260206000209081019282156105c6579160200282015b828111156105c55782518260006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055509160200191906001019061056d565b5b5090506105d39190610629565b5090565b828054828255906000526020600020908101928215610618579160200282015b82811115610617578251829060ff169055916020019190600101906105f7565b5b5090506106259190610629565b5090565b5b8082111561064257600081600090555060010161062a565b5090565b610460806106556000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c80634a5d76cd1461003b578063ba77b06c1461006b575b600080fd5b610055600480360381019061005091906101e1565b610089565b60405161006291906102f6565b60405180910390f35b6100736100e4565b6040516100809190610408565b60405180910390f35b606060018054806020026020016040519081016040528092919081815260200182805480156100d757602002820191906000526020600020905b8154815260200190600101908083116100c3575b5050505050905092915050565b6060600080548060200260200160405190810160405280929190818152602001828054801561016857602002820191906000526020600020905b8160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001906001019080831161011e575b5050505050905090565b600080fd5b600080fd5b600080fd5b600080fd5b600080fd5b60008083601f8401126101a1576101a061017c565b5b8235905067ffffffffffffffff8111156101be576101bd610181565b5b6020830191508360208202830111156101da576101d9610186565b5b9250929050565b600080602083850312156101f8576101f7610172565b5b600083013567ffffffffffffffff81111561021657610215610177565b5b6102228582860161018b565b92509250509250929050565b600081519050919050565b600082825260208201905092915050565b6000819050602082019050919050565b6000819050919050565b61026d8161025a565b82525050565b600061027f8383610264565b60208301905092915050565b6000602082019050919050565b60006102a38261022e565b6102ad8185610239565b93506102b88361024a565b8060005b838110156102e95781516102d08882610273565b97506102db8361028b565b9250506001810190506102bc565b5085935050505092915050565b600060208201905081810360008301526103108184610298565b905092915050565b600081519050919050565b600082825260208201905092915050565b6000819050602082019050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b600061036f82610344565b9050919050565b61037f81610364565b82525050565b60006103918383610376565b60208301905092915050565b6000602082019050919050565b60006103b582610318565b6103bf8185610323565b93506103ca83610334565b8060005b838110156103fb5781516103e28882610385565b97506103ed8361039d565b9250506001810190506103ce565b5085935050505092915050565b6000602082019050818103600083015261042281846103aa565b90509291505056fea26469706673582212204aab6725297603e4a73daf57fffdd55330e60293dfde78dcbbddb9178144583064736f6c63430008110033`

/*
*
wrapupCode is used to call sortValidators precompiled contract with the following code
```
// SPDX-License-Identifier: GPL-3.0
pragma solidity >=0.8.0 <0.9.0;

	contract Wrapup {
	    constructor() {}

	    address[] _validators = [
	        0x0000000000000000000000000000000000000064,
	        0x0000000000000000000000000000000000000065,
	        0x0000000000000000000000000000000000000066,
	        0x0000000000000000000000000000000000000067,
	        0x0000000000000000000000000000000000000068,
	        0x0000000000000000000000000000000000000069,
	        0x000000000000000000000000000000000000006a,
	        0x000000000000000000000000000000000000006b,
	        0x000000000000000000000000000000000000006C,
	        0x000000000000000000000000000000000000006D,
	        0x000000000000000000000000000000000000006E,
	        0x000000000000000000000000000000000000006F,
	        0x0000000000000000000000000000000000000070,
	        0x0000000000000000000000000000000000000071,
	        0x0000000000000000000000000000000000000072,
	        0x0000000000000000000000000000000000000073,
	        0x0000000000000000000000000000000000000074,
	        0x0000000000000000000000000000000000000075,
	        0x0000000000000000000000000000000000000076,
	        0x0000000000000000000000000000000000000077,
	        0x0000000000000000000000000000000000000078
	    ];
	    uint256[] _weights = [
	        uint256(1000),
	        uint256(2000),
	        uint256(3000),
	        uint256(4000),
	        uint256(5000),
	        uint256(6000),
	        uint256(7000),
	        uint256(8000),
	        uint256(9000),
	        uint256(10000),
	        uint256(11000),
	        uint256(12000),
	        uint256(13000),
	        uint256(14000),
	        uint256(15000),
	        uint256(16000),
	        uint256(17000),
	        uint256(18000),
	        uint256(19000),
	        uint256(20000),
	        uint256(21000)
	    ];

	    function sortValidators() public view returns (address[3] memory _result) {
	        bytes memory payload = abi.encodeWithSignature(
	            "sortValidators(address[],uint256[])",
	            _validators,
	            _weights
	        );

	        uint256 payloadLength = payload.length;
	        uint256 validatorsLength = _validators.length;
	        uint256 validatorsLengthHex = validatorsLength * 0x20;
	        uint256 validatorsLengthHex64 = validatorsLength * 0x20 + 64;
	        address _smc = address(0x66);

	        assembly {
	            let payloadStart := add(payload, 32)
	            if iszero(
	                staticcall(
	                    0,
	                    _smc,
	                    payloadStart,
	                    payloadLength,
	                    _result,
	                    validatorsLengthHex64
	                )
	            ) {
	                revert(0, 0)
	            }
	            returndatacopy(_result, 64, validatorsLengthHex)
	            return(_result, validatorsLengthHex)
	        }
	    }
	}

```
*/
const (
	wrapupCode = `6080604052604051806102a00160405280606473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606873ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606973ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606a73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606c73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606d73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606e73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606f73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001607073ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001607173ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001607273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001607373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001607473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001607573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001607673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001607773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001607873ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815250600090601561044d929190610523565b50604051806102a001604052806103e881526020016107d08152602001610bb88152602001610fa0815260200161138881526020016117708152602001611b588152602001611f40815260200161232881526020016127108152602001612af88152602001612ee081526020016132c881526020016136b08152602001613a988152602001613e80815260200161426881526020016146508152602001614a388152602001614e20815260200161520881525060019060156105109291906105ad565b5034801561051d57600080fd5b50610617565b82805482825590600052602060002090810192821561059c579160200282015b8281111561059b5782518260006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555091602001919060010190610543565b5b5090506105a991906105fa565b5090565b8280548282559060005260206000209081019282156105e9579160200282015b828111156105e85782518255916020019190600101906105cd565b5b5090506105f691906105fa565b5090565b5b808211156106135760008160009055506001016105fb565b5090565b6105bc806106266000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c806327167aec14610030575b600080fd5b61003861004e565b60405161004591906102de565b60405180910390f35b61005661014c565b600080600160405160240161006c9291906102fa565b6040516020818303038152906040527f788341af000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff19166020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff8381831617835250505050905060008151905060008080549050905060006020826101099190610430565b90506000604060208461011c9190610430565b61012691906103da565b90506000606690506020860182888783856000fa61014357600080fd5b836040893e8388f35b604051806102a00160405280601590602082028036833780820191505090505090565b600061017b838361019f565b60208301905092915050565b600061019383836102cf565b60208301905092915050565b6101a8816104b4565b82525050565b6101b781610365565b6101c181846103ad565b92506101cc82610331565b8060005b838110156101fd5781516101e4878261016f565b96506101ef83610386565b9250506001810190506101d0565b505050505050565b600061021082610370565b61021a81856103b8565b93506102258361033b565b8060005b8381101561025d5761023a82610553565b610244888261016f565b975061024f83610393565b925050600181019050610229565b5085935050505092915050565b60006102758261037b565b61027f81856103c9565b935061028a83610350565b8060005b838110156102c25761029f82610566565b6102a98882610187565b97506102b4836103a0565b92505060018101905061028e565b5085935050505092915050565b6102d8816104e6565b82525050565b60006102a0820190506102f460008301846101ae565b92915050565b600060408201905081810360008301526103148185610205565b90508181036020830152610328818461026a565b90509392505050565b6000819050919050565b60008190508160005260206000209050919050565b60008190508160005260206000209050919050565b600060159050919050565b600081549050919050565b600081549050919050565b6000602082019050919050565b6000600182019050919050565b6000600182019050919050565b600081905092915050565b600082825260208201905092915050565b600082825260208201905092915050565b60006103e5826104e6565b91506103f0836104e6565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0382111561042557610424610524565b5b828201905092915050565b600061043b826104e6565b9150610446836104e6565b9250817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff048311821515161561047f5761047e610524565b5b828202905092915050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b60006104bf826104c6565b9050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b60006105036104fe83610579565b61048a565b9050919050565b600061051d61051883610579565b6104aa565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600061055f82546104f0565b9050919050565b6000610572825461050a565b9050919050565b60008160001c905091905056fea2646970667358221220bf18e78f84c245c19e8996d9afbcaa3ec60e02e413932b3848d7eed59feb161264736f6c63430008070033`
	wrapupAbi  = `[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"inputs":[],"name":"sortValidators","outputs":[{"internalType":"address[21]","name":"_validators","type":"address[21]"}],"stateMutability":"view","type":"function"}]`
)

/*
*
// SPDX-License-Identifier: GPL-3.0
pragma solidity >=0.8.0 <0.9.0;

	contract SyncNewValidatorSet {
	    constructor() {}

	    address[] _candidates = [
	        0x0000000000000000000000000000000000000064,
	        0x0000000000000000000000000000000000000065,
	        0x0000000000000000000000000000000000000066,
	        0x0000000000000000000000000000000000000067,
	        0x0000000000000000000000000000000000000068,
	        0x0000000000000000000000000000000000000069,
	        0x000000000000000000000000000000000000006a,
	        0x000000000000000000000000000000000000006b,
	        0x000000000000000000000000000000000000006C,
	        0x000000000000000000000000000000000000006D,
	        0x000000000000000000000000000000000000006E,
	        0x000000000000000000000000000000000000006F,
	        0x0000000000000000000000000000000000000070,
	        0x0000000000000000000000000000000000000071,
	        0x0000000000000000000000000000000000000072,
	        0x0000000000000000000000000000000000000073,
	        0x0000000000000000000000000000000000000074,
	        0x0000000000000000000000000000000000000075,
	        0x0000000000000000000000000000000000000076,
	        0x0000000000000000000000000000000000000077,
	        0x0000000000000000000000000000000000000078
	    ];
	    uint256[] _weights = [
	        uint256(1000000),
	        uint256(2000000),
	        uint256(3000000),
	        uint256(4000000),
	        uint256(5000000),
	        uint256(6000000),
	        uint256(7000000),
	        uint256(8000000),
	        uint256(9000000),
	        uint256(10000000),
	        uint256(11000000),
	        uint256(12000000),
	        uint256(13000000),
	        uint256(14000000),
	        uint256(15000000),
	        uint256(16000000),
	        uint256(17000000),
	        uint256(18000000),
	        uint256(19000000),
	        uint256(20000000),
	        uint256(21000000)
	    ];
		uint256[] _isTrustedOrganizations = [
			uint256(0),
			uint256(0),
			uint256(0),
			uint256(0),
			uint256(0),
			uint256(0),
			uint256(0),
			uint256(0),
			uint256(0),
			uint256(0),
			uint256(1),
			uint256(1),
			uint256(1),
			uint256(1),
			uint256(1),
			uint256(1),
			uint256(1),
			uint256(1),
			uint256(1),
			uint256(1),
			uint256(1)
		];
	    uint256 _maxValidatorNumber = 21;
	    uint256 _maxPrioritizedValidatorNumber = 11;

	    function pickValidatorSet() public view returns (address[] memory _result) {
	        bytes memory _payload = abi.encodeWithSignature(
	            "pickValidatorSet(address[],uint256[],uint256[],uint256,uint256)",
	            _candidates,
	            _weights,
	            _isTrustedOrganizations,
	            _maxValidatorNumber,
	            _maxPrioritizedValidatorNumber
	        );
	        bool _success = true;

	        uint256 _payloadLength = _payload.length;
	        uint256 _resultLength = 0x20 * _candidates.length + 0x40;
	        address _smc = address(0x68);

	        assembly {
	            let _payloadStart := add(_payload, 0x20)

	            if iszero(staticcall(gas(), _smc, _payloadStart, _payloadLength, _result, _resultLength)) {
	                _success := 0
	            }

	            if iszero(returndatasize()) {
	                _success := 0
	            }

	            _result := add(_result, 0x20)
	        }

	        require(_success, "PrecompileUsageSortValidators: call to precompile fails");
	    }
	}
*/
const (
	syncNewValidatorSetCode = `6080604052604051806102a00160405280606473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606873ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606973ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606a73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606c73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606d73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606e73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001606f73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001607073ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001607173ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001607273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001607373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001607473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001607573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001607673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001607773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001607873ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815250600090601561044d9291906105f5565b50604051806102a00160405280620f42408152602001621e84808152602001622dc6c08152602001623d09008152602001624c4b408152602001625b8d808152602001626acfc08152602001627a1200815260200162895440815260200162989680815260200162a7d8c0815260200162b71b00815260200162c65d40815260200162d59f80815260200162e4e1c0815260200162f42400815260200163010366408152602001630112a8808152602001630121eac081526020016301312d0081526020016301406f40815250600190601561052a92919061067f565b50604051806102a001604052806000815260200160008152602001600081526020016000815260200160008152602001600081526020016000815260200160008152602001600081526020016000815260200160018152602001600181526020016001815260200160018152602001600181526020016001815260200160018152602001600181526020016001815260200160018152602001600181525060029060156105d892919061067f565b506015600355600b6004553480156105ef57600080fd5b506106e9565b82805482825590600052602060002090810192821561066e579160200282015b8281111561066d5782518260006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555091602001919060010190610615565b5b50905061067b91906106cc565b5090565b8280548282559060005260206000209081019282156106bb579160200282015b828111156106ba57825182559160200191906001019061069f565b5b5090506106c891906106cc565b5090565b5b808211156106e55760008160009055506001016106cd565b5090565b6106de806106f86000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c8063c80c0b6c14610030575b600080fd5b61003861004f565b604051610046929190610340565b60405180910390f35b6060600080600060016002600354600454604051602401610074959493929190610370565b6040516020818303038152906040527f77941512000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff19166020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff838183161783525050505090506000600190506000825190506000604060008054905060206101159190610503565b61011f91906104ad565b90506000606890506020850182888583855afa61013b57600094505b3d61014557600094505b879650602088019750508361018f576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610186906103d8565b60405180910390fd5b50505050509091565b60006101a483836101c8565b60208301905092915050565b60006101bc8383610322565b60208301905092915050565b6101d181610587565b82525050565b60006101e282610432565b6101ec818561047a565b93506101f7836103f8565b8060005b8381101561022857815161020f8882610198565b975061021a83610453565b9250506001810190506101fb565b5085935050505092915050565b60006102408261043d565b61024a818561047a565b935061025583610408565b8060005b8381101561028d5761026a82610626565b6102748882610198565b975061027f83610460565b925050600181019050610259565b5085935050505092915050565b60006102a582610448565b6102af818561048b565b93506102ba8361041d565b8060005b838110156102f2576102cf82610639565b6102d988826101b0565b97506102e48361046d565b9250506001810190506102be565b5085935050505092915050565b600061030c60378361049c565b915061031782610659565b604082019050919050565b61032b816105b9565b82525050565b61033a816105b9565b82525050565b6000604082019050818103600083015261035a81856101d7565b90506103696020830184610331565b9392505050565b600060a082019050818103600083015261038a8188610235565b9050818103602083015261039e818761029a565b905081810360408301526103b2818661029a565b90506103c16060830185610331565b6103ce6080830184610331565b9695505050505050565b600060208201905081810360008301526103f1816102ff565b9050919050565b6000819050602082019050919050565b60008190508160005260206000209050919050565b60008190508160005260206000209050919050565b600081519050919050565b600081549050919050565b600081549050919050565b6000602082019050919050565b6000600182019050919050565b6000600182019050919050565b600082825260208201905092915050565b600082825260208201905092915050565b600082825260208201905092915050565b60006104b8826105b9565b91506104c3836105b9565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff038211156104f8576104f76105f7565b5b828201905092915050565b600061050e826105b9565b9150610519836105b9565b9250817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0483118215151615610552576105516105f7565b5b828202905092915050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b600061059282610599565b9050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b60006105d66105d18361064c565b61055d565b9050919050565b60006105f06105eb8361064c565b61057d565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600061063282546105c3565b9050919050565b600061064582546105dd565b9050919050565b60008160001c9050919050565b7f507265636f6d70696c655573616765536f727456616c696461746f72733a206360008201527f616c6c20746f20707265636f6d70696c65206661696c7300000000000000000060208201525056fea264697066735822122085ebbb8c16b63d8e71a6315b0c7b1ad88e35ddb36d250664b845bfe23027c0c064736f6c63430008070033`
	syncNewValidatorSetAbi  = `[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"inputs":[],"name":"pickValidatorSet","outputs":[{"internalType":"address[]","name":"_result","type":"address[]"},{"internalType":"uint256","name":"_newValidatorCount","type":"uint256"}],"stateMutability":"view","type":"function"}]`
)

var (
	caller             = common.BytesToAddress([]byte("sender"))
	expectedValidators = []common.Address{
		common.BytesToAddress([]byte{120}),
		common.BytesToAddress([]byte{119}),
		common.BytesToAddress([]byte{118}),
		common.BytesToAddress([]byte{117}),
		common.BytesToAddress([]byte{116}),
		common.BytesToAddress([]byte{115}),
		common.BytesToAddress([]byte{114}),
		common.BytesToAddress([]byte{113}),
		common.BytesToAddress([]byte{112}),
		common.BytesToAddress([]byte{111}),
		common.BytesToAddress([]byte{110}),
		common.BytesToAddress([]byte{109}),
		common.BytesToAddress([]byte{108}),
		common.BytesToAddress([]byte{107}),
		common.BytesToAddress([]byte{106}),
		common.BytesToAddress([]byte{105}),
		common.BytesToAddress([]byte{104}),
		common.BytesToAddress([]byte{103}),
		common.BytesToAddress([]byte{102}),
		common.BytesToAddress([]byte{101}),
		common.BytesToAddress([]byte{100}),
	}
)

/*
*
verifyHeadersTestCode represents the following smart contract code
// SPDX-License-Identifier: GPL-3.0
pragma solidity >=0.8.0 <0.9.0;

contract VerifyHeaderTestContract {

	    constructor() {}

	    function verify(bytes memory header1, bytes memory header2) public view returns (bool) {
	        bytes memory payload = abi.encodeWithSignature("validatingDoubleSignProof(bytes,bytes)", header1, header2);
	        uint payloadLength = payload.length;
	        address _smc = address(0x67);
	        uint[1] memory _output;
	        assembly {
	            let payloadStart := add(payload, 32)
	            if iszero(staticcall(0, _smc, payloadStart, payloadLength, _output, 0x20)) {
	                revert(0, 0)
	            }
	        }
	        return (_output[0] != 0);
	    }
	}
*/
const (
	verifyHeadersTestCode = "608060405234801561001057600080fd5b50610299806100206000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c8063f7e83aee14610030575b600080fd5b61004361003e36600461018b565b610057565b604051901515815260200160405180910390f35b600080838360405160240161006d929190610235565b60408051601f198184030181529190526020810180516001600160e01b031663580a316360e01b179052805190915060676100a66100ca565b602084016020828583866000fa6100bc57600080fd5b505115159695505050505050565b60405180602001604052806001906020820280368337509192915050565b634e487b7160e01b600052604160045260246000fd5b600082601f83011261010f57600080fd5b813567ffffffffffffffff8082111561012a5761012a6100e8565b604051601f8301601f19908116603f01168101908282118183101715610152576101526100e8565b8160405283815286602085880101111561016b57600080fd5b836020870160208301376000602085830101528094505050505092915050565b6000806040838503121561019e57600080fd5b823567ffffffffffffffff808211156101b657600080fd5b6101c2868387016100fe565b935060208501359150808211156101d857600080fd5b506101e5858286016100fe565b9150509250929050565b6000815180845260005b81811015610215576020818501810151868301820152016101f9565b506000602082860101526020601f19601f83011685010191505092915050565b60408152600061024860408301856101ef565b828103602084015261025a81856101ef565b9594505050505056fea2646970667358221220e689890bbe17c2e97389470ed4baa21af25fd9cd6348d7511924615440d967d364736f6c63430008110033"
	verifyHeadersTestAbi  = `[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"inputs":[{"internalType":"bytes","name":"header1","type":"bytes"},{"internalType":"bytes","name":"header2","type":"bytes"}],"name":"verify","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"}]`
)

type TestScenario struct {
	Address   common.Address
	Weight    *big.Int
	IsTrusted *big.Int
}

func getTestScenarios() []TestScenario {
	return []TestScenario{
		{
			Address:   common.BytesToAddress([]byte{100}),
			Weight:    big.NewInt(1_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{101}),
			Weight:    big.NewInt(2_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{102}),
			Weight:    big.NewInt(3_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{103}),
			Weight:    big.NewInt(4_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{104}),
			Weight:    big.NewInt(5_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{105}),
			Weight:    big.NewInt(6_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{106}),
			Weight:    big.NewInt(7_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{107}),
			Weight:    big.NewInt(8_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{108}),
			Weight:    big.NewInt(9_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{109}),
			Weight:    big.NewInt(10_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{110}),
			Weight:    big.NewInt(11_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{111}),
			Weight:    big.NewInt(12_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{112}),
			Weight:    big.NewInt(13_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{113}),
			Weight:    big.NewInt(14_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{114}),
			Weight:    big.NewInt(15_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{115}),
			Weight:    big.NewInt(16_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{116}),
			Weight:    big.NewInt(17_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{117}),
			Weight:    big.NewInt(18_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{118}),
			Weight:    big.NewInt(19_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{119}),
			Weight:    big.NewInt(20_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{120}),
			Weight:    big.NewInt(21_000_000),
			IsTrusted: big0,
		},
	}
}

func getAddresses(scenarios []TestScenario) []common.Address {
	var result []common.Address
	for _, s := range scenarios {
		result = append(result, s.Address)
	}

	return result
}

func getWeights(scenarios []TestScenario) []*big.Int {
	var result []*big.Int
	for _, s := range scenarios {
		result = append(result, s.Weight)
	}

	return result
}

func getIsTrusted(scenarios []TestScenario) []*big.Int {
	var result []*big.Int
	for _, s := range scenarios {
		result = append(result, s.IsTrusted)
	}

	return result
}

func setIsTrustedAt(scenarios []TestScenario, indexes ...int64) {
	for _, i := range indexes {
		scenarios[i].IsTrusted = big1
	}
}

var (
	addressesTest = []common.Address{
		common.BytesToAddress([]byte{100}),
		common.BytesToAddress([]byte{101}),
		common.BytesToAddress([]byte{102}),
		common.BytesToAddress([]byte{103}),
		common.BytesToAddress([]byte{104}),
		common.BytesToAddress([]byte{105}),
		common.BytesToAddress([]byte{106}),
		common.BytesToAddress([]byte{107}),
		common.BytesToAddress([]byte{108}),
		common.BytesToAddress([]byte{109}),
		common.BytesToAddress([]byte{110}),
		common.BytesToAddress([]byte{111}),
		common.BytesToAddress([]byte{112}),
		common.BytesToAddress([]byte{113}),
		common.BytesToAddress([]byte{114}),
		common.BytesToAddress([]byte{115}),
		common.BytesToAddress([]byte{116}),
		common.BytesToAddress([]byte{117}),
		common.BytesToAddress([]byte{118}),
		common.BytesToAddress([]byte{119}),
		common.BytesToAddress([]byte{120}),
	}
	weightsTest = []*big.Int{
		big.NewInt(1_000_000),
		big.NewInt(2_000_000),
		big.NewInt(3_000_000),
		big.NewInt(4_000_000),
		big.NewInt(5_000_000),
		big.NewInt(6_000_000),
		big.NewInt(7_000_000),
		big.NewInt(8_000_000),
		big.NewInt(9_000_000),
		big.NewInt(10_000_000),
		big.NewInt(11_000_000),
		big.NewInt(12_000_000),
		big.NewInt(13_000_000),
		big.NewInt(14_000_000),
		big.NewInt(15_000_000),
		big.NewInt(16_000_000),
		big.NewInt(17_000_000),
		big.NewInt(18_000_000),
		big.NewInt(19_000_000),
		big.NewInt(20_000_000),
		big.NewInt(22_000_000),
	}
)

func TestSort(t *testing.T) {
	fmt.Println(common.BytesToAddress([]byte{101}))
	addrs := []common.Address{
		common.BytesToAddress([]byte{102}),
		common.BytesToAddress([]byte{103}),
		common.BytesToAddress([]byte{104}),
		common.BytesToAddress([]byte{105}),
		common.BytesToAddress([]byte{106}),
	}

	totalBalances := []*big.Int{
		big.NewInt(10),
		big.NewInt(2),
		big.NewInt(11),
		big.NewInt(15),
		big.NewInt(1),
	}

	expectedAddrs := []common.Address{
		common.BytesToAddress([]byte{105}),
		common.BytesToAddress([]byte{104}),
		common.BytesToAddress([]byte{102}),
		common.BytesToAddress([]byte{103}),
		common.BytesToAddress([]byte{106}),
	}

	expectedBalances := []*big.Int{
		big.NewInt(15),
		big.NewInt(11),
		big.NewInt(10),
		big.NewInt(2),
		big.NewInt(1),
	}

	sortValidators(addrs, totalBalances)
	for i, val := range addrs {
		if expectedBalances[i].Cmp(totalBalances[i]) != 0 {
			t.Fatal(fmt.Sprintf("mismatched balance at %d, expected:%s got:%s", i, expectedBalances[i].String(), totalBalances[i].String()))
		}
		if expectedAddrs[i].Hex() != val.Hex() {
			t.Fatal(fmt.Sprintf("mismatched addr at %d, expected:%s got:%s", i, expectedValidators[i].Hex(), val.Hex()))
		}
	}
}

// TestConsortiumValidatorSorting_Run sorts 21 validators successfully
func TestConsortiumValidatorSorting_Run(t *testing.T) {
	var (
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	)

	smcAbi, err := abi.JSON(strings.NewReader(consortiumSortValidatorAbi))
	if err != nil {
		t.Fatal(err)
	}

	input, err := smcAbi.Pack(sortValidatorsMethod, addressesTest, weightsTest)

	if err != nil {
		t.Fatal(err)
	}

	evm, err := newEVM(caller, statedb)
	if err != nil {
		t.Fatal(err)
	}
	c := &consortiumValidatorSorting{caller: AccountRef(caller), evm: evm}
	output, err := c.Run(input)
	if err != nil {
		t.Fatal(err)
	}
	//println(common.Bytes2Hex(output))

	res, err := smcAbi.Methods[sortValidatorsMethod].Outputs.Unpack(output)
	if err != nil {
		t.Fatal(err)
	}
	sortedValidators := *abi.ConvertType(res[0], new([21]common.Address)).(*[21]common.Address)
	if len(expectedValidators) != len(sortedValidators) {
		t.Fatal(fmt.Sprintf("expected len %d, got %v", 21, len(sortedValidators)))
	}
	for i, addr := range sortedValidators {
		//println(addr.Hex())
		if expectedValidators[i].Hex() != addr.Hex() {
			t.Fatal(fmt.Sprintf("mismatched addr at %d, expected:%s got:%s", i, expectedValidators[i].Hex(), addr.Hex()))
		}
	}
}

// TestConsortiumValidatorSorting_Run2 simulates a call from a user who trigger system contract to call `sort` precompiled contract
func TestConsortiumValidatorSorting_Run2(t *testing.T) {
	var (
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	)
	smcAbi, err := abi.JSON(strings.NewReader(wrapupAbi))
	if err != nil {
		t.Fatal(err)
	}
	input, err := smcAbi.Pack(sortValidatorsMethod)
	if err != nil {
		t.Fatal(err)
	}

	evm, err := newEVM(caller, statedb)
	if err != nil {
		t.Fatal(err)
	}
	_, contract, _, err := evm.Create(AccountRef(caller), common.FromHex(wrapupCode), math.MaxUint64/2, big0)
	if err != nil {
		t.Fatal(err)
	}

	// set this contract into consortiumV2Contracts to make it become a system contract
	evm.chainConfig.ConsortiumV2Contracts.SlashIndicator = contract

	ret, _, err := evm.StaticCall(AccountRef(caller), contract, input, 1000000)
	if err != nil {
		t.Fatal(err)
	}

	res, err := smcAbi.Methods[sortValidatorsMethod].Outputs.Unpack(ret)
	if err != nil {
		t.Fatal(err)
	}

	sortedValidators := *abi.ConvertType(res[0], new([21]common.Address)).(*[21]common.Address)
	if len(expectedValidators) != len(sortedValidators) {
		t.Fatal(fmt.Sprintf("expected len 21, got %v", len(sortedValidators)))
	}
	for i, addr := range sortedValidators {
		if expectedValidators[i].Hex() != addr.Hex() {
			t.Fatal(fmt.Sprintf("mismatched addr at %d, expected:%s got:%s", i, expectedValidators[i].Hex(), addr.Hex()))
		}
	}
}

// TestConsortiumVerifyHeaders_verify tests verify function
func TestConsortiumVerifyHeaders_verify(t *testing.T) {
	header1, header2, err := prepareHeader(big1)
	if err != nil {
		t.Fatal(err)
	}
	c := &consortiumVerifyHeaders{evm: &EVM{chainConfig: &params.ChainConfig{ChainID: big1}}}
	if !c.verify(*types.FromHeader(header1, big1), *types.FromHeader(header2, big1)) {
		t.Fatal("expected true, got false")
	}

	if c.verify(*types.FromHeader(header1, big1), *types.FromHeader(header1, big1)) {
		t.Fatal("expected false, got true")
	}
}

// TestConsortiumVerifyHeaders_Run init 2 headers, pack them and call `Run` function directly
func TestConsortiumVerifyHeaders_Run(t *testing.T) {
	var (
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	)
	smcAbi, err := abi.JSON(strings.NewReader(consortiumVerifyHeadersAbi))
	if err != nil {
		t.Fatal(err)
	}
	evm, err := newEVM(caller, statedb)
	if err != nil {
		t.Fatal(err)
	}
	header1, header2, err := prepareHeader(evm.chainConfig.ChainID)
	if err != nil {
		t.Fatal(err)
	}
	encodedHeader1, err := types.FromHeader(header1, big1).Bytes(consortiumVerifyHeadersAbi, getHeader)
	if err != nil {
		t.Fatal(err)
	}
	encodedHeader2, err := types.FromHeader(header2, big1).Bytes(consortiumVerifyHeadersAbi, getHeader)
	if err != nil {
		t.Fatal(err)
	}
	input, err := smcAbi.Pack(verifyHeaders, encodedHeader1, encodedHeader2)
	if err != nil {
		t.Fatal(err)
	}
	c := &consortiumVerifyHeaders{evm: evm, caller: AccountRef(caller)}
	result, err := c.Run(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 32 {
		t.Fatal(fmt.Sprintf("expected len 32 got %d", len(result)))
	}
	if result[len(result)-1] != 1 {
		t.Fatal(fmt.Sprintf("expected 1 (true) got %d", result[len(result)-1]))
	}
}

// TestConsortiumVerifyHeaders_Run2 deploys smart contract and call precompiled contracts via this contract
func TestConsortiumVerifyHeaders_Run2(t *testing.T) {
	var (
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	)
	smcAbi, err := abi.JSON(strings.NewReader(verifyHeadersTestAbi))
	if err != nil {
		t.Fatal(err)
	}
	header1, header2, err := prepareHeader(big1)
	if err != nil {
		t.Fatal(err)
	}
	encodedHeader1, err := types.FromHeader(header1, big1).Bytes(consortiumVerifyHeadersAbi, getHeader)
	if err != nil {
		t.Fatal(err)
	}
	encodedHeader2, err := types.FromHeader(header2, big1).Bytes(consortiumVerifyHeadersAbi, getHeader)
	if err != nil {
		t.Fatal(err)
	}
	input, err := smcAbi.Pack("verify", encodedHeader1, encodedHeader2)
	if err != nil {
		t.Fatal(err)
	}
	evm, err := newEVM(caller, statedb)
	if err != nil {
		t.Fatal(err)
	}
	_, contract, _, err := evm.Create(AccountRef(caller), common.FromHex(verifyHeadersTestCode), math.MaxUint64/2, big0)
	if err != nil {
		t.Fatal(err)
	}

	// set this contract into consortiumV2Contracts to make it become a system contract
	evm.chainConfig.ConsortiumV2Contracts.SlashIndicator = contract

	ret, _, err := evm.StaticCall(AccountRef(caller), contract, input, 1000000)
	if err != nil {
		t.Fatal(err)
	}

	res, err := smcAbi.Methods["verify"].Outputs.Unpack(ret)
	if err != nil {
		t.Fatal(err)
	}
	result := *abi.ConvertType(res[0], new(bool)).(*bool)
	if !result {
		t.Fatal("expected true got false")
	}
}

// TestArrangeValidatorCandidates arranges 21 candidates with 11 trusted nodes (ordering)
func TestArrangeValidatorCandidates(t *testing.T) {
	scenarios := getTestScenarios()
	setIsTrustedAt(scenarios, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20)
	isTrustedOrganizations := getIsTrusted(scenarios)
	candidates := getAddresses(scenarios)
	weights := getWeights(scenarios)

	newValidatorCount := uint64(21)
	maxPrioritizedValidatorNumber := big.NewInt(11)

	candidateMap := createCandidateMap(candidates, isTrustedOrganizations)
	sortValidators(candidates, weights)
	updateIsTrustedOrganizations(candidates, isTrustedOrganizations, candidateMap)
	arrangeValidatorCandidates(candidates, newValidatorCount, isTrustedOrganizations, maxPrioritizedValidatorNumber)

	expectedCandidates := []common.Address{
		common.BytesToAddress([]byte{120}),
		common.BytesToAddress([]byte{119}),
		common.BytesToAddress([]byte{118}),
		common.BytesToAddress([]byte{117}),
		common.BytesToAddress([]byte{116}),
		common.BytesToAddress([]byte{115}),
		common.BytesToAddress([]byte{114}),
		common.BytesToAddress([]byte{113}),
		common.BytesToAddress([]byte{112}),
		common.BytesToAddress([]byte{111}),

		common.BytesToAddress([]byte{110}),
		common.BytesToAddress([]byte{109}),
		common.BytesToAddress([]byte{108}),
		common.BytesToAddress([]byte{107}),
		common.BytesToAddress([]byte{106}),
		common.BytesToAddress([]byte{105}),
		common.BytesToAddress([]byte{104}),
		common.BytesToAddress([]byte{103}),
		common.BytesToAddress([]byte{102}),
		common.BytesToAddress([]byte{101}),
		common.BytesToAddress([]byte{100}),
	}
	for i, candidate := range candidates {
		if !bytes.Equal(expectedCandidates[i].Bytes(), candidate.Bytes()) {
			t.Fatal(fmt.Sprintf("mismatched candidate address at %d, expected:%s got:%s", i, expectedCandidates[i].Hex(), candidate.Hex()))
		}
	}
}

// TestArrangeValidatorCandidates_RandomTrustedOrganizations arranges 21 candidates with 11 trusted nodes (randomly)
func TestArrangeValidatorCandidates_RandomTrustedOrganizations(t *testing.T) {
	scenarios := getTestScenarios()
	setIsTrustedAt(scenarios, 2, 5, 8, 10, 11, 13, 14, 17, 18, 19, 20)
	isTrustedOrganizations := getIsTrusted(scenarios)
	candidates := getAddresses(scenarios)
	weights := getWeights(scenarios)

	newValidatorCount := uint64(21)
	maxPrioritizedValidatorNumber := big.NewInt(11)

	candidateMap := createCandidateMap(candidates, isTrustedOrganizations)
	sortValidators(candidates, weights)

	updateIsTrustedOrganizations(candidates, isTrustedOrganizations, candidateMap)
	arrangeValidatorCandidates(candidates, newValidatorCount, isTrustedOrganizations, maxPrioritizedValidatorNumber)

	expectedCandidates := []common.Address{
		common.BytesToAddress([]byte{120}),
		common.BytesToAddress([]byte{119}),
		common.BytesToAddress([]byte{118}),
		common.BytesToAddress([]byte{117}),
		common.BytesToAddress([]byte{114}),
		common.BytesToAddress([]byte{113}),
		common.BytesToAddress([]byte{111}),
		common.BytesToAddress([]byte{110}),
		common.BytesToAddress([]byte{108}),
		common.BytesToAddress([]byte{105}),
		common.BytesToAddress([]byte{102}),

		common.BytesToAddress([]byte{116}),
		common.BytesToAddress([]byte{115}),
		common.BytesToAddress([]byte{112}),
		common.BytesToAddress([]byte{109}),
		common.BytesToAddress([]byte{107}),
		common.BytesToAddress([]byte{106}),
		common.BytesToAddress([]byte{104}),
		common.BytesToAddress([]byte{103}),
		common.BytesToAddress([]byte{101}),
		common.BytesToAddress([]byte{100}),
	}

	for i, candidate := range candidates {
		if !bytes.Equal(expectedCandidates[i].Bytes(), candidate.Bytes()) {
			t.Fatal(fmt.Sprintf("mismatched candidate address at %d, expected:%s got:%s", i, expectedCandidates[i].Hex(), candidate.Hex()))
		}
	}
}

// TestArrangeValidatorCandidates_Max5Prioritized arranges 21 candidates with 5 trusted nodes (ordering)
func TestArrangeValidatorCandidates_Max5Prioritized(t *testing.T) {
	scenarios := getTestScenarios()
	setIsTrustedAt(scenarios, 16, 17, 18, 19, 20)
	isTrustedOrganizations := getIsTrusted(scenarios)
	candidates := getAddresses(scenarios)
	weights := getWeights(scenarios)

	newValidatorCount := uint64(21)
	maxPrioritizedValidatorNumber := big.NewInt(11)

	candidateMap := createCandidateMap(candidates, isTrustedOrganizations)
	sortValidators(candidates, weights)
	updateIsTrustedOrganizations(candidates, isTrustedOrganizations, candidateMap)
	arrangeValidatorCandidates(candidates, newValidatorCount, isTrustedOrganizations, maxPrioritizedValidatorNumber)

	expectedCandidates := []common.Address{
		common.BytesToAddress([]byte{120}),
		common.BytesToAddress([]byte{119}),
		common.BytesToAddress([]byte{118}),
		common.BytesToAddress([]byte{117}),
		common.BytesToAddress([]byte{116}),
		common.BytesToAddress([]byte{115}),
		common.BytesToAddress([]byte{114}),
		common.BytesToAddress([]byte{113}),
		common.BytesToAddress([]byte{112}),
		common.BytesToAddress([]byte{111}),

		common.BytesToAddress([]byte{110}),
		common.BytesToAddress([]byte{109}),
		common.BytesToAddress([]byte{108}),
		common.BytesToAddress([]byte{107}),
		common.BytesToAddress([]byte{106}),
		common.BytesToAddress([]byte{105}),
		common.BytesToAddress([]byte{104}),
		common.BytesToAddress([]byte{103}),
		common.BytesToAddress([]byte{102}),
		common.BytesToAddress([]byte{101}),
		common.BytesToAddress([]byte{100}),
	}

	for i, candidate := range candidates {
		if !bytes.Equal(expectedCandidates[i].Bytes(), candidate.Bytes()) {
			t.Fatal(fmt.Sprintf("mismatched candidate address at %d, expected:%s got:%s", i, expectedCandidates[i].Hex(), candidate.Hex()))
		}
	}
}

// TestArrangeValidatorCandidates_Miss5Nodes arranges 16 candidates with 5 trusted nodes (ordering)
func TestArrangeValidatorCandidates_Miss5Nodes(t *testing.T) {
	scenarios := getTestScenarios()
	setIsTrustedAt(scenarios, 11, 12, 13, 14, 15)
	isTrustedOrganizations := getIsTrusted(scenarios)[5:]
	candidates := getAddresses(scenarios)[5:]
	weights := getWeights(scenarios)[5:]
	newValidatorCount := uint64(16)
	maxPrioritizedValidatorNumber := big.NewInt(11)

	candidateMap := createCandidateMap(candidates, isTrustedOrganizations)
	sortValidators(candidates, weights)
	updateIsTrustedOrganizations(candidates, isTrustedOrganizations, candidateMap)
	arrangeValidatorCandidates(candidates, newValidatorCount, isTrustedOrganizations, maxPrioritizedValidatorNumber)

	expectedCandidates := []common.Address{
		common.BytesToAddress([]byte{115}),
		common.BytesToAddress([]byte{114}),
		common.BytesToAddress([]byte{113}),
		common.BytesToAddress([]byte{112}),
		common.BytesToAddress([]byte{111}),

		common.BytesToAddress([]byte{120}),
		common.BytesToAddress([]byte{119}),
		common.BytesToAddress([]byte{118}),
		common.BytesToAddress([]byte{117}),
		common.BytesToAddress([]byte{116}),

		common.BytesToAddress([]byte{110}),
		common.BytesToAddress([]byte{109}),
		common.BytesToAddress([]byte{108}),
		common.BytesToAddress([]byte{107}),
		common.BytesToAddress([]byte{106}),
		common.BytesToAddress([]byte{105}),
	}
	for i, candidate := range candidates {
		if !bytes.Equal(expectedCandidates[i].Bytes(), candidate.Bytes()) {
			t.Fatal(fmt.Sprintf("mismatched candidate address at %d, expected:%s got:%s", i, expectedCandidates[i].Hex(), candidate.Hex()))
		}
	}
}

// TestArrangeValidatorCandidates_Has15TrustedNodes arranges 25 candidates with 15 trusted nodes (ordering)
// while maxPrioritizedValidatorNumber is 11
func TestArrangeValidatorCandidates_Has15TrustedNodes(t *testing.T) {
	scenarios := getTestScenarios()
	scenarios = append(scenarios, []TestScenario{
		{
			Address:   common.BytesToAddress([]byte{121}),
			Weight:    big.NewInt(8_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{122}),
			Weight:    big.NewInt(8_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{123}),
			Weight:    big.NewInt(8_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{124}),
			Weight:    big.NewInt(8_000_000),
			IsTrusted: big0,
		},
	}...)
	candidates := getAddresses(scenarios)
	weights := getWeights(scenarios)
	setIsTrustedAt(scenarios, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24)
	isTrustedOrganizations := getIsTrusted(scenarios)
	newValidatorCount := uint64(21)
	maxPrioritizedValidatorNumber := big.NewInt(11)

	candidateMap := createCandidateMap(candidates, isTrustedOrganizations)
	sortValidators(candidates, weights)
	updateIsTrustedOrganizations(candidates, isTrustedOrganizations, candidateMap)
	arrangeValidatorCandidates(candidates, newValidatorCount, isTrustedOrganizations, maxPrioritizedValidatorNumber)

	expectedCandidates := []common.Address{
		common.BytesToAddress([]byte{120}),
		common.BytesToAddress([]byte{119}),
		common.BytesToAddress([]byte{118}),
		common.BytesToAddress([]byte{117}),
		common.BytesToAddress([]byte{116}),
		common.BytesToAddress([]byte{115}),
		common.BytesToAddress([]byte{114}),
		common.BytesToAddress([]byte{113}),
		common.BytesToAddress([]byte{112}),
		common.BytesToAddress([]byte{111}),

		common.BytesToAddress([]byte{110}),
		common.BytesToAddress([]byte{109}),
		common.BytesToAddress([]byte{108}),
		common.BytesToAddress([]byte{124}),
		common.BytesToAddress([]byte{123}),
		common.BytesToAddress([]byte{122}),
		common.BytesToAddress([]byte{121}),
		common.BytesToAddress([]byte{107}),
		common.BytesToAddress([]byte{106}),
		common.BytesToAddress([]byte{105}),
		common.BytesToAddress([]byte{104}),
	}

	for i, candidate := range candidates[:newValidatorCount] {
		if !bytes.Equal(expectedCandidates[i].Bytes(), candidate.Bytes()) {
			t.Fatal(fmt.Sprintf("mismatched candidate address at %d, expected:%s got:%s", i, expectedCandidates[i].Hex(), candidate.Hex()))
		}
	}
}

// TestArrangeValidatorCandidates_Has15TrustedNodes arranges 25 candidates with 15 trusted nodes (ordering)
// while maxPrioritizedValidatorNumber is 11
func TestArrangeValidatorCandidates_TrustedNodesAtBeginningArray(t *testing.T) {
	scenarios := getTestScenarios()
	scenarios = append(scenarios, []TestScenario{
		{
			Address:   common.BytesToAddress([]byte{121}),
			Weight:    big.NewInt(8_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{122}),
			Weight:    big.NewInt(8_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{123}),
			Weight:    big.NewInt(8_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{124}),
			Weight:    big.NewInt(80_000_000),
			IsTrusted: big0,
		},
	}...)
	candidates := getAddresses(scenarios)
	weights := getWeights(scenarios)
	setIsTrustedAt(scenarios, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	isTrustedOrganizations := getIsTrusted(scenarios)
	newValidatorCount := uint64(21)
	maxPrioritizedValidatorNumber := big.NewInt(11)

	candidateMap := createCandidateMap(candidates, isTrustedOrganizations)
	sortValidators(candidates, weights)
	updateIsTrustedOrganizations(candidates, isTrustedOrganizations, candidateMap)
	arrangeValidatorCandidates(candidates, newValidatorCount, isTrustedOrganizations, maxPrioritizedValidatorNumber)

	expectedCandidates := []common.Address{
		common.BytesToAddress([]byte{110}),
		common.BytesToAddress([]byte{109}),
		common.BytesToAddress([]byte{108}),
		common.BytesToAddress([]byte{107}),
		common.BytesToAddress([]byte{106}),
		common.BytesToAddress([]byte{105}),
		common.BytesToAddress([]byte{104}),
		common.BytesToAddress([]byte{103}),
		common.BytesToAddress([]byte{102}),
		common.BytesToAddress([]byte{101}),
		common.BytesToAddress([]byte{100}),

		common.BytesToAddress([]byte{124}),
		common.BytesToAddress([]byte{120}),
		common.BytesToAddress([]byte{119}),
		common.BytesToAddress([]byte{118}),
		common.BytesToAddress([]byte{117}),
		common.BytesToAddress([]byte{116}),
		common.BytesToAddress([]byte{115}),
		common.BytesToAddress([]byte{114}),
		common.BytesToAddress([]byte{113}),
		common.BytesToAddress([]byte{112}),
		common.BytesToAddress([]byte{103}),
	}
	for i, candidate := range candidates[:newValidatorCount] {
		if !bytes.Equal(expectedCandidates[i].Bytes(), candidate.Bytes()) {
			t.Fatal(fmt.Sprintf("mismatched candidate address at %d, expected:%s got:%s", i, expectedCandidates[i].Hex(), candidate.Hex()))
		}
	}
}

// TestConsortiumPickValidatorSet_Run init 2 headers, pack them and call `Run` function directly
// Returns 21 candidates if candidates have 21 nodes and 11/21 trusted nodes
func TestConsortiumPickValidatorSet_Run(t *testing.T) {
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)

	smcAbi, err := abi.JSON(strings.NewReader(consortiumPickValidatorSetAbi))
	if err != nil {
		t.Fatal(err)
	}
	scenarios := getTestScenarios()
	candidates := getAddresses(scenarios)
	weights := getWeights(scenarios)
	setIsTrustedAt(scenarios, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20)
	isTrustedOrganizations := getIsTrusted(scenarios)
	maxValidatorNumber := big.NewInt(21)
	maxPrioritizedValidatorNumber := big.NewInt(11)

	expectedCandidates := []common.Address{
		common.BytesToAddress([]byte{120}),
		common.BytesToAddress([]byte{119}),
		common.BytesToAddress([]byte{118}),
		common.BytesToAddress([]byte{117}),
		common.BytesToAddress([]byte{116}),
		common.BytesToAddress([]byte{115}),
		common.BytesToAddress([]byte{114}),
		common.BytesToAddress([]byte{113}),
		common.BytesToAddress([]byte{112}),
		common.BytesToAddress([]byte{111}),
		common.BytesToAddress([]byte{110}),

		common.BytesToAddress([]byte{109}),
		common.BytesToAddress([]byte{108}),
		common.BytesToAddress([]byte{107}),
		common.BytesToAddress([]byte{106}),
		common.BytesToAddress([]byte{105}),
		common.BytesToAddress([]byte{104}),
		common.BytesToAddress([]byte{103}),
		common.BytesToAddress([]byte{102}),
		common.BytesToAddress([]byte{101}),
		common.BytesToAddress([]byte{100}),
	}

	input, err := smcAbi.Pack(pickValidatorSetMethod, candidates, weights, isTrustedOrganizations, maxValidatorNumber, maxPrioritizedValidatorNumber)

	if err != nil {
		t.Fatal(err)
	}

	evm, err := newEVM(caller, statedb)
	if err != nil {
		t.Fatal(err)
	}
	c := &consortiumPickValidatorSet{caller: AccountRef(caller), evm: evm}
	output, err := c.Run(input)
	if err != nil {
		t.Fatal(err)
	}

	res, err := smcAbi.Methods[pickValidatorSetMethod].Outputs.Unpack(output)
	if err != nil {
		t.Fatal(err)
	}
	validators := *abi.ConvertType(res[0], new([21]common.Address)).(*[21]common.Address)
	if len(expectedCandidates) != len(validators) {
		t.Fatal(fmt.Sprintf("expected len %d, got %v", len(expectedCandidates), len(validators)))
	}

	for i, addr := range validators {
		if expectedCandidates[i].Hex() != addr.Hex() {
			t.Fatal(fmt.Sprintf("mismatched addr at %d, expected:%s got:%s", i, expectedCandidates[i].Hex(), addr.Hex()))
		}
	}
}

// TestConsortiumPickValidatorSet_Run2 init 2 headers, pack them and call `Run` function directly
// Returns 15 candidates if candidates have 15 nodes and 5/15 trusted nodes
func TestConsortiumPickValidatorSet_Run2(t *testing.T) {
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)

	smcAbi, err := abi.JSON(strings.NewReader(consortiumPickValidatorSetAbi))
	if err != nil {
		t.Fatal(err)
	}
	scenarios := getTestScenarios()
	candidates := getAddresses(scenarios)[:15]
	weights := getWeights(scenarios)[:15]
	setIsTrustedAt(scenarios, 10, 11, 12, 13, 14)
	isTrustedOrganizations := getIsTrusted(scenarios)[:15]
	maxValidatorNumber := big.NewInt(21)
	maxPrioritizedValidatorNumber := big.NewInt(11)

	expectedCandidates := []common.Address{
		common.BytesToAddress([]byte{114}),
		common.BytesToAddress([]byte{113}),
		common.BytesToAddress([]byte{112}),
		common.BytesToAddress([]byte{111}),
		common.BytesToAddress([]byte{110}),

		common.BytesToAddress([]byte{109}),
		common.BytesToAddress([]byte{108}),
		common.BytesToAddress([]byte{107}),
		common.BytesToAddress([]byte{106}),
		common.BytesToAddress([]byte{105}),
		common.BytesToAddress([]byte{104}),
		common.BytesToAddress([]byte{103}),
		common.BytesToAddress([]byte{102}),
		common.BytesToAddress([]byte{101}),
		common.BytesToAddress([]byte{100}),
	}

	input, err := smcAbi.Pack(pickValidatorSetMethod, candidates, weights, isTrustedOrganizations, maxValidatorNumber, maxPrioritizedValidatorNumber)

	if err != nil {
		t.Fatal(err)
	}

	evm, err := newEVM(caller, statedb)
	if err != nil {
		t.Fatal(err)
	}
	c := &consortiumPickValidatorSet{caller: AccountRef(caller), evm: evm}
	output, err := c.Run(input)
	if err != nil {
		t.Fatal(err)
	}

	res, err := smcAbi.Methods[pickValidatorSetMethod].Outputs.Unpack(output)
	if err != nil {
		t.Fatal(err)
	}
	validators := *abi.ConvertType(res[0], new([15]common.Address)).(*[15]common.Address)
	if len(expectedCandidates) != len(validators) {
		t.Fatal(fmt.Sprintf("expected len %d, got %v", len(expectedCandidates), len(validators)))
	}

	for i, addr := range validators {
		if expectedCandidates[i].Hex() != addr.Hex() {
			t.Fatal(fmt.Sprintf("mismatched addr at %d, expected:%s got:%s", i, expectedCandidates[i].Hex(), addr.Hex()))
		}
	}
}

// TestConsortiumPickValidatorSet_Run3 init 2 headers, pack them and call `Run` function directly
// If the length of trusted nodes reach the maxPrioritizedValidatorNumber, then the other trusted nodes
// will be treated as normal nodes
// Returns 21 candidates if candidates have 25 nodes and 15/25 trusted nodes
func TestConsortiumPickValidatorSet_Run3(t *testing.T) {
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)

	smcAbi, err := abi.JSON(strings.NewReader(consortiumPickValidatorSetAbi))
	if err != nil {
		t.Fatal(err)
	}
	scenarios := getTestScenarios()
	scenarios = append(scenarios, []TestScenario{
		{
			Address:   common.BytesToAddress([]byte{121}),
			Weight:    big.NewInt(8_000_000),
			IsTrusted: big1,
		},
		{
			Address:   common.BytesToAddress([]byte{122}),
			Weight:    big.NewInt(8_000_000),
			IsTrusted: big1,
		},
		{
			Address:   common.BytesToAddress([]byte{123}),
			Weight:    big.NewInt(8_000_000),
			IsTrusted: big1,
		},
		{
			Address:   common.BytesToAddress([]byte{124}),
			Weight:    big.NewInt(80_000_000), // must be the most priority since the weight is largest
			IsTrusted: big1,
		},
	}...)
	candidates := getAddresses(scenarios)
	weights := getWeights(scenarios)
	setIsTrustedAt(scenarios, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24)
	isTrustedOrganizations := getIsTrusted(scenarios)
	maxValidatorNumber := big.NewInt(21)
	maxPrioritizedValidatorNumber := big.NewInt(11)

	expectedCandidates := []common.Address{
		common.BytesToAddress([]byte{124}),
		common.BytesToAddress([]byte{120}),
		common.BytesToAddress([]byte{119}),
		common.BytesToAddress([]byte{118}),
		common.BytesToAddress([]byte{117}),
		common.BytesToAddress([]byte{116}),
		common.BytesToAddress([]byte{115}),
		common.BytesToAddress([]byte{114}),
		common.BytesToAddress([]byte{113}),
		common.BytesToAddress([]byte{112}),
		common.BytesToAddress([]byte{111}),
		common.BytesToAddress([]byte{110}),
		common.BytesToAddress([]byte{109}),
		common.BytesToAddress([]byte{108}),

		common.BytesToAddress([]byte{123}),
		common.BytesToAddress([]byte{122}),
		common.BytesToAddress([]byte{121}),

		common.BytesToAddress([]byte{107}),
		common.BytesToAddress([]byte{106}),
		common.BytesToAddress([]byte{105}),
		common.BytesToAddress([]byte{104}),
	}

	input, err := smcAbi.Pack(pickValidatorSetMethod, candidates, weights, isTrustedOrganizations, maxValidatorNumber, maxPrioritizedValidatorNumber)

	if err != nil {
		t.Fatal(err)
	}

	evm, err := newEVM(caller, statedb)
	if err != nil {
		t.Fatal(err)
	}
	c := &consortiumPickValidatorSet{caller: AccountRef(caller), evm: evm}
	output, err := c.Run(input)
	if err != nil {
		t.Fatal(err)
	}

	res, err := smcAbi.Methods[pickValidatorSetMethod].Outputs.Unpack(output)
	if err != nil {
		t.Fatal(err)
	}
	validators := *abi.ConvertType(res[0], new([21]common.Address)).(*[21]common.Address)
	fmt.Println(addressesToByte(validators[:]))
	if len(expectedCandidates) != len(validators) {
		t.Fatal(fmt.Sprintf("expected len %d, got %v", len(expectedCandidates), len(validators)))
	}

	for i, addr := range validators {
		if expectedCandidates[i].Hex() != addr.Hex() {
			t.Fatal(fmt.Sprintf("mismatched addr at %d, expected:%s got:%s", i, expectedCandidates[i].Hex(), addr.Hex()))
		}
	}
}

// TestConsortiumPickValidatorSet_Run4 init 2 headers, pack them and call `Run` function directly
// Returns 21 candidates if candidates have 25 nodes and 15/25 trusted nodes
// If the length of trusted nodes reach the maxPrioritizedValidatorNumber, then the other trusted nodes
// will be treated as normal nodes
// In case all trusted nodes is at the beginning of the array, the waitingCandidates must not out of bound
func TestConsortiumPickValidatorSet_Run4(t *testing.T) {
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)

	smcAbi, err := abi.JSON(strings.NewReader(consortiumPickValidatorSetAbi))
	if err != nil {
		t.Fatal(err)
	}
	scenarios := getTestScenarios()
	scenarios = append(scenarios, []TestScenario{
		{
			Address:   common.BytesToAddress([]byte{121}),
			Weight:    big.NewInt(8_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{122}),
			Weight:    big.NewInt(8_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{123}),
			Weight:    big.NewInt(8_000_000),
			IsTrusted: big0,
		},
		{
			Address:   common.BytesToAddress([]byte{124}),
			Weight:    big.NewInt(80_000_000), // must be the most priority since the weight is largest
			IsTrusted: big0,
		},
	}...)
	candidates := getAddresses(scenarios)
	weights := getWeights(scenarios)
	setIsTrustedAt(scenarios, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	isTrustedOrganizations := getIsTrusted(scenarios)
	maxValidatorNumber := big.NewInt(21)
	maxPrioritizedValidatorNumber := big.NewInt(11)

	expectedCandidates := []common.Address{
		common.BytesToAddress([]byte{110}),
		common.BytesToAddress([]byte{109}),
		common.BytesToAddress([]byte{108}),
		common.BytesToAddress([]byte{107}),
		common.BytesToAddress([]byte{106}),
		common.BytesToAddress([]byte{105}),
		common.BytesToAddress([]byte{104}),
		common.BytesToAddress([]byte{103}),
		common.BytesToAddress([]byte{102}),
		common.BytesToAddress([]byte{101}),
		common.BytesToAddress([]byte{100}),

		common.BytesToAddress([]byte{124}),
		common.BytesToAddress([]byte{120}),
		common.BytesToAddress([]byte{119}),
		common.BytesToAddress([]byte{118}),
		common.BytesToAddress([]byte{117}),
		common.BytesToAddress([]byte{116}),
		common.BytesToAddress([]byte{115}),
		common.BytesToAddress([]byte{114}),
		common.BytesToAddress([]byte{113}),
		common.BytesToAddress([]byte{112}),
	}

	input, err := smcAbi.Pack(pickValidatorSetMethod, candidates, weights, isTrustedOrganizations, maxValidatorNumber, maxPrioritizedValidatorNumber)

	if err != nil {
		t.Fatal(err)
	}

	evm, err := newEVM(caller, statedb)
	if err != nil {
		t.Fatal(err)
	}
	c := &consortiumPickValidatorSet{caller: AccountRef(caller), evm: evm}
	output, err := c.Run(input)
	if err != nil {
		t.Fatal(err)
	}

	res, err := smcAbi.Methods[pickValidatorSetMethod].Outputs.Unpack(output)
	if err != nil {
		t.Fatal(err)
	}
	validators := *abi.ConvertType(res[0], new([21]common.Address)).(*[21]common.Address)
	if len(expectedCandidates) != len(validators) {
		t.Fatal(fmt.Sprintf("expected len %d, got %v", len(expectedCandidates), len(validators)))
	}

	for i, addr := range validators {
		if expectedCandidates[i].Hex() != addr.Hex() {
			t.Fatal(fmt.Sprintf("mismatched addr at %d, expected:%s got:%s", i, expectedCandidates[i].Hex(), addr.Hex()))
		}
	}
}

// TestConsortiumPickValidatorSet_Run5 simulates a call from a user who trigger system contract
// to call `sort` precompiled contract
func TestConsortiumPickValidatorSet_Run5(t *testing.T) {
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	smcAbi, err := abi.JSON(strings.NewReader(syncNewValidatorSetAbi))
	if err != nil {
		t.Fatal(err)
	}
	input, err := smcAbi.Pack(pickValidatorSetMethod)
	if err != nil {
		t.Fatal(err)
	}

	evm, err := newEVM(caller, statedb)
	if err != nil {
		t.Fatal(err)
	}
	_, contract, _, err := evm.Create(AccountRef(caller), common.FromHex(syncNewValidatorSetCode), math.MaxUint64/2, big0)
	if err != nil {
		t.Fatal(err)
	}

	// set this contract into consortiumV2Contracts to make it become a system contract
	evm.chainConfig.ConsortiumV2Contracts.SlashIndicator = contract

	ret, _, err := evm.StaticCall(AccountRef(caller), contract, input, 1000000)
	if err != nil {
		t.Fatal(err)
	}

	res, err := smcAbi.Methods[pickValidatorSetMethod].Outputs.Unpack(ret)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(res))
	validators := *abi.ConvertType(res[0], new([21]common.Address)).(*[21]common.Address)
	count := *abi.ConvertType(res[1], new(big.Int)).(*big.Int)
	fmt.Println(validators)
	fmt.Println(count.Uint64())

	expectedCandidates := []common.Address{
		common.BytesToAddress([]byte{120}),
		common.BytesToAddress([]byte{119}),
		common.BytesToAddress([]byte{118}),
		common.BytesToAddress([]byte{117}),
		common.BytesToAddress([]byte{116}),
		common.BytesToAddress([]byte{115}),
		common.BytesToAddress([]byte{114}),
		common.BytesToAddress([]byte{113}),
		common.BytesToAddress([]byte{112}),
		common.BytesToAddress([]byte{111}),
		common.BytesToAddress([]byte{110}),

		common.BytesToAddress([]byte{109}),
		common.BytesToAddress([]byte{108}),
		common.BytesToAddress([]byte{107}),
		common.BytesToAddress([]byte{106}),
		common.BytesToAddress([]byte{105}),
		common.BytesToAddress([]byte{104}),
		common.BytesToAddress([]byte{103}),
		common.BytesToAddress([]byte{102}),
		common.BytesToAddress([]byte{101}),
		common.BytesToAddress([]byte{100}),
	}

	if len(expectedCandidates) != len(validators) {
		t.Fatal(fmt.Sprintf("expected len 21, got %v", len(validators)))
	}

	for i, addr := range validators {
		if expectedCandidates[i].Hex() != addr.Hex() {
			t.Fatal(fmt.Sprintf("mismatched addr at %d, expected:%s got:%s", i, expectedCandidates[i].Hex(), addr.Hex()))
		}
	}
}

func prepareHeader(chainId *big.Int) (*types.Header, *types.Header, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, nil, err
	}
	// init extraData with extraVanity
	extraData := bytes.Repeat([]byte{0x00}, extraVanity)

	// append to extraData with validators set
	extraData = append(extraData, common.BytesToAddress([]byte("validator1")).Bytes()...)
	extraData = append(extraData, common.BytesToAddress([]byte("validator2")).Bytes()...)

	// add extra seal space
	extraData = append(extraData, make([]byte, crypto.SignatureLength)...)

	// create header1
	header1 := &types.Header{
		ParentHash:  common.BytesToHash([]byte("11")),
		UncleHash:   common.Hash{},
		Coinbase:    crypto.PubkeyToAddress(privateKey.PublicKey),
		Root:        common.BytesToHash([]byte("123")),
		TxHash:      common.BytesToHash([]byte("abc")),
		ReceiptHash: common.BytesToHash([]byte("def")),
		Bloom:       types.Bloom{},
		Difficulty:  big.NewInt(1000),
		Number:      big.NewInt(1000),
		GasLimit:    100000000,
		GasUsed:     0,
		Time:        1000,
		Extra:       make([]byte, len(extraData)),
		MixDigest:   common.Hash{},
		Nonce:       types.EncodeNonce(1000),
	}

	// create header2
	header2 := &types.Header{
		ParentHash:  common.BytesToHash([]byte("11")),
		UncleHash:   common.Hash{},
		Coinbase:    crypto.PubkeyToAddress(privateKey.PublicKey),
		Root:        common.BytesToHash([]byte("1232")),
		TxHash:      common.BytesToHash([]byte("abcd")),
		ReceiptHash: common.BytesToHash([]byte("defd")),
		Bloom:       types.Bloom{},
		Difficulty:  big.NewInt(1000),
		Number:      big.NewInt(1000),
		GasLimit:    100000000,
		GasUsed:     0,
		Time:        1000,
		Extra:       make([]byte, len(extraData)),
		MixDigest:   common.Hash{},
		Nonce:       types.EncodeNonce(1000),
	}

	// copy extraData
	copy(header1.Extra[:], extraData)
	copy(header2.Extra[:], extraData)

	// signing and add to extraData
	sig1, err := crypto.Sign(crypto.Keccak256(consortiumRlp(header1, chainId)), privateKey)
	if err != nil {
		return nil, nil, err
	}
	sig2, err := crypto.Sign(crypto.Keccak256(consortiumRlp(header2, chainId)), privateKey)
	if err != nil {
		return nil, nil, err
	}

	copy(header1.Extra[len(header1.Extra)-crypto.SignatureLength:], sig1)
	copy(header2.Extra[len(header2.Extra)-crypto.SignatureLength:], sig2)

	return header1, header2, nil
}

func consortiumRlp(header *types.Header, chainId *big.Int) []byte {
	b := new(bytes.Buffer)
	encodeSigHeader(b, header, chainId)
	return b.Bytes()
}

func newEVM(caller common.Address, statedb StateDB) (*EVM, error) {
	evm := &EVM{
		Context: BlockContext{
			CanTransfer: func(state StateDB, addr common.Address, value *big.Int) bool { return true },
			Transfer:    func(StateDB, common.Address, common.Address, *big.Int) {},
		},
		chainConfig: params.TestChainConfig,
		StateDB:     statedb,
		chainRules:  params.Rules{IsIstanbul: true, IsEIP150: true},
	}
	evm.interpreter = NewEVMInterpreter(evm, Config{NoBaseFee: true})
	_, contract, _, err := evm.Create(AccountRef(caller), common.FromHex(testSortCode), math.MaxUint64/2, big0)
	if err != nil {
		return nil, err
	}
	evm.chainConfig.ConsortiumV2Contracts = &params.ConsortiumV2Contracts{
		StakingContract:   contract,
		RoninValidatorSet: contract,
		SlashIndicator:    caller,
	}
	return evm, nil
}

func addressesToByte(addresses []common.Address) [][]byte {
	var result [][]byte
	for _, addr := range addresses {
		var value []byte
		for _, b := range addr.Bytes() {
			if b != 0 {
				value = append(value, b)
			}
		}
		result = append(result, value)
	}

	return result
}
