package common

import (
	"database/sql"

	"github.com/aws/aws-sdk-go/aws/session"
)

type databaseConfig struct {
	dsn      string
	database string
}

type Database struct {
	DataType string
	Pool     *sql.DB
}

type redisConfig struct {
	host     string
	password string
	prefix   string
}

type s3Config struct {
	Region  string
	Secret  string
	Token   string
	session *session.Session
}

type Server struct {
	ServerName string
	Port       string
	Secure     bool
	URL        string
}

type config struct {
	port     string
	database databaseConfig
	redis    redisConfig
	s3       s3Config
}
