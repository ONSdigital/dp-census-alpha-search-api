package api

import (
	"encoding/json"
	"net/http"

	errs "github.com/ONSdigital/dp-census-alpha-search-api/apierrors"
	"github.com/ONSdigital/log.go/log"
)

func (api *SearchAPI) getHierarchies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	setAccessControl(w, http.MethodGet)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	log.Event(ctx, "getHierarchies endpoint: incoming request", log.INFO)

	b, err := json.Marshal(api.hierarchies)
	if err != nil {
		log.Event(ctx, "getHierarchies endpoint: failed to marshal hierarchies resource into bytes", log.ERROR, log.Error(err))
		setErrorCode(w, errs.ErrInternalServer)
	}

	_, err = w.Write(b)
	if err != nil {
		log.Event(ctx, "getHierarchies endpoint: error writing response", log.ERROR, log.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Event(ctx, "getHierarchies endpoint: successfully retrieved geography hierarchies", log.INFO)
}
