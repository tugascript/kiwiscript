package controllers

import (
	"fmt"

	"github.com/google/uuid"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
)

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func NewAuthResponse(accessToken, refreshToken string, expiresIn int64) AuthResponse {
	return AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
	}
}

type MessageResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

func NewMessageResponse(message string) MessageResponse {
	return MessageResponse{
		ID:      uuid.NewString(),
		Message: message,
	}
}

type PaginatedResponse[T any] struct {
	Count    int64  `json:"count"`
	Next     string `json:"next,omitempty"`
	Previous string `json:"previous,omitempty"`
	Results  []T    `json:"results"`
}

func newPaginatedNavigationURL(frontendDomain string, limit, offset int32) string {
	if offset < 0 {
		offset = 0
	}

	return fmt.Sprintf("https://%s?limit=%d&offset=%d", frontendDomain, limit, offset)
}

func NewPaginatedResponse[T any, V any](
	frontendDomain string,
	limit,
	offset int32,
	count int64,
	entites []V,
	mapper func(V) T,
) PaginatedResponse[T] {
	results := make([]T, len(entites))

	for i, entity := range entites {
		results[i] = mapper(entity)
	}

	var next, prev string
	nextOffset := offset + limit
	prevOffset := offset - limit
	if int64(nextOffset) < count {
		next = newPaginatedNavigationURL(frontendDomain, limit, nextOffset)
	}
	if prevOffset > 0 {
		prev = newPaginatedNavigationURL(frontendDomain, limit, prevOffset)
	}

	return PaginatedResponse[T]{
		Count:    count,
		Next:     next,
		Previous: prev,
		Results:  results,
	}
}

type LanguageResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

func NewLanguageResponse(language db.Language) LanguageResponse {
	return LanguageResponse{
		ID:   language.ID,
		Name: language.Name,
		Icon: language.Icon,
	}
}
