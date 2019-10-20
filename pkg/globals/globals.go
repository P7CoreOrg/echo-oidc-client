package globals

import (
	badger "github.com/dgraph-io/badger"
)

var (
	_rootFolder string
	_db         *badger.DB
)

func OpenBadgerDb(rootFolder string) (err error, db *badger.DB) {
	_db, err = badger.Open(badger.DefaultOptions(rootFolder).WithTruncate(true))
	if err != nil {
		return
	}
	_rootFolder = rootFolder
	db = _db
	return
}
func Delete(key string) (err error) {
	txn := _db.NewTransaction(true)
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

func Put(key string, value *[]byte) (err error) {
	txn := _db.NewTransaction(true)
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

func Get(key string) (err error, value *[]byte) {
	err = _db.View(func(txn *badger.Txn) error {
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
