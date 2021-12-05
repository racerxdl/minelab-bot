package database

import (
	"encoding/json"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"time"
)

var log = logrus.New()

const markBucket = "marks"

type DB interface {
	AddPositionMark(username, mark string, position mgl32.Vec3) error
	GetPositionMark(username, mark string) (vec mgl32.Vec3, err error)
	Close() error
}

type database struct {
	db *bolt.DB
}

func MakeDB(filename string) (DB, error) {
	db, err := bolt.Open(filename, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Database opened at %q", filename)
	return &database{
		db: db,
	}, nil
}

func (d *database) Close() error {
	return d.db.Close()
}

func (d *database) AddPositionMark(username, mark string, position mgl32.Vec3) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		obj := positionMarker{
			Username: username,
			Mark:     mark,
			Position: position,
		}

		data, err := json.Marshal(obj)
		if err != nil {
			return fmt.Errorf("create serialize json: %s", err)
		}

		b, err := tx.CreateBucketIfNotExists([]byte(markBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		log.Infof("Saving marker %q", username+"-"+mark)
		return b.Put([]byte(username+"-"+mark), data)
	})
}

func (d *database) GetPositionMark(username, mark string) (vec mgl32.Vec3, err error) {
	k := []byte(username + "-" + mark)
	log.Infof("Reading marker %q", username+"-"+mark)
	err = d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(markBucket))
		v := b.Get(k)
		if v == nil {
			return fmt.Errorf("marker %s for user %s not found", mark, username)
		}

		m := positionMarker{}
		err = json.Unmarshal(v, &m)
		if err != nil {
			return fmt.Errorf("error deserializing json: %s", err)
		}

		vec = m.Position

		return nil
	})
	return vec, err
}
