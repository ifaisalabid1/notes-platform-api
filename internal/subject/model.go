package subject

import (
	"time"

	"github.com/google/uuid"
)

type Subject struct {
	ID          uuid.UUID `json:"id"`
	SemesterID  uuid.UUID `json:"semester_id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description"`
	SortOrder   int       `json:"sort_order"`
	IsPublished bool      `json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateSubjectInput struct {
	Title       string  `json:"title"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
	SortOrder   int     `json:"sort_order"`
	IsPublished bool    `json:"is_published"`
}

type UpdateSubjectInput struct {
	Title       string  `json:"title"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
	SortOrder   int     `json:"sort_order"`
	IsPublished bool    `json:"is_published"`
}
