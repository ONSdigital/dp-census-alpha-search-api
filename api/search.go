package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	errs "github.com/ONSdigital/dp-census-alpha-search-api/apierrors"
	"github.com/ONSdigital/dp-census-alpha-search-api/models"
	"github.com/ONSdigital/log.go/log"
)

const (
	defaultLimit  = 50
	defaultOffset = 0

	internalError         = "internal server error"
	exceedsDefaultMaximum = "the maximum offset has been reached, the offset cannot be more than"
	topicFilterError      = "invalid list of topics to filter by"
)

func (api *SearchAPI) getData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	setAccessControl(w, http.MethodGet)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var err error

	q := r.FormValue("q")
	requestedLimit := r.FormValue("limit")
	requestedOffset := r.FormValue("offset")
	dimensions := r.FormValue("dimensions")
	topics := r.FormValue("topics")

	logData := log.Data{
		"query_term":       q,
		"requested_limit":  requestedLimit,
		"requested_offset": requestedOffset,
		"topics":           topics,
		"dimensions":       dimensions,
	}

	log.Event(ctx, "getData endpoint: incoming request", log.INFO, logData)

	// Remove leading and/or trailing whitespace
	term := strings.TrimSpace(q)

	if term == "" {
		log.Event(ctx, "getData endpoint: query parameter \"q\" empty", log.ERROR, log.Error(errs.ErrEmptySearchTerm), logData)
		setErrorCode(w, errs.ErrEmptySearchTerm)
		return
	}

	limit := defaultLimit
	if requestedLimit != "" {
		limit, err = strconv.Atoi(requestedLimit)
		if err != nil {
			log.Event(ctx, "getData endpoint: request limit parameter error", log.ERROR, log.Error(err), logData)
			setErrorCode(w, errs.ErrParsingQueryParameters)
			return
		}
	}

	offset := defaultOffset
	if requestedOffset != "" {
		offset, err = strconv.Atoi(requestedOffset)
		if err != nil {
			log.Event(ctx, "getData endpoint: request offset parameter error", log.ERROR, log.Error(err), logData)
			setErrorCode(w, errs.ErrParsingQueryParameters)
			return
		}
	}

	page := &models.PageVariables{
		DefaultMaxResults: api.defaultMaxResults,
		Limit:             limit,
		Offset:            offset,
	}

	if err = page.Validate(); err != nil {
		log.Event(ctx, "getData endpoint: validate pagination", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	logData["limit"] = page.Limit
	logData["offset"] = page.Offset

	dimensionFilters, err := models.ValidateDimensions(dimensions)
	if err != nil {
		log.Event(ctx, "getData endpoint: validate filter by topics", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	topicFilters, err := models.ValidateTopics(topics)
	if err != nil {
		log.Event(ctx, "getData endpoint: validate filter by topics", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	log.Event(ctx, "getData endpoint: just before querying search index", log.INFO, logData)

	// build dataset search query
	query := buildSearchQuery(term, dimensionFilters, topicFilters, limit, offset)

	response, status, err := api.elasticsearch.QueryDatasetSearch(ctx, api.datasetIndex, query, limit, offset)
	if err != nil {
		logData["elasticsearch_status"] = status
		log.Event(ctx, "getData endpoint: failed to get search results", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	searchResults := &models.SearchResults{
		Limit:      page.Limit,
		Offset:     page.Offset,
		TotalCount: response.Hits.Total,
		Items:      []models.SearchResult{},
	}

	for _, result := range response.Hits.HitList {

		doc := result.Source
		doc.Matches = result.Matches

		searchResults.Items = append(searchResults.Items, doc)
	}

	searchResults.Count = len(searchResults.Items)

	b, err := json.Marshal(searchResults)
	if err != nil {
		log.Event(ctx, "getData endpoint: failed to marshal search resource into bytes", log.ERROR, log.Error(err), logData)
		setErrorCode(w, errs.ErrInternalServer)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.Event(ctx, "getData endpoint: error writing response", log.ERROR, log.Error(err), logData)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Event(ctx, "getData endpoint: successfully searched index", log.INFO, logData)
}

func setAccessControl(w http.ResponseWriter, method string) {
	w.Header().Set("Access-Control-Allow-Methods", method+",OPTIONS")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Max-Age", "86400")
	w.Header().Set("Content-Type", "application/json")
}

func setErrorCode(w http.ResponseWriter, err error) {

	switch {
	case errs.NotFoundMap[err]:
		http.Error(w, err.Error(), http.StatusNotFound)
	case errs.BadRequestMap[err]:
		http.Error(w, err.Error(), http.StatusBadRequest)
	case strings.Contains(err.Error(), exceedsDefaultMaximum):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case strings.Contains(err.Error(), topicFilterError):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, internalError, http.StatusInternalServerError)
	}
}

func buildSearchQuery(term string, dimensionFilters []models.Filter, topicFilters []models.Filter, limit, offset int) interface{} {
	var object models.Object
	highlight := make(map[string]models.Object)

	highlight["alias"] = object
	highlight["description"] = object
	highlight["title"] = object
	highlight["topic1"] = object
	highlight["topic2"] = object
	highlight["topic3"] = object
	highlight["dimensions.label"] = object
	highlight["dimensions.name"] = object

	alias := make(map[string]string)
	description := make(map[string]string)
	title := make(map[string]string)
	topic1 := make(map[string]string)
	topic2 := make(map[string]string)
	topic3 := make(map[string]string)
	dimensionLabels := make(map[string]string)
	dimensionNames := make(map[string]string)
	alias["alias"] = term
	description["description"] = term
	title["title"] = term
	topic1["topic1"] = term
	topic2["topic2"] = term
	topic3["topic3"] = term
	dimensionLabels["dimensions.label"] = term
	dimensionNames["dimensions.name"] = term

	aliasMatch := models.Match{
		Match: alias,
	}

	descriptionMatch := models.Match{
		Match: description,
	}

	titleMatch := models.Match{
		Match: title,
	}

	topic1Match := models.Match{
		Match: topic1,
	}

	topic2Match := models.Match{
		Match: topic2,
	}

	topic3Match := models.Match{
		Match: topic3,
	}

	scores := models.Scores{
		Score: models.Score{
			Order: "desc",
		},
	}

	listOfScores := []models.Scores{}
	listOfScores = append(listOfScores, scores)

	query := &models.Body{
		From: offset,
		Size: limit,
		Highlight: &models.Highlight{
			Fields:   highlight,
			PreTags:  []string{"<b><em>"},
			PostTags: []string{"</em></b>"},
		},
		Query: models.Query{
			Bool: &models.Bool{
				Should: []models.Match{
					aliasMatch,
					descriptionMatch,
					titleMatch,
					topic1Match,
					topic2Match,
					topic3Match,
					{
						Nested: &models.Nested{
							Path: "dimensions",
							Query: []models.NestedQuery{
								{
									Term: dimensionLabels,
								},
								{
									Term: dimensionNames,
								},
							},
						},
					},
				},
				MinimumShouldMatch: 1,
			},
		},
		Sort:      listOfScores,
		TotalHits: true,
	}

	if topicFilters != nil {
		query.Query.Bool.Filter = topicFilters
	}

	if dimensionFilters != nil && len(dimensionFilters) > 0 {
		query.Query.Bool.Filter = append(query.Query.Bool.Filter, dimensionFilters...)
	}

	return query
}
