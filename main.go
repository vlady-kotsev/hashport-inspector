package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"net/http"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

const (
	BurnMode       = "burn"
	MintMode       = "mint"
	Mainnet        = "mainnet"
	Testnet        = "testnet"
	TokenBurn      = "TOKENBURN"
	TokenMint      = "TOKENMINT"
	CryptoTransfer = "CRYPTOTRANSFER"
)

func main() {
	network := flag.String("network", "", "Hedera Network")
	accountID := flag.String("account", "", "The account ID")
	tokenID := flag.String("token", "", "The token ID")
	mode := flag.String("mode", "", "The mode: it can be burn or mint")
	flag.Parse()

	if *network == "" || (*network != Mainnet && *network != Testnet) {
		panic("invalid network")
	}

	if *mode == "" || (*mode != BurnMode && *mode != MintMode) {
		panic("invalid mode")
	}

	_, err := hedera.AccountIDFromString(*accountID)
	if err != nil {
		panic(err)
	}
	_, err = hedera.TokenIDFromString(*tokenID)
	if err != nil {
		panic(err)
	}

	baseUrl := fmt.Sprintf("https://%s.mirrornode.hedera.com", *network)
	// accountID := "0.0.540219"
	// tokenID := "0.0.624505"
	// 0.0.4418477
	// 0.0.540219
	url := fmt.Sprintf("/api/v1/transactions/?account.id=%s&limit=100&order=asc&transactiontype=cryptotransfer&transactiontype=tokenburn&transactiontype=tokenmint", *accountID)
	//url := fmt.Sprintf("/api/v1/transactions/?account.id=%s&limit=100&order=asc&transactiontype=cryptotransfer", accountID)

	done := make(chan struct{})
	errorChan := make(chan error)
	cryptoTransferChan := make(chan FoundTransaction)
	cryptoBurnOrMintChan := make(chan FoundTransaction)
	txsChan := make(chan Transaction, 5)

	fmt.Println("Processing...")
	go getTransactions(url, baseUrl, errorChan, txsChan)
	if *mode == BurnMode {
		go processBurnTransactions(txsChan, *tokenID, *accountID, cryptoTransferChan, cryptoBurnOrMintChan, done)
	} else if *mode == MintMode {
		go processMintTransactions(txsChan, *tokenID, *accountID, cryptoTransferChan, cryptoBurnOrMintChan, done)
	}

	cryptoTransfers := []FoundTransaction{}
	cryptoBurnsOrMints := []FoundTransaction{}
outer:
	for {
		select {
		case err := <-errorChan:
			fmt.Println(err)
			return
		case <-done:
			break outer
		case tx := <-cryptoTransferChan:
			//fmt.Printf("Crypto transfer: %s\n", txID)
			cryptoTransfers = append(cryptoTransfers, tx)
		case tx := <-cryptoBurnOrMintChan:
			//fmt.Printf("Crypto burn: %s\n", txID)
			cryptoBurnsOrMints = append(cryptoBurnsOrMints, tx)
		}
	}
	fmt.Printf("==== Done ===\n\n")
	// for _, txID := range cryptoTransfers {
	// 	fmt.Println(txID)
	// }
	// for _, txID := range cryptoBurnsIDs {
	// 	fmt.Println(txID)
	// }
	var diffs []FoundTransaction
	if *mode == BurnMode {
		diffs = findDiff(cryptoTransfers, cryptoBurnsOrMints)
	}
	if *mode == MintMode {
		diffs = findDiff(cryptoBurnsOrMints, cryptoTransfers)
	}

	totalAmount := int64(0)
	for _, diff := range diffs {
		fmt.Println(diff.TransactionID)
		totalAmount += diff.Amount
	}
	fmt.Printf("Crypto Transfers: %d Crypto burns/mints: %d\n", len(cryptoTransfers), len(cryptoBurnsOrMints))
	fmt.Printf("Total Amount: %d\n", totalAmount)
}

func getTransactions(url, baseUrl string, errorChan chan<- error, txsChan chan<- Transaction) {
	//txCounter := 0
	for {
		//txCounter++
		//fmt.Println(txCounter)
		completeUrl := fmt.Sprintf("%s%s", baseUrl, url)
		//fmt.Printf("Calling: %s\n", completeUrl)
		resp, err := http.Get(completeUrl)
		if err != nil {
			errorChan <- err
		}

		var response Response
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			defer resp.Body.Close()
			errorChan <- err
		}
		resp.Body.Close()

		for _, tx := range response.Transactions {
			//fmt.Println(tx)
			txsChan <- tx
		}

		if response.Links.Next == nil {
			close(txsChan)
			return
		}

		url = *response.Links.Next
	}
}

type FoundTransaction struct {
	TransactionID string
	Amount        int64
}

func processBurnTransactions(txsChan <-chan Transaction, tokenID, accountID string, cryptoTransferChan, cryptoBurnChan chan<- FoundTransaction, done chan struct{}) {
	for {
		tx, ok := <-txsChan
		//fmt.Printf("Received tx: %s\n", tx.TransactionID)
		if !ok {
			close(done)
			return
		}

		for _, tokenTransfer := range tx.TokenTransfers {
			if tokenTransfer.TokenID == tokenID && tokenTransfer.Account == accountID {
				if tx.Name == CryptoTransfer && tokenTransfer.Amount > 0 {
					cryptoTransferChan <- FoundTransaction{
						TransactionID: tx.TransactionID,
						Amount:        tokenTransfer.Amount,
					}
				}
				if tx.Name == TokenBurn {
					cryptoBurnChan <- FoundTransaction{
						TransactionID: tx.TransactionID,
						Amount:        int64(math.Abs(float64(tokenTransfer.Amount))),
					}
				}
			}
		}
	}
}

func processMintTransactions(txsChan <-chan Transaction, tokenID, accountID string, cryptoTransferChan, cryptoMintChan chan<- FoundTransaction, done chan struct{}) {
	for {
		tx, ok := <-txsChan
		//fmt.Printf("Received tx: %s\n", tx.TransactionID)
		if !ok {
			close(done)
			return
		}

		for _, tokenTransfer := range tx.TokenTransfers {
			if tokenTransfer.TokenID == tokenID && tokenTransfer.Account == accountID {
				if tx.Name == CryptoTransfer && tokenTransfer.Amount < 0 {
					cryptoTransferChan <- FoundTransaction{
						TransactionID: tx.TransactionID,
						Amount:        int64(math.Abs(float64(tokenTransfer.Amount))),
					}
				}
				if tx.Name == TokenMint {
					cryptoMintChan <- FoundTransaction{
						TransactionID: tx.TransactionID,
						Amount:        tokenTransfer.Amount,
					}
				}
			}
		}
	}
}

func findDiff(first, second []FoundTransaction) []FoundTransaction {
	diffs := []FoundTransaction{}
	unique := make(map[int64]bool)
	for _, s := range second {
		unique[s.Amount] = true
	}

	for _, f := range first {
		if !unique[f.Amount] {
			diffs = append(diffs, f)
		}
	}
	return diffs
}
