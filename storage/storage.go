package storage

import (
	"github.com/ansriaz/redzilla/model"
	"github.com/sirupsen/logrus"
)

var store *Store

//GetStore return the store instance
func GetStore(collection string, cfg *model.Config) *Store {
	if store == nil {
		logrus.Debugf("Initializing store at %s", cfg.StorePath)
		store = NewStore(collection, cfg.StorePath)
	}
	return store
}
