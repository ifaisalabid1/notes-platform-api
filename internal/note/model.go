package note

import (
	"time"

	"github.com/google/uuid"
)

type Note struct {
	ID               uuid.UUID  `json:"id"`
	ChapterID        uuid.UUID  `json:"chapter_id"`
	Title            string     `json:"title"`
	Slug             string     `json:"slug"`
	Description      *string    `json:"description"`
	OriginalFileName string     `json:"original_file_name"`
	StoredObjectKey  string     `json:"stored_object_key"`
	FileContentType  string     `json:"file_content_type"`
	FileSizeBytes    int64      `json:"file_size_bytes"`
	IsWatermarked    bool       `json:"is_watermarked"`
	IsPublished      bool       `json:"is_published"`
	SortOrder        int        `json:"sort_order"`
	UploadedBy       *uuid.UUID `json:"uploaded_by"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type CreateNoteInput struct {
	Title            string  `json:"title"`
	Slug             string  `json:"slug"`
	Description      *string `json:"description"`
	OriginalFileName string  `json:"original_file_name"`
	StoredObjectKey  string  `json:"stored_object_key"`
	FileContentType  string  `json:"file_content_type"`
	FileSizeBytes    int64   `json:"file_size_bytes"`
	IsWatermarked    bool    `json:"is_watermarked"`
	IsPublished      bool    `json:"is_published"`
	SortOrder        int     `json:"sort_order"`
}

type UpdateNoteInput struct {
	Title       string  `json:"title"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
	IsPublished bool    `json:"is_published"`
	SortOrder   int     `json:"sort_order"`
}
