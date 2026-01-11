package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db               *sql.DB
	dbCriticalLogger *log.Logger
}

func NewDB(connectStr string) (*PostgresRepository, error) {

	dataBase := &PostgresRepository{
		dbCriticalLogger: log.New(os.Stderr, "[DB CONNECTION ERROR] ", log.Ldate|log.Ltime),
	}
	db, err := sql.Open("postgres", connectStr)
	if err != nil {
		dataBase.dbCriticalLogger.Printf("CRITICAL: error in connect dataBase: %s\n", err.Error())
		return nil, fmt.Errorf("error in connect dataBase: %s\n", err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		dataBase.dbCriticalLogger.Printf("CRITICAL: error in ping dataBase: %s\n", err.Error())
		return nil, fmt.Errorf("error in ping dataBase: %s\n", err.Error())
	}
	log.Println("Database connected successfully")
	dataBase.db = db
	return dataBase, nil
}

func (pr *PostgresRepository) Close() error {
	return pr.db.Close()
}

func (pr *PostgresRepository) PingWithTimeout(duration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	return pr.db.PingContext(ctx)
}

func (pr *PostgresRepository) PingWithCtx(ctx context.Context) error {
	return pr.db.PingContext(ctx)
}

func (pr *PostgresRepository) Begin() (*sql.Tx, error) {
	return pr.db.Begin()
}

func (pr *PostgresRepository) Name() string {
	return "PostgreSQL"
}
