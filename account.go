package main

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/adshao/go-binance/v2"
)

type accountManager struct {
	client *binance.Client
}

func NewAccountManager(client *binance.Client) *accountManager {
	return &accountManager{client: client}
}

func (am *accountManager) GetBalance(symbol string) float64 {
	resp, err := am.client.NewGetAccountService().Do(context.Background())
	if err != nil {
		log.Fatalf("error getting balance: %v\n", err)
	}

	var balance float64
	for _, b := range resp.Balances {
		if (b.Asset == "USDT" && symbol == "USDT") || b.Asset == strings.Replace(symbol, "USDT", "", 1) {
			f, err := strconv.ParseFloat(b.Free, 32)
			if err != nil {
				panic(err)
			}

			balance = f
		}
	}

	return balance
}
