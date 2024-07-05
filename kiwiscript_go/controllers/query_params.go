package controllers

const (
	OffsetDefault int = 0
	LimitDefault  int = 25
)

type GetLanguagesQueryParams struct {
	Offset int32  `validate:"gte=1"`
	Limit  int32  `validate:"gte=1,lte=100"`
	Search string `validate:"min=1,max=50"`
}
