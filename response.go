package main

type Response struct {
	Transactions []Transaction `json:"transactions"`
	Links        Links         `json:"links"`
}

type Transaction struct {
	Bytes                    *string                 `json:"bytes"`
	ChargedTxFee             int64                   `json:"charged_tx_fee"`
	ConsensusTimestamp       string                  `json:"consensus_timestamp"`
	EntityID                 string                  `json:"entity_id"`
	MaxFee                   string                  `json:"max_fee"`
	MemoBase64               *string                 `json:"memo_base64"`
	Name                     string                  `json:"name"`
	NFTTransfers             []NFTTransfer           `json:"nft_transfers"`
	Node                     string                  `json:"node"`
	Nonce                    int64                   `json:"nonce"`
	ParentConsensusTimestamp string                  `json:"parent_consensus_timestamp"`
	Result                   string                  `json:"result"`
	Scheduled                bool                    `json:"scheduled"`
	StakingRewardTransfers   []StakingRewardTransfer `json:"staking_reward_transfers"`
	TransactionHash          string                  `json:"transaction_hash"`
	TransactionID            string                  `json:"transaction_id"`
	TokenTransfers           []TokenTransfer         `json:"token_transfers"`
	Transfers                []Transfer              `json:"transfers"`
	ValidDurationSeconds     string                  `json:"valid_duration_seconds"`
	ValidStartTimestamp      string                  `json:"valid_start_timestamp"`
}

type NFTTransfer struct {
	IsApproval        bool   `json:"is_approval"`
	ReceiverAccountID string `json:"receiver_account_id"`
	SenderAccountID   string `json:"sender_account_id"`
	SerialNumber      int64  `json:"serial_number"`
	TokenID           string `json:"token_id"`
}

type StakingRewardTransfer struct {
	Account string `json:"account"`
	Amount  int64 `json:"amount"`
}

type TokenTransfer struct {
	TokenID    string `json:"token_id"`
	Account    string `json:"account"`
	Amount     int64  `json:"amount"`
	IsApproval bool   `json:"is_approval"`
}

type Transfer struct {
	Account    string `json:"account"`
	Amount     int64  `json:"amount"`
	IsApproval bool   `json:"is_approval"`
}

type Links struct {
	Next *string `json:"next"`
}
