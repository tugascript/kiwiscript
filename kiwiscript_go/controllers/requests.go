package controllers

type SignUpRequest struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name" validate:"required,min=2,max=50"`
	LastName  string `json:"last_name" validate:"required,min=2,max=50"`
	Location  string `json:"location" validate:"required,min=3,max=3"`
	BirthDate string `json:"birth_date" validate:"required"`
	Password1 string `json:"password" validate:"required,min=8,max=50"`
	Password2 string `json:"password2" validate:"required,eqfield=Password1"`
}

type SignInRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=1"`
}

type ConfirmSignInRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,min=1"`
}

type SignOutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,jwt"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,jwt"`
}

type ConfirmRequest struct {
	ConfirmationToken string `json:"confirmation_token" validate:"required,jwt"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	ResetToken string `json:"reset_token" validate:"required,jwt"`
	Password1  string `json:"password" validate:"required,min=8,max=50"`
	Password2  string `json:"password2" validate:"required,eqfield=Password1"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required,min=1"`
	Password1   string `json:"password" validate:"required,min=8,max=50"`
	Password2   string `json:"password2" validate:"required,eqfield=Password1"`
}

type UpdateEmailRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=1"`
}

type CreateLanguageRequest struct {
	Name string `json:"name" validate:"required,min=2,max=50"`
	Icon string `json:"icon" validate:"required"`
}
