package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

type Database struct {
	conn *pgx.Conn
}

func New(ctx context.Context, url string) *Database {
	if url == "" {
		log.Fatal("database url is empty")
	}

	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	return &Database{
		conn: conn,
	}
}

func (db *Database) Ping(ctx context.Context) error {
	return db.conn.Ping(ctx)
}
