package validation

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrTitleRequired = errors.New("title is required")
	ErrSlugRequired  = errors.New("slug is required")
	ErrInvalidSlug   = errors.New("slug may only contain lowercase letters, numbers, and hyphens")
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type TitleSlugInput struct {
	Title string
	Slug  string
}

func NormalizeTitleSlug(input TitleSlugInput) TitleSlugInput {
	return TitleSlugInput{
		Title: strings.TrimSpace(input.Title),
		Slug:  strings.TrimSpace(input.Slug),
	}
}

func ValidateTitleSlug(input TitleSlugInput) error {
	input = NormalizeTitleSlug(input)

	if input.Title == "" {
		return ErrTitleRequired
	}

	if input.Slug == "" {
		return ErrSlugRequired
	}

	if !slugPattern.MatchString(input.Slug) {
		return ErrInvalidSlug
	}

	return nil
}
