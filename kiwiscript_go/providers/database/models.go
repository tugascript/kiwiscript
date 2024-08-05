// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package db

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuthProvider struct {
	ID        int32
	Email     string
	Provider  string
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
}

type Certificate struct {
	ID           uuid.UUID
	UserID       int32
	LanguageSlug string
	SeriesTitle  string
	SeriesSlug   string
	CompletedAt  pgtype.Timestamp
	CreatedAt    pgtype.Timestamp
	UpdatedAt    pgtype.Timestamp
}

type Donation struct {
	ID           int32
	UserID       int32
	Amount       int64
	Currency     string
	Recurring    bool
	RecurringRef pgtype.Text
	CreatedAt    pgtype.Timestamp
	UpdatedAt    pgtype.Timestamp
}

type Language struct {
	ID          int32
	Name        string
	Slug        string
	Icon        string
	SeriesCount int16
	AuthorID    int32
	CreatedAt   pgtype.Timestamp
	UpdatedAt   pgtype.Timestamp
}

type LanguageProgress struct {
	ID              int32
	UserID          int32
	LanguageSlug    string
	CompletedSeries int16
	ViewedAt        pgtype.Timestamp
	CreatedAt       pgtype.Timestamp
	UpdatedAt       pgtype.Timestamp
}

type Lesson struct {
	ID               int32
	Title            string
	Position         int16
	IsPublished      bool
	WatchTimeSeconds int32
	ReadTimeSeconds  int32
	AuthorID         int32
	LanguageSlug     string
	SeriesSlug       string
	SectionID        int32
	CreatedAt        pgtype.Timestamp
	UpdatedAt        pgtype.Timestamp
}

type LessonArticle struct {
	ID              int32
	LessonID        int32
	AuthorID        int32
	Content         string
	ReadTimeSeconds int32
	CreatedAt       pgtype.Timestamp
	UpdatedAt       pgtype.Timestamp
}

type LessonFile struct {
	ID        uuid.UUID
	LessonID  int32
	AuthorID  int32
	Ext       string
	Name      string
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
}

type LessonProgress struct {
	ID                 int32
	UserID             int32
	LanguageSlug       string
	SeriesSlug         string
	SectionID          int32
	LessonID           int32
	LanguageProgressID int32
	SeriesProgressID   int32
	SectionProgressID  int32
	CompletedAt        pgtype.Timestamp
	ViewedAt           pgtype.Timestamp
	CreatedAt          pgtype.Timestamp
	UpdatedAt          pgtype.Timestamp
}

type LessonVideo struct {
	ID               int32
	LessonID         int32
	AuthorID         int32
	Url              string
	WatchTimeSeconds int32
	CreatedAt        pgtype.Timestamp
	UpdatedAt        pgtype.Timestamp
}

type Payment struct {
	ID         int32
	PaymentRef string
	UserID     int32
	DonationID int32
	Amount     int64
	Currency   string
	CreatedAt  pgtype.Timestamp
	UpdatedAt  pgtype.Timestamp
}

type Section struct {
	ID               int32
	Title            string
	LanguageSlug     string
	SeriesSlug       string
	Description      string
	Position         int16
	LessonsCount     int16
	WatchTimeSeconds int32
	ReadTimeSeconds  int32
	IsPublished      bool
	AuthorID         int32
	CreatedAt        pgtype.Timestamp
	UpdatedAt        pgtype.Timestamp
}

type SectionProgress struct {
	ID                 int32
	UserID             int32
	LanguageSlug       string
	SeriesSlug         string
	SectionID          int32
	LanguageProgressID int32
	SeriesProgressID   int32
	CompletedLessons   int16
	CompletedAt        pgtype.Timestamp
	ViewedAt           pgtype.Timestamp
	CreatedAt          pgtype.Timestamp
	UpdatedAt          pgtype.Timestamp
}

type Series struct {
	ID               int32
	Title            string
	Slug             string
	Description      string
	SectionsCount    int16
	LessonsCount     int16
	WatchTimeSeconds int32
	ReadTimeSeconds  int32
	IsPublished      bool
	LanguageSlug     string
	AuthorID         int32
	CreatedAt        pgtype.Timestamp
	UpdatedAt        pgtype.Timestamp
}

type SeriesImage struct {
	ID        int32
	SeriesID  int32
	AuthorID  int32
	File      uuid.UUID
	Ext       string
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
}

type SeriesProgress struct {
	ID                 int32
	UserID             int32
	SeriesSlug         string
	LanguageSlug       string
	LanguageProgressID int32
	CompletedSections  int16
	CompletedLessons   int16
	PartsCount         int16
	CompletedAt        pgtype.Timestamp
	ViewedAt           pgtype.Timestamp
	CreatedAt          pgtype.Timestamp
	UpdatedAt          pgtype.Timestamp
}

type User struct {
	ID          int32
	FirstName   string
	LastName    string
	Location    string
	Email       string
	BirthDate   pgtype.Date
	Version     int16
	IsAdmin     bool
	IsStaff     bool
	IsConfirmed bool
	Password    pgtype.Text
	CreatedAt   pgtype.Timestamp
	UpdatedAt   pgtype.Timestamp
}
