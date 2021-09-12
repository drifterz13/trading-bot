package main

import (
	"context"
	"log"
	"strconv"

	"github.com/adshao/go-binance/v2"
)

type klineManager struct {
	client *binance.Client
	symbol string
}

func NewKlineManager(client *binance.Client) *klineManager {
	return &klineManager{client: client}
}

func (km *klineManager) GetAvgOHLC(symbol string, interval string, limit int) *Kline {
	klines, err := km.client.NewKlinesService().Symbol(symbol).
		Interval(interval).Limit(limit).Do(context.Background())

	if err != nil {
		log.Fatalf("error GetAvgOHLC: %v\n", err)
	}

	var avgopen float64
	var avghigh float64
	var avglow float64
	var avgclose float64

	for _, k := range klines {
		o, err := strconv.ParseFloat(k.Open, 32)
		if err != nil {
			panic(err)
		}
		avgopen = avgopen + o

		h, err := strconv.ParseFloat(k.High, 32)
		if err != nil {
			panic(err)
		}
		avghigh = avghigh + h

		l, err := strconv.ParseFloat(k.Low, 32)
		if err != nil {
			panic(err)
		}
		avglow = avglow + l

		c, err := strconv.ParseFloat(k.Close, 32)
		if err != nil {
			panic(err)
		}
		avgclose = avgclose + c

	}

	avgopen = avgopen / float64(len(klines))
	avghigh = avghigh / float64(len(klines))
	avglow = avglow / float64(len(klines))
	avgclose = avgclose / float64(len(klines))

	return &Kline{Open: avgopen, High: avghigh, Low: avglow, Close: avgclose}
}
