// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: users.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const confirmUser = `-- name: ConfirmUser :one
UPDATE "users" SET
  "is_confirmed" = true,
  "version" = "version" + 1
WHERE "id" = $1
RETURNING id, first_name, last_name, location, email, birth_date, version, is_admin, is_staff, is_confirmed, password, created_at, updated_at
`

func (q *Queries) ConfirmUser(ctx context.Context, id int32) (User, error) {
	row := q.db.QueryRow(ctx, confirmUser, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Location,
		&i.Email,
		&i.BirthDate,
		&i.Version,
		&i.IsAdmin,
		&i.IsStaff,
		&i.IsConfirmed,
		&i.Password,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createUserWithPassword = `-- name: CreateUserWithPassword :one
INSERT INTO "users" (
  "first_name",
  "last_name",
  "location",
  "email",
  "birth_date",
  "password",
  "is_confirmed"
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  false
) RETURNING id, first_name, last_name, location, email, birth_date, version, is_admin, is_staff, is_confirmed, password, created_at, updated_at
`

type CreateUserWithPasswordParams struct {
	FirstName string
	LastName  string
	Location  string
	Email     string
	BirthDate pgtype.Date
	Password  pgtype.Text
}

func (q *Queries) CreateUserWithPassword(ctx context.Context, arg CreateUserWithPasswordParams) (User, error) {
	row := q.db.QueryRow(ctx, createUserWithPassword,
		arg.FirstName,
		arg.LastName,
		arg.Location,
		arg.Email,
		arg.BirthDate,
		arg.Password,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Location,
		&i.Email,
		&i.BirthDate,
		&i.Version,
		&i.IsAdmin,
		&i.IsStaff,
		&i.IsConfirmed,
		&i.Password,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createUserWithoutPassword = `-- name: CreateUserWithoutPassword :one
INSERT INTO "users" (
  "first_name",
  "last_name",
  "location",
  "email",
  "birth_date",
  "is_confirmed"
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  true
) RETURNING id, first_name, last_name, location, email, birth_date, version, is_admin, is_staff, is_confirmed, password, created_at, updated_at
`

type CreateUserWithoutPasswordParams struct {
	FirstName string
	LastName  string
	Location  string
	Email     string
	BirthDate pgtype.Date
}

func (q *Queries) CreateUserWithoutPassword(ctx context.Context, arg CreateUserWithoutPasswordParams) (User, error) {
	row := q.db.QueryRow(ctx, createUserWithoutPassword,
		arg.FirstName,
		arg.LastName,
		arg.Location,
		arg.Email,
		arg.BirthDate,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Location,
		&i.Email,
		&i.BirthDate,
		&i.Version,
		&i.IsAdmin,
		&i.IsStaff,
		&i.IsConfirmed,
		&i.Password,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteAllUsers = `-- name: DeleteAllUsers :exec
DELETE FROM "users"
`

func (q *Queries) DeleteAllUsers(ctx context.Context) error {
	_, err := q.db.Exec(ctx, deleteAllUsers)
	return err
}

const deleteUserById = `-- name: DeleteUserById :exec
DELETE FROM "users"
WHERE "id" = $1
`

func (q *Queries) DeleteUserById(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, deleteUserById, id)
	return err
}

const findUserByEmail = `-- name: FindUserByEmail :one
SELECT id, first_name, last_name, location, email, birth_date, version, is_admin, is_staff, is_confirmed, password, created_at, updated_at FROM "users"
WHERE "email" = $1 LIMIT 1
`

func (q *Queries) FindUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, findUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Location,
		&i.Email,
		&i.BirthDate,
		&i.Version,
		&i.IsAdmin,
		&i.IsStaff,
		&i.IsConfirmed,
		&i.Password,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const findUserById = `-- name: FindUserById :one
SELECT id, first_name, last_name, location, email, birth_date, version, is_admin, is_staff, is_confirmed, password, created_at, updated_at FROM "users"
WHERE "id" = $1 LIMIT 1
`

func (q *Queries) FindUserById(ctx context.Context, id int32) (User, error) {
	row := q.db.QueryRow(ctx, findUserById, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Location,
		&i.Email,
		&i.BirthDate,
		&i.Version,
		&i.IsAdmin,
		&i.IsStaff,
		&i.IsConfirmed,
		&i.Password,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateUser = `-- name: UpdateUser :exec
UPDATE "users" SET
  "first_name" = $1,
  "last_name" = $2,
  "location" = $3
WHERE "id" = $4
`

type UpdateUserParams struct {
	FirstName string
	LastName  string
	Location  string
	ID        int32
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) error {
	_, err := q.db.Exec(ctx, updateUser,
		arg.FirstName,
		arg.LastName,
		arg.Location,
		arg.ID,
	)
	return err
}

const updateUserEmail = `-- name: UpdateUserEmail :one
UPDATE "users" SET
  "email" = $1,
  "version" = "version" + 1
WHERE "id" = $2
RETURNING id, first_name, last_name, location, email, birth_date, version, is_admin, is_staff, is_confirmed, password, created_at, updated_at
`

type UpdateUserEmailParams struct {
	Email string
	ID    int32
}

func (q *Queries) UpdateUserEmail(ctx context.Context, arg UpdateUserEmailParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUserEmail, arg.Email, arg.ID)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Location,
		&i.Email,
		&i.BirthDate,
		&i.Version,
		&i.IsAdmin,
		&i.IsStaff,
		&i.IsConfirmed,
		&i.Password,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateUserPassword = `-- name: UpdateUserPassword :one
UPDATE "users" SET
  "password" = $1,
  "version" = "version" + 1
WHERE "id" = $2
RETURNING id, first_name, last_name, location, email, birth_date, version, is_admin, is_staff, is_confirmed, password, created_at, updated_at
`

type UpdateUserPasswordParams struct {
	Password pgtype.Text
	ID       int32
}

func (q *Queries) UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUserPassword, arg.Password, arg.ID)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Location,
		&i.Email,
		&i.BirthDate,
		&i.Version,
		&i.IsAdmin,
		&i.IsStaff,
		&i.IsConfirmed,
		&i.Password,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
