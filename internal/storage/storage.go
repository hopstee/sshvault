package storage

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
	ID      string
	Name    string
	Address string
	Port    int
	User    string
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

func (s *Storage) Records() ([]Record, error) {
	var records []Record
	err := s.db.View(func(tx *bbolt.Tx) error {
		recordsBucket := tx.Bucket([]byte(RECORDS_BACKET))
		if recordsBucket == nil {
			return RecordBucketNotFound
		}

		return recordsBucket.ForEach(func(k, v []byte) error {
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

func (s *Storage) Create(name, address, user string, port int) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		recordsBucket := tx.Bucket([]byte(RECORDS_BACKET))
		if recordsBucket == nil {
			return RecordBucketNotFound
		}

		indexBucket := tx.Bucket([]byte(INDEX_BUCKET))
		if indexBucket == nil {
			return IndexBucketNotFound
		}

		if v := indexBucket.Get([]byte(name)); v != nil {
			return ConnectionExists
		}

		k := uuid.New()
		r := Record{
			ID:      k.String(),
			Name:    name,
			Address: address,
			Port:    port,
			User:    user,
		}
		v, err := json.Marshal(r)
		if err != nil {
			return err
		}
		if err := recordsBucket.Put([]byte(k.String()), v); err != nil {
			return err
		}

		return indexBucket.Put([]byte(name), []byte(k.String()))
	})
}

func (s *Storage) Find(name string) (Record, error) {
	var record Record
	err := s.db.View(func(tx *bbolt.Tx) error {
		recordBucket := tx.Bucket([]byte(RECORDS_BACKET))
		if recordBucket == nil {
			return nil
		}

		indexBucket := tx.Bucket([]byte(INDEX_BUCKET))
		if indexBucket == nil {
			return IndexBucketNotFound
		}

		id := indexBucket.Get([]byte(name))
		if id == nil {
			return ConnectionNotExists
		}

		v := recordBucket.Get(id)
		if v == nil {
			return ErrNotFound
		}
		if err := json.Unmarshal(v, &record); err != nil {
			return err
		}
		return nil
	})
	return record, err
}

func (s *Storage) Delete(name string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		recordBucket := tx.Bucket([]byte(RECORDS_BACKET))
		if recordBucket == nil {
			return nil
		}

		indexBucket := tx.Bucket([]byte(INDEX_BUCKET))
		if indexBucket == nil {
			return IndexBucketNotFound
		}

		id := indexBucket.Get([]byte(name))
		if id == nil {
			return ConnectionNotExists
		}

		if err := recordBucket.Delete(id); err != nil {
			return err
		}

		return indexBucket.Delete([]byte(name))
	})
}
