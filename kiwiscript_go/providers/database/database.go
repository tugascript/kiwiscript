package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	connPool *pgxpool.Pool
	*Queries
}

func NewDatabase(connPool *pgxpool.Pool) *Database {
	return &Database{
		connPool: connPool,
		Queries:  New(connPool),
	}
}

func (database *Database) BeginTx(ctx context.Context) (*Queries, pgx.Tx, error) {
	txn, err := database.connPool.BeginTx(ctx, pgx.TxOptions{
		DeferrableMode: pgx.Deferrable,
		IsoLevel:       pgx.ReadCommitted,
		AccessMode:     pgx.ReadWrite,
	})

	if err != nil {
		return nil, nil, err
	}

	return database.WithTx(txn), txn, nil
}
