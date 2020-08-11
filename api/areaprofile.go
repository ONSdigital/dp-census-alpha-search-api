package api

import (
	"encoding/json"
	"net/http"

	errs "github.com/ONSdigital/dp-census-alpha-search-api/apierrors"
	"github.com/ONSdigital/dp-census-alpha-search-api/models"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

func (api *SearchAPI) getAreaProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	setAccessControl(w, http.MethodGet)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	logData := log.Data{"topic": id}

	log.Event(ctx, "getAreaProfile endpoint: incoming request", log.INFO, logData)

	query := models.AreaProfileQuery{
		Query: models.Query{
			Term: map[string]string{
				"id": id,
			},
		},
	}

	response, status, err := api.elasticsearch.GetAreaProfile(ctx, api.areaProfileIndex, query)
	if err != nil {
		logData["elasticsearch_status"] = status
		log.Event(ctx, "searchData endpoint: failed to get all datat type search results", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	b, err := json.Marshal(response)
	if err != nil {
		log.Event(ctx, "searchData endpoint: failed to marshal search resource into bytes", log.ERROR, log.Error(err), logData)
		setErrorCode(w, errs.ErrInternalServer)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.Event(ctx, "searchData endpoint: error writing response", log.ERROR, log.Error(err), logData)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Event(ctx, "searchData endpoint: successfully searched index", log.INFO, logData)
}
