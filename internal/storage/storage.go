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

var (
	ErrNotFound             = errors.New("record not found")
	ErrRecordBucketNotFound = errors.New("record bucket not found")
	ErrIndexBucketNotFound  = errors.New("index bucket not found")
	ErrConnectionExists     = errors.New("connection with name already exists")
	ErrConnectionNotExists  = errors.New("connection with name does not exist")
)

type AuthType string

const (
	PasswordAuth AuthType = "password"
	KeyAuth      AuthType = "sshkey"
	AgentAuth    AuthType = "agent"
)

type Record struct {
	ID          string
	Name        string
	Address     string
	Port        int
	User        string
	AuthType    AuthType
	PathToKey   string
	PasswordKey string
	// TODO: add tags for classifications
	// Tags []string
}

type UpsertDto struct {
	Name        string
	Address     string
	User        string
	PathToKey   string
	PasswordKey string
	Port        int
	AuthType    AuthType
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
			return ErrRecordBucketNotFound
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

func (s *Storage) Create(dto UpsertDto) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		recordsBucket := tx.Bucket([]byte(RECORDS_BACKET))
		if recordsBucket == nil {
			return ErrRecordBucketNotFound
		}

		indexBucket := tx.Bucket([]byte(INDEX_BUCKET))
		if indexBucket == nil {
			return ErrIndexBucketNotFound
		}

		if v := indexBucket.Get([]byte(dto.Name)); v != nil {
			return ErrConnectionExists
		}

		ID := uuid.New()
		r := Record{
			ID:          ID.String(),
			Name:        dto.Name,
			Address:     dto.Address,
			Port:        dto.Port,
			User:        dto.User,
			AuthType:    dto.AuthType,
			PathToKey:   dto.PathToKey,
			PasswordKey: dto.PasswordKey,
		}
		v, err := json.Marshal(r)
		if err != nil {
			return err
		}
		if err := recordsBucket.Put([]byte(ID.String()), v); err != nil {
			return err
		}

		return indexBucket.Put([]byte(dto.Name), []byte(ID.String()))
	})
}

func (s *Storage) Update(name string, dto UpsertDto) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		recordsBucket := tx.Bucket([]byte(RECORDS_BACKET))
		if recordsBucket == nil {
			return ErrRecordBucketNotFound
		}

		indexBucket := tx.Bucket([]byte(INDEX_BUCKET))
		if indexBucket == nil {
			return ErrIndexBucketNotFound
		}

		id := indexBucket.Get([]byte(name))
		if id == nil {
			return ErrConnectionNotExists
		}

		v := recordsBucket.Get(id)
		if v == nil {
			return ErrNotFound
		}

		var record Record
		if err := json.Unmarshal(v, &record); err != nil {
			return err
		}

		updated := Record{
			ID:          record.ID,
			Name:        dto.Name,
			Address:     dto.Address,
			Port:        dto.Port,
			User:        dto.User,
			AuthType:    dto.AuthType,
			PathToKey:   dto.PathToKey,
			PasswordKey: dto.PasswordKey,
		}

		data, err := json.Marshal(updated)
		if err != nil {
			return err
		}

		if err := recordsBucket.Put(id, data); err != nil {
			return err
		}

		if record.Name != dto.Name {
			if err := indexBucket.Delete([]byte(record.Name)); err != nil {
				return err
			}

			if err := indexBucket.Put([]byte(dto.Name), id); err != nil {
				return err
			}
		}

		return nil
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
			return ErrIndexBucketNotFound
		}

		id := indexBucket.Get([]byte(name))
		if id == nil {
			return ErrConnectionNotExists
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
			return ErrIndexBucketNotFound
		}

		id := indexBucket.Get([]byte(name))
		if id == nil {
			return ErrConnectionNotExists
		}

		if err := recordBucket.Delete(id); err != nil {
			return err
		}

		return indexBucket.Delete([]byte(name))
	})
}
