package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// Templates handles the functionality for the /api/templates endpoint
func (as *Server) Templates(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		// Parse query parameters
		params := ParseQueryParams(r)
		log.Debugf("Query params: %v", params)

		// Get user ID from context
		userID := ctx.Get(r, "user_id")
		if userID == nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid user context"}, http.StatusBadRequest)
			return
		}
		uid := userID.(int64)

		// Check if using new query system
		if params.HasPaging || len(params.Filters) > 0 || params.OrderBy != "id" || params.OrderDir != "desc" {
			// Use new query system
			modelParams := models.ConvertAPIQueryParams(params)
			result, err := models.GetTemplatesWithQuery(uid, modelParams)
			if err != nil {
				log.Error(err)
				JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			// Extract data and metadata
			data := ExtractDataFromQueryResult(result)
			meta := ConvertQueryResultToMeta(result)
			JSONResponseWithPagination(w, data, meta, http.StatusOK)
		} else {
			// Legacy support - use existing logic
			ts, err := models.GetTemplates(uid)
			if err != nil {
				log.Error(err)
			}
			JSONResponse(w, ts, http.StatusOK)
		}
	//POST: Create a new template and return it as JSON
	case r.Method == "POST":
		t := models.Template{}
		// Put the request into a template
		err := json.NewDecoder(r.Body).Decode(&t)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
			return
		}
		_, err = models.GetTemplateByName(t.Name, ctx.Get(r, "user_id").(int64))
		if err != gorm.ErrRecordNotFound {
			JSONResponse(w, models.Response{Success: false, Message: "Template name already in use"}, http.StatusConflict)
			return
		}
		t.ModifiedDate = time.Now().UTC()
		t.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PostTemplate(&t)
		if err == models.ErrTemplateNameNotSpecified {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		if err == models.ErrTemplateMissingParameter {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		if err == models.ErrTemplateInvalidLanguage {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error inserting template into database"}, http.StatusInternalServerError)
			log.Error(err)
			return
		}
		JSONResponse(w, t, http.StatusCreated)
	}
}

// Template handles the functions for the /api/templates/:id endpoint
func (as *Server) Template(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	t, err := models.GetTemplate(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Template not found"}, http.StatusNotFound)
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, t, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeleteTemplate(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting template"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Template deleted successfully!"}, http.StatusOK)
	case r.Method == "PUT":
		t = models.Template{}
		err = json.NewDecoder(r.Body).Decode(&t)
		if err != nil {
			log.Error(err)
		}
		if t.Id != id {
			JSONResponse(w, models.Response{Success: false, Message: "Error: /:id and template_id mismatch"}, http.StatusBadRequest)
			return
		}
		t.ModifiedDate = time.Now().UTC()
		t.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PutTemplate(&t)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		JSONResponse(w, t, http.StatusOK)
	}
}
