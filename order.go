package main

import (
	"context"
	"log"
	"os"

	"github.com/adshao/go-binance/v2"
)

var appEnv = os.Getenv("APP_ENV")

type orderManager struct {
	orderSrv     *binance.CreateOrderService
	listOrderSrv *binance.ListOrdersService
}

func NewOrderManager(client *binance.Client) *orderManager {
	orderSrv := client.
		NewCreateOrderService().
		Type(binance.OrderTypeLimit).
		TimeInForce(binance.TimeInForceTypeGTC)

	listOrderSrv := client.NewListOrdersService()

	return &orderManager{
		orderSrv:     orderSrv,
		listOrderSrv: listOrderSrv,
	}
}

func (om *orderManager) Sell(order *Order) {
	sellSrv := om.orderSrv.Symbol(order.Symbol).Price(order.Price).Quantity(order.Quantity).Side(binance.SideTypeSell)

	ctx := context.Background()
	if appEnv == "dev" {
		err := sellSrv.Test(ctx)
		if err != nil {
			panic(err)
		}
	} else {
		_, err := sellSrv.Do(ctx)
		if err != nil {
			panic(err)
		}
	}
}

func (om *orderManager) Buy(order *Order) {
	buySrv := om.orderSrv.Symbol(order.Symbol).Price(order.Price).Quantity(order.Quantity).Side(binance.SideTypeBuy)

	ctx := context.Background()
	if appEnv == "dev" {
		err := buySrv.Test(ctx)
		if err != nil {
			panic(err)
		}
	} else {
		_, err := buySrv.Do(ctx)
		if err != nil {
			panic(err)
		}
	}
}

func (om *orderManager) GetRecentOrder(symbol string) *Order {
	orders, err := om.listOrderSrv.Symbol(symbol).Do(context.Background())
	if err != nil {
		log.Fatalf("error getting recent order %v", err)
	}

	if len(orders) == 0 {
		return &Order{}
	}

	o := orders[len(orders)-1]
	return &Order{
		Price:    o.Price,
		Quantity: o.OrigQuantity,
		Symbol:   o.Symbol,
		Type:     string(o.Side),
	}
}

func (om *orderManager) IsOrderOpen(symbol string) bool {
	openOrders, err := om.listOrderSrv.Symbol(symbol).
		Do(context.Background())
	if err != nil {
		panic(err)
	}

	for _, o := range openOrders {
		if o.Symbol == symbol && o.Status == binance.OrderStatusTypeNew {
			log.Printf("[%v] order status: %v, side: %v\n", o.Symbol, o.Status, o.Side)
			return true
		}
	}

	return false
}
