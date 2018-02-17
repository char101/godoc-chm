package main

import (
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

type Cache struct {
	db *bolt.DB
}

func newCache() *Cache {
	db, err := bolt.Open("cache.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("cache"))
		if err != nil {
			return fmt.Errorf("create bucket (cache): %v", err)
		}
		return nil
	})
	return &Cache{
		db: db,
	}
}

func (c *Cache) set(k string, v []byte) {
	err := c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("cache"))
		err := b.Put([]byte(k), v)
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Cache) get(k string) []byte {
	var val []byte
	c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("cache"))
		val = b.Get([]byte(k))
		return nil
	})
	return val
}

func (c *Cache) close() error {
	return c.db.Close()
}
