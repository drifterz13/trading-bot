package dto

import "strconv"

var (
	BuyType  = "BUY"
	SellType = "SELL"
)

type Kline struct {
	Open  float64
	High  float64
	Low   float64
	Close float64
}

type Order struct {
	Symbol   string
	Quantity string
	Price    string
	Type     string
}

type OrderFloat64 struct {
	Symbol   string
	Quantity float64
	Price    float64
	Type     string
}

func (o *Order) IsEmpty() bool {
	return *o == (Order{})
}

func (o *Order) ToFloat64() *OrderFloat64 {
	var err error
	var qty float64
	var price float64

	qty, err = strconv.ParseFloat(o.Quantity, 64)
	price, err = strconv.ParseFloat(o.Price, 64)
	if err != nil {
		panic(err)
	}

	return &OrderFloat64{
		Symbol:   o.Symbol,
		Quantity: qty,
		Price:    price,
		Type:     o.Type,
	}
}

func (o *OrderFloat64) ToString() *Order {
	return &Order{
		Symbol:   o.Symbol,
		Quantity: strconv.FormatFloat(o.Quantity, 'f', -1, 64),
		Price:    strconv.FormatFloat(o.Price, 'f', -1, 64),
		Type:     o.Type,
	}
}
