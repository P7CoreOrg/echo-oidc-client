package globals

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	bolt "github.com/boltdb/bolt"
	badger "github.com/dgraph-io/badger"
)

var (
	_rootFolder string
	_dbBadger   *badger.DB
	_dbBolt     *bolt.DB
	useBolt     = true
)

func UseBolt() bool {
	return useBolt
}
func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func OpenBoltDb(rootFolder string) (err error, db *bolt.DB) {
	_rootFolder = rootFolder
	createDirIfNotExist(_rootFolder)
	dbPath := path.Join(_rootFolder, "_afx_kv.db")

	_dbBolt, err = bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})

	if err != nil {
		log.Fatal(err)
	}
	err = _dbBolt.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("MyBucket"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	db = _dbBolt
	return
}
func OpenBadgerDb(rootFolder string) (err error, db *badger.DB) {
	_dbBadger, err = badger.Open(badger.DefaultOptions(rootFolder).WithTruncate(true))
	if err != nil {
		return
	}
	_rootFolder = rootFolder
	db = _dbBadger
	return
}

func BoltPut(key string, value *[]byte) (err error) {
	err = _dbBolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyBucket"))
		err := b.Put([]byte(key), *value)
		return err
	})
	return
}
func BoltGet(key string) (err error, value *[]byte) {

	err = _dbBolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyBucket"))
		v := b.Get([]byte(key))
		value = &v
		return nil
	})
	return
}

func BoltDelete(key string) (err error) {
	err = _dbBolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyBucket"))
		err := b.Delete([]byte(key))
		return err
	})
	return
}

func Put(key string, value *[]byte) (err error) {
	if useBolt {
		return BoltPut(key, value)
	} else {
		return BadgerPut(key, value)
	}
}
func Get(key string) (err error, value *[]byte) {
	if useBolt {
		return BoltGet(key)
	} else {
		return BadgerGet(key)
	}

}
func Delete(key string) (err error) {
	if useBolt {
		return BoltDelete(key)
	} else {
		return BadgerDelete(key)
	}

}

func BadgerDelete(key string) (err error) {
	txn := _dbBadger.NewTransaction(true)
	defer txn.Discard()
	err = txn.Delete([]byte(key))
	if err != nil {
		return
	}
	// Commit the transaction and check for error.
	if err = txn.Commit(); err != nil {
		return
	}
	return
}

func BadgerPut(key string, value *[]byte) (err error) {
	txn := _dbBadger.NewTransaction(true)
	defer txn.Discard()

	// Use the transaction...
	err = txn.Set([]byte(key), *value)
	if err != nil {
		return
	}

	// Commit the transaction and check for error.
	if err = txn.Commit(); err != nil {
		return
	}
	return
}

func BadgerGet(key string) (err error, value *[]byte) {
	err = _dbBadger.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		var valCopy []byte
		err = item.Value(func(val []byte) error {
			// This func with val would only be called if item.Value encounters no error.

			// Copying or parsing val is valid.
			valCopy = append([]byte{}, val...)

			// Assigning val slice to another variable is NOT OK.
			//		  valNot = val // Do not do this.
			return nil
		})
		if err != nil {
			return err
		}
		value = &valCopy

		return nil
	})

	return
}
