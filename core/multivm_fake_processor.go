// +build !sputnikvm

package core

import (
	"math/big"

	"github.com/demedallo/demedallod/core/state"
	"github.com/demedallo/demedallod/core/types"
	evm "github.com/demedallo/demedallod/core/vm"
)

const SputnikVMExists = false

var UseSputnikVM = false

func ApplyMultiVmTransaction(config *ChainConfig, bc *BlockChain, gp *GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, totalUsedGas *big.Int) (*types.Receipt, evm.Logs, *big.Int, error) {
	panic("not implemented")
}
