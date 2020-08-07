package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	es "github.com/ONSdigital/dp-census-alpha-search-api/internal/elasticsearch"
	"github.com/ONSdigital/dp-census-alpha-search-api/models"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
)

const (
	defaultDatasetIndex        = "datasets"
	defaultElasticsearchAPIURL = "http://localhost:9200"
	defaultFilename            = "cmd-datasets.csv"
	defaultDimensionFile       = "../data/dimensions.json"
	defaultTaxonomyFile        = "../data/taxonomy.json"
	mappingsFile               = "dataset-mappings.json"
)

var (
	datasetIndex, elasticsearchAPIURL, filename, dimensionsFilename, taxonomyFilename string
	taxonomy                                                                          models.Taxonomy
	topicLevels                                                                       = make(map[string]TopicLevels)
)

// Dataset represents the data stored against a resource in elasticsearch index
type Dataset struct {
	Alias       string      `json:"alias"`
	Description string      `json:"description"`
	Dimensions  []Dimension `json:"dimensions"`
	GeoLocation GeoLocation `json:"location"`
	Links       Links       `json:"links"`
	Title       string      `json:"title"`
	Topic1      string      `json:"topic1,omitempty"`
	Topic2      string      `json:"topic2,omitempty"`
	Topic3      string      `json:"topic3,omitempty"`
}

// Dimension is an object representing a single dimension
type Dimension struct {
	Label string `json:"label"`
	Name  string `json:"name"`
}

// GeoLocation is an object that describes the geometry of the area a dataset relates to
type GeoLocation struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"`
}

// Links represents a set of links related to the dataset
type Links struct {
	Self Self `json:"self"`
}

// Self represents a link to a unique dataset resourse
type Self struct {
	HRef string `json:"href"`
	ID   string `json:"id"`
}

// TopicLevels represent the levels within the topic hierarchy (aka taxonomy)
type TopicLevels struct {
	TopicLevel1 string
	TopicLevel2 string
	TopicLevel3 string
}

func main() {
	ctx := context.Background()
	flag.StringVar(&datasetIndex, "dataset-index", defaultDatasetIndex, "the elasticsearch index that datasets will be uploaded to")
	flag.StringVar(&elasticsearchAPIURL, "elasticsearch-url", defaultElasticsearchAPIURL, "the elasticsearch url")
	flag.StringVar(&filename, "filename", defaultFilename, "the csv filename that contains data to upload to elasticsearch")
	flag.StringVar(&dimensionsFilename, "dimensions-filename", defaultDimensionFile, "the file locataion and name that contains a list of dataset dimensions")
	flag.StringVar(&taxonomyFilename, "taxonomy-filename", defaultTaxonomyFile, "the file locataion and name that contains the taxonomy hierarchy")
	flag.Parse()

	if datasetIndex == "" {
		datasetIndex = defaultDatasetIndex
	}

	if elasticsearchAPIURL == "" {
		elasticsearchAPIURL = defaultElasticsearchAPIURL
	}

	if filename == "" {
		filename = defaultFilename
	}

	if dimensionsFilename == "" {
		dimensionsFilename = defaultDimensionFile
	}

	if taxonomyFilename == "" {
		taxonomyFilename = defaultTaxonomyFile
	}

	log.Event(ctx, "script variables", log.INFO, log.Data{"dataset_index": datasetIndex, "elasticsearch_api_url": elasticsearchAPIURL, "filename": filename, "dimensions-file": dimensionsFilename, "taxonomy-file": taxonomyFilename})

	cli := dphttp.NewClient()
	esAPI := es.NewElasticSearchAPI(cli, elasticsearchAPIURL)

	// Read in Taxonomy into memory
	taxonomyFile, err := ioutil.ReadFile(taxonomyFilename)
	if err != nil {
		log.Event(ctx, "failed to read taxonomy file", log.ERROR, log.Error(err), log.Data{"taxonomy_filename": taxonomyFilename})
		os.Exit(1)
	}

	if err = json.Unmarshal([]byte(taxonomyFile), &taxonomy); err != nil {
		log.Event(ctx, "unable to unmarshal taxonomy into struct", log.ERROR, log.Error(err), log.Data{"taxonomy_filename": taxonomyFilename})
		os.Exit(1)
	}

	// Invert taxonomy so each topic has a list of parent topics and store in map
	for _, topic := range taxonomy.Topics {
		topicLevels[topic.FormattedTitle] = TopicLevels{
			TopicLevel1: topic.FormattedTitle,
		}

		for _, topic2 := range topic.ChildTopics {
			topicLevels[topic2.FormattedTitle] = TopicLevels{
				TopicLevel1: topic.FormattedTitle,
				TopicLevel2: topic2.FormattedTitle,
			}

			for _, topic3 := range topic2.ChildTopics {
				topicLevels[topic3.FormattedTitle] = TopicLevels{
					TopicLevel1: topic.FormattedTitle,
					TopicLevel2: topic2.FormattedTitle,
					TopicLevel3: topic3.FormattedTitle,
				}
			}
		}
	}

	// delete existing elasticsearch index if already exists
	status, err := esAPI.DeleteSearchIndex(ctx, datasetIndex)
	if err != nil {
		if status != http.StatusNotFound {
			log.Event(ctx, "failed to delete index", log.ERROR, log.Error(err), log.Data{"status": status})
			os.Exit(1)
		}

		log.Event(ctx, "failed to delete index as index cannot be found, continuing", log.WARN, log.Error(err), log.Data{"status": status})
	}

	// create elasticsearch index with settings/mapping
	status, err = esAPI.CreateSearchIndex(ctx, datasetIndex, mappingsFile)
	if err != nil {
		log.Event(ctx, "failed to create index", log.ERROR, log.Error(err), log.Data{"status": status})
		os.Exit(1)
	}

	// upload geo locations from data/datasets-test.csv and manipulate data into models.GeoDoc
	if err = uploadDocs(ctx, esAPI, datasetIndex, filename); err != nil {
		log.Event(ctx, "failed to retrieve dataset docs", log.ERROR, log.Error(err))
		os.Exit(1)
	}

	log.Event(ctx, "successfully loaded in dataset docs", log.INFO)
}

func uploadDocs(ctx context.Context, esAPI *es.API, indexName, filename string) error {
	csvfile, err := os.Open(filename)
	if err != nil {
		log.Event(ctx, "failed to open the csv file", log.ERROR, log.Error(err))
		return err
	}

	// Parse the file
	r := csv.NewReader(csvfile)

	headerRow, err := r.Read()
	if err != nil {
		log.Event(ctx, "failed to read header row", log.ERROR, log.Error(err))
		return err
	}

	headerIndex, err := check(headerRow)
	if err != nil {
		log.Event(ctx, "header row missing expected headers", log.ERROR, log.Error(err))
		return err
	}

	count := 0

	dimensionMap := make(map[string]string)
	// Iterate through the records
	for {
		count++
		// Read each record from csv
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Event(ctx, "failed to read row", log.ERROR, log.Error(err))
		}

		datasetDoc := &Dataset{
			Alias:       row[headerIndex["alias"]],
			Description: row[headerIndex["description"]],
			GeoLocation: addUKBoundary(),
			Links: Links{
				Self: Self{
					HRef: row[headerIndex["ons-link"]],
				},
			},
			Title: row[headerIndex["title"]],
		}

		dimensionNames := row[headerIndex["dimension-names"]]
		dimensionLabels := row[headerIndex["dimension-labels"]]

		dn := strings.SplitN(dimensionNames, ":", -1)
		dl := strings.SplitN(dimensionLabels, ":", -1)

		if len(dn) != len(dl) {
			log.Event(ctx, "dimensions labels and names do not match up, unequal length", log.WARN, log.Data{"row": count + 1, "dataset": datasetDoc.Title})
			continue
		}

		var dimensions []Dimension
		for i := 0; i < len(dn); i++ {
			dimensions = append(dimensions, Dimension{
				Label: dl[i],
				Name:  dn[i],
			})
			dimensionMap[dn[i]] = dl[i]
		}

		datasetDoc.Dimensions = dimensions

		topic := row[headerIndex["topic"]]
		if topic != "" {
			log.Event(ctx, "topic?", log.Data{"topic": topic})
			// find topic hierarchy - using taxonomy map
			taxonomy := topicLevels[topic]

			datasetDoc.Topic1 = taxonomy.TopicLevel1
			datasetDoc.Topic2 = taxonomy.TopicLevel2
			datasetDoc.Topic3 = taxonomy.TopicLevel3
			log.Event(ctx, "dataset?", log.Data{"datasets": datasetDoc})
		}

		bytes, err := json.Marshal(datasetDoc)
		if err != nil {
			log.Event(ctx, "failed to marshal dataset document to bytes", log.ERROR, log.Error(err), log.Data{"count": count})
			return err
		}

		// Add document to elasticsearch index
		if _, err = esAPI.AddDocument(ctx, indexName, bytes); err != nil {
			log.Event(ctx, "failed to upload dataset document to index", log.ERROR, log.Error(err), log.Data{"count": count})
			return err
		}
	}

	log.Event(ctx, "dimensions?", log.Data{"dimensions": dimensionMap})

	dimensionList := createDimensionList(ctx, dimensionMap)
	// Store dimensions to a file
	file, err := json.MarshalIndent(dimensionList, "", "  ")
	if err != nil {
		log.Event(ctx, "failed to marshal taxonomy with indentation", log.FATAL, log.Error(err))
		os.Exit(1)
	}

	if err = ioutil.WriteFile(dimensionsFilename, file, 0644); err != nil {
		log.Event(ctx, "failed to write to file", log.FATAL, log.Error(err))
		os.Exit(1)
	}

	return nil
}

// DimensionsDoc represents a list of dimensions
type DimensionsDoc struct {
	Dimensions []DimensionObject `json:"items"`
	TotalCount int               `json:"total_count"`
}

// DimensionObject represents the structure of a dimension
type DimensionObject struct {
	Label string `json:"label,omitempty"`
	Name  string `json:"name,omitempty"`
}

func createDimensionList(ctx context.Context, dimensionMap map[string]string) DimensionsDoc {
	var dimensions []DimensionObject
	for k, v := range dimensionMap {
		dimensions = append(dimensions, DimensionObject{
			Label: v,
			Name:  k,
		})
	}

	return DimensionsDoc{
		Dimensions: dimensions,
		TotalCount: len(dimensions),
	}
}

var validHeaders = map[string]bool{
	"alias":            true,
	"description":      true,
	"dimension-names":  true,
	"dimension-labels": true,
	"ons-link":         true,
	"title":            true,
	"topic":            true,
}

func check(headerRow []string) (map[string]int, error) {
	hasHeaders := map[string]bool{
		"alias":            false,
		"description":      false,
		"dimension-names":  false,
		"dimension-labels": false,
		"ons-link":         false,
		"title":            false,
		"topic":            false,
	}

	if len(headerRow) < 1 {
		return nil, errors.New("empty header row")
	}

	var indexHeader = make(map[string]int)
	for i, header := range headerRow {
		if !validHeaders[header] {
			return nil, errors.New("invalid header: " + header)
		}

		hasHeaders[header] = true
		indexHeader[header] = i
	}

	var hasHeadersMissing bool
	var missingHeaders string
	for key, value := range hasHeaders {
		if !value {
			hasHeadersMissing = true
			missingHeaders = missingHeaders + key + " "
		}
	}

	if hasHeadersMissing {
		return nil, errors.New("missing header in row: " + missingHeaders)
	}

	return indexHeader, nil
}

func addUKBoundary() GeoLocation {
	ukBoundary := [][][]float64{
		{
			{2.629085, 50.576346},
			{-7.692067, 49.433325},
			{-9.782838, 60.860061},
			{0.891099, 60.991247},
			{2.629085, 50.576346},
		},
	}

	return GeoLocation{
		Type:        "polygon",
		Coordinates: ukBoundary,
	}
}
