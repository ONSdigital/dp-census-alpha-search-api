package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/ONSdigital/dp-census-alpha-search-api/models"
	"github.com/ONSdigital/log.go/log"
)

const (
	onsWebsite = "https://www.ons.gov.uk"

	taxonomyLandingPage = "taxonomy_landing_page"
)

var (
	filename string

	defaultFilename = "../data/taxonomy.json"
)

var topicList = make(map[string]int)

func main() {
	flag.StringVar(&filename, "filename", defaultFilename, "the name and path of the file location to create file")
	flag.Parse()

	if filename == "" {
		filename = defaultFilename
	}

	ctx := context.Background()

	log.Event(ctx, "script variables", log.INFO, log.Data{"ons_website": onsWebsite})

	// Call ons website for top level taxonomy
	taxonomy, err := callONSWebite(ctx, onsWebsite)
	if err != nil {
		log.Event(ctx, "failed to retrieve taxonomy data from ons website", log.FATAL, log.Error(err))
		os.Exit(1)
	}

	// Store doc to file
	file, err := json.MarshalIndent(taxonomy, "", "  ")
	if err != nil {
		log.Event(ctx, "failed to marshal taxonomy with indentation", log.FATAL, log.Error(err))
		os.Exit(1)
	}

	if err = ioutil.WriteFile(filename, file, 0644); err != nil {
		log.Event(ctx, "failed to write to file", log.FATAL, log.Error(err))
		os.Exit(1)
	}
}

func callONSWebite(ctx context.Context, url string) (*models.Taxonomy, error) {

	logData := log.Data{"url": url}

	resp, err := http.Get(url + "/data")
	if err != nil {
		log.Event(ctx, "request to ons website failed", log.ERROR, log.Error(err), logData)
		return nil, err
	}
	defer resp.Body.Close()

	topTaxonomy, err := CreateFirstLevelTaxonomy(ctx, resp.Body)
	if err != nil {
		return nil, err
	}

	// handle duplicate top level topics
	topLevelTopics := make(map[string]bool)

	var topics []models.Topic
	for _, section := range topTaxonomy.Sections {

		if !topLevelTopics[section.Theme.URI] {
			topLevelTopics[section.Theme.URI] = true
			topic, err := GetTopics(ctx, url, section.Theme.URI, 1)
			if err != nil {
				return nil, err
			}

			if topic == nil {
				continue
			}
			topics = append(topics, *topic)
		}
	}

	taxonomy := models.Taxonomy{
		Topics: topics,
	}

	log.Event(ctx, "data", log.INFO, log.Data{"top_list": topicList})

	return &taxonomy, nil
}

type TopTaxonomy struct {
	Sections    []Section   `json:"sections"`
	URI         string      `json:"uri"`
	Description Description `json:"description"`
}

type Section struct {
	Theme Theme `json:"theme"`
}

type Theme struct {
	URI string `json:"uri"`
}

func CreateFirstLevelTaxonomy(ctx context.Context, reader io.Reader) (*TopTaxonomy, error) {

	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.New("unable to read bytes")
	}

	var topTaxonomy TopTaxonomy

	if err = json.Unmarshal(b, &topTaxonomy); err != nil {
		return nil, errors.New("failed to parse json body")
	}

	return &topTaxonomy, nil
}

type ChildTaxonomy struct {
	Sections    []ChildSection `json:"sections"`
	Type        string         `json:"type"`
	URI         string         `json:"uri"`
	Description Description    `json:"description"`
}

type Description struct {
	Title string `json:"title"`
}

type ChildSection struct {
	URI string `json:"uri"`
}

func GetTopics(ctx context.Context, url, parentTopic string, level int) (*models.Topic, error) {
	extendedURL := onsWebsite + parentTopic + "/data"
	logData := log.Data{"url": extendedURL}

	resp, err := http.Get(extendedURL)
	if err != nil {
		log.Event(ctx, "GetTopics: unsuccessful request", log.ERROR, log.Error(err), logData)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		log.Event(ctx, "got a not found page", log.WARN, logData)
		return nil, nil
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Event(ctx, "GetTopics: failed to read response body", log.ERROR, log.Error(err), logData)
		return nil, errors.New("unable to read bytes")
	}

	var childTaxonomy ChildTaxonomy
	if err = json.Unmarshal(b, &childTaxonomy); err != nil {
		logData["bytes"] = string(b)
		log.Event(ctx, "GetTopics: unable to marshal response data into child taxonomy", log.ERROR, log.Error(err), logData)
		return nil, errors.New("failed to parse json body")
	}

	var topics []models.Topic
	if childTaxonomy.Type == taxonomyLandingPage {
		for _, section := range childTaxonomy.Sections {
			topic, err := GetTopics(ctx, extendedURL, section.URI, level+1)
			if err != nil {
				log.Event(ctx, "GetTopics: request to ons website failed", log.ERROR, log.Error(err), logData)
				return nil, err
			}

			if topic == nil {
				continue
			}

			topics = append(topics, *topic)
		}
	}

	title := strings.SplitAfter(parentTopic, "/")
	formattedTitle := title[len(title)-1]

	topic := models.Topic{
		Title:          childTaxonomy.Description.Title,
		FormattedTitle: formattedTitle,
		ChildTopics:    topics,
	}

	topicList[formattedTitle] = level

	return &topic, nil
}
