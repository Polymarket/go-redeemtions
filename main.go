package main

import (
	"context"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/playground/go-redemptions/conditional_tokens"
	"github.com/shopspring/decimal"
)

const (
	WS_URL = "wss://..." // TODO: Complete this with a WS RPC
)

var REDEEMER_ADDRESS = common.HexToAddress("0x5f211A24da4c005d9438A1eA269673b85eD0b376")
var HASHED_SIGNATURES = [...]common.Hash{crypto.Keccak256Hash([]byte("PayoutRedemption(address,address,bytes32,bytes32,uint256[],uint256)"))}
var CONTRACT_ADDRESS = []common.Address{common.HexToAddress("0x4D97DCd97eC945f40cF65F87097ACe5EA0476045")}
var FQ = ethereum.FilterQuery{
	Addresses: CONTRACT_ADDRESS[:],
	Topics:    [][]common.Hash{HASHED_SIGNATURES[:]},
}

func main() {
	ctx := context.Background()
	client, err := ethclient.DialContext(ctx, WS_URL)
	if err != nil {
		panic(err)
	}

	logs := make(chan types.Log)
	subs, err := client.SubscribeFilterLogs(ctx, FQ, logs)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case err, ok := <-subs.Err():
			if ok && err != nil {
				panic(err)
			}
		case log, ok := <-logs:
			if !ok {
				panic("logs channel closed")
			}

			redemption, err := getContract().ParsePayoutRedemption(log)
			if err != nil {
				panic(err)
			}

			if redemption.Redeemer.String() == REDEEMER_ADDRESS.String() {
				fmt.Printf("redemption: %s\n", normalizeAmounts(redemption.Payout).String())
			}
		}
	}
}

// gets a conditional token instance
func getContract() *conditional_tokens.ConditionalTokensFilterer {
	c, err := conditional_tokens.NewConditionalTokensFilterer(CONTRACT_ADDRESS[0], nil)
	if err != nil {
		panic(err)
	}
	return c
}

// removes the 6 usdc 0s
func normalizeAmounts(amount *big.Int) decimal.Decimal {
	return decimal.NewFromBigInt(amount, 0).Div(decimal.NewFromFloat(math.Pow(10, 6)))
}
