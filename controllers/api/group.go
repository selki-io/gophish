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

// Groups returns a list of groups if requested via GET.
// If requested via POST, APIGroups creates a new group and returns a reference to it.
func (as *Server) Groups(w http.ResponseWriter, r *http.Request) {
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

		// Check if using new query system or legacy id__in
		if params.HasPaging || len(params.Filters) > 1 || (len(params.Filters) == 1 && !params.HasFilter("id")) {
			// Use new query system
			modelParams := models.ConvertAPIQueryParams(params)
			result, err := models.GetGroupsWithQuery(uid, modelParams)
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
			// Legacy support - use existing logic for id__in
			var groupIds []int64
			if params.HasFilter("id") {
				if filter, exists := params.GetFilterValue("id"); exists && filter.Operator == "in" {
					if ids, ok := filter.Value.([]interface{}); ok {
						for _, id := range ids {
							if idInt, ok := id.(int64); ok {
								groupIds = append(groupIds, idInt)
							}
						}
					}
				}
			}

			// Get groups with optional ID filter
			gs, err := models.GetGroups(uid, groupIds...)
			if err != nil {
				JSONResponse(w, models.Response{Success: false, Message: "No groups found"}, http.StatusNotFound)
				return
			}
			JSONResponse(w, gs, http.StatusOK)
		}
	//POST: Create a new group and return it as JSON
	case r.Method == "POST":
		g := models.Group{}
		// Put the request into a group
		err := json.NewDecoder(r.Body).Decode(&g)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
			return
		}
		_, err = models.GetGroupByName(g.Name, ctx.Get(r, "user_id").(int64))
		if err != gorm.ErrRecordNotFound {
			JSONResponse(w, models.Response{Success: false, Message: "Group name already in use"}, http.StatusConflict)
			return
		}
		g.ModifiedDate = time.Now().UTC()
		g.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PostGroup(&g)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		JSONResponse(w, g, http.StatusCreated)
	}
}

// GroupsSummary returns a summary of the groups owned by the current user.
func (as *Server) GroupsSummary(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		gs, err := models.GetGroupSummaries(ctx.Get(r, "user_id").(int64))
		if err != nil {
			log.Error(err)
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, gs, http.StatusOK)
	}
}

// Group returns details about the requested group.
// If the group is not valid, Group returns null.
func (as *Server) Group(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	g, err := models.GetGroup(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Group not found"}, http.StatusNotFound)
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, g, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeleteGroup(&g)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting group"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Group deleted successfully!"}, http.StatusOK)
	case r.Method == "PUT":
		// Change this to get from URL and uid (don't bother with id in r.Body)
		g = models.Group{}
		err = json.NewDecoder(r.Body).Decode(&g)
		if err != nil {
			log.Errorf("error decoding group: %v", err)
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		if g.Id != id {
			JSONResponse(w, models.Response{Success: false, Message: "Error: /:id and group_id mismatch"}, http.StatusInternalServerError)
			return
		}
		g.ModifiedDate = time.Now().UTC()
		g.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PutGroup(&g)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		JSONResponse(w, g, http.StatusOK)
	}
}

// GroupSummary returns a summary of the groups owned by the current user.
func (as *Server) GroupSummary(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 0, 64)
		g, err := models.GetGroupSummary(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Group not found"}, http.StatusNotFound)
			return
		}
		JSONResponse(w, g, http.StatusOK)
	}
}
