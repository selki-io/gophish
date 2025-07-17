package api

import (
	"encoding/json"
	"net/http"
	"reflect"
)

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data    interface{} `json:"data"`
	Meta    *MetaData   `json:"meta,omitempty"`
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
}

// MetaData contains pagination metadata
type MetaData struct {
	Total      int64 `json:"total"`
	Limit      int   `json:"limit"`
	Offset     int   `json:"offset"`
	Page       int   `json:"page"`
	TotalPages int   `json:"total_pages"`
	HasMore    bool  `json:"has_more"`
}

// PaginatedJSONResponse sends a paginated JSON response
func PaginatedJSONResponse(w http.ResponseWriter, data interface{}, meta *MetaData, status int) {
	response := PaginatedResponse{
		Data:    data,
		Meta:    meta,
		Success: status >= 200 && status < 300,
	}

	if !response.Success {
		response.Message = "An error occurred"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// JSONResponseWithPagination sends either a paginated response or a simple response
// based on whether pagination parameters were provided
func JSONResponseWithPagination(w http.ResponseWriter, data interface{}, meta *MetaData, status int) {
	// If no pagination metadata, send simple response (backward compatibility)
	if meta == nil {
		JSONResponse(w, data, status)
		return
	}

	// Send paginated response
	PaginatedJSONResponse(w, data, meta, status)
}

// ConvertQueryResultToMeta converts a QueryResult to MetaData
func ConvertQueryResultToMeta(result interface{}) *MetaData {
	if result == nil {
		return nil
	}

	// Use reflection to check if the result has the expected fields
	v := reflect.ValueOf(result)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	// Extract metadata fields if they exist
	totalField := v.FieldByName("Total")
	limitField := v.FieldByName("Limit")
	offsetField := v.FieldByName("Offset")
	pageField := v.FieldByName("Page")
	totalPagesField := v.FieldByName("TotalPages")
	hasMoreField := v.FieldByName("HasMore")

	// Check if all required fields exist
	if !totalField.IsValid() || !limitField.IsValid() || !offsetField.IsValid() {
		return nil
	}

	meta := &MetaData{
		Total:  totalField.Int(),
		Limit:  int(limitField.Int()),
		Offset: int(offsetField.Int()),
	}

	if pageField.IsValid() {
		meta.Page = int(pageField.Int())
	}

	if totalPagesField.IsValid() {
		meta.TotalPages = int(totalPagesField.Int())
	}

	if hasMoreField.IsValid() {
		meta.HasMore = hasMoreField.Bool()
	}

	return meta
}

// ExtractDataFromQueryResult extracts the data field from a QueryResult
func ExtractDataFromQueryResult(result interface{}) interface{} {
	if result == nil {
		return nil
	}

	v := reflect.ValueOf(result)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return result
	}

	// Extract the Data field if it exists
	dataField := v.FieldByName("Data")
	if !dataField.IsValid() {
		return result
	}

	return dataField.Interface()
}

// CreateMetaFromParams creates MetaData from query parameters and total count
func CreateMetaFromParams(params *QueryParams, total int64, actualCount int) *MetaData {
	if params == nil || !params.HasPaging {
		return nil
	}

	totalPages := int((total + int64(params.Limit) - 1) / int64(params.Limit))
	hasMore := params.Offset+actualCount < int(total)

	return &MetaData{
		Total:      total,
		Limit:      params.Limit,
		Offset:     params.Offset,
		Page:       params.Page,
		TotalPages: totalPages,
		HasMore:    hasMore,
	}
}
