package common

import (
	"fmt"
	"log"
	"os"

	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
)

type DatabaseConfig struct {
	dbType   string
	host     string
	port     string
	user     string
	dbname   string
	password string
}

func (d *DatabaseConfig) BuildDSN() string {
	var dsn string
	switch os.Getenv("DATABASE_TYPE") {
	case "mysql":
		return fmt.Sprintf("%s:%s@/%s", d.user, d.password, d.dbname)
	default:
	}
	return dsn
}

func (d *DatabaseConfig) OpenDbConn() (db.Session, error) {
	var db db.Session
	var err error
	switch os.Getenv("DATABASE_TYPE") {
	case "mysql":
		settings := mysql.ConnectionURL{
			Database: d.dbname,
			Host:     d.host,
			User:     d.user,
			Password: d.password,
		}
		db, err = mysql.Open(settings)
		if err != nil {
			log.Fatal("db connection Open: ", err)
			return nil, err
		}
		return db, nil
	default:
		return nil, nil
	}
}
