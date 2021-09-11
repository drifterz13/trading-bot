package manager

import (
	"context"
	"strconv"

	"github.com/adshao/go-binance/v2"
)

type priceManager struct {
	client binance.Client
}

func NewPriceManager(client *binance.Client) *priceManager {
	return &priceManager{client: *client}
}

func (pm *priceManager) GetLatestPrice(symbol string) float64 {
	prices, err := pm.client.NewListPricesService().Symbol(symbol).Do(context.Background())
	if err != nil {
		panic(err)
	}

	var price string
	for _, p := range prices {
		price = p.Price
	}

	p, err := strconv.ParseFloat(price, 64)
	if err != nil {
		panic(err)
	}

	return p
}
