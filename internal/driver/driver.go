package driver

import (
	"database/sql"
	"time"
	_ "github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB holds database connection pool
type DB struct {
	SQL *sql.DB
}

var dbConn = &DB{}

const (
	maxOpenDbConn     = 10
	maxIdleDbConn     = 5
	maxDbConnLifetime = 5 * time.Minute
)

// Creates a new database connection for Postgresql
func ConnectSql(dsn string) (*DB, error) {
	d, err := NewDB(dsn)
	if err != nil {
		panic(err)
	}
	d.SetMaxOpenConns(maxOpenDbConn)
	d.SetConnMaxIdleTime(maxIdleDbConn)
	d.SetConnMaxLifetime(maxDbConnLifetime)
	dbConn.SQL = d
	err = TestDb(d)
	if err != nil {
		return nil, err
	}
	return dbConn, nil
}

// Tries to ping the database
func TestDb(d *sql.DB) error {
	err := d.Ping()
	if err != nil {
		return err
	}
	return nil
}

// Creates a new database for the application
func NewDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
