package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// QueryParams represents parsed query parameters for pagination, filtering, and ordering
type QueryParams struct {
	Limit     int                    `json:"limit"`
	Offset    int                    `json:"offset"`
	Page      int                    `json:"page"`
	OrderBy   string                 `json:"order_by"`
	OrderDir  string                 `json:"order_dir"`
	Filters   map[string]FilterParam `json:"filters"`
	HasPaging bool                   `json:"has_paging"`
}

// FilterParam represents a filter parameter with operator and value
type FilterParam struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// FilterOperator constants for supported filter operations
const (
	FilterEq       = "eq"       // equals (default)
	FilterNe       = "ne"       // not equals
	FilterContains = "contains" // contains (LIKE %value%)
	FilterIn       = "in"       // in (value1,value2,...)
	FilterGt       = "gt"       // greater than
	FilterGte      = "gte"      // greater than or equal
	FilterLt       = "lt"       // less than
	FilterLte      = "lte"      // less than or equal
)

// ParseQueryParams parses HTTP query parameters into a QueryParams struct
func ParseQueryParams(r *http.Request) *QueryParams {
	params := &QueryParams{
		Limit:     10000, // Default high limit for frontend compatibility
		Offset:    0,
		Page:      1,
		OrderBy:   "id",
		OrderDir:  "desc",
		Filters:   make(map[string]FilterParam),
		HasPaging: false,
	}

	query := r.URL.Query()

	// Parse pagination parameters
	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			params.Limit = limit
			params.HasPaging = true
		}
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			params.Offset = offset
			params.HasPaging = true
		}
	}

	if pageStr := query.Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
			params.Offset = (page - 1) * params.Limit
			params.HasPaging = true
		}
	}

	// Parse ordering parameters
	if orderBy := query.Get("order_by"); orderBy != "" {
		params.OrderBy = orderBy
	}

	if orderDir := query.Get("order_dir"); orderDir != "" {
		if orderDir == "asc" || orderDir == "desc" {
			params.OrderDir = orderDir
		}
	}

	// Parse filter parameters
	for key, values := range query {
		if len(values) == 0 {
			continue
		}

		// Skip non-filter parameters and invalid keys
		if key == "limit" || key == "offset" || key == "page" || key == "order_by" || key == "order_dir" || key == "api_key" || key == "{}" || key == "" {
			continue
		}

		// Handle special case for existing id__in parameter
		if key == "id__in" {
			params.Filters["id"] = FilterParam{
				Field:    "id",
				Operator: FilterIn,
				Value:    values[0],
			}
			continue
		}

		// Parse filter parameters with operator syntax (field__operator)
		parts := strings.Split(key, "__")
		field := parts[0]
		operator := FilterEq // default operator

		if len(parts) > 1 {
			operator = parts[1]
		}

		// Validate operator
		if !isValidOperator(operator) {
			operator = FilterEq
		}

		// Parse value based on operator
		value := parseFilterValue(operator, values[0])

		params.Filters[field] = FilterParam{
			Field:    field,
			Operator: operator,
			Value:    value,
		}
	}

	return params
}

// isValidOperator checks if the operator is supported
func isValidOperator(operator string) bool {
	validOperators := []string{
		FilterEq, FilterNe, FilterContains, FilterIn,
		FilterGt, FilterGte, FilterLt, FilterLte,
	}

	for _, valid := range validOperators {
		if operator == valid {
			return true
		}
	}
	return false
}

// parseFilterValue parses the filter value based on the operator
func parseFilterValue(operator string, value string) interface{} {
	switch operator {
	case FilterIn:
		// Split comma-separated values
		parts := strings.Split(value, ",")
		var result []interface{}
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				// Try to parse as int64, fallback to string
				if intVal, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
					result = append(result, intVal)
				} else {
					result = append(result, trimmed)
				}
			}
		}
		return result
	case FilterGt, FilterGte, FilterLt, FilterLte:
		// Try to parse as int64 for numeric comparisons
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
		return value
	default:
		return value
	}
}

// GetFilterValue returns the filter value for a specific field
func (qp *QueryParams) GetFilterValue(field string) (FilterParam, bool) {
	filter, exists := qp.Filters[field]
	return filter, exists
}

// HasFilter checks if a filter exists for a specific field
func (qp *QueryParams) HasFilter(field string) bool {
	_, exists := qp.Filters[field]
	return exists
}

// String returns a string representation of the query parameters
func (qp *QueryParams) String() string {
	return fmt.Sprintf("QueryParams{Limit:%d, Offset:%d, Page:%d, OrderBy:%s, OrderDir:%s, Filters:%v}",
		qp.Limit, qp.Offset, qp.Page, qp.OrderBy, qp.OrderDir, qp.Filters)
}
