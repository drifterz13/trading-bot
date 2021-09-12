package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	bolt "go.etcd.io/bbolt"
)

func NewDB() (*bolt.DB, error) {
	var dbPath string
	if os.Getenv("APP_ENV") == "dev" {
		dbPath = "./data/dev.db"
		log.Println("use dev db")
	} else {
		dbPath = "./data/prod.db"
		log.Println("use prod db")
	}

	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, err
	}

	return db, err
}

type DataStore interface {
	CreateBucket(name string)
	Save(order *Order)
	Last(bucket string) *Order
	GetAll(bucket string)
}

type dataStore struct {
	db *bolt.DB
}

func NewDataStore(db *bolt.DB) DataStore {
	return &dataStore{db}
}

func (r *dataStore) CreateBucket(name string) {
	r.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}

func (r *dataStore) Save(order *Order) {
	r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(order.Symbol))
		now := time.Now().Format(time.RFC3339)

		byteOrder, err := json.Marshal(&order)
		if err != nil {
			panic(err)
		}

		if err := b.Put([]byte(now), []byte(byteOrder)); err != nil {
			return err
		}

		return nil
	})
}

func (r *dataStore) Last(bucket string) *Order {
	var order Order
	r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		c := b.Cursor()

		_, v := c.Last()

		if err := json.Unmarshal(v, &order); err != nil {
			return fmt.Errorf("unmarshal value: %v\n", err)
		}

		return nil
	})

	return &order
}

func (r *dataStore) GetAll(bucket string) {
	r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var order Order
			if err := json.Unmarshal(v, &order); err != nil {
				panic(err)
			}
			log.Printf("get all: %v, price: %v, quantity: %v, type: %v\n", order.Symbol, order.Price, order.Quantity, order.Type)
		}

		return nil
	})
}
