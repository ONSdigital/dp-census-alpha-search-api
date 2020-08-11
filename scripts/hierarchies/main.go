package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/ONSdigital/log.go/log"
)

const hierarchyFilename = "../data/hierarchy.json"

var geoHierarchyMap = map[string]string{
	"Lower Layer Super Output Areas":  "lowerlayersuperoutputareas",
	"Middle Layer Super Output Areas": "middlelayersuperoutputareas",
	"Output Areas":                    "outputareas",
	"Major Towns and Cities":          "majortownsandcities",
	"Countries":                       "countries",
}

func main() {
	ctx := context.Background()

	hierarchyList := createGeoHierarchyList(ctx, geoHierarchyMap)
	// Store hierarchies to a file
	file, err := json.MarshalIndent(hierarchyList, "", "  ")
	if err != nil {
		log.Event(ctx, "failed to marshal taxonomy with indentation", log.FATAL, log.Error(err))
		os.Exit(1)
	}

	if err = ioutil.WriteFile(hierarchyFilename, file, 0644); err != nil {
		log.Event(ctx, "failed to write to file", log.FATAL, log.Error(err))
		os.Exit(1)
	}

	log.Event(ctx, "successfully created hierarchies.json", log.INFO)
}

// GeoHierarchiesDoc represents a list of geography hierarchies
type GeoHierarchiesDoc struct {
	Items      []GeographyObject `json:"items"`
	TotalCount int               `json:"total_count"`
}

// GeographyObject represents the structure of a dimension
type GeographyObject struct {
	Hierarchy           string `json:"hierarchy"`
	FilterableHierarchy string `json:"filterable_hierarchy"`
}

func createGeoHierarchyList(ctx context.Context, geoHierarchyMap map[string]string) GeoHierarchiesDoc {
	var hierarchies []GeographyObject
	for k, v := range geoHierarchyMap {
		hierarchies = append(hierarchies, GeographyObject{
			Hierarchy:           k,
			FilterableHierarchy: v,
		})
	}

	return GeoHierarchiesDoc{
		Items:      hierarchies,
		TotalCount: len(hierarchies),
	}
}
