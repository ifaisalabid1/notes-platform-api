package note

import (
	"mime/multipart"
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

type PublicNote struct {
	ID               uuid.UUID `json:"id"`
	ChapterID        uuid.UUID `json:"chapter_id"`
	Title            string    `json:"title"`
	Slug             string    `json:"slug"`
	Description      *string   `json:"description"`
	OriginalFileName string    `json:"original_file_name"`
	FileURL          string    `json:"file_url"`
	IsPublished      bool      `json:"is_published"`
	SortOrder        int       `json:"sort_order"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type FileMetadata struct {
	NoteID           uuid.UUID `json:"note_id"`
	Title            string    `json:"title"`
	OriginalFileName string    `json:"original_file_name"`
	StoredObjectKey  string    `json:"stored_object_key"`
	FileContentType  string    `json:"file_content_type"`
	FileSizeBytes    int64     `json:"file_size_bytes"`
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

type UploadNoteInput struct {
	Title       string
	Slug        string
	Description *string
	SortOrder   int
	IsPublished bool
	File        multipart.File
	FileHeader  *multipart.FileHeader
}

type UpdateNoteInput struct {
	Title       string  `json:"title"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
	IsPublished bool    `json:"is_published"`
	SortOrder   int     `json:"sort_order"`
}
