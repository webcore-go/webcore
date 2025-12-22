package mongo

import (
	"github.com/semanggilab/webcore-go/app/loader"
)

type MongoLoader struct {
	name string
}

func (a *MongoLoader) SetClassName(name string) {
	a.name = name
}

func (a *MongoLoader) ClassName() string {
	return a.name
}

func (l *MongoLoader) Init(args ...any) (loader.Library, error) {
	// config := args[1].(config.DatabaseConfig)

	db := &MongoDatabase{}
	err := db.Install(args...)
	if err != nil {
		return nil, err
	}

	err = db.Connect()
	if err != nil {
		return nil, err
	}

	return db, nil
}
