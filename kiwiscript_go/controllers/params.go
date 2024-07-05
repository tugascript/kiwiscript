package controllers

type GetLanguageParams struct {
	Name string `validate:"required,min=2,max=50,alphanum"`
}
