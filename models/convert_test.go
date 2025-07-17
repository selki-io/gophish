package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock API QueryParams to test conversion
type MockAPIQueryParams struct {
	Limit     int                           `json:"limit"`
	Offset    int                           `json:"offset"`
	OrderBy   string                        `json:"order_by"`
	OrderDir  string                        `json:"order_dir"`
	Filters   map[string]MockAPIFilterParam `json:"filters"`
	HasPaging bool                          `json:"has_paging"`
}

type MockAPIFilterParam struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

func TestConvertAPIQueryParams(t *testing.T) {
	// Test basic conversion
	mockParams := &MockAPIQueryParams{
		Limit:     50,
		Offset:    100,
		OrderBy:   "name",
		OrderDir:  "asc",
		HasPaging: true,
		Filters: map[string]MockAPIFilterParam{
			"status": {
				Field:    "status",
				Operator: "eq",
				Value:    "active",
			},
		},
	}

	converted := ConvertAPIQueryParams(mockParams)
	assert.NotNil(t, converted)
	assert.Equal(t, 50, converted.Limit)
	assert.Equal(t, 100, converted.Offset)
	assert.Equal(t, "name", converted.OrderBy)
	assert.Equal(t, "asc", converted.OrderDir)
	assert.True(t, converted.HasPaging)
	assert.Equal(t, 1, len(converted.Filters))

	statusFilter := converted.Filters["status"]
	assert.Equal(t, "status", statusFilter.Field)
	assert.Equal(t, "eq", statusFilter.Operator)
	assert.Equal(t, "active", statusFilter.Value)
}

func TestConvertAPIQueryParamsNil(t *testing.T) {
	// Test nil input
	converted := ConvertAPIQueryParams(nil)
	assert.Nil(t, converted)
}
