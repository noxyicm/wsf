package resource

import (
	"wsf/config"
	"wsf/db"
)

// TYPEDb id of resource
const TYPEDb = "db"

func init() {
	Register(TYPEDb, NewDbResource)
}

// NewDbResource creates a new resource of type Db
func NewDbResource(options config.Config) (Interface, error) {
	dbb, err := db.NewDB(options)
	if err != nil {
		return nil, err
	}

	db.SetInstance(dbb)
	return dbb, nil
}
