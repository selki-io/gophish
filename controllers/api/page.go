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

// Pages handles requests for the /api/pages/ endpoint
func (as *Server) Pages(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		// Parse query parameters
		params := ParseQueryParams(r)

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
			result, err := models.GetPagesWithQuery(uid, modelParams)
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
			ps, err := models.GetPages(uid)
			if err != nil {
				log.Error(err)
			}
			JSONResponse(w, ps, http.StatusOK)
		}
	//POST: Create a new page and return it as JSON
	case r.Method == "POST":
		p := models.Page{}
		// Put the request into a page
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid request"}, http.StatusBadRequest)
			return
		}
		// Check to make sure the name is unique
		_, err = models.GetPageByName(p.Name, ctx.Get(r, "user_id").(int64))
		if err != gorm.ErrRecordNotFound {
			JSONResponse(w, models.Response{Success: false, Message: "Page name already in use"}, http.StatusConflict)
			log.Error(err)
			return
		}
		p.ModifiedDate = time.Now().UTC()
		p.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PostPage(&p)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, p, http.StatusCreated)
	}
}

// Page contains functions to handle the GET'ing, DELETE'ing, and PUT'ing
// of a Page object
func (as *Server) Page(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	p, err := models.GetPage(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Page not found"}, http.StatusNotFound)
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, p, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeletePage(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting page"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Page Deleted Successfully"}, http.StatusOK)
	case r.Method == "PUT":
		p = models.Page{}
		err = json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			log.Error(err)
		}
		if p.Id != id {
			JSONResponse(w, models.Response{Success: false, Message: "/:id and /:page_id mismatch"}, http.StatusBadRequest)
			return
		}
		p.ModifiedDate = time.Now().UTC()
		p.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PutPage(&p)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error updating page: " + err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, p, http.StatusOK)
	}
}
