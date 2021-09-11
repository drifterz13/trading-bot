package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/adshao/go-binance/v2"
	bot "github.com/drifterz13/trading-bot/internal/bot"
	db "github.com/drifterz13/trading-bot/internal/database"
	bolt "go.etcd.io/bbolt"
)

var (
	apiKey    = os.Getenv("BINANCE_API_KEY")
	secretKey = os.Getenv("BINANCE_SECRET_KEY")
	symbols   = []string{"ALGOUSDT", "SOLUSDT", "MATICUSDT", "ADAUSDT", "BTCUSDT"}
	delay     = 15 * time.Minute
)

func main() {
	flag.Parse()

	var dbPath string
	if os.Getenv("APP_ENV") == "dev" {
		dbPath = "./data/dev.db"
	} else {
		dbPath = "./data/prod.db"
	}

	boltDB, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer boltDB.Close()

	repo := db.NewBoltRepository(boltDB)
	client := binance.NewClient(apiKey, secretKey)

	for {
		for _, symbol := range symbols {
			repo.CreateBucket(symbol)
			b := bot.NewBot(client, repo)
			b.Run(symbol)
		}

		time.Sleep(delay)
	}
}
