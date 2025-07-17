package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryBuilderFieldMapping(t *testing.T) {
	// Test field mapping logic
	campaign := Campaign{}
	builder := NewQueryBuilder(campaign, 1).
		WithAllowedFields(GetAllowedFieldsForCampaign())

	// Test allowed field mapping
	assert.Equal(t, "name", builder.getDBFieldName("name"))
	assert.Equal(t, "status", builder.getDBFieldName("status"))
	assert.Equal(t, "created_date", builder.getDBFieldName("created_date"))
}

func TestQueryBuilderTemplateFields(t *testing.T) {
	// Test template field mapping
	template := Template{}
	builder := NewQueryBuilder(template, 1).
		WithAllowedFields(GetAllowedFieldsForTemplate())

	// Test allowed field mapping
	assert.Equal(t, "name", builder.getDBFieldName("name"))
	assert.Equal(t, "lang", builder.getDBFieldName("lang"))
	assert.Equal(t, "modified_date", builder.getDBFieldName("modified_date"))
}

func TestCamelToSnake(t *testing.T) {
	builder := &QueryBuilder{}

	assert.Equal(t, "test_field", builder.camelToSnake("testField"))
	assert.Equal(t, "test_field_name", builder.camelToSnake("testFieldName"))
	assert.Equal(t, "id", builder.camelToSnake("id"))
	assert.Equal(t, "created_date", builder.camelToSnake("createdDate"))
}

func TestAllowedFields(t *testing.T) {
	// Test that allowed fields are returned correctly
	campaignFields := GetAllowedFieldsForCampaign()
	assert.Contains(t, campaignFields, "id")
	assert.Contains(t, campaignFields, "name")
	assert.Contains(t, campaignFields, "status")

	templateFields := GetAllowedFieldsForTemplate()
	assert.Contains(t, templateFields, "id")
	assert.Contains(t, templateFields, "name")
	assert.Contains(t, templateFields, "lang")

	groupFields := GetAllowedFieldsForGroup()
	assert.Contains(t, groupFields, "id")
	assert.Contains(t, groupFields, "name")

	userFields := GetAllowedFieldsForUser()
	assert.Contains(t, userFields, "id")
	assert.Contains(t, userFields, "username")

	smtpFields := GetAllowedFieldsForSMTP()
	assert.Contains(t, smtpFields, "id")
	assert.Contains(t, smtpFields, "name")
	assert.Contains(t, smtpFields, "host")

	pageFields := GetAllowedFieldsForPage()
	assert.Contains(t, pageFields, "id")
	assert.Contains(t, pageFields, "name")
	assert.Contains(t, pageFields, "capture_credentials")
}
