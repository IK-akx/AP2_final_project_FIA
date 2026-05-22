package postgres

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type DB struct {
	Conn *sql.DB
}

func NewDB(dsn string) (*DB, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err = conn.Ping(); err != nil {
		return nil, err
	}

	return &DB{Conn: conn}, nil
}
