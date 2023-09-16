package privacy

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

type RPCTransaction struct {
	BlockHash        *common.Hash      `json:"blockHash"`
	BlockNumber      *hexutil.Big      `json:"blockNumber"`
	From             common.Address    `json:"from"`
	Gas              hexutil.Uint64    `json:"gas"`
	GasPrice         *hexutil.Big      `json:"gasPrice"`
	GasFeeCap        *hexutil.Big      `json:"maxFeePerGas,omitempty"`
	GasTipCap        *hexutil.Big      `json:"maxPriorityFeePerGas,omitempty"`
	Hash             common.Hash       `json:"hash"`
	Input            hexutil.Bytes     `json:"input"`
	Nonce            hexutil.Uint64    `json:"nonce"`
	To               *common.Address   `json:"to"`
	TransactionIndex *hexutil.Uint64   `json:"transactionIndex"`
	Value            *hexutil.Big      `json:"value"`
	Type             hexutil.Uint64    `json:"type"`
	Accesses         *types.AccessList `json:"accessList,omitempty"`
	ChainID          *hexutil.Big      `json:"chainId,omitempty"`
	V                *hexutil.Big      `json:"v"`
	R                *hexutil.Big      `json:"r"`
	S                *hexutil.Big      `json:"s"`

	// Arbitrum fields:
	RequestId           *common.Hash    `json:"requestId,omitempty"`           // Contract SubmitRetryable Deposit
	TicketId            *common.Hash    `json:"ticketId,omitempty"`            // Retry
	MaxRefund           *hexutil.Big    `json:"maxRefund,omitempty"`           // Retry
	SubmissionFeeRefund *hexutil.Big    `json:"submissionFeeRefund,omitempty"` // Retry
	RefundTo            *common.Address `json:"refundTo,omitempty"`            // SubmitRetryable Retry
	L1BaseFee           *hexutil.Big    `json:"l1BaseFee,omitempty"`           // SubmitRetryable
	DepositValue        *hexutil.Big    `json:"depositValue,omitempty"`        // SubmitRetryable
	RetryTo             *common.Address `json:"retryTo,omitempty"`             // SubmitRetryable
	RetryValue          *hexutil.Big    `json:"retryValue,omitempty"`          // SubmitRetryable
	RetryData           *hexutil.Bytes  `json:"retryData,omitempty"`           // SubmitRetryable
	Beneficiary         *common.Address `json:"beneficiary,omitempty"`         // SubmitRetryable
	MaxSubmissionFee    *hexutil.Big    `json:"maxSubmissionFee,omitempty"`    // SubmitRetryable
}

type Transactions []RPCTransaction

type Receipt types.Receipt

type RPCBlock struct {
	BaseFeePerGas    string       `json:"baseFeePerGas"`
	Difficulty       string       `json:"difficulty"`
	ExtraData        string       `json:"extraData"`
	GasLimit         string       `json:"gasLimit"`
	GasUsed          string       `json:"gasUsed"`
	Hash             string       `json:"hash"`
	L1BlockNumber    string       `json:"l1BlockNumber"`
	LogsBloom        string       `json:"logsBloom"`
	Miner            string       `json:"miner"`
	MixHash          string       `json:"mixHash"`
	Nonce            string       `json:"nonce"`
	Number           string       `json:"number"`
	ParentHash       string       `json:"parentHash"`
	ReceiptsRoot     string       `json:"receiptsRoot"`
	SendCount        string       `json:"sendCount"`
	SendRoot         string       `json:"sendRoot"`
	Sha3Uncles       string       `json:"sha3Uncles"`
	Size             string       `json:"size"`
	StateRoot        string       `json:"stateRoot"`
	Timestamp        string       `json:"timestamp"`
	TotalDifficulty  string       `json:"totalDifficulty"`
	Transactions     Transactions `json:"transactions"`
	TransactionsRoot string       `json:"transactionsRoot"`
	Uncles           []string     `json:"uncles"`
}

type Blocks []RPCBlock

type JsonrpcMessage struct {
	Version string          `json:"jsonrpc,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *jsonError      `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

type jsonError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type AuthToken struct {
	Token        string
	Address      []string
	CreationTime uint64
}
