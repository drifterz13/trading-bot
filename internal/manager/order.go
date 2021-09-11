package manager

import (
	"context"
	"os"

	"github.com/adshao/go-binance/v2"
	db "github.com/drifterz13/trading-bot/internal/database"
	"github.com/drifterz13/trading-bot/internal/dto"
)

var appEnv = os.Getenv("APP_ENV")

type orderManager struct {
	orderSrv     *binance.CreateOrderService
	listOrderSrv *binance.ListOrdersService
	repo         db.BoltRepository
}

func NewOrderManager(client *binance.Client, repo db.BoltRepository) *orderManager {
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

func (om *orderManager) Sell(order *dto.Order) {
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

func (om *orderManager) Buy(order *dto.Order) {
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
		if o.Symbol == symbol {
			return true
		}
	}

	return false
}
