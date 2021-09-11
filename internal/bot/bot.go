package internal

import (
	"log"
	"strconv"

	"github.com/adshao/go-binance/v2"
	db "github.com/drifterz13/trading-bot/internal/database"
	"github.com/drifterz13/trading-bot/internal/dto"
	"github.com/drifterz13/trading-bot/internal/manager"
	"github.com/drifterz13/trading-bot/internal/utils"
)

var (
	symbols = []string{"ALGOUSDT", "SOLUSDT", "MATICUSDT", "ADAUSDT", "BTCUSDT"}
)

type KlineManager interface {
	GetAvgOHLC(symbol string, interval string, limit int) *dto.Kline
}
type AccountManager interface {
	GetBalance(symbol string) float64
}
type OrderManager interface {
	Buy(order *dto.Order)
	Sell(order *dto.Order)
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
	repo           db.BoltRepository
}

func NewBot(client *binance.Client, repo db.BoltRepository) *Bot {
	return &Bot{
		klineManager:   manager.NewKlineManager(client),
		accountManager: manager.NewAccountManager(client),
		orderManager:   manager.NewOrderManager(client, repo),
		priceManager:   manager.NewPriceManager(client),
		repo:           repo,
	}
}

func (s *Bot) Run(symbol string) {
	ohlc := s.klineManager.GetAvgOHLC(symbol, "15m", 48)
	latestPrice := s.priceManager.GetLatestPrice(symbol)
	balance := s.accountManager.GetBalance(symbol)
	latestPriceRatio := utils.GetPriceRatio(latestPrice, ohlc.High)

	if balance == 0 && latestPriceRatio < -3 {
		if s.orderManager.IsOrderOpen(symbol) {
			return
		}
		// consider buying
		qty := s.GetBuyQuantity(symbol, latestPrice)
		order := &dto.Order{
			Symbol:   symbol,
			Price:    strconv.FormatFloat(latestPrice, 'f', -1, 64),
			Quantity: qty,
			Type:     dto.BuyType,
		}
		s.orderManager.Sell(order)
		return
	}

	lastOrder := s.repo.Last(symbol)
	if lastOrder.IsEmpty() {
		log.Fatalln("last order not found.")
	}

	if balance > 0 && latestPriceRatio <= -20 {
		if s.orderManager.IsOrderOpen(symbol) {
			return
		}

		// stop loss
		order := &dto.Order{
			Symbol:   symbol,
			Price:    strconv.FormatFloat(latestPrice, 'f', -1, 64),
			Quantity: lastOrder.Quantity,
			Type:     dto.SellType,
		}
		s.orderManager.Sell(order)

		return
	}

	boughtPriceRatio := utils.GetPriceRatio(lastOrder.ToFloat64().Price, latestPrice)

	if balance > 0 && boughtPriceRatio >= 5 {
		if s.orderManager.IsOrderOpen(symbol) {
			return
		}

		// taking profit
		order := &dto.Order{
			Symbol:   symbol,
			Price:    strconv.FormatFloat(latestPrice, 'f', -1, 64),
			Quantity: lastOrder.Quantity,
			Type:     dto.SellType,
		}
		s.orderManager.Sell(order)
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

func (s *Bot) GetBuyQuantity(symbol string, price float64) string {
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

	budget := s.GetAffordableBudget()
	qty := utils.ToFixed((budget / price), dec)

	return strconv.FormatFloat(qty, 'f', -1, 64)
}
