package native

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/eth/tracers/internal"
)

func init() {
	tracers.DefaultDirectory.Register("tokenBalanceTracer", newContractTracer, false)
}

type tokenBalanceTracer struct {
	contracts    map[common.Address]map[string]struct{}
	topContracts map[common.Address]common.Address // contractAddress: topContractAddress
	checkTop     bool
}

func newContractTracer(ctx *tracers.Context, cfg json.RawMessage) (*tracers.Tracer, error) {
	ct := &tokenBalanceTracer{
		contracts:    make(map[common.Address]map[string]struct{}),
		topContracts: make(map[common.Address]common.Address),
		checkTop:     false,
	}
	return &tracers.Tracer{
		Hooks: &tracing.Hooks{
			OnOpcode: ct.OnOpcode,
		},
		GetResult: ct.GetResult,
	}, nil
}

func (ct *tokenBalanceTracer) OnOpcode(pc uint64, opcode byte, gas, cost uint64, scope tracing.OpContext, rData []byte, depth int, err error) {
	if err != nil {
		return
	}
	op := vm.OpCode(opcode)
	stackData := scope.StackData()
	stackLen := len(stackData)

	contractAddress := scope.Address()
	if _, ok := ct.contracts[contractAddress]; !ok {
		ct.contracts[contractAddress] = make(map[string]struct{})
	}
	if ct.checkTop {
		caller := scope.Caller()
		if caller != tracers.TokenContractCaller && caller != tracers.TokenContractAddress {
			if _, ok := ct.topContracts[contractAddress]; !ok {
				if topContract, hasCaller := ct.topContracts[caller]; !hasCaller {
					ct.topContracts[contractAddress] = caller
				} else {
					ct.topContracts[contractAddress] = topContract
				}
			}
		}
	}

	switch {
	case stackLen >= 2 && op == vm.KECCAK256:
		offset := stackData[stackLen-1]
		size := stackData[stackLen-2]
		data, err := internal.GetMemoryCopyPadded(scope.MemoryData(), int64(offset.Uint64()), int64(size.Uint64()))
		if err != nil {
			return
		}
		if _, ok := ct.contracts[contractAddress]; !ok {
			ct.contracts[contractAddress] = make(map[string]struct{})
		}
		ct.contracts[contractAddress][hexutil.Encode(data)] = struct{}{}
	}
}

type TokenBalanceResult struct {
	Contracts    map[common.Address][]string       `json:"contracts"`
	TopContracts map[common.Address]common.Address `json:"topContracts"`
}

func (ct *tokenBalanceTracer) GetResult() (json.RawMessage, error) {
	// remove empty key
	for k, v := range ct.contracts {
		if len(v) == 0 {
			delete(ct.contracts, k)
		}
	}

	contracts := make(map[common.Address][]string)
	for k, vs := range ct.contracts {
		contracts[k] = make([]string, 0)
		for v := range vs {
			contracts[k] = append(contracts[k], v)
		}
	}

	tbr := TokenBalanceResult{
		Contracts:    contracts,
		TopContracts: ct.topContracts,
	}

	res, err := json.Marshal(tbr)
	if err != nil {
		return nil, err
	}
	// clear result
	ct.contracts = make(map[common.Address]map[string]struct{})
	ct.topContracts = make(map[common.Address]common.Address)

	return json.RawMessage(res), nil
}
