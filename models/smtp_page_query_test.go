package models

import (
	"testing"

	check "gopkg.in/check.v1"
)

func (s *ModelsSuite) TestSMTPQueryFunctionality(c *check.C) {
	// Create test SMTP profiles
	smtp1 := SMTP{
		Name:        "Global SMTP Server",
		Host:        "smtp.global.com",
		FromAddress: "noreply@global.com",
		UserId:      1,
	}
	err := PostSMTP(&smtp1)
	c.Assert(err, check.Equals, nil)
	defer DeleteSMTP(smtp1.Id, smtp1.UserId)

	smtp2 := SMTP{
		Name:        "Local SMTP Server",
		Host:        "smtp.local.com",
		FromAddress: "noreply@local.com",
		UserId:      1,
	}
	err = PostSMTP(&smtp2)
	c.Assert(err, check.Equals, nil)
	defer DeleteSMTP(smtp2.Id, smtp2.UserId)

	// Test GetSMTPsWithQuery with filtering
	params := &QueryParams{
		Filters: map[string]FilterParam{
			"name": {Field: "name", Operator: "contains", Value: "Global"},
		},
	}
	result, err := GetSMTPsWithQuery(1, params)
	c.Assert(err, check.Equals, nil)
	c.Assert(result.Total, check.Equals, int64(1))

	smtps := result.Data.(*[]SMTP)
	c.Assert(len(*smtps), check.Equals, 1)
	c.Assert((*smtps)[0].Name, check.Equals, "Global SMTP Server")

	// Test ordering
	params = &QueryParams{
		OrderBy:  "name",
		OrderDir: "asc",
	}
	result, err = GetSMTPsWithQuery(1, params)
	c.Assert(err, check.Equals, nil)

	smtps = result.Data.(*[]SMTP)
	c.Assert(len(*smtps) >= 2, check.Equals, true)
	// First should be "Global" (alphabetically before "Local")
	foundGlobal := false
	foundLocal := false
	for _, smtp := range *smtps {
		if smtp.Name == "Global SMTP Server" {
			foundGlobal = true
		}
		if smtp.Name == "Local SMTP Server" {
			foundLocal = true
		}
	}
	c.Assert(foundGlobal && foundLocal, check.Equals, true)
}

func (s *ModelsSuite) TestPagesQueryFunctionality(c *check.C) {
	// Create test pages
	page1 := Page{
		Name:               "Login Page",
		HTML:               "<html>Login</html>",
		CaptureCredentials: true,
		UserId:             1,
	}
	err := PostPage(&page1)
	c.Assert(err, check.Equals, nil)
	defer DeletePage(page1.Id, page1.UserId)

	page2 := Page{
		Name:               "Survey Page",
		HTML:               "<html>Survey</html>",
		CaptureCredentials: false,
		UserId:             1,
	}
	err = PostPage(&page2)
	c.Assert(err, check.Equals, nil)
	defer DeletePage(page2.Id, page2.UserId)

	// Test filtering by capture_credentials
	params := &QueryParams{
		Filters: map[string]FilterParam{
			"capture_credentials": {Field: "capture_credentials", Operator: "eq", Value: true},
		},
	}
	result, err := GetPagesWithQuery(1, params)
	c.Assert(err, check.Equals, nil)
	c.Assert(result.Total >= 1, check.Equals, true)

	pages := result.Data.(*[]Page)
	for _, page := range *pages {
		c.Assert(page.CaptureCredentials, check.Equals, true)
	}

	// Test pagination
	params = &QueryParams{
		Limit:     1,
		Offset:    0,
		HasPaging: true,
	}
	result, err = GetPagesWithQuery(1, params)
	c.Assert(err, check.Equals, nil)

	pages = result.Data.(*[]Page)
	c.Assert(len(*pages), check.Equals, 1)
	c.Assert(result.HasMore, check.Equals, result.Total > 1)
}

func (s *ModelsSuite) TestSMTPCount(c *check.C) {
	// Create test SMTP profiles
	smtp1 := SMTP{
		Name:        "Count Test SMTP 1",
		Host:        "smtp.test1.com",
		FromAddress: "test1@test.com",
		UserId:      1,
	}
	PostSMTP(&smtp1)
	defer DeleteSMTP(smtp1.Id, smtp1.UserId)

	smtp2 := SMTP{
		Name:        "Count Test SMTP 2",
		Host:        "smtp.test2.com",
		FromAddress: "test2@test.com",
		UserId:      1,
	}
	PostSMTP(&smtp2)
	defer DeleteSMTP(smtp2.Id, smtp2.UserId)

	// Test count without filters
	params := &QueryParams{}
	countResult, err := GetSMTPsCount(1, params)
	c.Assert(err, check.Equals, nil)
	c.Assert(countResult.Count >= 2, check.Equals, true)

	// Test count with filter
	params = &QueryParams{
		Filters: map[string]FilterParam{
			"name": {Field: "name", Operator: "contains", Value: "Count Test"},
		},
	}
	countResult, err = GetSMTPsCount(1, params)
	c.Assert(err, check.Equals, nil)
	c.Assert(countResult.Count, check.Equals, int64(2))
}

func (s *ModelsSuite) TestPagesCount(c *check.C) {
	// Create test pages
	page1 := Page{
		Name:               "Count Test Page 1",
		HTML:               "<html>Test 1</html>",
		CaptureCredentials: true,
		UserId:             1,
	}
	PostPage(&page1)
	defer DeletePage(page1.Id, page1.UserId)

	page2 := Page{
		Name:               "Count Test Page 2",
		HTML:               "<html>Test 2</html>",
		CaptureCredentials: false,
		UserId:             1,
	}
	PostPage(&page2)
	defer DeletePage(page2.Id, page2.UserId)

	// Test count without filters
	params := &QueryParams{}
	countResult, err := GetPagesCount(1, params)
	c.Assert(err, check.Equals, nil)
	c.Assert(countResult.Count >= 2, check.Equals, true)

	// Test count with filter
	params = &QueryParams{
		Filters: map[string]FilterParam{
			"capture_credentials": {Field: "capture_credentials", Operator: "eq", Value: true},
		},
	}
	countResult, err = GetPagesCount(1, params)
	c.Assert(err, check.Equals, nil)
	c.Assert(countResult.Count >= 1, check.Equals, true)
}

func TestSMTPAllowedFields(t *testing.T) {
	fields := GetAllowedFieldsForSMTP()

	// Check essential fields are allowed
	assert := func(field string) {
		if _, ok := fields[field]; !ok {
			t.Errorf("Expected field %s to be allowed for SMTP", field)
		}
	}

	assert("id")
	assert("name")
	assert("host")
	assert("username")
	assert("from_address")
	assert("interface_type")
	assert("modified_date")
}

func TestPageAllowedFields(t *testing.T) {
	fields := GetAllowedFieldsForPage()

	// Check essential fields are allowed
	assert := func(field string) {
		if _, ok := fields[field]; !ok {
			t.Errorf("Expected field %s to be allowed for Page", field)
		}
	}

	assert("id")
	assert("name")
	assert("capture_credentials")
	assert("capture_passwords")
	assert("redirect_url")
	assert("modified_date")
}
