package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/ONSdigital/log.go/log"
	"github.com/globalsign/mgo"
	"gopkg.in/mgo.v2/bson"
)

const (
	collection = "datasets"
	database   = "datasets"

	onsSite = "https://www.ons.gov.uk"

	defaultBindAddr  = "localhost:27017"
	defaultFilename  = "cmd-datasets.csv"
	missingFileError = "no such file or directory"
)

var bindAddr, filename string

func main() {
	ctx := context.Background()
	flag.StringVar(&bindAddr, "mongodb-bind-addr", defaultBindAddr, "the address including authorisation if needed to bind to mongo database")
	flag.StringVar(&filename, "filename", defaultFilename, "the name and path of the file location to create file")
	flag.Parse()

	if bindAddr == "" {
		bindAddr = defaultBindAddr
	}

	if filename == "" {
		filename = defaultFilename
	}

	log.Event(ctx, "script variables", log.INFO, log.Data{"mongodb_bind_addr": bindAddr})

	mongo := Mongo{
		Collection: collection,
		Database:   database,
		URI:        bindAddr,
	}

	session, err := mongo.Init()
	if err != nil {
		log.Event(ctx, "unable to connect to mongo database", log.ERROR, log.Error(err), log.Data{"mongodb-bind-addr": bindAddr})
		os.Exit(1)
	}

	mongo.Session = session

	datasets, err := mongo.getDatasets(ctx)
	if err != nil {
		log.Event(ctx, "unable to retrieve list of datasets from mongo db", log.ERROR, log.Error(err))
		os.Exit(1)
	}

	// Remove existing file
	if err := os.Remove(filename); err != nil {
		if strings.Contains(err.Error(), missingFileError) {
			log.Event(ctx, "unable to remove existing file as it doesn't exist, continue processing", log.INFO)
		} else {
			log.Event(ctx, "unable to remove existing file before adding recreating it", log.ERROR, log.Error(err))
			os.Exit(1)
		}
	}

	headerLine := "title,alias,description,topic,ons-link,dimension-names,dimension-labels\n"
	writeToFile(headerLine)

	for i, dataset := range datasets {
		var url, edition string

		// Check data exists on ons website as we are building links to that website
		if dataset.Current != nil {
			if dataset.Current.Links != nil {
				if dataset.Current.Links.LatestVersion != nil {
					if dataset.Current.Links.LatestVersion.HRef != "" {
						href := dataset.Current.Links.LatestVersion.HRef
						splitHReF := strings.SplitAfter(href, "datasets")
						endOfPath := strings.SplitAfter(splitHReF[1], "versions")
						fullPath := "/datasets" + endOfPath[0]
						url = onsSite + fullPath

						// Get edition
						getEdition := strings.SplitAfter(endOfPath[0], "editions/")
						edition = strings.TrimSuffix(getEdition[1], "/versions")
						log.Event(context.Background(), "found dataset", log.INFO, log.Data{"index": i})

						resp, err := http.Get(url)
						if err != nil {
							log.Event(ctx, "failed to call url", log.ERROR, log.Error(err), log.Data{"url": url})
							continue
						}

						if resp.StatusCode != http.StatusOK {
							log.Event(ctx, "request to ons website failed", log.ERROR, log.Error(errors.New("status code unexpected")), log.Data{"url": url})
							continue
						}
					}
				}
			}
		}

		// continue if url is still empty
		if url == "" {
			continue
		}

		var topic string
		if dataset.Current.QMI != nil && dataset.Current.QMI.HRef != "" {

			// Remove host from qmi, leaving the path to methodology only
			qmiPath := strings.SplitAfter(dataset.Current.QMI.HRef, "https://www.ons.gov.uk/")

			// Check length to determine if qmi is an ons.gov.uk url
			if len(qmiPath) > 1 {
				// Split path in two leaving the dataset name separate from taxonomy of topics
				qmiArray1 := strings.SplitAfter(qmiPath[1], "methodologies")
				qmiArray2 := strings.SplitAfter(qmiArray1[0], "qmis")

				// Create a list of topics
				list := strings.SplitAfter(qmiArray2[0], "/")
				// Find lowest level topic in list, this will be the second from last value due to "methodologies" keyword being the last value in list
				topic = list[len(list)-2]

				// Remove trailing whitespace off topic
				topic = strings.TrimRight(topic, "/")
			}
		}

		// Retrieve dataset dimensions by finding latest published instance of dataset
		latestVersion := dataset.Current.Links.LatestVersion.ID

		// Get latest dataset instance
		instance, err := mongo.getDatasetInstance(ctx, dataset.ID, edition, latestVersion)
		if err != nil {
			log.Event(ctx, "unable to retrieve dataset instance from mongo db", log.ERROR, log.Error(err))
			os.Exit(1)
		}

		var dimensionNames, dimensionLabels string
		for i, dimension := range instance.Dimensions {
			label := dimension.Label
			if label == "" {
				label = dimension.Name
			}

			if i == 0 {
				dimensionNames = dimensionNames + dimension.Name
				dimensionLabels = dimensionLabels + label
			} else {
				dimensionNames = dimensionNames + ":" + dimension.Name
				dimensionLabels = dimensionLabels + ":" + label
			}
		}

		row := `"` + dataset.Current.Title + `",` + dataset.ID + `,"` + dataset.Current.Description + `","` + topic + `",` + url + `,"` + dimensionNames + `","` + dimensionLabels + `"` + "\n"

		writeToFile(row)
	}
}

// Mongo represents a simplistic MongoDB configuration.
type Mongo struct {
	Collection string
	Database   string
	Session    *mgo.Session
	URI        string
}

// Init creates a new mgo.Session with a strong consistency and a write mode of "majortiy".
func (m *Mongo) Init() (session *mgo.Session, err error) {
	if session != nil {
		return nil, errors.New("session already exists")
	}

	if session, err = mgo.Dial(m.URI); err != nil {
		return nil, err
	}

	session.EnsureSafe(&mgo.Safe{WMode: "majority"})
	session.SetMode(mgo.Strong, true)

	return session, nil
}

// DatasetUpdate represents an evolving dataset with the current dataset and the updated dataset
type DatasetUpdate struct {
	ID      string   `bson:"_id,omitempty"         json:"id,omitempty"`
	Current *Dataset `bson:"current,omitempty"     json:"current,omitempty"`
}

// Dataset represents information related to a single dataset
type Dataset struct {
	Description string        `bson:"description,omitempty" json:"description,omitempty"`
	Links       *DatasetLinks `bson:"links,omitempty"       json:"links,omitempty"`
	QMI         *QMIObject    `bson:"qmi,omitempty"       json:"qmi,omitempty"`
	Title       string        `bson:"title,omitempty"       json:"title,omitempty"`
}

// DatasetLinks represents a list of specific links related to the dataset resource
type DatasetLinks struct {
	LatestVersion *LinkObject `bson:"latest_version,omitempty"  json:"latest_version,omitempty"`
}

// LinkObject represents a generic structure for all links
type LinkObject struct {
	HRef string `bson:"href,omitempty"  json:"href,omitempty"`
	ID   string `bson:"id,omitempty"  json:"id,omitempty"`
}

type QMIObject struct {
	HRef string `bson:"href,omitempty"  json:"href,omitempty"`
}

// Version represents information related to a single version for an edition of a dataset
type Version struct {
	Dimensions []Dimension `bson:"dimensions,omitempty"     json:"dimensions,omitempty"`
	Edition    string      `bson:"edition,omitempty"        json:"edition,omitempty"`
	ID         string      `bson:"id,omitempty"             json:"id,omitempty"`
	Version    int         `bson:"version,omitempty"        json:"version,omitempty"`
}

// Dimension represents an overview for a single dimension. This includes a link to the code list API
// which provides metadata about the dimension and all possible values.
type Dimension struct {
	Label string `bson:"label,omitempty"         json:"label,omitempty"`
	Name  string `bson:"name,omitempty"          json:"name,omitempty"`
}

// getDatasets retrieves all dataset documents
func (m *Mongo) getDatasets(ctx context.Context) ([]DatasetUpdate, error) {
	s := m.Session.Copy()
	defer s.Close()

	iter := s.DB(m.Database).C(m.Collection).Find(nil).Iter()
	defer func() {
		err := iter.Close()
		if err != nil {
			log.Event(ctx, "error closing iterator", log.ERROR, log.Error(err))
		}
	}()

	datasets := []DatasetUpdate{}
	if err := iter.All(&datasets); err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.New("dataset not found")
		}
		return nil, err
	}

	return datasets, nil
}

// getDatasetsInstance retrieves a single instance of a dataset
func (m *Mongo) getDatasetInstance(ctx context.Context, datasetID, edition, version string) (*Version, error) {
	s := m.Session.Copy()
	defer s.Close()

	versionNumber, err := strconv.Atoi(version)
	if err != nil {
		return nil, err
	}

	selector := bson.M{
		"links.dataset.id": datasetID,
		"edition":          edition,
		"version":          versionNumber,
	}

	var datasetVersion Version
	err = s.DB(m.Database).C("instances").Find(selector).One(&datasetVersion)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.New("instance not found")
		}
		return nil, err
	}
	return &datasetVersion, nil
}

// writeToFile add csv row to file
func writeToFile(line string) error {
	connection, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	_, err = connection.WriteString(line)
	if err != nil {
		return err
	}

	return nil
}
