package api

import (
	"context"

	"github.com/ONSdigital/dp-census-alpha-search-api/models"
	"github.com/ONSdigital/go-ns/server"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

var httpServer *server.Server

// SearchAPI manages searches across indices
type SearchAPI struct {
	areaProfileIndex  string
	datasetIndex      string
	defaultMaxResults int
	dimensions        models.DimensionsDoc
	elasticsearch     Elasticsearcher
	postcodeIndex     string
	router            *mux.Router
	taxonomy          models.Taxonomy
}

// CreateAndInitialiseSearchAPI manages all the routes configured to API
func CreateAndInitialiseSearchAPI(ctx context.Context, bindAddr string, esAPI Elasticsearcher, defaultMaxResults int, datasetIndex, areaProfileIndex, postcodeIndex string, dimensions models.DimensionsDoc, taxonomy models.Taxonomy, errorChan chan error) {

	router := mux.NewRouter()
	routes(ctx,
		router,
		esAPI,
		defaultMaxResults,
		datasetIndex,
		areaProfileIndex,
		postcodeIndex,
		dimensions,
		taxonomy,
	)

	httpServer = server.New(bindAddr, router)

	// Disable this here to allow service to manage graceful shutdown of the entire app.
	httpServer.HandleOSSignals = false

	go func() {
		log.Event(ctx, "Starting api...", log.INFO)
		if err := httpServer.ListenAndServe(); err != nil {
			log.Event(ctx, "api http server returned error", log.ERROR, log.Error(err))
			errorChan <- err
		}
	}()
}

func routes(ctx context.Context,
	router *mux.Router,
	elasticsearch Elasticsearcher,
	defaultMaxResults int,
	datasetIndex, areaProfileIndex, postcodeIndex string,
	dimensions models.DimensionsDoc,
	taxonomy models.Taxonomy) *SearchAPI {

	api := SearchAPI{
		areaProfileIndex:  areaProfileIndex,
		datasetIndex:      datasetIndex,
		defaultMaxResults: defaultMaxResults,
		dimensions:        dimensions,
		elasticsearch:     elasticsearch,
		postcodeIndex:     postcodeIndex,
		router:            router,
		taxonomy:          taxonomy,
	}

	// api.router.HandleFunc("/search", api.getAllSearch).Methods("GET", "OPTIONS")
	api.router.HandleFunc("/dimensions", api.getDimensions).Methods("GET", "OPTIONS")
	api.router.HandleFunc("/taxonomy", api.getTaxonomy).Methods("GET", "OPTIONS")
	api.router.HandleFunc("/taxonomy/{topic}", api.getTopic).Methods("GET", "OPTIONS")

	return &api
}

// Close represents the graceful shutting down of the http server
func Close(ctx context.Context) error {
	if err := httpServer.Shutdown(ctx); err != nil {
		return err
	}
	log.Event(ctx, "graceful shutdown of http server complete", log.INFO)
	return nil
}