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
	tracers.DefaultDirectory.Register("contractTracer", newContractTracer, false)
}

type contractTracer struct {
	contracts map[common.Address]map[string]struct{}
}

func newContractTracer(ctx *tracers.Context, cfg json.RawMessage) (*tracers.Tracer, error) {
	ct := &contractTracer{
		contracts: make(map[common.Address]map[string]struct{}),
	}
	return &tracers.Tracer{
		Hooks: &tracing.Hooks{
			OnOpcode: ct.OnOpcode,
		},
		GetResult: ct.GetResult,
	}, nil
}

func (ct *contractTracer) OnOpcode(pc uint64, opcode byte, gas, cost uint64, scope tracing.OpContext, rData []byte, depth int, err error) {
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

func (ct *contractTracer) GetResult() (json.RawMessage, error) {
	// remove empty key
	for k, v := range ct.contracts {
		if len(v) == 0 {
			delete(ct.contracts, k)
		}
	}

	res, err := json.Marshal(ct.contracts)
	if err != nil {
		return nil, err
	}

	// clear result
	ct.contracts = make(map[common.Address]map[string]struct{})

	return json.RawMessage(res), nil
}
