package main

import (
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
}
type PriceManager interface {
	GetLatestPrice(symbol string) float64
}

type Bot struct {
	klineManager   KlineManager
	accountManager AccountManager
	orderManager   OrderManager
	priceManager   PriceManager
	repo           DataStore
}

func NewBot(client *binance.Client, repo DataStore) *Bot {
	return &Bot{
		klineManager:   NewKlineManager(client),
		accountManager: NewAccountManager(client),
		orderManager:   NewOrderManager(client, repo),
		priceManager:   NewPriceManager(client),
		repo:           repo,
	}
}

func (s *Bot) Run(symbol string) {
	ohlc := s.klineManager.GetAvgOHLC(symbol, "15m", 48)
	latestPrice := s.priceManager.GetLatestPrice(symbol)
	balance := s.accountManager.GetBalance(symbol)
	latestPriceRatio := GetPriceRatio(latestPrice, ohlc.High)

	log.Printf("[%v] latest price: %.2f, high: %.2f, and ratio: %.2f", symbol, latestPrice, ohlc.High, latestPriceRatio)

	if balance == 0 && latestPriceRatio < -3 {
		log.Printf("[%v] buying...", symbol)
		if s.orderManager.IsOrderOpen(symbol) {
			log.Printf("[%v] order is already open.", symbol)

			return
		}
		// consider buying
		qty := s.GetBuyQuantity(symbol, latestPrice)
		order := &Order{
			Symbol:   symbol,
			Price:    strconv.FormatFloat(latestPrice, 'f', -1, 64),
			Quantity: qty,
			Type:     BuyType,
		}
		s.orderManager.Buy(order)
		log.Printf("[%v] buy order price: %v, quantity: %v, type: %v", order.Symbol, order.Price, order.Quantity, order.Type)

		return
	}

	lastOrder := s.repo.Last(symbol)

	if lastOrder.IsEmpty() {
		log.Printf("[%v] last order not found.", symbol)
		return
	}

	log.Printf("[%v] last order price: %v, quantity: %v, type: %v", lastOrder.Symbol, lastOrder.Price, lastOrder.Quantity, lastOrder.Type)

	boughtPrice := lastOrder.ToFloat64().Price
	boughtPriceRatio := GetPriceRatio(boughtPrice, latestPrice)

	log.Printf("[%v] bought price: %v, latest price: %.2f, and ratio: %v", symbol, boughtPrice, latestPrice, boughtPriceRatio)

	if balance > 0 && boughtPriceRatio <= -20 {
		log.Printf("[%v] going to stop loss...", symbol)
		if s.orderManager.IsOrderOpen(symbol) {
			log.Printf("[%v] order is already open.", symbol)

			return
		}

		// stop loss
		log.Printf("[%v] balance: %.2f", symbol, balance)
		p := ToFixed(latestPrice, 2)
		q := ToFixed(balance/latestPrice, s.GetQuantityDecimal(symbol))

		order := &Order{
			Symbol:   symbol,
			Price:    strconv.FormatFloat(p, 'f', -1, 64),
			Quantity: strconv.FormatFloat(q, 'f', -1, 64),
			Type:     SellType,
		}

		if p*q < 10 {
			log.Printf("[%v] too low volume: %.2f", symbol, p*q)
			return
		}

		s.orderManager.Sell(order)
		log.Printf("[%v] stop loss order price: %v, quantity: %v, type: %v", order.Symbol, order.Price, order.Quantity, order.Type)

		return
	}

	if balance > 0 && boughtPriceRatio >= 5 {
		log.Printf("[%v] going to take profit...", symbol)
		if s.orderManager.IsOrderOpen(symbol) {
			log.Printf("[%v] order is already open.", symbol)

			return
		}

		// taking profit
		log.Printf("[%v] balance: %.2f", symbol, balance)
		p := ToFixed(latestPrice, 2)
		q := ToFixed(balance/latestPrice, s.GetQuantityDecimal(symbol))
		order := &Order{
			Symbol:   symbol,
			Price:    strconv.FormatFloat(p, 'f', -1, 64),
			Quantity: strconv.FormatFloat(q, 'f', -1, 64),
			Type:     SellType,
		}
		if p*q < 10 {
			log.Printf("[%v] too low volume: %.2f", symbol, p*q)
			return
		}

		s.orderManager.Sell(order)
		log.Printf("[%v] taking porift order price: %v, quantity: %v, type: %v", order.Symbol, order.Price, order.Quantity, order.Type)
	}
}

func (s *Bot) GetAffordableBudget() float64 {
	var bought float64
	usdt := s.accountManager.GetBalance("USDT")
	for _, sym := range symbols {
		bal := s.accountManager.GetBalance(sym)
		if bal > 0 {
			bought = bought + 1
		}
	}

	budget := (float64(len(symbols)) - bought) / usdt

	// Binance not allow order that less than 10$.
	if budget < 10 {
		return 0
	}

	return budget
}

func (s *Bot) GetQuantityDecimal(symbol string) int {
	var dec int

	switch symbol {
	case "ADAUSDT":
		fallthrough
	case "MATICUSDT":
		fallthrough
	case "ALGOUSDT":
		dec = 0
		break
	case "SOLUSDT":
		dec = 2
		break
	case "BTCUSDT":
		dec = 5
		break
	}

	return dec
}

func (s *Bot) GetBuyQuantity(symbol string, price float64) string {
	dec := s.GetQuantityDecimal(symbol)
	budget := s.GetAffordableBudget()
	qty := ToFixed((budget / price), dec)

	return strconv.FormatFloat(qty, 'f', -1, 64)
}
