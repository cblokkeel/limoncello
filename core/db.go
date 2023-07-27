package core

import (
	"fmt"

	"go.etcd.io/bbolt"
)

type Pair struct {
	K []byte
	V []byte
}

type BoltDB struct {
	db *bbolt.DB
}

func NewBoltDB() (*BoltDB, error) {
	db, err := bbolt.Open("data.db", 0666, nil)
	if err != nil {
		return nil, err
	}

	return &BoltDB{
		db,
	}, nil
}

func (bdb *BoltDB) Close() error {
	return bdb.db.Close()
}

func (bdb *BoltDB) ReadCollections(colls []string) ([]*Pair, error) {
	pairs := []*Pair{}
	for _, coll := range colls {
		collPairs, err := bdb.ReadCollection(coll)
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, collPairs...)
	}
	return pairs, nil
}

func (bdb *BoltDB) ReadCollection(collName string) ([]*Pair, error) {
	pairs := []*Pair{}
	if err := bdb.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(collName))
		if b == nil {
			return fmt.Errorf("Collection %s does not exist", collName)
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			pairs = append(pairs, &Pair{k, v})
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return pairs, nil
}

func (bdb *BoltDB) CreateCollection(collName string) error {
	if err := bdb.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(collName))
		if err != nil {
			return fmt.Errorf("Error while creating new collection: %s", err)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (bdb *BoltDB) ReadKeyInCollection(collName string, k string) ([]byte, error) {
	var v []byte
	if err := bdb.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(collName))
		if b == nil {
			return fmt.Errorf("Collection %s does not exist", collName)
		}
		v = b.Get([]byte(k))
		return nil
	}); err != nil {
		return nil, err
	}
	return v, nil
}

func (bdb *BoltDB) UpsertInCollection(collName string, k string, v string) error {
	if err := bdb.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(collName))
		if b == nil {
			return fmt.Errorf("Invalid collection: %s", collName)
		}
		b.Put([]byte(k), []byte(v))
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (bdb *BoltDB) DeleteKeyInCollection(collName string, k string) error {
	if err := bdb.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(collName))
		if b == nil {
			return fmt.Errorf("Collection %s does not exist", collName)
		}
		err := b.Delete([]byte(k))
		if err != nil {
			return fmt.Errorf("Error while deleting key: %s in collection: %s. Error: %s", k, collName, err)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (bdb *BoltDB) DeleteCollection(collName string) error {
	if err := bdb.db.Update(func(tx *bbolt.Tx) error {
		err := tx.DeleteBucket([]byte(collName))
		if err != nil {
			return fmt.Errorf("Collection %s does not exist", collName)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
