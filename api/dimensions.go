package api

import (
	"encoding/json"
	"net/http"

	errs "github.com/ONSdigital/dp-census-alpha-search-api/apierrors"
	"github.com/ONSdigital/log.go/log"
)

func (api *SearchAPI) getDimensions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	setAccessControl(w, http.MethodGet)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	log.Event(ctx, "getDimensions endpoint: incoming request", log.INFO)

	b, err := json.Marshal(api.dimensions)
	if err != nil {
		log.Event(ctx, "getDimensions endpoint: failed to marshal dimensions resource into bytes", log.ERROR, log.Error(err))
		setErrorCode(w, errs.ErrInternalServer)
	}

	_, err = w.Write(b)
	if err != nil {
		log.Event(ctx, "getDimensions endpoint: error writing response", log.ERROR, log.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Event(ctx, "getDimensions endpoint: successfully searched index", log.INFO)
}
