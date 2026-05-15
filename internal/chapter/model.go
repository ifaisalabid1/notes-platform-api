package chapter

import (
	"time"

	"github.com/google/uuid"
)

type Chapter struct {
	ID          uuid.UUID `json:"id"`
	UnitID      uuid.UUID `json:"unit_id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description"`
	SortOrder   int       `json:"sort_order"`
	IsPublished bool      `json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateChapterInput struct {
	Title       string  `json:"title"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
	SortOrder   int     `json:"sort_order"`
	IsPublished bool    `json:"is_published"`
}

type UpdateChapterInput struct {
	Title       string  `json:"title"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
	SortOrder   int     `json:"sort_order"`
	IsPublished bool    `json:"is_published"`
}
