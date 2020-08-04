package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	errs "github.com/ONSdigital/dp-census-search-prototypes/apierrors"
	es "github.com/ONSdigital/dp-census-search-prototypes/elasticsearch"
	"github.com/ONSdigital/dp-census-search-prototypes/scripts/geojson/models"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
	"github.com/tamerh/jsparser"
)

const (
	elasticsearchAPIURL = "http://localhost:9200"
	features            = "features"
	geoFileIndex        = "test_geo"
	geoJSONPath         = "../geojson/"
)

var (
	countCh             = make(chan int)
	polygonCountCh      = make(chan int)
	multiPolygonCountCh = make(chan int)

	geojsonfiles = []string{
		"Output_Areas_(December_2011)_Boundaries_EW_BFC.geojson",
		"Output_Areas_(December_2011)_Boundaries_EW_BFE.geojson",
		"Output_Areas_(December_2011)_Boundaries_EW_BGC.geojson",
	}
)

func main() {
	ctx := context.Background()

	cli := dphttp.NewClient()
	esAPI := es.NewElasticSearchAPI(cli, elasticsearchAPIURL)

	go trackCounts(ctx)

	log.Event(ctx, "about to read in geojson", log.INFO)

	for _, filename := range geojsonfiles {
		fileLocation := geoJSONPath + filename
		f, err := os.Open(fileLocation)
		if err != nil {
			log.Event(ctx, "failed to open oa file", log.FATAL, log.Error(err))
			os.Exit(1)
		}

		br := bufio.NewReaderSize(f, 65536)
		parser := jsparser.NewJSONParser(br, features)

		log.Event(ctx, "about to store docs in elastic search", log.INFO)

		// Iterate items for individual geo boundaries and store documents in elasticsearch
		if err = storeDocs(ctx, esAPI, geoFileIndex, parser); err != nil {
			log.Event(ctx, "failed to store lsoa data in elasticsearch", log.FATAL, log.Error(err))
			os.Exit(1)
		}
	}

	log.Event(ctx, "successfully added 2011 oa data to "+geoFileIndex+" index", log.INFO)
}

func trackCounts(ctx context.Context) {
	var (
		totalCounter        = 0
		polygonCounter      = 0
		multiPolygonCounter = 0
	)

	t := time.NewTicker(5 * time.Second)

	for {
		select {
		case n := <-countCh:
			totalCounter += n
		case n := <-polygonCountCh:
			polygonCounter += n
		case n := <-multiPolygonCountCh:
			multiPolygonCounter += n
		case n := <-countCh:
			totalCounter += n
		case <-t.C:
			log.Event(ctx, "Total uploaded: "+strconv.Itoa(totalCounter)+" | Polygons: "+strconv.Itoa(polygonCounter)+" | MultiPolygons: "+strconv.Itoa(multiPolygonCounter), log.INFO)
		}
	}
}

func createGeoDoc(reader io.Reader) (*models.GeoDocs, error) {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errs.ErrUnableToReadMessage
	}

	var geoDocs models.GeoDocs

	err = json.Unmarshal(b, &geoDocs)
	if err != nil {
		return nil, errs.ErrUnableToParseJSON
	}
	return &geoDocs, nil
}

func storeDocs(ctx context.Context, esAPI *es.API, indexName string, parser *jsparser.JsonParser) error {
	count := 0
	polygonCount := 0
	multiPolygonCount := 0
	var geoDocs []interface{}

	// Iterate through the records
	for feature := range parser.Stream() {
		count++

		sA := feature.ObjectVals["properties"].(*jsparser.JSON).ObjectVals["Shape__Area"].(string)
		shapeArea, err := strconv.ParseFloat(sA, 64)
		if err != nil {
			log.Event(ctx, "failed to caste interface to float64", log.ERROR, log.Error(err), log.Data{"shape_area": sA})
			return err
		}

		sL := feature.ObjectVals["properties"].(*jsparser.JSON).ObjectVals["Shape__Length"].(string)
		shapeLength, err := strconv.ParseFloat(sL, 64)
		if err != nil {
			log.Event(ctx, "failed to caste interface to float64", log.ERROR, log.Error(err), log.Data{"shape_length": sL})
			return err
		}

		newDoc := &models.GeoDoc{
			Code:        feature.ObjectVals["properties"].(*jsparser.JSON).ObjectVals["LAD11CD"].(string),
			Hierarchy:   "Output Areas",
			LAD11CD:     feature.ObjectVals["properties"].(*jsparser.JSON).ObjectVals["LAD11CD"].(string),
			OA11CD:      feature.ObjectVals["properties"].(*jsparser.JSON).ObjectVals["OA11CD"].(string),
			ShapeArea:   shapeArea,
			ShapeLength: shapeLength,
			Location: models.GeoLocation{
				Type: feature.ObjectVals["geometry"].(*jsparser.JSON).ObjectVals["type"].(string),
			},
		}

		if newDoc.Location.Type == "MultiPolygon" {
			newDoc.Location.Coordinates, err = getMultiPolygonCoordinates(ctx, feature.ObjectVals["geometry"].(*jsparser.JSON).ObjectVals["coordinates"])
			multiPolygonCount++
		} else {
			newDoc.Location.Coordinates, err = getPolygonCoordinates(ctx, feature.ObjectVals["geometry"].(*jsparser.JSON).ObjectVals["coordinates"])
			polygonCount++

		}
		if err != nil {
			log.Event(ctx, "failed to get coordinates", log.ERROR, log.Error(err), log.Data{"count": count})
			return err
		}

		geoDocs = append(geoDocs, newDoc)

		if count == 100 {
			if _, err := esAPI.BulkRequest(ctx, indexName, geoDocs); err != nil {
				log.Event(ctx, "failed to upload document to index", log.ERROR, log.Error(err), log.Data{"count": count})
				return err
			}

			countCh <- count
			polygonCountCh <- polygonCount
			multiPolygonCountCh <- multiPolygonCount

			count = 0
			polygonCount = 0
			multiPolygonCount = 0
			geoDocs = nil
		}
	}

	// Capture last bulk
	if count != 0 {
		if _, err := esAPI.BulkRequest(ctx, indexName, geoDocs); err != nil {
			log.Event(ctx, "failed to upload document to index", log.ERROR, log.Error(err), log.Data{"count": count})
			return err
		}

		countCh <- count
		polygonCountCh <- polygonCount
		multiPolygonCountCh <- multiPolygonCount

		count = 0
		polygonCount = 0
		multiPolygonCount = 0
		geoDocs = nil
	}

	return nil
}

func getPolygonCoordinates(ctx context.Context, geometry interface{}) ([][][]float64, error) {
	var g [][][]float64
	for i := 0; i < len(geometry.(*jsparser.JSON).ArrayVals); i++ {
		var coordinates [][]float64
		for j := 0; j < len(geometry.(*jsparser.JSON).ArrayVals[i].(*jsparser.JSON).ArrayVals); j++ {
			k1 := geometry.(*jsparser.JSON).ArrayVals[i].(*jsparser.JSON).ArrayVals[j].(*jsparser.JSON).ArrayVals[0].(string)
			k2 := geometry.(*jsparser.JSON).ArrayVals[i].(*jsparser.JSON).ArrayVals[j].(*jsparser.JSON).ArrayVals[1].(string)

			lat, err := strconv.ParseFloat(k1, 64)
			if err != nil {
				log.Event(ctx, "failed to caste interface to float64", log.ERROR, log.Error(err), log.Data{"lat": k1})
				return g, err
			}

			lon, err := strconv.ParseFloat(k2, 64)
			if err != nil {
				log.Event(ctx, "failed to caste interface to float64", log.ERROR, log.Error(err), log.Data{"lon": k2})
				return g, err
			}

			coordinates = append(coordinates, []float64{lat, lon})
		}

		g = append(g, coordinates)
	}

	return g, nil
}

func getMultiPolygonCoordinates(ctx context.Context, geometry interface{}) ([][][][]float64, error) {
	var g [][][][]float64
	for i := 0; i < len(geometry.(*jsparser.JSON).ArrayVals); i++ {
		var multiCoordinates [][][]float64
		for j := 0; j < len(geometry.(*jsparser.JSON).ArrayVals[i].(*jsparser.JSON).ArrayVals); j++ {
			var coordinates [][]float64
			for k := 0; k < len(geometry.(*jsparser.JSON).ArrayVals[i].(*jsparser.JSON).ArrayVals[j].(*jsparser.JSON).ArrayVals); k++ {
				k1 := geometry.(*jsparser.JSON).ArrayVals[i].(*jsparser.JSON).ArrayVals[j].(*jsparser.JSON).ArrayVals[k].(*jsparser.JSON).ArrayVals[0].(string)
				k2 := geometry.(*jsparser.JSON).ArrayVals[i].(*jsparser.JSON).ArrayVals[j].(*jsparser.JSON).ArrayVals[k].(*jsparser.JSON).ArrayVals[1].(string)

				lat, err := strconv.ParseFloat(k1, 64)
				if err != nil {
					log.Event(ctx, "failed to caste interface to float64", log.ERROR, log.Error(err), log.Data{"lat": k1})
					return g, err
				}

				lon, err := strconv.ParseFloat(k2, 64)
				if err != nil {
					log.Event(ctx, "failed to caste interface to float64", log.ERROR, log.Error(err), log.Data{"lon": k2})
					return g, err
				}

				coordinates = append(coordinates, []float64{lat, lon})
			}
			multiCoordinates = append(multiCoordinates, coordinates)
		}
		g = append(g, multiCoordinates)
	}

	return g, nil
}
