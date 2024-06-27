package tracers

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/internal/ethapi"
	"strings"
)

var (
	_jsondata = `
[
	{
		"inputs": [
			{
				"internalType": "bytes",
				"name": "data",
				"type": "bytes"
			}
		],
		"name": "_decodeBalance",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "pure",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address[]",
				"name": "tokens",
				"type": "address[]"
			}
		],
		"name": "balance",
		"outputs": [
			{
				"internalType": "bool[]",
				"name": "",
				"type": "bool[]"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address[]",
				"name": "tokens",
				"type": "address[]"
			},
			{
				"internalType": "address[][]",
				"name": "users",
				"type": "address[][]"
			}
		],
		"name": "tokenBalance",
		"outputs": [
			{
				"internalType": "uint256[][]",
				"name": "",
				"type": "uint256[][]"
			}
		],
		"stateMutability": "view",
		"type": "function"
	}
]
`
	_deployCode = "0x608060405234801561001057600080fd5b50600436106100415760003560e01c806328523bc314610046578063cadda4ba14610076578063dfca5f24146100a6575b600080fd5b610060600480360361005b91908101906107cc565b6100d6565b60405161006d9190610aef565b60405180910390f35b610090600480360361008b919081019061078b565b6103ef565b60405161009d9190610b11565b60405180910390f35b6100c060048036036100bb9190810190610838565b610576565b6040516100cd9190610b55565b60405180910390f35b606080835160405190808252806020026020018201604052801561010e57816020015b60608152602001906001900390816100f95790505b50905060008090505b84518110156103e457606084828151811061012e57fe5b602002602001015190506060815160405190808252806020026020018201604052801561016a5781602001602082028038833980820191505090505b50905060008090505b82518110156103bc576000606089868151811061018c57fe5b602002602001015173ffffffffffffffffffffffffffffffffffffffff168584815181106101b657fe5b60200260200101516040516024016101ce9190610ad4565b6040516020818303038152906040527f70a08231000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff19166020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff83818316178352505050506040516102589190610abd565b600060405180830381855afa9150503d8060008114610293576040519150601f19603f3d011682016040523d82523d6000602084013e610298565b606091505b50915091508115610392576000815114156102cc5760008484815181106102bb57fe5b60200260200101818152505061038d565b3073ffffffffffffffffffffffffffffffffffffffff1663dfca5f24826040518263ffffffff1660e01b81526004016103059190610b33565b60206040518083038186803b15801561031d57600080fd5b505afa92505050801561034e57506040513d601f19601f8201168201806040525061034b9190810190610879565b60015b61037157600084848151811061036057fe5b60200260200101818152505061038c565b8085858151811061037e57fe5b602002602001018181525050505b5b6103ad565b60008484815181106103a057fe5b6020026020010181815250505b50508080600101915050610173565b50808484815181106103ca57fe5b602002602001018190525050508080600101915050610117565b508091505092915050565b60608082516040519080825280602002602001820160405280156104225781602001602082028038833980820191505090505b50905060008090505b835181101561056c57600084828151811061044257fe5b602002602001015173ffffffffffffffffffffffffffffffffffffffff16306040516024016104719190610ad4565b6040516020818303038152906040527f70a08231000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff19166020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff83818316178352505050506040516104fb9190610abd565b600060405180830381855afa9150503d8060008114610536576040519150601f19603f3d011682016040523d82523d6000602084013e61053b565b606091505b505090508083838151811061054c57fe5b60200260200101901515908115158152505050808060010191505061042b565b5080915050919050565b60008180602001905161058c9190810190610879565b9050919050565b6000813590506105a281610dae565b92915050565b600082601f8301126105b957600080fd5b81356105cc6105c782610b9d565b610b70565b915081818352602084019350602081019050838560208402820111156105f157600080fd5b60005b8381101561062157816106078882610593565b8452602084019350602083019250506001810190506105f4565b5050505092915050565b600082601f83011261063c57600080fd5b813561064f61064a82610bc5565b610b70565b9150818183526020840193506020810190508385602084028201111561067457600080fd5b60005b838110156106a4578161068a8882610593565b845260208401935060208301925050600181019050610677565b5050505092915050565b600082601f8301126106bf57600080fd5b81356106d26106cd82610bed565b610b70565b9150818183526020840193506020810190508360005b8381101561071857813586016106fe88826105a8565b8452602084019350602083019250506001810190506106e8565b5050505092915050565b600082601f83011261073357600080fd5b813561074661074182610c15565b610b70565b9150808252602083016020830185838301111561076257600080fd5b61076d838284610d5b565b50505092915050565b60008151905061078581610dc5565b92915050565b60006020828403121561079d57600080fd5b600082013567ffffffffffffffff8111156107b757600080fd5b6107c38482850161062b565b91505092915050565b600080604083850312156107df57600080fd5b600083013567ffffffffffffffff8111156107f957600080fd5b6108058582860161062b565b925050602083013567ffffffffffffffff81111561082257600080fd5b61082e858286016106ae565b9150509250929050565b60006020828403121561084a57600080fd5b600082013567ffffffffffffffff81111561086457600080fd5b61087084828501610722565b91505092915050565b60006020828403121561088b57600080fd5b600061089984828501610776565b91505092915050565b60006108ae83836109c8565b905092915050565b60006108c28383610a26565b60208301905092915050565b60006108da8383610a9f565b60208301905092915050565b6108ef81610d13565b82525050565b600061090082610c71565b61090a8185610cc4565b93508360208202850161091c85610c41565b8060005b85811015610958578484038952815161093985826108a2565b945061094483610c9d565b925060208a01995050600181019050610920565b50829750879550505050505092915050565b600061097582610c7c565b61097f8185610cd5565b935061098a83610c51565b8060005b838110156109bb5781516109a288826108b6565b97506109ad83610caa565b92505060018101905061098e565b5085935050505092915050565b60006109d382610c87565b6109dd8185610ce6565b93506109e883610c61565b8060005b83811015610a19578151610a0088826108ce565b9750610a0b83610cb7565b9250506001810190506109ec565b5085935050505092915050565b610a2f81610d25565b82525050565b6000610a4082610c92565b610a4a8185610cf7565b9350610a5a818560208601610d6a565b610a6381610d9d565b840191505092915050565b6000610a7982610c92565b610a838185610d08565b9350610a93818560208601610d6a565b80840191505092915050565b610aa881610d51565b82525050565b610ab781610d51565b82525050565b6000610ac98284610a6e565b915081905092915050565b6000602082019050610ae960008301846108e6565b92915050565b60006020820190508181036000830152610b0981846108f5565b905092915050565b60006020820190508181036000830152610b2b818461096a565b905092915050565b60006020820190508181036000830152610b4d8184610a35565b905092915050565b6000602082019050610b6a6000830184610aae565b92915050565b6000604051905081810181811067ffffffffffffffff82111715610b9357600080fd5b8060405250919050565b600067ffffffffffffffff821115610bb457600080fd5b602082029050602081019050919050565b600067ffffffffffffffff821115610bdc57600080fd5b602082029050602081019050919050565b600067ffffffffffffffff821115610c0457600080fd5b602082029050602081019050919050565b600067ffffffffffffffff821115610c2c57600080fd5b601f19601f8301169050602081019050919050565b6000819050602082019050919050565b6000819050602082019050919050565b6000819050602082019050919050565b600081519050919050565b600081519050919050565b600081519050919050565b600081519050919050565b6000602082019050919050565b6000602082019050919050565b6000602082019050919050565b600082825260208201905092915050565b600082825260208201905092915050565b600082825260208201905092915050565b600082825260208201905092915050565b600081905092915050565b6000610d1e82610d31565b9050919050565b60008115159050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b82818337600083830152505050565b60005b83811015610d88578082015181840152602081019050610d6d565b83811115610d97576000848401525b50505050565b6000601f19601f8301169050919050565b610db781610d13565b8114610dc257600080fd5b50565b610dce81610d51565b8114610dd957600080fd5b5056fea2646970667358221220f9b9f5f4ef7f72e17db320a4f779b0d67dd72824394842e6895523dbf78709e464736f6c63430006000033"
	_caller     = "0x1992111111111111111111111111111111111110"
	_contract   = "0x1992111111111111111111111111111111111111"
)

type TokenContract struct {
	caller  common.Address
	address common.Address
	code    []byte
	abi     abi.ABI
}

func NewTokenContract() TokenContract {
	_abi, _ := abi.JSON(strings.NewReader(_jsondata))
	return TokenContract{
		caller:  common.HexToAddress(_caller),
		address: common.HexToAddress(_contract),
		code:    hexutil.MustDecode(_deployCode),
		abi:     _abi,
	}
}

func (tc *TokenContract) Override() *ethapi.StateOverride {
	return &ethapi.StateOverride{
		tc.address: ethapi.OverrideAccount{
			Nonce:     nil,
			Code:      (*hexutil.Bytes)(&tc.code),
			Balance:   nil,
			State:     nil,
			StateDiff: nil,
		},
	}
}
