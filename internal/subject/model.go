package subject

import (
	"time"

	"github.com/google/uuid"
)

type Subject struct {
	ID          uuid.UUID  `json:"id"`
	SemesterID  uuid.UUID  `json:"semester_id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Description *string    `json:"description"`
	SortOrder   int        `json:"sort_order"`
	IsPublished bool       `json:"is_published"`
	CreatedBy   *uuid.UUID `json:"created_by"`
	UpdatedBy   *uuid.UUID `json:"updated_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type AdminSubject struct {
	ID            uuid.UUID  `json:"id"`
	SemesterID    uuid.UUID  `json:"semester_id"`
	Title         string     `json:"title"`
	Slug          string     `json:"slug"`
	Description   *string    `json:"description"`
	SortOrder     int        `json:"sort_order"`
	IsPublished   bool       `json:"is_published"`
	CreatedBy     *uuid.UUID `json:"created_by"`
	CreatedByName *string    `json:"created_by_name"`
	UpdatedBy     *uuid.UUID `json:"updated_by"`
	UpdatedByName *string    `json:"updated_by_name"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
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
