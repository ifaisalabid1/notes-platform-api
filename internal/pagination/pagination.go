package pagination

import (
	"math"
	"net/http"
	"strconv"
	"strings"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type Params struct {
	Page     int
	PageSize int
	Search   string
}

type Metadata struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

func FromRequest(r *http.Request) Params {
	query := r.URL.Query()

	page := parsePositiveInt(query.Get("page"), DefaultPage)
	pageSize := parsePositiveInt(query.Get("page_size"), DefaultPageSize)

	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	search := strings.TrimSpace(query.Get("q"))

	return Params{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}
}

func (p Params) Limit() int {
	if p.PageSize <= 0 {
		return DefaultPageSize
	}

	if p.PageSize > MaxPageSize {
		return MaxPageSize
	}

	return p.PageSize
}

func (p Params) Offset() int {
	page := p.Page
	if page <= 0 {
		page = DefaultPage
	}

	return (page - 1) * p.Limit()
}

func NewMetadata(params Params, totalItems int) Metadata {
	page := params.Page
	if page <= 0 {
		page = DefaultPage
	}

	pageSize := params.Limit()

	totalPages := 0
	if totalItems > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(pageSize)))
	}

	return Metadata{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

func parsePositiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}
