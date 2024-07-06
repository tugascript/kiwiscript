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
