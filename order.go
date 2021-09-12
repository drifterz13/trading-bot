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
	repo         DataStore
}

func NewOrderManager(client *binance.Client, repo DataStore) *orderManager {
	orderSrv := client.
		NewCreateOrderService().
		Type(binance.OrderTypeLimit).
		TimeInForce(binance.TimeInForceTypeGTC)

	listOrderSrv := client.NewListOrdersService()

	return &orderManager{
		orderSrv:     orderSrv,
		listOrderSrv: listOrderSrv,
		repo:         repo,
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

	om.repo.Save(order)
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

	om.repo.Save(order)
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
