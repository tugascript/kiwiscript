package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	cc "github.com/kiwiscript/kiwiscript_go/providers/cache"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/providers/email"
	"github.com/kiwiscript/kiwiscript_go/providers/tokens"
	"golang.org/x/crypto/argon2"
)

const memory uint32 = 65_536
const iterations uint32 = 3
const parallelism uint8 = 4
const saltSize uint32 = 16
const keySize uint32 = 32

func generateSalt() ([]byte, error) {
	salt := make([]byte, saltSize)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

func hashPassword(password string) (string, error) {
	salt, err := generateSalt()
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keySize)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	return b64Salt + "." + b64Hash, nil
}

func verifyPassword(password, hash string) bool {
	parts := strings.Split(hash, ".")

	if len(parts) != 2 {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	comparisonHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keySize)
	return bytes.Equal(decodedHash, comparisonHash)
}

type SignUpOptions struct {
	Email     string
	FirstName string
	LastName  string
	Location  string
	BirthDate string
	Password  string
}

func (s *Services) SignUp(ctx context.Context, options SignUpOptions) *ServiceError {
	log := s.log.WithGroup("services.auth.SignUp").With("email", options.Email)
	log.InfoContext(ctx, "Sign up")

	birthDate, err := time.Parse(time.DateOnly, options.BirthDate)
	if err != nil {
		log.WarnContext(ctx, "Invalid date format", "error", err)
		return NewValidationError("'birthdate' is invalid date format")
	}

	password, err := hashPassword(options.Password)
	if err != nil {
		errMsg := "Failed to hash password"
		log.ErrorContext(ctx, errMsg, "error", err)
		return NewServerError(errMsg)
	}

	user, serviceErr := s.CreateUser(ctx, CreateUserOptions{
		Email:     options.Email,
		FirstName: options.FirstName,
		LastName:  options.LastName,
		Location:  options.Location,
		BirthDate: birthDate,
		Password:  password,
		Provider:  ProviderEmail,
	})
	if serviceErr != nil {
		log.WarnContext(ctx, "Failed to create user", "error", serviceErr)
		return serviceErr
	}

	confirmationToken, err := s.jwt.CreateEmailToken(tokens.EmailTokenConfirmation, user)
	if err != nil {
		errMsg := "Failed to create confirmation token"
		log.ErrorContext(ctx, errMsg, "error", err)
		return NewServerError(errMsg)
	}

	go func() {
		opts := email.ConfirmationEmailOptions{
			Email:             user.Email,
			FirstName:         user.FirstName,
			LastName:          user.LastName,
			ConfirmationToken: confirmationToken,
		}
		if err := s.mail.SendConfirmationEmail(opts); err != nil {
			log.WarnContext(ctx, "Failed to send confirmation email", "error", err)
		}
	}()

	log.InfoContext(ctx, "Sign up successfully")
	return nil
}

func (s *Services) generateAuthResponse(ctx context.Context, log *slog.Logger, successMsg string, user db.User) (AuthResponse, *ServiceError) {
	var authResponse AuthResponse

	accessToken, err := s.jwt.CreateAccessToken(user)
	if err != nil {
		errMsg := "Failed to create access token"
		log.ErrorContext(ctx, errMsg, "error", err)
		return authResponse, NewServerError(errMsg)
	}

	refreshToken, err := s.jwt.CreateRefreshToken(user)
	if err != nil {
		errMsg := "Failed to create refresh token"
		log.ErrorContext(ctx, errMsg, "error", err)
		return authResponse, NewServerError(errMsg)
	}

	authResponse = AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.jwt.GetAccessTtl(),
	}
	log.InfoContext(ctx, successMsg)
	return authResponse, nil
}

type AuthResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

func (s *Services) ConfirmEmail(ctx context.Context, token string) (AuthResponse, *ServiceError) {
	log := s.log.WithGroup("services.auth.CofirmEmail").With("token", token)
	log.InfoContext(ctx, "Confirm email")
	var authResponse AuthResponse

	tokenType, claims, err := s.jwt.VerifyEmailToken(token)
	if err != nil {
		log.WarnContext(ctx, "Invalid token", "error", err)
		return authResponse, NewUnauthorizedError()
	}

	if tokenType != tokens.EmailTokenConfirmation {
		log.WarnContext(ctx, "Invalid token type")
		return authResponse, NewUnauthorizedError()
	}

	user, serviceErr := s.FindUserByID(ctx, claims.ID)
	if serviceErr != nil {
		log.WarnContext(ctx, "User not found", "error", serviceErr)
		return authResponse, NewUnauthorizedError()
	}
	if user.IsConfirmed {
		log.WarnContext(ctx, "User already confirmed")
		return authResponse, NewUnauthorizedError()
	}
	if user.Version != claims.Version {
		log.WarnContext(ctx, "Invalid token version")
		return authResponse, NewUnauthorizedError()
	}

	user, serviceErr = s.ConfirmUser(ctx, user.ID)
	if serviceErr != nil {
		log.ErrorContext(ctx, "Failed to confirm user", "error", serviceErr)
		return authResponse, serviceErr
	}

	return s.generateAuthResponse(ctx, log, "Confirmed email successfully", user)
}

func generateCode() (string, error) {
	const codeLength = 6
	const digits = "0123456789"
	code := make([]byte, codeLength)

	for i := 0; i < codeLength; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[num.Int64()]
	}

	return string(code), nil
}

func (s *Services) SaveTwoFactorCode(userID int32) (string, *ServiceError) {
	code, err := generateCode()
	if err != nil {
		return "", NewServerError("Failed to generate two factor code")
	}

	hashedCode, err := hashPassword(code)
	if err != nil {
		return "", NewServerError("Failed to hash two factor code")
	}

	if err := s.cache.AddTwoFactorCode(cc.AddTwoFactorCodeOptions{UserID: userID, Code: hashedCode}); err != nil {
		return "", NewServerError("Failed to save two factor code")
	}

	return code, nil
}

type SignInOptions struct {
	Email    string
	Password string
}

func (s *Services) SignIn(ctx context.Context, options SignInOptions) *ServiceError {
	log := s.log.WithGroup("services.auth.SignIn").With("email", options.Email)
	log.InfoContext(ctx, "Sign in")

	prms := db.FindAuthProviderByEmailAndProviderParams{
		Email:    options.Email,
		Provider: ProviderEmail,
	}
	if _, err := s.database.FindAuthProviderByEmailAndProvider(ctx, prms); err != nil {
		log.WarnContext(ctx, "Failed to find auth provider", "error", err)
		return NewUnauthorizedError()
	}

	user, serviceErr := s.FindUserByEmail(ctx, options.Email)
	if serviceErr != nil {
		log.WarnContext(ctx, "Failed to find user", "error", serviceErr)
		return NewUnauthorizedError()
	}

	if !user.IsConfirmed {
		errMsg := "User not confirmed"
		log.WarnContext(ctx, errMsg)
		return NewValidationError(errMsg)
	}
	if !verifyPassword(options.Password, user.Password.String) {
		log.WarnContext(ctx, "Invalid password")
		return NewUnauthorizedError()
	}

	code, serviceErr := s.SaveTwoFactorCode(user.ID)
	if serviceErr != nil {
		log.ErrorContext(ctx, "Failed to save two factor code", "error", serviceErr)
		return serviceErr
	}

	go func() {
		opts := email.CodeEmailOptions{
			Email: user.Email,
			Code:  code,
		}
		if err := s.mail.SendCodeEmail(opts); err != nil {
			log.ErrorContext(ctx, "Failed to send two factor email", "error", err)
		}
	}()

	log.InfoContext(ctx, "Sign in successful")
	return nil
}

type TwoFactorOptions struct {
	Email string
	Code  string
}

func (s *Services) TwoFactor(ctx context.Context, options TwoFactorOptions) (AuthResponse, *ServiceError) {
	log := s.log.WithGroup("services.auth.TwoFactor").With("email", options.Email)
	log.InfoContext(ctx, "Two factor confirmation")
	var authResponse AuthResponse

	user, serviceErr := s.FindUserByEmail(ctx, options.Email)
	if serviceErr != nil {
		log.WarnContext(ctx, "Failed to find user", "error", serviceErr)
		return authResponse, NewUnauthorizedError()
	}

	hashedCode, err := s.cache.GetTwoFactorCode(user.ID)
	if err != nil {
		log.WarnContext(ctx, "Failed to get two factor code", "error", err)
		return authResponse, NewUnauthorizedError()
	}
	if !verifyPassword(options.Code, hashedCode) {
		log.WarnContext(ctx, "Invalid two factor code")
		return authResponse, NewUnauthorizedError()
	}
	if err := s.cache.DeleteTwoFactorCode(user.ID); err != nil {
		errMsg := "Failed to delete two factor code"
		log.WarnContext(ctx, errMsg, "error", err)
		return authResponse, NewServerError(errMsg)
	}

	return s.generateAuthResponse(ctx, log, "Confirmed two factor successfully", user)
}

func (s *Services) Refresh(ctx context.Context, token string) (AuthResponse, *ServiceError) {
	log := s.log.WithGroup("services.auth.RefreshToken").With("token", token)
	log.InfoContext(ctx, "Refresh token")
	var authResponse AuthResponse

	claims, id, _, err := s.jwt.VerifyRefreshToken(token)
	if err != nil {
		log.WarnContext(ctx, "Invalid token", "error", err)
		return authResponse, NewUnauthorizedError()
	}
	if s.cache.IsBlackListed(ctx, id) {
		log.WarnContext(ctx, "Token black listed")
		return authResponse, NewUnauthorizedError()
	}

	user, serviceErr := s.FindUserByID(ctx, claims.ID)
	if serviceErr != nil {
		log.WarnContext(ctx, "Failed to find user", "error", err)
		return authResponse, NewUnauthorizedError()
	}
	if user.Version != claims.Version {
		log.WarnContext(ctx, "Invalid token version")
		return authResponse, NewUnauthorizedError()
	}

	return s.generateAuthResponse(ctx, log, "Refreshed token successfully", user)
}

func (s *Services) SignOut(ctx context.Context, token string) *ServiceError {
	log := s.log.WithGroup("services.auth.SignOut")
	log.InfoContext(ctx, "Sign out")

	_, id, exp, err := s.jwt.VerifyRefreshToken(token)
	if err != nil {
		log.WarnContext(ctx, "Invalid token", "error", err)
		return NewUnauthorizedError()
	}

	err = s.cache.AddBlackList(cc.AddBlackListOptions{
		ID:    id,
		Token: token,
		Exp:   exp,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to add black list", "error", err)
		return NewServerError("Failed to add black list")
	}

	log.InfoContext(ctx, "Signed out successfully")
	return nil
}

type UpdatePasswordOptions struct {
	UserID      int32
	UserVersion int16
	OldPassword string
	NewPassword string
}

func (s *Services) UpdatePassword(ctx context.Context, options UpdatePasswordOptions) (AuthResponse, *ServiceError) {
	log := s.log.WithGroup("services.auth.UpdatePassword").With("userID", options.UserID)
	log.InfoContext(ctx, "Update password")
	var authResponse AuthResponse

	user, serviceErr := s.FindUserByID(ctx, options.UserID)
	if serviceErr != nil {
		log.WarnContext(ctx, "Failed to find user", "error", serviceErr)
		return authResponse, NewUnauthorizedError()
	}
	if user.Version != options.UserVersion {
		log.WarnContext(ctx, "Invalid user version")
		return authResponse, NewUnauthorizedError()
	}

	_, err := s.database.FindAuthProviderByEmailAndProvider(ctx, db.FindAuthProviderByEmailAndProviderParams{
		Email:    user.Email,
		Provider: ProviderEmail,
	})
	if err == nil {
		if !verifyPassword(options.OldPassword, user.Password.String) {
			log.WarnContext(ctx, "Invalid old password")
			return authResponse, NewUnauthorizedError()
		}

		password, err := hashPassword(options.NewPassword)
		if err != nil {
			log.ErrorContext(ctx, "Failed to hash new password", "error", err)
			return authResponse, NewServerError("Failed to hash new password")
		}

		user, serviceErr = s.UpdateUserPassword(ctx, UpdateUserPasswordOptions{
			ID:       options.UserID,
			Password: password,
		})
		if serviceErr != nil {
			log.ErrorContext(ctx, "Failed to update password", "error", serviceErr)
			return authResponse, serviceErr
		}

		return s.generateAuthResponse(ctx, log, "Updated password successfully", user)
	}

	serviceErr = FromDBError(err)
	if serviceErr.Code != CodeNotFound {
		log.WarnContext(ctx, "Failed to find auth provider", "error", err)
		return authResponse, serviceErr
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return authResponse, FromDBError(err)
	}

	defer func() {
		if p := recover(); p != nil {
			txn.Rollback(ctx)
			panic(p)
		}
		if err != nil || serviceErr != nil {
			txn.Rollback(ctx)
			return
		}
		if commitErr := txn.Commit(ctx); commitErr != nil {
			panic(commitErr)
		}
	}()

	err = qrs.CreateAuthProvider(ctx, db.CreateAuthProviderParams{
		Email:    user.Email,
		Provider: ProviderEmail,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to create auth provider", "error", err)
		return authResponse, FromDBError(err)
	}

	passwordHash, err := hashPassword(options.NewPassword)
	if err != nil {
		log.ErrorContext(ctx, "Failed to hash new password", "error", err)
		return authResponse, NewServerError("Failed to hash new password")
	}

	var password pgtype.Text
	if err = password.Scan(passwordHash); err != nil {
		log.ErrorContext(ctx, "Failed to scan password", "error", err)
		return authResponse, NewServerError("Failed to scan password")
	}

	user, err = qrs.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:       options.UserID,
		Password: password,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update password", "error", err)
		return authResponse, FromDBError(err)
	}

	return s.generateAuthResponse(ctx, log, "Updated password successfully", user)
}

func (s *Services) ForgotPassword(ctx context.Context, userEmail string) *ServiceError {
	log := s.log.WithGroup("services.auth.ResetPassword").With("email", userEmail)
	log.InfoContext(ctx, "Reset password")

	user, serviceErr := s.FindUserByEmail(ctx, userEmail)
	if serviceErr != nil {
		if serviceErr.Code == CodeNotFound {
			log.InfoContext(ctx, "User not found, skip reset password", "email", userEmail)
			return nil
		}

		log.WarnContext(ctx, "Failed to find user", "error", serviceErr)
		return serviceErr
	}

	emailToken, err := s.jwt.CreateEmailToken(tokens.EmailTokenReset, user)
	if err != nil {
		errMsg := "Failed to create email token"
		log.ErrorContext(ctx, errMsg, "error", err)
		return NewServerError(errMsg)
	}

	go func() {
		opts := email.ResetEmailOptions{
			Email:      user.Email,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			ResetToken: emailToken,
		}
		if err := s.mail.SendResetEmail(opts); err != nil {
			log.ErrorContext(ctx, "Failed to send two factor email", "error", err)
		}
	}()

	log.InfoContext(ctx, "Reset password successfully")
	return nil
}

type ResetPasswordOptions struct {
	ResetToken  string
	NewPassword string
}

func (s *Services) ResetPassword(ctx context.Context, options ResetPasswordOptions) *ServiceError {
	log := s.log.WithGroup("services.auth.ResetPassword")
	log.InfoContext(ctx, "Reset password")

	tokenType, claims, err := s.jwt.VerifyEmailToken(options.ResetToken)
	if err != nil {
		log.WarnContext(ctx, "Invalid token", "error", err)
		return NewUnauthorizedError()
	}

	if tokenType != tokens.EmailTokenReset {
		log.WarnContext(ctx, "Invalid token type")
		return NewUnauthorizedError()
	}

	user, serviceErr := s.FindUserByID(ctx, claims.ID)
	if serviceErr != nil {
		log.WarnContext(ctx, "Failed to find user", "error", serviceErr)
		return NewUnauthorizedError()
	}
	if user.Version != claims.Version {
		log.WarnContext(ctx, "Invalid token version")
		return NewUnauthorizedError()
	}

	passwordHash, err := hashPassword(options.NewPassword)
	if err != nil {
		log.ErrorContext(ctx, "Failed to hash new password", "error", err)
		return NewServerError("Failed to hash new password")
	}

	_, serviceErr = s.UpdateUserPassword(ctx, UpdateUserPasswordOptions{
		ID:       user.ID,
		Password: passwordHash,
	})
	if serviceErr != nil {
		log.ErrorContext(ctx, "Failed to update password", "error", serviceErr)
		return serviceErr
	}

	log.InfoContext(ctx, "Reset password successful")
	return nil
}

type UpdateEmailOptions struct {
	UserID      int32
	UserVersion int16
	NewEmail    string
	Password    string
}

func (s *Services) UpdateEmail(ctx context.Context, options UpdateEmailOptions) (AuthResponse, *ServiceError) {
	log := s.log.WithGroup("services.auth.UpdateEmail")
	log.InfoContext(ctx, "Update email", "userID", options.UserID)
	var authResponse AuthResponse

	user, serviceErr := s.FindUserByID(ctx, options.UserID)
	if serviceErr != nil {
		log.WarnContext(ctx, "Failed to find user", "error", serviceErr)
		return authResponse, NewUnauthorizedError()
	}
	if user.Version != options.UserVersion {
		log.WarnContext(ctx, "Invalid user version")
		return authResponse, NewUnauthorizedError()
	}

	if _, err := s.database.FindAuthProviderByEmailAndProvider(ctx, db.FindAuthProviderByEmailAndProviderParams{
		Email:    user.Email,
		Provider: ProviderEmail,
	}); err != nil {
		serviceErr = FromDBError(err)

		if serviceErr.Code == CodeNotFound {
			log.WarnContext(ctx, "Email auth provider not found", "error", err)
			return authResponse, NewUnauthorizedError()
		}

		log.ErrorContext(ctx, "Failed to find auth provider", "error", err)
		return authResponse, serviceErr
	}
	if !verifyPassword(options.Password, user.Password.String) {
		log.WarnContext(ctx, "Invalid password")
		return authResponse, NewUnauthorizedError()
	}

	if _, err := s.database.FindAuthProviderByEmailAndProvider(ctx, db.FindAuthProviderByEmailAndProviderParams{
		Email:    options.NewEmail,
		Provider: ProviderEmail,
	}); err == nil {
		log.WarnContext(ctx, "Email already exists")
		return authResponse, NewValidationError("Email already in use")
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return authResponse, FromDBError(err)
	}

	defer func() {
		if p := recover(); p != nil {
			txn.Rollback(ctx)
			panic(p)
		}
		if err != nil || serviceErr != nil {
			txn.Rollback(ctx)
			return
		}
		if commitErr := txn.Commit(ctx); commitErr != nil {
			panic(commitErr)
		}
	}()

	user, err = qrs.UpdateUserEmail(ctx, db.UpdateUserEmailParams{
		ID:    user.ID,
		Email: options.NewEmail,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update email", "error", err)
		return authResponse, FromDBError(err)
	}
	if err = qrs.DeleteProviderByEmailAndNotProvider(ctx, db.DeleteProviderByEmailAndNotProviderParams{
		Email:    user.Email,
		Provider: ProviderEmail,
	}); err != nil {
		log.ErrorContext(ctx, "Failed to delete auth provider", "error", err)
		return authResponse, FromDBError(err)
	}

	return s.generateAuthResponse(ctx, log, "Update email successfully", user)
}

func (s *Services) ProcessAuthHeader(authHeader string) (tokens.AccessUserClaims, *ServiceError) {
	authHeaderSlice := strings.Split(authHeader, " ")
	var userClaims tokens.AccessUserClaims

	if len(authHeaderSlice) != 2 {
		return userClaims, NewUnauthorizedError()
	}
	if strings.ToLower(authHeaderSlice[0]) != "bearer" {
		return userClaims, NewUnauthorizedError()
	}

	userClaims, err := s.jwt.VerifyAccessToken(authHeaderSlice[1])
	if err != nil {
		return userClaims, NewUnauthorizedError()
	}
	return userClaims, nil
}
