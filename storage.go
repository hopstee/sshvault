package main

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

const RECORDS_BACKET = "records"
const INDEX_BUCKET = "names_index"

var ErrNotFound = errors.New("record not found")
var RecordBucketNotFound = errors.New("record bucket not found")
var IndexBucketNotFound = errors.New("index bucket not found")
var ConnectionExists = errors.New("connection with name already exists")
var ConnectionNotExists = errors.New("connection with name does not exist")

type Record struct {
	id            string
	name          string
	connectionCmd string
}

type Storage struct {
	db *bbolt.DB
}

func initDB(storePath string) (*bbolt.DB, error) {
	return bbolt.Open(storePath, 0600, &bbolt.Options{Timeout: 1 * time.Second})
}

func NewStorage(storePath string) (*Storage, error) {
	db, err := initDB(storePath)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(RECORDS_BACKET))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(INDEX_BUCKET))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}

func (s *Storage) records() ([]Record, error) {
	var records []Record
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(RECORDS_BACKET))
		if b == nil {
			return RecordBucketNotFound
		}

		return b.ForEach(func(k, v []byte) error {
			var r Record
			if err := json.Unmarshal(v, &r); err != nil {
				return err
			}

			records = append(records, r)
			return nil
		})
	})
	return records, err
}

func (s *Storage) create(name, connectionCmd string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		recordsB := tx.Bucket([]byte(RECORDS_BACKET))
		if recordsB == nil {
			return RecordBucketNotFound
		}

		indexB := tx.Bucket([]byte(INDEX_BUCKET))
		if indexB == nil {
			return IndexBucketNotFound
		}

		if v := indexB.Get([]byte(name)); v != nil {
			return ConnectionExists
		}

		k := uuid.New()
		v, err := json.Marshal(Record{
			id:            k.String(),
			name:          name,
			connectionCmd: connectionCmd,
		})
		if err != nil {
			return err
		}

		if err := recordsB.Put([]byte(k.String()), v); err != nil {
			return err
		}

		return indexB.Put([]byte(name), []byte(k.String()))
	})
}

func (s *Storage) update(name, connectionCmd string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		recordB := tx.Bucket([]byte(RECORDS_BACKET))
		if recordB == nil {
			return RecordBucketNotFound
		}

		indexB := tx.Bucket([]byte(INDEX_BUCKET))
		if indexB == nil {
			return IndexBucketNotFound
		}

		id := indexB.Get([]byte(name))
		if id == nil {
			return ConnectionNotExists
		}

		record := recordB.Get(id)
		if record == nil {
			return ConnectionNotExists
		}

		return recordB.Put(id, []byte(connectionCmd))
	})
}

func (s *Storage) find(name string) (Record, error) {
	var record Record
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(RECORDS_BACKET))
		if b == nil {
			return nil
		}
		v := b.Get([]byte(name))
		if v == nil {
			return ErrNotFound
		}
		record = Record{
			name:          name,
			connectionCmd: string(v),
		}
		return nil
	})
	return record, err
}

func (s *Storage) delete(name string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(RECORDS_BACKET))
		if b == nil {
			return nil
		}
		return b.Delete([]byte(name))
	})
}
