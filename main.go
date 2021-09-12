package main

import (
	"log"
	"os"
	"time"

	"github.com/adshao/go-binance/v2"
)

var (
	apiKey    = os.Getenv("BINANCE_API_KEY")
	secretKey = os.Getenv("BINANCE_SECRET_KEY")
	symbols   = []string{"ALGOUSDT", "SOLUSDT", "MATICUSDT", "ADAUSDT", "BTCUSDT"}
	delay     = 15 * time.Minute
)

func main() {
	client := binance.NewClient(apiKey, secretKey)

	for {
		for _, symbol := range symbols {
			bot := NewBot(client)
			bot.Run(symbol)
		}

		log.Println("going to sleep for 15 minutes.")
		time.Sleep(delay)
	}
}
