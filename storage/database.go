package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type Database struct {
	Storage               Storage
	db                    *sql.DB
	host                  string
	port                  int
	user                  string
	password              string
	name                  string
	maxOpenConnections    int
	maxIdleConnections    int
	connectionMaxLifetime time.Duration
	connectionMaxIdleTime time.Duration
	log                   *zap.Logger
}

type NewDatabaseOptions struct {
	Host                   string
	Port                   int
	User                   string
	Password               string
	Name                   string
	MaxOpenConnections     int
	MaxIdleConnections     int
	ConnectionMaxLifetime  time.Duration
	ConnnectionMaxIdleTime time.Duration
	Log                    *zap.Logger
}

func NewDatabase(opts NewDatabaseOptions) *Database {
	if opts.Log == nil {
		opts.Log = zap.NewNop()
	}

	return &Database{
		host:               opts.Host,
		port:               opts.Port,
		user:               opts.User,
		password:           opts.Password,
		name:               opts.Name,
		maxOpenConnections: opts.MaxOpenConnections,
		maxIdleConnections: opts.MaxIdleConnections,
		log:                opts.Log,
	}
}

func (d *Database) dsn() string {
	appEnv := os.Getenv("APP_ENV")
	ssl := ""
	if appEnv == "" || appEnv == "local" || appEnv == "development" {
		ssl = "sslmode=disable"
	}

	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s %s", // sslmode=disable",
		d.host, d.port, d.user, d.password, d.name, ssl,
	)
}

func (d *Database) Connect() error {
	d.log.Info("Connecting to database", zap.String("url", d.dsn()))

	var err error
	d.db, err = sql.Open("postgres", d.dsn())
	if err != nil {
		return err
	}

	d.Storage = NewStorage(d.db)

	d.log.Debug(
		"Setting connection pool options",
		zap.Int("max open connections", d.maxOpenConnections),
		zap.Int("max idle connections", d.maxIdleConnections),
		zap.Duration("connection max lifetime", d.connectionMaxLifetime),
		zap.Duration("connection max idle time", d.connectionMaxIdleTime),
	)

	d.db.SetMaxOpenConns(d.maxOpenConnections)
	d.db.SetMaxIdleConns(d.maxIdleConnections)
	d.db.SetConnMaxLifetime(d.connectionMaxIdleTime)
	d.db.SetConnMaxIdleTime(d.connectionMaxIdleTime)

	return nil
}

func (d *Database) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	if err := d.db.PingContext(ctx); err != nil {
		return err
	}
	_, err := d.db.ExecContext(ctx, `select 1`)
	return err
}
