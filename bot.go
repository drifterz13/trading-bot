package main

import (
	"errors"
	"log"
	"strconv"

	"github.com/adshao/go-binance/v2"
)

type KlineManager interface {
	GetAvgOHLC(symbol string, interval string, limit int) *Kline
}
type AccountManager interface {
	GetBalance(symbol string) float64
}
type OrderManager interface {
	Buy(order *Order)
	Sell(order *Order)
	IsOrderOpen(symbol string) bool
	GetRecentOrder(symbol string) *Order
}
type PriceManager interface {
	GetLatestPrice(symbol string) float64
}

type Bot struct {
	klineManager   KlineManager
	accountManager AccountManager
	orderManager   OrderManager
	priceManager   PriceManager
}

func NewBot(client *binance.Client) *Bot {
	return &Bot{
		klineManager:   NewKlineManager(client),
		accountManager: NewAccountManager(client),
		orderManager:   NewOrderManager(client),
		priceManager:   NewPriceManager(client),
	}
}

func (b *Bot) Run(symbol string) {
	ohlc := b.klineManager.GetAvgOHLC(symbol, "15m", 48)
	latestPrice := b.priceManager.GetLatestPrice(symbol)
	balance := b.accountManager.GetBalance(symbol)
	latestPriceRatio := GetPriceRatio(latestPrice, ohlc.High)
	recentOrder := b.orderManager.GetRecentOrder(symbol)

	value := balance * latestPrice

	log.Printf("[%v] balance: %.2f, total: %.2f", symbol, balance, value)
	log.Printf("[%v] current: %.2f, high: %.2f, ratio: %.2f%%", symbol, latestPrice, ohlc.High, latestPriceRatio)

	if value < 10 && latestPriceRatio < -3 {
		if b.orderManager.IsOrderOpen(symbol) {
			log.Printf("[%v] order is already open.", symbol)
			return
		}

		qty, err := b.GetBuyQuantity(symbol, latestPrice)
		if err != nil {
			panic(err)
		}
		order := &Order{
			Symbol:   symbol,
			Price:    strconv.FormatFloat(latestPrice, 'f', -1, 64),
			Quantity: qty,
			Type:     BuyType,
		}
		b.orderManager.Buy(order)
		log.Printf("[%v] buying at: %v, quantity: %v", order.Symbol, order.Price, order.Quantity)
		return
	}

	if recentOrder.IsEmpty() {
		log.Printf("[%v] recent order not found.", symbol)
		return
	}

	boughtPrice := recentOrder.ToFloat64().Price
	boughtPriceRatio := GetPriceRatio(latestPrice, boughtPrice)
	log.Printf("[%v] bought: %v, current: %.2f, ratio: %.2f%%", symbol, boughtPrice, latestPrice, boughtPriceRatio)

	if value >= 10 && (boughtPriceRatio <= -20 || boughtPriceRatio >= 5) {
		log.Printf("[%v] selling because gain/loss :%.2f%%", symbol, boughtPriceRatio)
		if b.orderManager.IsOrderOpen(symbol) {
			log.Printf("[%v] order is already open.", symbol)
			return
		}

		p := ToFixed(latestPrice, 2)
		q := ToFixed(balance/latestPrice, b.GetQuantityDecimal(symbol))

		order := &Order{
			Symbol:   symbol,
			Price:    strconv.FormatFloat(p, 'f', -1, 64),
			Quantity: strconv.FormatFloat(q, 'f', -1, 64),
			Type:     SellType,
		}

		b.orderManager.Sell(order)
		log.Printf("[%v] selling at: %v, quantity: %v", order.Symbol, order.Price, order.Quantity)
	}
}

func (b *Bot) GetAffordableBudget() float64 {
	var bought float64
	usdt := b.accountManager.GetBalance("USDT")
	log.Printf("usdt balance: %.2f", usdt)

	for _, sym := range symbols {
		bal := b.accountManager.GetBalance(sym)
		if bal > 0 {
			bought = bought + 1
		}
	}

	totalSymbols := float64(len(symbols))
	if bought == totalSymbols {
		return 0
	}

	budget := usdt / (totalSymbols - bought)
	// Binance not allow order that less than 10$.
	if budget < 10 {
		return 0
	}

	return budget
}

func (b *Bot) GetQuantityDecimal(symbol string) int {
	m := map[string]int{
		"ADAUSDT":   1,
		"MATICUSDT": 1,
		"ALGOUSDT":  1,
		"SOLUSDT":   2,
		"BTCUSDT":   5,
	}
	return m[symbol]
}

func (b *Bot) GetBuyQuantity(symbol string, price float64) (string, error) {
	var included bool
	for _, s := range symbols {
		if s == symbol {
			included = true
		}
	}

	if !included {
		return "", errors.New("GetBuyQuantity: invalid symbol")
	}

	dec := b.GetQuantityDecimal(symbol)
	budget := b.GetAffordableBudget()
	qty := ToFixed((budget / price), dec)

	return strconv.FormatFloat(qty, 'f', -1, 64), nil
}
