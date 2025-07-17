package api

import (
	"net/http"

	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
)

// CountHandler is a generic function type for count endpoints
type CountHandler func(uid int64, params *models.QueryParams) (*models.CountResult, error)

// GenericCountEndpoint creates a generic count endpoint handler
func GenericCountEndpoint(countFunc CountHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusMethodNotAllowed)
			return
		}

		// Parse query parameters
		params := ParseQueryParams(r)
		log.Debugf("Count query params: %v", params)

		// Get user ID from context
		userID := ctx.Get(r, "user_id")
		if userID == nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid user context"}, http.StatusBadRequest)
			return
		}
		uid := userID.(int64)

		// Convert to models params
		modelParams := models.ConvertAPIQueryParams(params)

		// Execute count query
		result, err := countFunc(uid, modelParams)
		if err != nil {
			log.Error(err)
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}

		// Return count result
		JSONResponse(w, models.Response{
			Success: true,
			Message: "Count retrieved successfully",
			Data:    result,
		}, http.StatusOK)
	}
}

// TemplatesCount endpoint
func (as *Server) TemplatesCount(w http.ResponseWriter, r *http.Request) {
	GenericCountEndpoint(models.GetTemplatesCount)(w, r)
}

// CampaignsCount endpoint
func (as *Server) CampaignsCount(w http.ResponseWriter, r *http.Request) {
	GenericCountEndpoint(models.GetCampaignsCount)(w, r)
}

// GroupsCount endpoint
func (as *Server) GroupsCount(w http.ResponseWriter, r *http.Request) {
	GenericCountEndpoint(models.GetGroupsCount)(w, r)
}

// UsersCount endpoint
func (as *Server) UsersCount(w http.ResponseWriter, r *http.Request) {
	GenericCountEndpoint(models.GetUsersCount)(w, r)
}

// SMTPCount endpoint
func (as *Server) SMTPCount(w http.ResponseWriter, r *http.Request) {
	GenericCountEndpoint(models.GetSMTPsCount)(w, r)
}

// PagesCount endpoint
func (as *Server) PagesCount(w http.ResponseWriter, r *http.Request) {
	GenericCountEndpoint(models.GetPagesCount)(w, r)
}
