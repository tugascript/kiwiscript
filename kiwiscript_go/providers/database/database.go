// Copyright (C) 2024 Afonso Barracha
//
// This file is part of KiwiScript.
//
// KiwiScript is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// KiwiScript is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with KiwiScript.  If not, see <https://www.gnu.org/licenses/>.

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

func (database *Database) FinalizeTx(ctx context.Context, txn pgx.Tx, err error) {
	if p := recover(); p != nil {
		if err := txn.Rollback(ctx); err != nil {
			panic(err)
		}
		panic(p)
	}
	if err != nil {
		if err := txn.Rollback(ctx); err != nil {
			panic(err)
		}
		return
	}
	if commitErr := txn.Commit(ctx); commitErr != nil {
		panic(commitErr)
	}
}

func (database *Database) RawQuery(ctx context.Context, sql string, args []interface{}) (pgx.Rows, error) {
	return database.connPool.Query(ctx, sql, args...)
}

func (database *Database) RawQueryRow(ctx context.Context, sql string, args []interface{}) pgx.Row {
	return database.connPool.QueryRow(ctx, sql, args...)
}
