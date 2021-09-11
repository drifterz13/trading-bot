package database

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/drifterz13/trading-bot/internal/dto"
	bolt "go.etcd.io/bbolt"
)

type BoltRepository interface {
	CreateBucket(name string)
	Save(order *dto.Order)
	Last(bucket string) *dto.Order
}

type boltRepository struct {
	db *bolt.DB
}

func NewBoltRepository(db *bolt.DB) BoltRepository {
	return &boltRepository{db}
}

func (r *boltRepository) CreateBucket(name string) {
	r.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}

func (r *boltRepository) Save(order *dto.Order) {
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

func (r *boltRepository) Last(bucket string) *dto.Order {
	var order dto.Order
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
