package store

import (
	"bytes"
	"encoding/gob"
	"github.com/dgraph-io/badger/v2"
	"github.com/finitum/AAAAA/pkg/models"
	"github.com/pkg/errors"
)

type Badger struct {
	db *badger.DB
}

const pkgPrefix = "pkg_"
const userPrefix = "user_"

func OpenBadgerStore(path string) (*Badger, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, errors.Wrap(err, "opening badger store")
	}

	return &Badger{
		db,
	}, nil
}

func (b *Badger) GetPackage(name string) (*models.Pkg, error) {
	var pkg *models.Pkg

	return pkg, b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(pkgPrefix + name))
		if err != nil {
			return errors.Wrap(err, "badger get")
		}

		return errors.Wrap(item.Value(func(val []byte) error {
			buf := bytes.NewBuffer(val)

			dec := gob.NewDecoder(buf)
			return errors.Wrap(dec.Decode(&pkg), "gob decode")
		}), "badger read")
	})
}

func (b *Badger) AddPackage(pkg *models.Pkg) error {
	return b.db.Update(func(txn *badger.Txn) error {
		var value bytes.Buffer

		enc := gob.NewEncoder(&value)
		err := enc.Encode(pkg)
		if err != nil {
			return errors.Wrap(err, "gob encode")
		}

		return errors.Wrap(txn.Set([]byte(pkgPrefix+pkg.Name), value.Bytes()), "badger transaction")
	})
}

func (b *Badger) DelPackage(pkg *models.Pkg) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return errors.Wrap(txn.Delete([]byte(pkgPrefix+pkg.Name)), "badger transaction")
	})
}

func (b *Badger) AllPackages() (pkgs []*models.Pkg, err error) {
	return pkgs, b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(pkgPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			var pkg *models.Pkg
			err := item.Value(func(val []byte) error {
				buf := bytes.NewBuffer(val)

				dec := gob.NewDecoder(buf)
				return errors.Wrap(dec.Decode(&pkg), "gob decode")
			})
			pkgs = append(pkgs, pkg)
			if err != nil {
				return errors.Wrap(err, "badger iteration")
			}
		}
		return nil
	})
}

func (b *Badger) AllPackageNames() ([]string, error) {
	pkgs, err := b.AllPackages()
	if err != nil {
		return nil, errors.Wrap(err, "all packages")
	}
	var pkgNames []string
	for _, v := range pkgs {
		pkgNames = append(pkgNames, v.Name)
	}
	return pkgNames, nil
}

func (b *Badger) GetUser(name string) (user *models.User, err error) {
	err = b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(userPrefix + name))
		if err != nil {
			return errors.Wrap(err, "badger get")
		}

		return errors.Wrap(item.Value(func(val []byte) error {
			buf := bytes.NewBuffer(val)

			dec := gob.NewDecoder(buf)
			return errors.Wrap(dec.Decode(&user), "gob decode")
		}), "badger read")
	})
	return
}

func (b *Badger) AddUser(user *models.User) error {
	return b.db.Update(func(txn *badger.Txn) error {
		var value bytes.Buffer

		enc := gob.NewEncoder(&value)
		err := enc.Encode(user)
		if err != nil {
			return errors.Wrap(err, "gob encode")
		}

		return errors.Wrap(txn.Set([]byte(userPrefix+user.Username), value.Bytes()), "badger transaction")
	})
}

func (b *Badger) DelUser(user *models.User) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return errors.Wrap(txn.Delete([]byte(userPrefix+user.Username)), "badger transaction")
	})
}
