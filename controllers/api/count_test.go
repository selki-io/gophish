package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ctx "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/models"
)

func makeCountRequest(testCtx *testContext, endpoint string) (int64, error) {
	req := httptest.NewRequest(http.MethodGet, endpoint, nil)
	req = ctx.Set(req, "user_id", testCtx.admin.Id)
	req = ctx.Set(req, "user", testCtx.admin)

	response := httptest.NewRecorder()

	// Call the handler directly based on the endpoint
	// Extract path without query params
	path := endpoint
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	switch path {
	case "/api/templates/_count":
		testCtx.apiServer.TemplatesCount(response, req)
	case "/api/campaigns/_count":
		testCtx.apiServer.CampaignsCount(response, req)
	case "/api/groups/_count":
		testCtx.apiServer.GroupsCount(response, req)
	case "/api/smtp/_count":
		testCtx.apiServer.SMTPCount(response, req)
	case "/api/pages/_count":
		testCtx.apiServer.PagesCount(response, req)
	case "/api/users/_count":
		testCtx.apiServer.UsersCount(response, req)
	default:
		return 0, fmt.Errorf("unknown endpoint: %s", path)
	}

	if response.Code != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d, body: %s", response.Code, response.Body.String())
	}

	var resp models.Response
	err := json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		return 0, fmt.Errorf("failed to decode response: %v, body: %s", err, response.Body.String())
	}

	if !resp.Success {
		return 0, fmt.Errorf("request failed: %s", resp.Message)
	}

	// Extract count from response
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("unexpected data format: %T", resp.Data)
	}
	countFloat, ok := data["count"].(float64)
	if !ok {
		return 0, fmt.Errorf("count not found or invalid type: %v", data["count"])
	}
	return int64(countFloat), nil
}

func TestTemplatesCount(t *testing.T) {
	testCtx := setupTest(t)

	// Create test templates
	t1 := models.Template{Name: "Marketing Template", UserId: testCtx.admin.Id, Lang: "en", Text: "Test content"}
	err := models.PostTemplate(&t1)
	if err != nil {
		t.Fatalf("Failed to create template 1: %v", err)
	}
	defer models.DeleteTemplate(t1.Id, testCtx.admin.Id)

	t2 := models.Template{Name: "Sales Template", UserId: testCtx.admin.Id, Lang: "es", Text: "Test content"}
	err = models.PostTemplate(&t2)
	if err != nil {
		t.Fatalf("Failed to create template 2: %v", err)
	}
	defer models.DeleteTemplate(t2.Id, testCtx.admin.Id)

	t3 := models.Template{Name: "Marketing Update", UserId: testCtx.admin.Id, Lang: "en", Text: "Test content"}
	err = models.PostTemplate(&t3)
	if err != nil {
		t.Fatalf("Failed to create template 3: %v", err)
	}
	defer models.DeleteTemplate(t3.Id, testCtx.admin.Id)

	// Test 1: Count all templates
	count, err := makeCountRequest(testCtx, "/api/templates/_count")
	if err != nil {
		t.Fatalf("Failed to get templates count: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 templates, got %d", count)
	}

	// Test 2: Count templates with name containing "Marketing"
	count, err = makeCountRequest(testCtx, "/api/templates/_count?name__contains=Marketing")
	if err != nil {
		t.Fatalf("Failed to get filtered count: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 templates with 'Marketing' in name, got %d", count)
	}

	// Test 3: Count templates with lang=en
	count, err = makeCountRequest(testCtx, "/api/templates/_count?lang=en")
	if err != nil {
		t.Fatalf("Failed to get lang filtered count: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 English templates, got %d", count)
	}

	// Test 4: Count templates with combined filters
	// Should match only "Marketing Template" (has lang=en AND name contains "Marketing")
	// Should NOT match "Marketing Update" if it has lang=en but we're looking for exact match
	count, err = makeCountRequest(testCtx, "/api/templates/_count?lang=en&name__contains=Marketing")
	if err != nil {
		t.Fatalf("Failed to get combined filtered count: %v", err)
	}
	// We have "Marketing Template" and "Marketing Update", both with lang=en
	if count != 2 {
		t.Errorf("Expected 2 English templates with 'Marketing' in name, got %d", count)
	}
}

func TestCampaignsCount(t *testing.T) {
	testCtx := setupTest(t)

	// Create test data
	g := models.Group{
		Name:   "Test Group",
		UserId: testCtx.admin.Id,
		Targets: []models.Target{
			{BaseRecipient: models.BaseRecipient{Email: "test@example.com"}},
		},
	}
	err := models.PostGroup(&g)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	defer models.DeleteGroup(&g)

	template := models.Template{Name: "Test Template", UserId: testCtx.admin.Id, Text: "Test content"}
	err = models.PostTemplate(&template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}
	defer models.DeleteTemplate(template.Id, testCtx.admin.Id)

	page := models.Page{Name: "Test Page", UserId: testCtx.admin.Id, HTML: "<html>Test</html>"}
	err = models.PostPage(&page)
	if err != nil {
		t.Fatalf("Failed to create page: %v", err)
	}
	defer models.DeletePage(page.Id, testCtx.admin.Id)

	smtp := models.SMTP{Name: "Test SMTP", Host: "smtp.test.com", FromAddress: "test@test.com", UserId: testCtx.admin.Id}
	err = models.PostSMTP(&smtp)
	if err != nil {
		t.Fatalf("Failed to create SMTP: %v", err)
	}
	defer models.DeleteSMTP(smtp.Id, smtp.UserId)

	// Create test campaigns
	c1 := models.Campaign{
		Name:     "Active Campaign",
		UserId:   testCtx.admin.Id,
		Template: template,
		Page:     page,
		SMTP:     smtp,
		Groups:   []models.Group{g},
	}
	err = models.PostCampaign(&c1, testCtx.admin.Id)
	if err != nil {
		t.Fatalf("Failed to create campaign 1: %v", err)
	}
	defer models.DeleteCampaign(c1.Id)

	c2 := models.Campaign{
		Name:     "Another Campaign",
		UserId:   testCtx.admin.Id,
		Template: template,
		Page:     page,
		SMTP:     smtp,
		Groups:   []models.Group{g},
	}
	err = models.PostCampaign(&c2, testCtx.admin.Id)
	if err != nil {
		t.Fatalf("Failed to create campaign 2: %v", err)
	}
	defer models.DeleteCampaign(c2.Id)

	// Test count endpoint
	count, err := makeCountRequest(testCtx, "/api/campaigns/_count")
	if err != nil {
		t.Fatalf("Failed to get campaigns count: %v", err)
	}
	if count < 2 {
		t.Errorf("Expected at least 2 campaigns, got %d", count)
	}

	// Test with filter
	count, err = makeCountRequest(testCtx, "/api/campaigns/_count?name__contains=Active")
	if err != nil {
		t.Fatalf("Failed to get filtered campaigns count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 campaign with 'Active' in name, got %d", count)
	}
}

func TestGroupsCount(t *testing.T) {
	testCtx := setupTest(t)

	// Create test groups
	g1 := models.Group{
		Name:   "Marketing Group",
		UserId: 1,
		Targets: []models.Target{
			{BaseRecipient: models.BaseRecipient{Email: "test1@example.com"}},
		},
	}
	models.PostGroup(&g1)
	defer models.DeleteGroup(&g1)

	g2 := models.Group{
		Name:   "Sales Group",
		UserId: 1,
		Targets: []models.Target{
			{BaseRecipient: models.BaseRecipient{Email: "test2@example.com"}},
		},
	}
	models.PostGroup(&g2)
	defer models.DeleteGroup(&g2)

	// Test count all
	count, err := makeCountRequest(testCtx, "/api/groups/_count")
	if err != nil {
		t.Fatalf("Failed to get groups count: %v", err)
	}
	if count < 2 {
		t.Errorf("Expected at least 2 groups, got %d", count)
	}

	// Test with filter
	count, err = makeCountRequest(testCtx, "/api/groups/_count?name__contains=Marketing")
	if err != nil {
		t.Fatalf("Failed to get filtered groups count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 group with 'Marketing' in name, got %d", count)
	}
}

func TestSMTPCount(t *testing.T) {
	testCtx := setupTest(t)

	// Create test SMTP profiles
	smtp1 := models.SMTP{
		Name:        "Global SMTP",
		Host:        "smtp.global.com",
		FromAddress: "test@global.com",
		UserId:      1,
	}
	models.PostSMTP(&smtp1)
	defer models.DeleteSMTP(smtp1.Id, smtp1.UserId)

	smtp2 := models.SMTP{
		Name:        "Local SMTP",
		Host:        "smtp.local.com",
		FromAddress: "test@local.com",
		UserId:      1,
	}
	models.PostSMTP(&smtp2)
	defer models.DeleteSMTP(smtp2.Id, smtp2.UserId)

	// Test count all
	count, err := makeCountRequest(testCtx, "/api/smtp/_count")
	if err != nil {
		t.Fatalf("Failed to get SMTP count: %v", err)
	}
	if count < 2 {
		t.Errorf("Expected at least 2 SMTP profiles, got %d", count)
	}

	// Test with filter
	count, err = makeCountRequest(testCtx, "/api/smtp/_count?name__contains=Global")
	if err != nil {
		t.Fatalf("Failed to get filtered SMTP count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 SMTP profile with 'Global' in name, got %d", count)
	}
}

func TestPagesCount(t *testing.T) {
	testCtx := setupTest(t)

	// Create test pages
	p1 := models.Page{
		Name:               "Login Page",
		CaptureCredentials: true,
		UserId:             testCtx.admin.Id,
		HTML:               "<html>Login</html>",
	}
	err := models.PostPage(&p1)
	if err != nil {
		t.Fatalf("Failed to create page 1: %v", err)
	}
	defer models.DeletePage(p1.Id, testCtx.admin.Id)

	p2 := models.Page{
		Name:               "Survey Page",
		CaptureCredentials: false,
		UserId:             testCtx.admin.Id,
		HTML:               "<html>Survey</html>",
	}
	err = models.PostPage(&p2)
	if err != nil {
		t.Fatalf("Failed to create page 2: %v", err)
	}
	defer models.DeletePage(p2.Id, testCtx.admin.Id)

	// Test count all
	count, err := makeCountRequest(testCtx, "/api/pages/_count")
	if err != nil {
		t.Fatalf("Failed to get pages count: %v", err)
	}
	if count < 2 {
		t.Errorf("Expected at least 2 pages, got %d", count)
	}

	// Test with filter
	count, err = makeCountRequest(testCtx, "/api/pages/_count?capture_credentials=true")
	if err != nil {
		t.Fatalf("Failed to get filtered pages count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 page with capture_credentials=true, got %d", count)
	}
}

func TestCountEndpointIgnoresPagination(t *testing.T) {
	testCtx := setupTest(t)

	// Create test templates to ensure we have data
	for i := 0; i < 5; i++ {
		template := models.Template{Name: fmt.Sprintf("Test %d", i), UserId: testCtx.admin.Id, Text: "Test content"}
		err := models.PostTemplate(&template)
		if err != nil {
			t.Fatalf("Failed to create template %d: %v", i, err)
		}
		defer models.DeleteTemplate(template.Id, testCtx.admin.Id)
	}

	// Test that pagination params are ignored
	count, err := makeCountRequest(testCtx, "/api/templates/_count?limit=1&offset=2")
	if err != nil {
		t.Fatalf("Failed to get count with pagination params: %v", err)
	}
	// Should return total count, not limited by pagination
	if count < 5 {
		t.Errorf("Expected at least 5 templates (pagination should be ignored), got %d", count)
	}
}
