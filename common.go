package common

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"github.com/tsawler/celeritas/cache"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const version = "1.0.0"

var myRedisCache *cache.RedisCache
var redisPool *redis.Pool

type Common struct {
	AppName       string
	Debug         bool
	Version       string
	RootPath      string
	Routes        *chi.Mux
	DB            *gorm.DB
	S3            *session.Session
	config        config
	EncryptionKey string
	Cache         cache.Cache
	Server        Server
}

func (c *Common) checkDotEnv(path string) error {
	err := c.CreateFileIfNotExists(fmt.Sprintf("%s/.env", path))
	if err != nil {
		return err
	}
	return nil
}

// BuildDSN builds the datasource name for our database, and returns it as a string
func (c *Common) BuildDSN() string {
	var dsn string

	switch os.Getenv("DATABASE_TYPE") {
	case "postgres", "postgresql":
		dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_PORT"),
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_NAME"),
			os.Getenv("DATABASE_SSL_MODE"))

		// we check to see if a database passsword has been supplied, since including "password=" with nothing
		// after it sometimes causes postgres to fail to allow a connection.
		if os.Getenv("DATABASE_PASS") != "" {
			dsn = fmt.Sprintf("%s password=%s", dsn, os.Getenv("DATABASE_PASS"))
		}
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", os.Getenv("DATABASE_USER"), os.Getenv("DATABASE_PASS"), os.Getenv("DATABASE_HOST"), os.Getenv("DATABASE_PORT"), os.Getenv("DATABASE_NAME"))
	default:

	}
	return dsn
}

func (c *Common) New(rootPath string) error {
	//check env file exists
	err := c.checkDotEnv(rootPath)
	if err != nil {
		return err
	}

	// read .env
	err = godotenv.Load(rootPath + "/.env")
	if err != nil {
		return err
	}
	c.config = config{}

	// connect to database
	if os.Getenv("DATABASE_TYPE") != "" {
		sqldb, err := c.OpenDB(os.Getenv("DATABASE_TYPE"), c.BuildDSN())
		if err != nil {
			fmt.Println("Error connecting database")
			os.Exit(1)
		}
		sqldb.SetConnMaxLifetime(3 * time.Second)
		db, err := gorm.Open(mysql.New(mysql.Config{
			Conn: sqldb,
		}), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})

		if err != nil {
			fmt.Println("Error connecting database 2")
			os.Exit(1)
		}
		c.config.database = databaseConfig{
			dsn:      c.BuildDSN(),
			database: os.Getenv("DATABASE_TYPE"),
		}

		c.DB = db
		// defer func() {
		// 	dbInstance, _ := c.DB.DB()
		// 	_ = dbInstance.Close()
		// }()

	}

	if os.Getenv("CACHE") == "redis" || os.Getenv("SESSION_TYPE") == "redis" {
		myRedisCache = c.createClientRedisCache()
		c.Cache = myRedisCache
		redisPool = myRedisCache.Conn
		c.config.redis = redisConfig{
			host:     os.Getenv("REDIS_HOST"),
			password: os.Getenv("REDIS_PASSWORD"),
			prefix:   os.Getenv("REDIS_PREFIX"),
		}
		defer redisPool.Close()
	}

	if os.Getenv("AWS_SECRET") != "" || os.Getenv("AWS_TOKEN") != "" {
		s3session, err := c.createS2Session()
		if err != nil {
			fmt.Println("Error conecting AWS")
			os.Exit(1)
		}
		c.config.s3 = s3Config{
			Region:  os.Getenv("AWS_S3_REGION"),
			Token:   os.Getenv("AWS_TOKEN"),
			Secret:  os.Getenv("AWS_SECRET"),
			session: s3session,
		}
		c.S3 = s3session
	}

	c.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	c.Version = version
	c.RootPath = rootPath
	c.Routes = c.routes().(*chi.Mux)
	c.config.port = os.Getenv("PORT")

	secure := true
	if strings.ToLower(os.Getenv("SECURE")) == "false" {
		secure = false
	}

	c.Server = Server{
		ServerName: os.Getenv("SERVER_NAME"),
		Port:       os.Getenv("PORT"),
		Secure:     secure,
		URL:        os.Getenv("APP_URL"),
	}
	c.EncryptionKey = os.Getenv("KEY")
	return nil
}

// ListenAndServe starts the web server
func (c *Common) ListenAndServe() {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler:      c.Routes,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
	}

	fmt.Printf("Listening on port %s \n\n", os.Getenv("PORT"))
	err := srv.ListenAndServe()
	if err != nil {
		fmt.Println("Error ", err)
	}
	//c.ErrorLog.Fatal(err)
}
