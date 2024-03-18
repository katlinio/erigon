package searcher

import (
	"fmt"
	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/common/hexutility"
	types2 "github.com/ledgerwatch/erigon-lib/types"
	"github.com/ledgerwatch/erigon/core/state"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/core/vm/evmtypes"
	"github.com/ledgerwatch/erigon/rpc"
	"math/big"
)

type OverrideAccount struct {
	Nonce     *uint64                      `json:"nonce,omitempty"`
	Code      *hexutility.Bytes            `json:"code,omitempty"`
	Balance   *big.Int                     `json:"balance,omitempty"`
	State     *map[common.Hash]common.Hash `json:"state,omitempty"`
	StateDiff *map[common.Hash]common.Hash `json:"stateDiff,omitempty"`
}

type StateOverride map[common.Address]OverrideAccount

// Apply overrides the fields of specified accounts into the given state.
func (diff *StateOverride) Apply(s *state.IntraBlockState) error {
	if diff == nil {
		return nil
	}
	for addr, account := range *diff {
		// Override account nonce.
		if account.Nonce != nil {
			s.SetNonce(addr, *account.Nonce)
		}
		// Override account(contract) code.
		if account.Code != nil {
			s.SetCode(addr, *account.Code)
		}
		// Override account balance.
		if account.Balance != nil {
			balance, overflow := uint256.FromBig(account.Balance)
			if overflow {
				return fmt.Errorf("account %s balance %s overflow", addr.Hex(), account.Balance.String())
			}
			s.SetBalance(addr, balance)
		}
		if account.State != nil && account.StateDiff != nil {
			return fmt.Errorf("account %s has both 'state' and 'stateDiff'", addr.Hex())
		}
		// Replace entire state if caller requires.
		if account.State != nil {
			storage := make(state.Storage)
			for key, value := range *account.State {
				v, overflow := uint256.FromBig(value.Big())
				if overflow {
					return fmt.Errorf("account %s storage %s=%s overflow", addr.Hex(), key.Hex(), value.String())
				}
				storage[key] = *v
			}
			s.SetStorage(addr, storage)
		}
		// Apply state diff into specified accounts.
		if account.StateDiff != nil {
			for key, value := range *account.StateDiff {
				v, overflow := uint256.FromBig(value.Big())
				if overflow {
					return fmt.Errorf("account %s storage %s=%s overflow", addr.Hex(), key.Hex(), value.String())
				}
				s.SetState(addr, &key, *v)
			}
		}
	}
	return nil
}

type BlockOverrides struct {
	Number      *big.Int        `json:"number,omitempty"`
	NumberShift int64           `json:"numberShift,omitempty"`
	Difficulty  *big.Int        `json:"difficulty,omitempty"`
	Time        *uint64         `json:"time,omitempty"`
	TimeShift   int64           `json:"timeShift,omitempty"`
	GasLimit    *uint64         `json:"gasLimit,omitempty"`
	Coinbase    *common.Address `json:"coinbase,omitempty"`
	Random      *common.Hash    `json:"random,omitempty"`
	BaseFee     *big.Int        `json:"baseFee,omitempty"`
}

// Apply overrides the given header fields into the given block context.
func (diff *BlockOverrides) Apply(blockCtx *evmtypes.BlockContext) {
	if diff == nil {
		return
	}
	if diff.Number != nil {
		blockCtx.BlockNumber = uint64(diff.Number.Int64())
	}
	if diff.NumberShift != 0 {
		blockCtx.BlockNumber = blockCtx.BlockNumber + uint64(diff.NumberShift)
	}
	if diff.Difficulty != nil {
		blockCtx.Difficulty = diff.Difficulty
	}
	if diff.Time != nil {
		blockCtx.Time = *diff.Time
	}
	if diff.TimeShift != 0 {
		blockCtx.Time += uint64(diff.TimeShift)
	}
	if diff.GasLimit != nil {
		blockCtx.GasLimit = *diff.GasLimit
	}
	if diff.Coinbase != nil {
		blockCtx.Coinbase = *diff.Coinbase
	}
	if diff.Random != nil {
		blockCtx.PrevRanDao = diff.Random
	}
	if diff.BaseFee != nil {
		blockCtx.BaseFee = uint256.MustFromBig(diff.BaseFee)
	}
}

type CallMsg struct {
	From      common.Address   `json:"from,omitempty"`
	To        *common.Address  `json:"to,omitempty"`
	Gas       uint64           `json:"gas,omitempty"`
	GasPrice  *big.Int         `json:"gasPrice,omitempty"`
	GasFeeCap *big.Int         `json:"gasFeeCap,omitempty"`
	GasTipCap *big.Int         `json:"gasTipCap,omitempty"`
	Value     *big.Int         `json:"value,omitempty"`
	Nonce     *uint64          `json:"nonce,omitempty"`
	Data      hexutility.Bytes `json:"data,omitempty"`

	// Introduced by AccessListTxType transaction.
	AccessList       types2.AccessList `json:"accessList,omitempty"`
	EnableAccessList bool              `json:"enableAccessList,omitempty"`
}

type CallArgs struct {
	Txs                    []*CallMsg            `json:"txs,omitempty"`
	StateBlockNumberOrHash rpc.BlockNumberOrHash `json:"stateBlockNumber,omitempty"`
	Timeout                *int64                `json:"timeout,omitempty"`
	EnableBaseFee          bool                  `json:"enableBaseFee,omitempty"`
	EnableCallTracer       bool                  `json:"enableCallTracer,omitempty"`
	BlockOverrides         *BlockOverrides       `json:"blockOverrides,omitempty"`
	StateOverrides         *StateOverride        `json:"stateOverrides,omitempty"`
}

type CallResult struct {
	StateBlockNumber int64       `json:"stateBlockNumber,omitempty"`
	TotalGasUsed     uint64      `json:"totalGasUsed,omitempty"`
	Txs              []*TxResult `json:"txs,omitempty"`
}

type TxResult struct {
	GasUsed    uint64            `json:"gasUsed,omitempty"`
	Error      string            `json:"error,omitempty"`
	ReturnData hexutility.Bytes  `json:"returnData,omitempty"`
	Logs       []*types.Log      `json:"logs,omitempty"`
	CallFrame  *CallFrame        `json:"callFrame,omitempty"`
	AccessList types2.AccessList `json:"accessList,omitempty"`
}

type CallBundleArgs struct {
	Txs                    []hexutility.Bytes    `json:"txs,omitempty"`
	StateBlockNumberOrHash rpc.BlockNumberOrHash `json:"stateBlockNumber,omitempty"`
	Timeout                *int64                `json:"timeout,omitempty"`
	EnableCallTracer       bool                  `json:"enableCallTracer,omitempty"`
	EnableAccessList       bool                  `json:"enableAccessList,omitempty"`
	BlockOverrides         *BlockOverrides       `json:"blockOverrides,omitempty"`
	StateOverrides         *StateOverride        `json:"stateOverrides,omitempty"`
}

type CallBundleResult struct {
	BundleGasPrice    *big.Int          `json:"bundleGasPrice,omitempty"`
	BundleHash        common.Hash       `json:"bundleHash,omitempty"`
	CoinbaseDiff      *big.Int          `json:"coinbaseDiff,omitempty"`
	GasFees           *big.Int          `json:"gasFees,omitempty"`
	EthSentToCoinbase *big.Int          `json:"ethSentToCoinbase,omitempty"`
	StateBlockNumber  int64             `json:"stateBlockNumber,omitempty"`
	TotalGasUsed      uint64            `json:"totalGasUsed,omitempty"`
	Txs               []*BundleTxResult `json:"txs,omitempty"`
}

type BundleTxResult struct {
	TxHash            common.Hash       `json:"txHash,omitempty"`
	GasUsed           uint64            `json:"gasUsed,omitempty"`
	Error             string            `json:"error,omitempty"`
	ReturnData        hexutility.Bytes  `json:"returnData,omitempty"`
	Logs              []*types.Log      `json:"logs,omitempty"`
	CoinbaseDiff      *big.Int          `json:"coinbaseDiff,omitempty"`
	GasFees           *big.Int          `json:"gasFees,omitempty"`
	EthSentToCoinbase *big.Int          `json:"ethSentToCoinbase,omitempty"`
	GasPrice          *big.Int          `json:"gasPrice,omitempty"`
	CallMsg           *CallMsg          `json:"callMsg,omitempty"`
	CallFrame         *CallFrame        `json:"callFrame,omitempty"`
	AccessList        types2.AccessList `json:"accessList,omitempty"`
}

type CallType string

const (
	CallTypeCall         = "CALL"
	CallTypeCallCode     = "CALLCODE"
	CallTypeStaticCall   = "STATICCALL"
	CallTypeCreate       = "CREATE"
	CallTypeCreate2      = "CREATE2"
	CallTypeSelfDestruct = "SELFDESTRUCT"
	CallTypeDelegateCall = "DELEGATECALL"
)

type CallFrame struct {
	Type         CallType         `json:"type,omitempty"`
	From         common.Address   `json:"from,omitempty"`
	To           common.Address   `json:"to,omitempty"`
	Value        *big.Int         `json:"value,omitempty"`
	Gas          uint64           `json:"gas,omitempty"`
	GasUsed      uint64           `json:"gasUsed,omitempty"`
	Error        string           `json:"error,omitempty"`
	RevertReason string           `json:"revertReason,omitempty"`
	Input        hexutility.Bytes `json:"input,omitempty"`
	Output       hexutility.Bytes `json:"output,omitempty"`
	Calls        []*CallFrame     `json:"calls,omitempty"`
	Logs         []*CallLog       `json:"logs,omitempty"`

	Parent *CallFrame `json:"-"`
}

func (f *CallFrame) ToLogs() []*types.Log {
	if f == nil {
		return nil
	}
	logs := make([]*types.Log, len(f.Logs))
	for i, l := range f.Logs {
		logs[i] = l.ToLog()
	}
	return logs
}

type CallLog struct {
	Address common.Address   `json:"address,omitempty"`
	Topics  []common.Hash    `json:"topics,omitempty"`
	Data    hexutility.Bytes `json:"data,omitempty"`
}

func (l *CallLog) ToLog() *types.Log {
	if l == nil {
		return nil
	}
	return &types.Log{
		Address: l.Address,
		Topics:  l.Topics,
		Data:    l.Data,
	}
}

type ChainDataArgs struct {
	StateBlockNumberOrHash rpc.BlockNumberOrHash            `json:"stateBlockNumber,omitempty"`
	Accounts               map[common.Address][]common.Hash `json:"accounts,omitempty"`
}

type ChainDataResult struct {
	Header      *types.Header               `json:"header,omitempty"`
	NextBaseFee *big.Int                    `json:"nextBaseFee,omitempty"`
	Accounts    map[common.Address]*Account `json:"accounts,omitempty"`
}

type Account struct {
	Nonce   uint64                      `json:"nonce,omitempty"`
	Balance *big.Int                    `json:"balance,omitempty"`
	State   map[common.Hash]common.Hash `json:"state,omitempty"`
}
