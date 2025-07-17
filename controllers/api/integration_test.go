package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gophish/gophish/models"
	"github.com/stretchr/testify/assert"
)

func TestTemplatesEndpointWithQueryParams(t *testing.T) {
	// Test that the endpoint doesn't panic when called with query parameters
	req := httptest.NewRequest("GET", "/api/templates", nil)
	w := httptest.NewRecorder()

	// Create a mock server
	server := &Server{}

	// This should not panic - test with simple request first
	server.Templates(w, req)

	// Check that we get a response (even if it's an error about missing API key)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// The response should contain error about invalid user context, not a panic
	body := w.Body.String()
	assert.Contains(t, strings.ToLower(body), "user context")
}

func TestParseQueryParamsIntegration(t *testing.T) {
	// Test that ParseQueryParams works correctly in the context of the API
	req := httptest.NewRequest("GET", "/api/templates?limit=25&page=2&name__contains=test&lang=en", nil)

	params := ParseQueryParams(req)

	assert.Equal(t, 25, params.Limit)
	assert.Equal(t, 2, params.Page)
	assert.Equal(t, 25, params.Offset) // Should be calculated from page and limit
	assert.True(t, params.HasPaging)

	// Check filters
	nameFilter, exists := params.GetFilterValue("name")
	assert.True(t, exists)
	assert.Equal(t, "contains", nameFilter.Operator)
	assert.Equal(t, "test", nameFilter.Value)

	langFilter, exists := params.GetFilterValue("lang")
	assert.True(t, exists)
	assert.Equal(t, "eq", langFilter.Operator)
	assert.Equal(t, "en", langFilter.Value)
}

func TestConvertAPIQueryParamsIntegration(t *testing.T) {
	// Test the conversion function with real API QueryParams
	req := httptest.NewRequest("GET", "/api/templates?limit=50&offset=100", nil)
	apiParams := ParseQueryParams(req)

	modelParams := models.ConvertAPIQueryParams(apiParams)

	assert.NotNil(t, modelParams)
	assert.Equal(t, 50, modelParams.Limit)
	assert.Equal(t, 100, modelParams.Offset)
	assert.True(t, modelParams.HasPaging)
}

func TestSMTPEndpointWithQueryParams(t *testing.T) {
	// Test SMTP endpoint with query parameters
	req := httptest.NewRequest("GET", "/api/smtp?name__contains=global&limit=10", nil)
	params := ParseQueryParams(req)

	assert.Equal(t, 10, params.Limit)

	nameFilter, exists := params.GetFilterValue("name")
	assert.True(t, exists)
	assert.Equal(t, "contains", nameFilter.Operator)
	assert.Equal(t, "global", nameFilter.Value)
}

func TestPagesEndpointWithQueryParams(t *testing.T) {
	// Test Pages endpoint with query parameters
	req := httptest.NewRequest("GET", "/api/pages?capture_credentials=true&order_by=name&order_dir=asc", nil)
	params := ParseQueryParams(req)

	assert.Equal(t, "name", params.OrderBy)
	assert.Equal(t, "asc", params.OrderDir)

	captureFilter, exists := params.GetFilterValue("capture_credentials")
	assert.True(t, exists)
	assert.Equal(t, "eq", captureFilter.Operator)
	assert.Equal(t, "true", captureFilter.Value)
}

func TestResponseWrapperWithPagination(t *testing.T) {
	// Test the JSONResponseWithPagination function
	w := httptest.NewRecorder()

	data := []map[string]interface{}{
		{"id": 1, "name": "Test 1"},
		{"id": 2, "name": "Test 2"},
	}

	meta := &MetaData{
		Total:      25,
		Limit:      10,
		Offset:     0,
		Page:       1,
		TotalPages: 3,
		HasMore:    true,
	}

	JSONResponseWithPagination(w, data, meta, http.StatusOK)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	// Check response structure
	body := w.Body.String()
	assert.Contains(t, body, `"data":[`)
	assert.Contains(t, body, `"meta":{`)
	assert.Contains(t, body, `"total":25`)
	assert.Contains(t, body, `"has_more":true`)
}

func TestEmptyBracesParameterHandling(t *testing.T) {
	// Test that {} parameter is properly ignored
	req := httptest.NewRequest("GET", "/api/templates?{}=&name=test", nil)
	params := ParseQueryParams(req)

	// {} should be ignored
	_, exists := params.GetFilterValue("{}")
	assert.False(t, exists)

	// name should be parsed
	nameFilter, exists := params.GetFilterValue("name")
	assert.True(t, exists)
	assert.Equal(t, "test", nameFilter.Value)
}

func TestAPIKeyIgnoredInFilters(t *testing.T) {
	// Test that api_key is not included in filters
	req := httptest.NewRequest("GET", "/api/templates?api_key=secret123&name=test", nil)
	params := ParseQueryParams(req)

	// api_key should be ignored
	_, exists := params.GetFilterValue("api_key")
	assert.False(t, exists)

	// name should be parsed
	nameFilter, exists := params.GetFilterValue("name")
	assert.True(t, exists)
	assert.Equal(t, "test", nameFilter.Value)
}
