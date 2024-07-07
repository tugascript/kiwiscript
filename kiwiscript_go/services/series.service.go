package services

type CreateSeriesOptions struct {
	Title       string
	Description string
	Tags        []string
}

func (s *Services) CreateSeries() {}
