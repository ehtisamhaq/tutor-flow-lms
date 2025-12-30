package handler

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/tutorflow/tutorflow-server/internal/middleware"
	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
	"github.com/tutorflow/tutorflow-server/internal/usecase/search"
)

// SearchHandler handles search HTTP requests
type SearchHandler struct {
	searchUC *search.UseCase
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(searchUC *search.UseCase) *SearchHandler {
	return &SearchHandler{searchUC: searchUC}
}

// RegisterRoutes registers search routes
func (h *SearchHandler) RegisterRoutes(g *echo.Group, optionalAuthMW echo.MiddlewareFunc) {
	s := g.Group("/search")
	s.GET("", h.Search, optionalAuthMW)
	s.GET("/suggestions", h.GetSuggestions)
	s.GET("/trending", h.GetTrendingSearches)
	s.GET("/facets", h.GetFacets)
	s.GET("/related/:courseId", h.GetRelatedCourses)
}

// Search godoc
// @Summary Search courses
// @Tags Search
// @Param q query string false "Search query"
// @Param category_id query string false "Category ID"
// @Param level query string false "Level (beginner, intermediate, advanced)"
// @Param min_price query number false "Minimum price"
// @Param max_price query number false "Maximum price"
// @Param min_rating query number false "Minimum rating"
// @Param is_free query boolean false "Free courses only"
// @Param sort_by query string false "Sort by (relevance, rating, price, students, newest)"
// @Param sort_order query string false "Sort order (asc, desc)"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Response{data=search.SearchResult}
// @Router /search [get]
func (h *SearchHandler) Search(c echo.Context) error {
	input := search.SearchInput{
		Query:     c.QueryParam("q"),
		SortBy:    c.QueryParam("sort_by"),
		SortOrder: c.QueryParam("sort_order"),
	}

	// Parse category ID
	if catID := c.QueryParam("category_id"); catID != "" {
		if id, err := uuid.Parse(catID); err == nil {
			input.CategoryID = &id
		}
	}

	// Parse level
	if level := c.QueryParam("level"); level != "" {
		input.Level = &level
	}

	// Parse prices
	if minPrice := c.QueryParam("min_price"); minPrice != "" {
		if val, err := strconv.ParseFloat(minPrice, 64); err == nil {
			input.MinPrice = &val
		}
	}
	if maxPrice := c.QueryParam("max_price"); maxPrice != "" {
		if val, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			input.MaxPrice = &val
		}
	}

	// Parse rating
	if minRating := c.QueryParam("min_rating"); minRating != "" {
		if val, err := strconv.ParseFloat(minRating, 64); err == nil {
			input.MinRating = &val
		}
	}

	// Parse is_free
	if isFree := c.QueryParam("is_free"); isFree == "true" {
		free := true
		input.IsFree = &free
	}

	// Parse pagination
	input.Page = 1
	input.Limit = 20
	if page := c.QueryParam("page"); page != "" {
		if val, err := strconv.Atoi(page); err == nil {
			input.Page = val
		}
	}
	if limit := c.QueryParam("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			input.Limit = val
		}
	}

	result, err := h.searchUC.Search(c.Request().Context(), input)
	if err != nil {
		return response.InternalError(c, "Search failed")
	}

	// Record search for analytics (optional)
	claims, ok := middleware.GetClaims(c)
	if ok && input.Query != "" {
		go func() {
			_ = h.searchUC.RecordSearch(c.Request().Context(), input.Query, &claims.UserID, result.Total)
		}()
	}

	return response.Success(c, result)
}

// GetSuggestions godoc
// @Summary Get search suggestions (autocomplete)
// @Tags Search
// @Param q query string true "Search query"
// @Param limit query int false "Limit (default 5)"
// @Success 200 {object} response.Response{data=[]string}
// @Router /search/suggestions [get]
func (h *SearchHandler) GetSuggestions(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return response.Success(c, []string{})
	}

	limit := 5
	if l := c.QueryParam("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}

	suggestions, err := h.searchUC.GetSuggestions(c.Request().Context(), query, limit)
	if err != nil {
		return response.Success(c, []string{})
	}

	return response.Success(c, suggestions)
}

// GetTrendingSearches godoc
// @Summary Get trending search terms
// @Tags Search
// @Param limit query int false "Limit (default 10)"
// @Success 200 {object} response.Response{data=[]string}
// @Router /search/trending [get]
func (h *SearchHandler) GetTrendingSearches(c echo.Context) error {
	limit := 10
	if l := c.QueryParam("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}

	trending, err := h.searchUC.GetTrendingSearches(c.Request().Context(), limit)
	if err != nil {
		return response.Success(c, []string{})
	}

	return response.Success(c, trending)
}

// GetFacets godoc
// @Summary Get search facets for filtering
// @Tags Search
// @Param q query string false "Search query to scope facets"
// @Success 200 {object} response.Response{data=search.SearchFacets}
// @Router /search/facets [get]
func (h *SearchHandler) GetFacets(c echo.Context) error {
	query := c.QueryParam("q")

	facets, err := h.searchUC.GetSearchFacets(c.Request().Context(), query)
	if err != nil {
		return response.InternalError(c, "Failed to get facets")
	}

	return response.Success(c, facets)
}

// GetRelatedCourses godoc
// @Summary Get courses related to a specific course
// @Tags Search
// @Param courseId path string true "Course ID"
// @Param limit query int false "Limit (default 4)"
// @Success 200 {object} response.Response
// @Router /search/related/{courseId} [get]
func (h *SearchHandler) GetRelatedCourses(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	limit := 4
	if l := c.QueryParam("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}

	courses, err := h.searchUC.GetRelatedCourses(c.Request().Context(), courseID, limit)
	if err != nil {
		return response.InternalError(c, "Failed to get related courses")
	}

	return response.Success(c, courses)
}
