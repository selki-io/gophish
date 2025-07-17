package models

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
)

// QueryBuilder provides a generic way to build GORM queries with pagination, filtering, and ordering
type QueryBuilder struct {
	db            *gorm.DB
	model         interface{}
	userID        int64
	baseWhere     string
	baseWhereArgs []interface{}
	allowedFields map[string]string // maps API field names to database column names
}

// NewQueryBuilder creates a new QueryBuilder instance
func NewQueryBuilder(model interface{}, userID int64) *QueryBuilder {
	return &QueryBuilder{
		db:            db,
		model:         model,
		userID:        userID,
		baseWhere:     "user_id = ?",
		allowedFields: make(map[string]string),
	}
}

// WithCustomBaseWhere allows customizing the base WHERE clause
func (qb *QueryBuilder) WithCustomBaseWhere(where string, args ...interface{}) *QueryBuilder {
	qb.baseWhere = where
	qb.baseWhereArgs = args
	return qb
}

// WithAllowedFields sets the mapping of API field names to database column names
func (qb *QueryBuilder) WithAllowedFields(fields map[string]string) *QueryBuilder {
	qb.allowedFields = fields
	return qb
}

// QueryParams represents the query parameters for building queries
type QueryParams struct {
	Limit     int                    `json:"limit"`
	Offset    int                    `json:"offset"`
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

// QueryResult represents the result of a query with metadata
type QueryResult struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
	Page       int         `json:"page"`
	TotalPages int         `json:"total_pages"`
	HasMore    bool        `json:"has_more"`
}

// BuildQuery constructs a GORM query based on the provided parameters
func (qb *QueryBuilder) BuildQuery(params *QueryParams) (*gorm.DB, error) {
	query := qb.db.Model(qb.model)

	// Apply base WHERE clause (usually user_id restriction)
	if qb.baseWhere != "" {
		if len(qb.baseWhereArgs) > 0 {
			// Use custom arguments
			args := append([]interface{}{qb.userID}, qb.baseWhereArgs...)
			query = query.Where(qb.baseWhere, args...)
		} else {
			// Use default userID
			query = query.Where(qb.baseWhere, qb.userID)
		}
	}

	// Apply filters
	for _, filter := range params.Filters {
		if err := qb.applyFilter(&query, filter); err != nil {
			return nil, err
		}
	}

	// Apply ordering
	if params.OrderBy != "" {
		orderField := qb.getDBFieldName(params.OrderBy)
		if orderField != "" {
			orderClause := fmt.Sprintf("%s %s", orderField, strings.ToUpper(params.OrderDir))
			query = query.Order(orderClause)
		}
	}

	return query, nil
}

// ExecuteQuery executes the query and returns results with metadata
func (qb *QueryBuilder) ExecuteQuery(params *QueryParams) (*QueryResult, error) {
	query, err := qb.BuildQuery(params)
	if err != nil {
		return nil, err
	}

	// Get total count (before pagination)
	var total int64
	countQuery := *query
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	if params.HasPaging {
		query = query.Offset(params.Offset).Limit(params.Limit)
	}

	// Execute query
	results := reflect.New(reflect.SliceOf(reflect.TypeOf(qb.model))).Interface()
	if err := query.Find(results).Error; err != nil {
		return nil, err
	}

	// Calculate pagination metadata
	page := 1
	totalPages := 1
	hasMore := false

	if params.Limit > 0 {
		page = (params.Offset / params.Limit) + 1
		totalPages = int((total + int64(params.Limit) - 1) / int64(params.Limit))
		hasMore = params.Offset+params.Limit < int(total)
	}

	return &QueryResult{
		Data:       results,
		Total:      total,
		Limit:      params.Limit,
		Offset:     params.Offset,
		Page:       page,
		TotalPages: totalPages,
		HasMore:    hasMore,
	}, nil
}

// applyFilter applies a single filter to the query
func (qb *QueryBuilder) applyFilter(query **gorm.DB, filter FilterParam) error {
	dbField := qb.getDBFieldName(filter.Field)
	if dbField == "" {
		return fmt.Errorf("invalid filter field: %s", filter.Field)
	}

	// Convert string boolean values to actual booleans if needed
	filterValue := filter.Value
	if strVal, ok := filter.Value.(string); ok {
		if strVal == "true" || strVal == "false" {
			filterValue = strVal == "true"
		}
	}

	switch filter.Operator {
	case "eq":
		*query = (*query).Where(fmt.Sprintf("%s = ?", dbField), filterValue)
	case "ne":
		*query = (*query).Where(fmt.Sprintf("%s != ?", dbField), filterValue)
	case "contains":
		*query = (*query).Where(fmt.Sprintf("%s LIKE ?", dbField), fmt.Sprintf("%%%s%%", filterValue))
	case "in":
		if values, ok := filterValue.([]interface{}); ok {
			*query = (*query).Where(fmt.Sprintf("%s IN (?)", dbField), values)
		} else {
			return fmt.Errorf("invalid value for 'in' operator: %v", filterValue)
		}
	case "gt":
		*query = (*query).Where(fmt.Sprintf("%s > ?", dbField), filterValue)
	case "gte":
		*query = (*query).Where(fmt.Sprintf("%s >= ?", dbField), filterValue)
	case "lt":
		*query = (*query).Where(fmt.Sprintf("%s < ?", dbField), filterValue)
	case "lte":
		*query = (*query).Where(fmt.Sprintf("%s <= ?", dbField), filterValue)
	default:
		return fmt.Errorf("unsupported filter operator: %s", filter.Operator)
	}

	return nil
}

// getDBFieldName maps API field names to database column names
func (qb *QueryBuilder) getDBFieldName(apiField string) string {
	// Check if there's a custom mapping
	if dbField, exists := qb.allowedFields[apiField]; exists {
		return dbField
	}

	// Default mapping: convert camelCase to snake_case
	return qb.camelToSnake(apiField)
}

// camelToSnake converts camelCase to snake_case
func (qb *QueryBuilder) camelToSnake(s string) string {
	var result strings.Builder
	for i, char := range s {
		if i > 0 && char >= 'A' && char <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(char)
	}
	return strings.ToLower(result.String())
}

// ConvertAPIQueryParams converts API QueryParams to models QueryParams
func ConvertAPIQueryParams(apiParams interface{}) *QueryParams {
	// Use reflection to convert between the two types
	if apiParams == nil {
		return nil
	}

	v := reflect.ValueOf(apiParams)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	params := &QueryParams{
		Filters: make(map[string]FilterParam),
	}

	// Extract fields using reflection
	if limitField := v.FieldByName("Limit"); limitField.IsValid() {
		params.Limit = int(limitField.Int())
	}

	if offsetField := v.FieldByName("Offset"); offsetField.IsValid() {
		params.Offset = int(offsetField.Int())
	}

	if orderByField := v.FieldByName("OrderBy"); orderByField.IsValid() {
		params.OrderBy = orderByField.String()
	}

	if orderDirField := v.FieldByName("OrderDir"); orderDirField.IsValid() {
		params.OrderDir = orderDirField.String()
	}

	if hasPagingField := v.FieldByName("HasPaging"); hasPagingField.IsValid() {
		params.HasPaging = hasPagingField.Bool()
	}

	if filtersField := v.FieldByName("Filters"); filtersField.IsValid() {
		if filtersField.Kind() == reflect.Map {
			for _, key := range filtersField.MapKeys() {
				value := filtersField.MapIndex(key)
				if value.IsValid() {
					// Convert the filter param
					filterParam := FilterParam{}
					filterValue := value

					// Handle both pointer and non-pointer cases
					if filterValue.Kind() == reflect.Ptr {
						filterValue = filterValue.Elem()
					}

					if fieldField := filterValue.FieldByName("Field"); fieldField.IsValid() {
						filterParam.Field = fieldField.String()
					}
					if operatorField := filterValue.FieldByName("Operator"); operatorField.IsValid() {
						filterParam.Operator = operatorField.String()
					}
					if valueField := filterValue.FieldByName("Value"); valueField.IsValid() {
						filterParam.Value = valueField.Interface()
					}

					params.Filters[key.String()] = filterParam
				}
			}
		}
	}

	return params
}

// CountResult represents the result of a count query
type CountResult struct {
	Count int64 `json:"count"`
}

// ExecuteCountQuery executes a count query and returns just the count with filters applied
func (qb *QueryBuilder) ExecuteCountQuery(params *QueryParams) (*CountResult, error) {
	query, err := qb.BuildQuery(params)
	if err != nil {
		return nil, err
	}

	// Get count with filters applied
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, err
	}

	return &CountResult{
		Count: count,
	}, nil
}

// GetAllowedFieldsForCampaign returns the allowed fields for campaign queries
func GetAllowedFieldsForCampaign() map[string]string {
	return map[string]string{
		"id":             "id",
		"name":           "name",
		"status":         "status",
		"created_date":   "created_date",
		"launch_date":    "launch_date",
		"completed_date": "completed_date",
		"template_id":    "template_id",
		"page_id":        "page_id",
		"smtp_id":        "smtp_id",
	}
}

// GetAllowedFieldsForGroup returns the allowed fields for group queries
func GetAllowedFieldsForGroup() map[string]string {
	return map[string]string{
		"id":            "id",
		"name":          "name",
		"modified_date": "modified_date",
	}
}

// GetAllowedFieldsForTemplate returns the allowed fields for template queries
func GetAllowedFieldsForTemplate() map[string]string {
	return map[string]string{
		"id":            "id",
		"name":          "name",
		"subject":       "subject",
		"modified_date": "modified_date",
		"lang":          "lang",
	}
}

// GetAllowedFieldsForUser returns the allowed fields for user queries
func GetAllowedFieldsForUser() map[string]string {
	return map[string]string{
		"id":             "id",
		"username":       "username",
		"role":           "role",
		"account_locked": "account_locked",
		"last_login":     "last_login",
	}
}

// GetAllowedFieldsForSMTP returns the allowed fields for SMTP queries
func GetAllowedFieldsForSMTP() map[string]string {
	return map[string]string{
		"id":                 "id",
		"name":               "name",
		"interface_type":     "interface_type",
		"host":               "host",
		"username":           "username",
		"from_address":       "from_address",
		"ignore_cert_errors": "ignore_cert_errors",
		"modified_date":      "modified_date",
	}
}

// GetAllowedFieldsForPage returns the allowed fields for page queries
func GetAllowedFieldsForPage() map[string]string {
	return map[string]string{
		"id":                  "id",
		"name":                "name",
		"capture_credentials": "capture_credentials",
		"capture_passwords":   "capture_passwords",
		"redirect_url":        "redirect_url",
		"modified_date":       "modified_date",
	}
}
