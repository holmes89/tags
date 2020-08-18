package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/holmes89/tags/internal"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

var (
	tagBucket = []byte("tags")
	resourceBucket = []byte("resources")
)

type KVStore interface {
	GetResource(id string) (internal.Resource, error)
	GetAllResources() ([]internal.Resource, error)
	PutResource(id string, resource internal.Resource) error
	GetTag(id string) (internal.Tag, error)
	GetAllTags() ([]internal.Tag, error)
	PutTag(id string, tag internal.Tag) error
}

type boltkv struct {
	conn *bolt.DB
}

func NewBoltConnection(lc fx.Lifecycle, configuration internal.Configuration) *bolt.DB {
	dbFile := configuration.DatabaseFile
	if dbFile == "" {
		logrus.Fatal("database file missing")
	}
	logrus.WithField("path", dbFile).Info("connecting to database")
	conn, err := bolt.Open( dbFile, 0600, nil)
	if err != nil {
		logrus.WithError(err).Fatal("unable to open database")
	}
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logrus.Info("closing database")
			return conn.Close()
		},
	})
	err = conn.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(resourceBucket)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_, err = tx.CreateBucketIfNotExists(tagBucket)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		logrus.WithError(err).Fatal("unable to create buckets")
	}
	return conn
}

func NewKVStore(conn *bolt.DB) KVStore {
	return  &boltkv{conn: conn}
}

func (b *boltkv) GetResource(id string) (internal.Resource, error) {
	var resource internal.Resource
	err := b.conn.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(resourceBucket)
		res := bucket.Get([]byte(id))
		if res == nil {
			return internal.ErrNotFound
		}
		if err := json.Unmarshal(res, &resource); err != nil {
			logrus.WithError(err).Error("unable to unmarshall resource")
			return err
		}
		return nil
	})
	return resource, err
}

func (b *boltkv) GetAllResources() ([]internal.Resource, error) {
	var resources []internal.Resource
	err := b.conn.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(resourceBucket)
		return bucket.ForEach(func(k, v []byte) error {
			var res internal.Resource
			if err := json.Unmarshal(v, &res); err != nil {
				return err
			}
			resources = append(resources, res)
			return nil
		})
	})
	if err != nil {
		logrus.WithError(err).Error("unable to fetch results for resources")
		return resources, errors.New("unable to fetch results")
	}
	return resources, nil
}

func (b *boltkv) PutResource(id string, resource internal.Resource) error {
	return b.conn.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(resourceBucket)
		rbytes, err := json.Marshal(resource)
		if err != nil {
			logrus.WithError(err).Error("unable to marshall resource")
			return errors.New("unable to store resource")
		}
		if err := bucket.Put([]byte(id), rbytes); err != nil {
			logrus.WithError(err).Error("unable to write resource")
			return errors.New("unable to store resource")
		}
		return nil
	})
}

func (b *boltkv) GetTag(id string) (internal.Tag, error) {
	var tag internal.Tag
	err := b.conn.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(tagBucket)
		res := bucket.Get([]byte(id))
		if res == nil {
			return internal.ErrNotFound
		}
		if err := json.Unmarshal(res, &tag); err != nil {
			logrus.WithError(err).Error("unable to unmarshall tag")
			return err
		}
		return nil
	})
	return tag, err
}

func (b *boltkv) GetAllTags() ([]internal.Tag, error) {
	var tags []internal.Tag
	err := b.conn.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(tagBucket)
		return bucket.ForEach(func(k, v []byte) error {
			var res internal.Tag
			if err := json.Unmarshal(v, &res); err != nil {
				return err
			}
			tags = append(tags, res)
			return nil
		})
	})
	if err != nil {
		logrus.WithError(err).Error("unable to fetch results for tags")
		return tags, errors.New("unable to fetch tags")
	}
	return tags, nil
}

func (b *boltkv) PutTag(id string, tag internal.Tag) error {
	return b.conn.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(tagBucket)
		tbytes, err := json.Marshal(tag)
		if err != nil {
			logrus.WithError(err).Error("unable to marshall resource")
			return errors.New("unable to store resource")
		}
		if err := bucket.Put([]byte(id), tbytes); err != nil {
			logrus.WithError(err).Error("unable to write resource")
			return errors.New("unable to store resource")
		}
		return nil
	})
}