package services

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/utils"
)

type CreateUserOptions struct {
	FirstName string
	LastName  string
	Location  string
	Email     string
	BirthDate time.Time
	Password  string
	Provider  string
}

func (s *Services) CreateUser(ctx context.Context, options CreateUserOptions) (db.User, *ServiceError) {
	log := s.log.WithGroup("services.users.CreateUser")
	var birthDate pgtype.Date
	var user db.User

	err := birthDate.Scan(options.BirthDate)
	if err != nil {
		log.InfoContext(ctx, "birthdate is invalid date format", "birthdate", options.BirthDate)
		return user, NewValidationError("'birthdate' is invalid date format")
	}

	var provider string
	var password pgtype.Text

	switch options.Provider {
	case utils.ProviderEmail:
		provider = options.Provider
		err := password.Scan(options.Password)
		if err != nil {
			return user, NewValidationError("'password' is invalid")
		}
	case utils.ProviderFacebook, utils.ProviderGoogle:
		provider = options.Provider
	default:
		return user, NewServerError("'provider' must be 'email', 'facebook' or 'google'")
	}

	location := strings.ToUpper(options.Location)
	if _, ok := utils.Location[location]; !ok {
		location = utils.LocationOTH
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		return user, FromDBError(err)
	}

	defer func() {
		if p := recover(); p != nil {
			txn.Rollback(ctx)
			panic(p)
		}
		if err != nil {
			txn.Rollback(ctx)
			return
		}
		if commitErr := txn.Commit(ctx); commitErr != nil {
			panic(commitErr)
		}
	}()

	if provider == utils.ProviderEmail {
		user, err = qrs.CreateUserWithPassword(ctx, db.CreateUserWithPasswordParams{
			FirstName: options.FirstName,
			LastName:  options.LastName,
			Email:     options.Email,
			BirthDate: birthDate,
			Password:  password,
			Location:  location,
		})
	} else {
		user, err = qrs.CreateUserWithoutPassword(ctx, db.CreateUserWithoutPasswordParams{
			FirstName: options.FirstName,
			LastName:  options.LastName,
			Email:     options.Email,
			BirthDate: birthDate,
			Location:  location,
		})
	}

	if err != nil {
		return user, FromDBError(err)
	}
	if err = qrs.CreateAuthProvider(ctx, db.CreateAuthProviderParams{
		Email:    user.Email,
		Provider: provider,
	}); err != nil {
		return db.User{}, FromDBError(err)
	}

	return user, nil
}

func (s *Services) FindUserByEmail(ctx context.Context, email string) (db.User, *ServiceError) {
	user, err := s.database.FindUserByEmail(ctx, email)

	if err != nil {
		return user, FromDBError(err)
	}

	return user, nil
}

func (s *Services) FindUserByID(ctx context.Context, id int32) (db.User, *ServiceError) {
	user, err := s.database.FindUserById(ctx, id)

	if err != nil {
		return db.User{}, FromDBError(err)
	}

	return user, nil
}

type UpdateUserEmailOptions struct {
	ID    int32
	Email string
}

func (s *Services) UpdateUserEmail(ctx context.Context, options UpdateUserEmailOptions) (db.User, *ServiceError) {
	user, err := s.database.UpdateUserEmail(ctx, db.UpdateUserEmailParams{
		Email: options.Email,
		ID:    options.ID,
	})

	if err != nil {
		return user, FromDBError(err)
	}

	return user, nil
}

type UpdateUserPasswordOptions struct {
	ID       int32
	Password string
}

func (s *Services) UpdateUserPassword(ctx context.Context, options UpdateUserPasswordOptions) (db.User, *ServiceError) {
	var password pgtype.Text
	err := password.Scan(options.Password)
	if err != nil {
		return db.User{}, NewValidationError("'password' is invalid")
	}

	var user db.User
	user, err = s.database.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		Password: password,
		ID:       options.ID,
	})
	if err != nil {
		return db.User{}, FromDBError(err)
	}

	return user, nil
}

func (s *Services) ConfirmUser(ctx context.Context, id int32) (db.User, *ServiceError) {
	user, err := s.database.ConfirmUser(ctx, id)
	if err != nil {
		return db.User{}, FromDBError(err)
	}

	return user, nil
}

func (s *Services) DeleteUser(ctx context.Context, id int32) *ServiceError {
	err := s.database.DeleteUserById(ctx, id)
	if err != nil {
		return FromDBError(err)
	}

	return nil
}
