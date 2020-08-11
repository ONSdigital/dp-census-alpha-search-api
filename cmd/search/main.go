package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/ONSdigital/dp-census-alpha-search-api/api"
	"github.com/ONSdigital/dp-census-alpha-search-api/config"
	es "github.com/ONSdigital/dp-census-alpha-search-api/internal/elasticsearch"
	"github.com/ONSdigital/dp-census-alpha-search-api/models"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
)

func main() {
	log.Namespace = "dp-search-api"
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Event(ctx, "application unexpectedly failed", log.ERROR, log.Error(err))
		os.Exit(1)
	}

	os.Exit(0)
}

func run(ctx context.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.Get()
	if err != nil {
		log.Event(ctx, "failed to retrieve configuration", log.FATAL, log.Error(err))
		return err
	}

	log.Event(ctx, "config on startup", log.INFO, log.Data{"config": cfg})

	// Read in Taxonomy JSON into memory
	taxonomyFile, err := ioutil.ReadFile(cfg.TaxonomyFilename)
	if err != nil {
		log.Event(ctx, "failed to read taxonomy file", log.ERROR, log.Error(err), log.Data{"taxonomy_filename": cfg.TaxonomyFilename})
		return err
	}

	var taxonomy models.Taxonomy

	if err = json.Unmarshal([]byte(taxonomyFile), &taxonomy); err != nil {
		log.Event(ctx, "unable to unmarshal taxonomy into struct", log.ERROR, log.Error(err), log.Data{"taxonomy_filename": cfg.TaxonomyFilename})
		return err
	}

	// Read in Dimensions JSON into memory
	dimensionsFile, err := ioutil.ReadFile(cfg.DimensionsFilename)
	if err != nil {
		log.Event(ctx, "failed to read dimensions file", log.ERROR, log.Error(err), log.Data{"dimensions_filename": cfg.DimensionsFilename})
		return err
	}

	var dimensions models.DimensionsDoc

	if err = json.Unmarshal([]byte(dimensionsFile), &dimensions); err != nil {
		log.Event(ctx, "unable to unmarshal taxonomy into struct", log.ERROR, log.Error(err), log.Data{"dimensions_filename": cfg.DimensionsFilename})
		return err
	}

	// Read in Hierarchies JSON into memory
	hierarchiesFile, err := ioutil.ReadFile(cfg.HierarchiesFilename)
	if err != nil {
		log.Event(ctx, "failed to read hierarchies file", log.ERROR, log.Error(err), log.Data{"hierarchies_filename": cfg.HierarchiesFilename})
		return err
	}

	var hierarchies models.GeoHierarchiesDoc

	if err = json.Unmarshal([]byte(hierarchiesFile), &hierarchies); err != nil {
		log.Event(ctx, "unable to unmarshal taxonomy into struct", log.ERROR, log.Error(err), log.Data{"hierarchies_filename": cfg.HierarchiesFilename})
		return err
	}

	cli := dphttp.NewClient()
	esAPI := es.NewElasticSearchAPI(cli, cfg.ElasticSearchAPIURL)

	_, status, err := esAPI.CallElastic(ctx, cfg.ElasticSearchAPIURL, "GET", nil)
	if err != nil {
		log.Event(ctx, "failed to start up, unable to connect to elastic search instance", log.ERROR, log.Error(err), log.Data{"http_status": status})
		return err
	}

	apiErrors := make(chan error, 1)

	api.CreateAndInitialiseSearchAPI(ctx, cfg.BindAddr, esAPI, cfg.MaxSearchResultsOffset, cfg.DatasetIndex, cfg.AreaProfileIndex, cfg.PoscodeIndex, dimensions, taxonomy, hierarchies, apiErrors)

	// block until a fatal error occurs
	select {
	case err := <-apiErrors:
		log.Event(ctx, "api error received", log.ERROR, log.Error(err))
		return err
	case <-signals:
		log.Event(ctx, "os signal received", log.INFO)
	}

	return nil
}
