package main

type mockKlineManager struct {
	klines map[string]*Kline
}

func (km *mockKlineManager) GetAvgOHLC(symbol string, interval string, limit int) *Kline {
	return km.klines[symbol]
}

type mockAccountManager struct {
	Balance map[string]float64
}

func (am *mockAccountManager) GetBalance(symbol string) float64 {
	return am.Balance[symbol]
}

type mockOrderManager struct {
	orderOpen bool
	called    map[string]int
	Orders    map[string]*Order
}

func (om *mockOrderManager) Buy(order *Order) {
	called := om.called["Buy"]
	om.called["Buy"] = called + 1
}
func (om *mockOrderManager) Sell(order *Order) {
	called := om.called["Sell"]
	om.called["Sell"] = called + 1
}

func (om *mockOrderManager) GetRecentOrder(symbol string) *Order {
	return om.Orders[symbol]
}

func (om *mockOrderManager) IsOrderOpen(symbol string) bool { return om.orderOpen }

type mockPriceManager struct {
	Prices map[string]float64
}

func (pm mockPriceManager) GetLatestPrice(symbol string) float64 {
	return pm.Prices[symbol]
}

var mockBot = &Bot{
	klineManager:   &mockKlineManager{},
	accountManager: &mockAccountManager{},
	orderManager:   &mockOrderManager{orderOpen: false, called: map[string]int{}},
	priceManager:   &mockPriceManager{},
}
