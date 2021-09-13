package main

import "testing"

func TestGetPriceRatio(t *testing.T) {
	tables := []struct {
		Price    float64
		Comparer float64
		Want     float64
	}{
		{Price: 1.5, Comparer: 1.3, Want: 15.38},
		{Price: 2.0, Comparer: 2.4, Want: -16.67},
	}

	for _, table := range tables {
		got := GetPriceRatio(table.Price, table.Comparer)
		if got != table.Want {
			t.Errorf("want %.2f; got %.2f", table.Want, got)
		}
	}
}

// TODO: support dynamic symbol.
func TestBotGetAffordableBudget(t *testing.T) {
	bal1 := map[string]float64{
		"USDT":     100,
		"ADAUSDT":  30,
		"SOLUSDT":  3,
		"ALGOUSDT": 10,
	}

	bal2 := map[string]float64{
		"USDT":      100,
		"ADAUSDT":   30,
		"SOLUSDT":   3,
		"ALGOUSDT":  22,
		"MATICUSDT": 29,
		"BTCUSDT":   0.000450,
	}

	tables := []struct {
		Balance map[string]float64
		Want    float64
	}{
		{Balance: bal1, Want: 50.0},
		{Balance: bal2, Want: 0.0},
	}

	for _, table := range tables {
		bot := &Bot{
			priceManager: &mockPriceManager{
				Prices: map[string]float64{
					"SOLUSDT":   180,
					"ADAUSDT":   2.4,
					"ALGOUSDT":  2.3,
					"MATICUSDT": 1.25,
					"BTCUSDT":   45500.50,
				},
			},
			accountManager: &mockAccountManager{Balance: table.Balance},
		}
		got := bot.GetAffordableBudget()

		if got != table.Want {
			t.Errorf("want %v; got %v", table.Want, got)
		}
	}
}

func TestBotGetBuyQuantity(t *testing.T) {
	bot := &Bot{
		priceManager: &mockPriceManager{
			Prices: map[string]float64{
				"SOLUSDT":   180,
				"ADAUSDT":   2.4,
				"ALGOUSDT":  2.2,
				"MATICUSDT": 1.3,
				"BTCUSDT":   47500.50,
			},
		},
		accountManager: &mockAccountManager{
			Balance: map[string]float64{
				"USDT": 100,
			},
		},
	}

	want := map[string]string{
		"ADAUSDT":   "8",
		"SOLUSDT":   "0.11",
		"ALGOUSDT":  "9",
		"MATICUSDT": "15",
		"BTCUSDT":   "0.0004",
	}

	for _, symbol := range symbols {
		price := bot.priceManager.GetLatestPrice(symbol)
		got := bot.GetOrderQuantity(symbol, price)

		if want[symbol] != got {
			t.Errorf("symbol %v; want %v; got %v", symbol, want[symbol], got)
		}
	}
}

func TestBotRun_BuyAll(t *testing.T) {
	km := &mockKlineManager{
		klines: map[string]*Kline{
			"ADAUSDT":   {High: 2.9, Low: 2.7, Open: 2.8, Close: 2.75},
			"SOLUSDT":   {High: 190, Low: 178, Open: 182, Close: 188},
			"ALGOUSDT":  {High: 2.4, Low: 2.1, Open: 2.2, Close: 2.3},
			"MATICUSDT": {High: 1.4, Low: 1.1, Open: 1.2, Close: 1.1},
			"BTCUSDT":   {High: 47500, Low: 45000, Open: 46500, Close: 47000},
		},
	}
	pm := &mockPriceManager{
		Prices: map[string]float64{
			"SOLUSDT":   180,
			"ADAUSDT":   2.4,
			"ALGOUSDT":  2,
			"MATICUSDT": 1.25,
			"BTCUSDT":   45500.50,
		},
	}

	am := &mockAccountManager{
		Balance: map[string]float64{"USDT": 100},
	}

	om := &mockOrderManager{orderOpen: false, called: map[string]int{}}

	bot := &Bot{
		klineManager:   km,
		accountManager: am,
		orderManager:   om,
		priceManager:   pm,
	}

	for _, symbol := range symbols {
		bot.Run(symbol)
	}

	got := om.called["Buy"]
	want := 5
	if got != want {
		t.Errorf("want %v; got %v", want, got)
	}
}

func TestBotRun_Sell(t *testing.T) {
	km := &mockKlineManager{
		klines: map[string]*Kline{
			"ADAUSDT":   {High: 2.9, Low: 2.7, Open: 2.8, Close: 2.75},
			"SOLUSDT":   {High: 190, Low: 178, Open: 182, Close: 188},
			"ALGOUSDT":  {High: 2.4, Low: 2.1, Open: 2.2, Close: 2.3},
			"MATICUSDT": {High: 1.4, Low: 1.1, Open: 1.2, Close: 1.1},
			"BTCUSDT":   {High: 47500, Low: 45000, Open: 46500, Close: 47000},
		},
	}
	pm := &mockPriceManager{
		Prices: map[string]float64{
			"SOLUSDT":   180,
			"ADAUSDT":   2.4,
			"ALGOUSDT":  2.8,
			"MATICUSDT": 1.45,
			"BTCUSDT":   48500.50,
		},
	}

	am := &mockAccountManager{
		Balance: map[string]float64{
			"USDT":    100,
			"SOLUSDT": 25.8,
			"ADAUSDT": 24,
		},
	}

	orders := map[string]*Order{
		// take profit
		"SOLUSDT": {
			Symbol:   "SOLUSDT",
			Price:    "152",
			Quantity: "0.15",
			Type:     BuyType,
		},
		// stop loss
		"ADAUSDT": {
			Symbol:   "ADAUSDT",
			Price:    "3.1",
			Quantity: "10",
			Type:     BuyType,
		},
		"ALGOUSDT":  {},
		"MATICUSDT": {},
		"BTCUSDT":   {},
	}

	om := &mockOrderManager{orderOpen: false, called: map[string]int{}, Orders: orders}

	bot := &Bot{
		klineManager:   km,
		accountManager: am,
		orderManager:   om,
		priceManager:   pm,
	}

	for _, symbol := range symbols {
		bot.Run(symbol)
	}

	got := om.called["Sell"]
	want := 2
	if got != want {
		t.Errorf("want %v; got %v", want, got)
	}
}
