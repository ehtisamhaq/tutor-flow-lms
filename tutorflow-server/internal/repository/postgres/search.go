package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// SearchRepository
type searchRepository struct {
	db *gorm.DB
}

func NewSearchRepository(db *gorm.DB) repository.SearchRepository {
	return &searchRepository{db: db}
}

func (r *searchRepository) SearchCourses(ctx context.Context, filters repository.SearchFilters) ([]domain.Course, int64, error) {
	var courses []domain.Course
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Course{}).
		Where("status = ?", domain.CourseStatusPublished)

	// Full-text search using PostgreSQL
	if filters.Query != "" {
		searchQuery := formatSearchQuery(filters.Query)
		query = query.Where(
			"to_tsvector('english', title || ' ' || COALESCE(short_description, '') || ' ' || COALESCE(description, '')) @@ to_tsquery('english', ?)",
			searchQuery,
		)
	}

	// Apply filters
	if filters.CategoryID != nil {
		query = query.Joins("JOIN course_categories cc ON cc.course_id = courses.id").
			Where("cc.category_id = ?", *filters.CategoryID)
	}
	if filters.Level != nil {
		query = query.Where("level = ?", *filters.Level)
	}
	if filters.MinPrice != nil {
		query = query.Where("price >= ?", *filters.MinPrice)
	}
	if filters.MaxPrice != nil {
		query = query.Where("price <= ?", *filters.MaxPrice)
	}
	if filters.MinRating != nil {
		query = query.Where("rating >= ?", *filters.MinRating)
	}
	if filters.IsFree != nil && *filters.IsFree {
		query = query.Where("is_free = true OR price = 0")
	}
	if filters.ExcludeID != nil {
		query = query.Where("id != ?", *filters.ExcludeID)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderClause := r.getOrderClause(filters.SortBy, filters.SortOrder, filters.Query)

	// Apply pagination
	offset := (filters.Page - 1) * filters.Limit
	err := query.
		Preload("Instructor").
		Preload("Categories").
		Order(orderClause).
		Offset(offset).
		Limit(filters.Limit).
		Find(&courses).Error

	if err != nil {
		return nil, 0, err
	}

	return courses, total, nil
}

func (r *searchRepository) getOrderClause(sortBy, sortOrder, query string) string {
	direction := "DESC"
	if sortOrder == "asc" {
		direction = "ASC"
	}

	switch sortBy {
	case "relevance":
		if query != "" {
			searchQuery := formatSearchQuery(query)
			return fmt.Sprintf("ts_rank(to_tsvector('english', title || ' ' || COALESCE(short_description, '')), to_tsquery('english', '%s')) DESC", searchQuery)
		}
		return "created_at DESC"
	case "rating":
		return "rating " + direction
	case "price":
		return "price " + direction
	case "students":
		return "total_students " + direction
	case "newest":
		return "created_at DESC"
	case "popular":
		return "total_students DESC, rating DESC"
	default:
		return "created_at DESC"
	}
}

func (r *searchRepository) GetFacets(ctx context.Context, query string) (*repository.SearchFacets, error) {
	facets := &repository.SearchFacets{}

	// Category facets
	var categoryFacets []struct {
		CategoryID uuid.UUID
		Name       string
		Count      int64
	}

	categoryQuery := r.db.WithContext(ctx).
		Table("courses").
		Select("categories.id as category_id, categories.name, COUNT(*) as count").
		Joins("JOIN course_categories cc ON cc.course_id = courses.id").
		Joins("JOIN categories ON categories.id = cc.category_id").
		Where("courses.status = ?", domain.CourseStatusPublished)

	if query != "" {
		searchQuery := formatSearchQuery(query)
		categoryQuery = categoryQuery.Where(
			"to_tsvector('english', courses.title || ' ' || COALESCE(courses.short_description, '')) @@ to_tsquery('english', ?)",
			searchQuery,
		)
	}

	categoryQuery.Group("categories.id, categories.name").
		Order("count DESC").
		Limit(10).
		Scan(&categoryFacets)

	for _, cf := range categoryFacets {
		facets.Categories = append(facets.Categories, repository.FacetItem{
			Label: cf.Name,
			Value: cf.CategoryID.String(),
			Count: cf.Count,
		})
	}

	// Level facets
	var levelFacets []struct {
		Level string
		Count int64
	}

	levelQuery := r.db.WithContext(ctx).
		Table("courses").
		Select("level, COUNT(*) as count").
		Where("status = ?", domain.CourseStatusPublished)

	if query != "" {
		searchQuery := formatSearchQuery(query)
		levelQuery = levelQuery.Where(
			"to_tsvector('english', title || ' ' || COALESCE(short_description, '')) @@ to_tsquery('english', ?)",
			searchQuery,
		)
	}

	levelQuery.Group("level").Order("count DESC").Scan(&levelFacets)

	for _, lf := range levelFacets {
		if lf.Level != "" {
			facets.Levels = append(facets.Levels, repository.FacetItem{
				Label: lf.Level,
				Value: lf.Level,
				Count: lf.Count,
			})
		}
	}

	// Price range facets
	facets.PriceRanges = []repository.FacetItem{
		{Label: "Free", Value: "free", Count: 0},
		{Label: "$0 - $25", Value: "0-25", Count: 0},
		{Label: "$25 - $50", Value: "25-50", Count: 0},
		{Label: "$50 - $100", Value: "50-100", Count: 0},
		{Label: "$100+", Value: "100+", Count: 0},
	}

	return facets, nil
}

func (r *searchRepository) GetSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	var suggestions []string

	searchQuery := formatPrefixQuery(query)

	err := r.db.WithContext(ctx).
		Model(&domain.Course{}).
		Select("DISTINCT title").
		Where("status = ?", domain.CourseStatusPublished).
		Where("to_tsvector('english', title) @@ to_tsquery('english', ?)", searchQuery).
		Limit(limit).
		Pluck("title", &suggestions).Error

	if err != nil {
		return nil, err
	}

	return suggestions, nil
}

func (r *searchRepository) GetTrendingSearches(ctx context.Context, limit int) ([]string, error) {
	var searches []string

	err := r.db.WithContext(ctx).
		Table("search_logs").
		Select("query, COUNT(*) as count").
		Where("created_at > NOW() - INTERVAL '7 days'").
		Group("query").
		Order("count DESC").
		Limit(limit).
		Pluck("query", &searches).Error

	if err != nil {
		// If search_logs table doesn't exist, return empty
		return []string{}, nil
	}

	return searches, nil
}

func (r *searchRepository) RecordSearch(ctx context.Context, query string, userID *uuid.UUID, resultCount int64) error {
	// Create search log entry (table may not exist yet)
	_ = r.db.WithContext(ctx).Exec(
		"INSERT INTO search_logs (query, user_id, result_count, created_at) VALUES (?, ?, ?, NOW())",
		query, userID, resultCount,
	)
	return nil
}

// formatSearchQuery formats for PostgreSQL to_tsquery
func formatSearchQuery(query string) string {
	words := strings.Fields(strings.TrimSpace(query))
	if len(words) == 0 {
		return ""
	}

	// Escape special characters and join with &
	escaped := make([]string, len(words))
	for i, word := range words {
		// Remove special tsquery characters
		word = strings.ReplaceAll(word, "'", "")
		word = strings.ReplaceAll(word, "&", "")
		word = strings.ReplaceAll(word, "|", "")
		word = strings.ReplaceAll(word, "!", "")
		word = strings.ReplaceAll(word, "(", "")
		word = strings.ReplaceAll(word, ")", "")
		if word != "" {
			escaped[i] = word
		}
	}

	return strings.Join(escaped, " & ")
}

// formatPrefixQuery formats for prefix matching (autocomplete)
func formatPrefixQuery(query string) string {
	words := strings.Fields(strings.TrimSpace(query))
	if len(words) == 0 {
		return ""
	}

	escaped := make([]string, len(words))
	for i, word := range words {
		word = strings.ReplaceAll(word, "'", "")
		if word != "" {
			escaped[i] = word + ":*"
		}
	}

	return strings.Join(escaped, " & ")
}
