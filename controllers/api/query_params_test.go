package api

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseQueryParams(t *testing.T) {
	// Test basic pagination parameters
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "limit=50&offset=100&page=3&order_by=name&order_dir=asc",
		},
	}

	params := ParseQueryParams(req)
	assert.Equal(t, 50, params.Limit)
	assert.Equal(t, 100, params.Offset)
	assert.Equal(t, 3, params.Page)
	assert.Equal(t, "name", params.OrderBy)
	assert.Equal(t, "asc", params.OrderDir)
	assert.True(t, params.HasPaging)
}

func TestParseQueryParamsWithFilters(t *testing.T) {
	// Test filter parameters
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "name__contains=test&status=active&id__in=1,2,3",
		},
	}

	params := ParseQueryParams(req)

	// Check name filter
	nameFilter, exists := params.GetFilterValue("name")
	assert.True(t, exists)
	assert.Equal(t, "contains", nameFilter.Operator)
	assert.Equal(t, "test", nameFilter.Value)

	// Check status filter (default eq operator)
	statusFilter, exists := params.GetFilterValue("status")
	assert.True(t, exists)
	assert.Equal(t, "eq", statusFilter.Operator)
	assert.Equal(t, "active", statusFilter.Value)

	// Check id filter with in operator
	idFilter, exists := params.GetFilterValue("id")
	assert.True(t, exists)
	assert.Equal(t, "in", idFilter.Operator)
	// For legacy id__in, the value is still a string, not parsed to []interface{}
	assert.Equal(t, "1,2,3", idFilter.Value)
}

func TestParseQueryParamsWithNewInSyntax(t *testing.T) {
	// Test new in syntax with proper parsing (using different field)
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "status__in=active,inactive,pending",
		},
	}

	params := ParseQueryParams(req)

	// Check status filter with in operator
	statusFilter, exists := params.GetFilterValue("status")
	assert.True(t, exists)
	assert.Equal(t, "in", statusFilter.Operator)
	// The parseFilterValue should parse it as []interface{} for new syntax
	values := statusFilter.Value.([]interface{})
	assert.Equal(t, 3, len(values))
	assert.Equal(t, "active", values[0])
	assert.Equal(t, "inactive", values[1])
	assert.Equal(t, "pending", values[2])
}

func TestParseQueryParamsDefaults(t *testing.T) {
	// Test default values
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "",
		},
	}

	params := ParseQueryParams(req)
	assert.Equal(t, 10000, params.Limit)
	assert.Equal(t, 0, params.Offset)
	assert.Equal(t, 1, params.Page)
	assert.Equal(t, "id", params.OrderBy)
	assert.Equal(t, "desc", params.OrderDir)
	assert.False(t, params.HasPaging)
}

func TestParseQueryParamsLegacyIdIn(t *testing.T) {
	// Test legacy id__in parameter
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "id__in=1,2,3",
		},
	}

	params := ParseQueryParams(req)

	// Check id filter
	idFilter, exists := params.GetFilterValue("id")
	assert.True(t, exists)
	assert.Equal(t, "in", idFilter.Operator)
	assert.Equal(t, "1,2,3", idFilter.Value)
}

func TestParseQueryParamsInvalidOperator(t *testing.T) {
	// Test invalid operator falls back to eq
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "name__invalid=test",
		},
	}

	params := ParseQueryParams(req)

	nameFilter, exists := params.GetFilterValue("name")
	assert.True(t, exists)
	assert.Equal(t, "eq", nameFilter.Operator)
	assert.Equal(t, "test", nameFilter.Value)
}

func TestParseQueryParamsIgnoresEmptyBraces(t *testing.T) {
	// Test that {} parameter is ignored
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "{}=&name=test&limit=20",
		},
	}

	params := ParseQueryParams(req)

	// Check that {} was ignored
	_, exists := params.GetFilterValue("{}")
	assert.False(t, exists)

	// Check that other params were parsed correctly
	nameFilter, exists := params.GetFilterValue("name")
	assert.True(t, exists)
	assert.Equal(t, "test", nameFilter.Value)
	assert.Equal(t, 20, params.Limit)
}

func TestParseQueryParamsIgnoresAPIKey(t *testing.T) {
	// Test that api_key parameter is ignored in filters
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "api_key=secret123&name=test",
		},
	}

	params := ParseQueryParams(req)

	// Check that api_key was ignored
	_, exists := params.GetFilterValue("api_key")
	assert.False(t, exists)

	// Check that other params were parsed correctly
	nameFilter, exists := params.GetFilterValue("name")
	assert.True(t, exists)
	assert.Equal(t, "test", nameFilter.Value)
}
