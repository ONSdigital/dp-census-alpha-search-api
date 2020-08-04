package main

import (
	"context"
	"net/http"
	"os"

	es "github.com/ONSdigital/dp-census-search-prototypes/elasticsearch"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
)

const (
	elasticsearchAPIURL = "http://localhost:9200"
	geoFileIndex        = "test_geo"
	mappingsFile        = "geography-mappings.json"
)

func main() {
	ctx := context.Background()

	cli := dphttp.NewClient()
	esAPI := es.NewElasticSearchAPI(cli, elasticsearchAPIURL)

	// delete existing elasticsearch index if already exists
	status, err := esAPI.DeleteSearchIndex(ctx, geoFileIndex)
	if err != nil {
		if status != http.StatusNotFound {
			log.Event(ctx, "failed to delete index", log.ERROR, log.Error(err), log.Data{"status": status})
			os.Exit(1)
		}

		log.Event(ctx, "failed to delete index as index cannot be found, continuing", log.WARN, log.Error(err), log.Data{"status": status})
	}

	// create elasticsearch index with settings/mapping
	status, err = esAPI.CreateSearchIndex(ctx, geoFileIndex, mappingsFile)
	if err != nil {
		log.Event(ctx, "failed to create index", log.ERROR, log.Error(err), log.Data{"status": status})
		os.Exit(1)
	}

	log.Event(ctx, "successfully refreshed "+geoFileIndex+" index", log.INFO)
}
