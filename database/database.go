package database

import (
	"encoding/json"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"strings"
	"time"
)

var log = logrus.New()

const markBucket = "marks"

type DB interface {
	AddPositionMark(username, mark string, dimension int, position mgl32.Vec3) error
	GetPositionMark(username, mark string, dimension int) (vec mgl32.Vec3, err error)
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

	d := &database{
		db: db,
	}

	err = d.migrate()

	if err != nil {
		_ = d.Close()
		return nil, err
	}

	return d, nil
}

func (d *database) migrate() error {
	err := d.db.Update(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(markBucket))
		if b != nil {
			c := b.Cursor()

			var keysToUpdate []string

			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				s := strings.Split(string(k), "-")
				if len(s) == 2 { // Old key without dimension
					keysToUpdate = append(keysToUpdate, string(k))
				}
			}

			for _, k := range keysToUpdate {
				v := b.Get([]byte(k))
				s := strings.Split(string(k), "-")
				newK := fmt.Sprintf("%s-0-%s", s[0], s[1])
				log.Infof("migrating key %s to %s", string(k), newK)
				err := b.Put([]byte(newK), v)
				if err != nil {
					return err
				}
				err = b.Delete([]byte(k))
				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	return err
}

func (d *database) Close() error {
	return d.db.Close()
}

func (d *database) AddPositionMark(username, mark string, dimension int, position mgl32.Vec3) error {
	ks := fmt.Sprintf("%s-%d-%s", username, dimension, mark)
	return d.db.Update(func(tx *bolt.Tx) error {
		obj := positionMarker{
			Username:  username,
			Mark:      mark,
			Dimension: dimension,
			Position:  position,
		}

		data, err := json.Marshal(obj)
		if err != nil {
			return fmt.Errorf("create serialize json: %s", err)
		}

		b, err := tx.CreateBucketIfNotExists([]byte(markBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		log.Infof("Saving marker %q", ks)
		return b.Put([]byte(ks), data)
	})
}

func (d *database) GetPositionMark(username, mark string, dimension int) (vec mgl32.Vec3, err error) {
	ks := fmt.Sprintf("%s-%d-%s", username, dimension, mark)
	k := []byte(ks)
	log.Infof("Reading marker %q", ks)
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
