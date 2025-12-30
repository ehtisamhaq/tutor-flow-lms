package search

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// UseCase defines search business logic
type UseCase struct {
	searchRepo   repository.SearchRepository
	courseRepo   repository.CourseRepository
	categoryRepo repository.CategoryRepository
}

// NewUseCase creates a new search use case
func NewUseCase(
	searchRepo repository.SearchRepository,
	courseRepo repository.CourseRepository,
	categoryRepo repository.CategoryRepository,
) *UseCase {
	return &UseCase{
		searchRepo:   searchRepo,
		courseRepo:   courseRepo,
		categoryRepo: categoryRepo,
	}
}

// SearchInput defines search parameters
type SearchInput struct {
	Query      string     `json:"query"`
	CategoryID *uuid.UUID `json:"category_id,omitempty"`
	Level      *string    `json:"level,omitempty"`
	MinPrice   *float64   `json:"min_price,omitempty"`
	MaxPrice   *float64   `json:"max_price,omitempty"`
	MinRating  *float64   `json:"min_rating,omitempty"`
	IsFree     *bool      `json:"is_free,omitempty"`
	SortBy     string     `json:"sort_by,omitempty"`    // relevance, rating, price, students, newest
	SortOrder  string     `json:"sort_order,omitempty"` // asc, desc
	Page       int        `json:"page,omitempty"`
	Limit      int        `json:"limit,omitempty"`
}

// SearchResult contains search results with metadata
type SearchResult struct {
	Courses    []CourseSearchResult `json:"courses"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	TotalPages int                  `json:"total_pages"`
	Facets     *SearchFacets        `json:"facets,omitempty"`
}

// CourseSearchResult is a course with search ranking
type CourseSearchResult struct {
	domain.Course
	Rank float64 `json:"rank,omitempty"`
}

// SearchFacets for filter UI
type SearchFacets struct {
	Categories  []FacetItem `json:"categories"`
	Levels      []FacetItem `json:"levels"`
	PriceRanges []FacetItem `json:"price_ranges"`
}

type FacetItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Count int64  `json:"count"`
}

// Search performs full-text search with filters
func (uc *UseCase) Search(ctx context.Context, input SearchInput) (*SearchResult, error) {
	// Set defaults
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Limit < 1 || input.Limit > 50 {
		input.Limit = 20
	}
	if input.SortBy == "" {
		if input.Query != "" {
			input.SortBy = "relevance"
		} else {
			input.SortBy = "newest"
		}
	}

	// Build search filters
	filters := repository.SearchFilters{
		Query:      input.Query,
		CategoryID: input.CategoryID,
		Level:      input.Level,
		MinPrice:   input.MinPrice,
		MaxPrice:   input.MaxPrice,
		MinRating:  input.MinRating,
		IsFree:     input.IsFree,
		SortBy:     input.SortBy,
		SortOrder:  input.SortOrder,
		Page:       input.Page,
		Limit:      input.Limit,
	}

	courses, total, err := uc.searchRepo.SearchCourses(ctx, filters)
	if err != nil {
		return nil, err
	}

	// Calculate pages
	totalPages := int(total) / input.Limit
	if int(total)%input.Limit > 0 {
		totalPages++
	}

	// Convert to search results
	results := make([]CourseSearchResult, len(courses))
	for i, c := range courses {
		results[i] = CourseSearchResult{Course: c}
	}

	return &SearchResult{
		Courses:    results,
		Total:      total,
		Page:       input.Page,
		TotalPages: totalPages,
	}, nil
}

// GetSearchFacets returns facets for filter UI
func (uc *UseCase) GetSearchFacets(ctx context.Context, query string) (*SearchFacets, error) {
	facets, err := uc.searchRepo.GetFacets(ctx, query)
	if err != nil {
		return nil, err
	}

	// Convert category facets
	categoryFacets := make([]FacetItem, len(facets.Categories))
	for i, f := range facets.Categories {
		categoryFacets[i] = FacetItem{Label: f.Label, Value: f.Value, Count: f.Count}
	}

	// Convert level facets
	levelFacets := make([]FacetItem, len(facets.Levels))
	for i, f := range facets.Levels {
		levelFacets[i] = FacetItem{Label: f.Label, Value: f.Value, Count: f.Count}
	}

	// Convert price range facets
	priceFacets := make([]FacetItem, len(facets.PriceRanges))
	for i, f := range facets.PriceRanges {
		priceFacets[i] = FacetItem{Label: f.Label, Value: f.Value, Count: f.Count}
	}

	return &SearchFacets{
		Categories:  categoryFacets,
		Levels:      levelFacets,
		PriceRanges: priceFacets,
	}, nil
}

// GetSuggestions returns autocomplete suggestions
func (uc *UseCase) GetSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	if len(query) < 2 {
		return []string{}, nil
	}
	if limit < 1 || limit > 10 {
		limit = 5
	}
	return uc.searchRepo.GetSuggestions(ctx, query, limit)
}

// GetTrendingSearches returns popular search terms
func (uc *UseCase) GetTrendingSearches(ctx context.Context, limit int) ([]string, error) {
	if limit < 1 || limit > 20 {
		limit = 10
	}
	return uc.searchRepo.GetTrendingSearches(ctx, limit)
}

// GetRelatedCourses returns courses similar to a given course
func (uc *UseCase) GetRelatedCourses(ctx context.Context, courseID uuid.UUID, limit int) ([]domain.Course, error) {
	if limit < 1 || limit > 10 {
		limit = 4
	}

	course, err := uc.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	// Build search from course keywords
	searchTerms := []string{course.Title}
	if course.ShortDescription != nil {
		searchTerms = append(searchTerms, *course.ShortDescription)
	}

	query := strings.Join(searchTerms, " ")

	filters := repository.SearchFilters{
		Query:     query,
		ExcludeID: &courseID,
		Limit:     limit,
		Page:      1,
	}

	courses, _, err := uc.searchRepo.SearchCourses(ctx, filters)
	return courses, err
}

// RecordSearch logs a search for analytics
func (uc *UseCase) RecordSearch(ctx context.Context, query string, userID *uuid.UUID, resultCount int64) error {
	if query == "" {
		return nil
	}
	return uc.searchRepo.RecordSearch(ctx, query, userID, resultCount)
}

// FormatSearchQuery normalizes query for FTS
func FormatSearchQuery(query string) string {
	// Clean and format for PostgreSQL tsquery
	words := strings.Fields(strings.TrimSpace(query))
	if len(words) == 0 {
		return ""
	}

	// Join with & for AND search
	return strings.Join(words, " & ")
}

// FormatWebSearchQuery formats for web search (prefix matching)
func FormatWebSearchQuery(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}

	words := strings.Fields(query)
	formatted := make([]string, len(words))
	for i, word := range words {
		formatted[i] = fmt.Sprintf("%s:*", word)
	}
	return strings.Join(formatted, " & ")
}
