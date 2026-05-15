package unit

import (
	"time"

	"github.com/google/uuid"
)

type Unit struct {
	ID          uuid.UUID `json:"id"`
	SubjectID   uuid.UUID `json:"subject_id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description"`
	SortOrder   int       `json:"sort_order"`
	IsPublished bool      `json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateUnitInput struct {
	Title       string  `json:"title"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
	SortOrder   int     `json:"sort_order"`
	IsPublished bool    `json:"is_published"`
}

type UpdateUnitInput struct {
	Title       string  `json:"title"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
	SortOrder   int     `json:"sort_order"`
	IsPublished bool    `json:"is_published"`
}
