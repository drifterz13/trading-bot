package main

import (
	"log"
	"os"
	"time"

	"github.com/adshao/go-binance/v2"
)

var symbols = []string{"ALGOUSDT", "SOLUSDT", "MATICUSDT", "ADAUSDT", "BTCUSDT"}

func main() {
	client := binance.NewClient(os.Getenv("BINANCE_API_KEY"), os.Getenv("BINANCE_SECRET_KEY"))

	for {
		for _, symbol := range symbols {
			bot := NewBot(client)
			bot.Run(symbol)
		}

		log.Println("going to sleep for 15 minutes.")
		time.Sleep(15 * time.Minute)
	}
}
