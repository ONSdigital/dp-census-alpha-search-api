package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	errs "github.com/ONSdigital/dp-census-alpha-search-api/apierrors"
	"github.com/ONSdigital/dp-census-alpha-search-api/models"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

const (
	defaultRelation = "intersects"

	relationError = "invalid relation value"
)

func (api *SearchAPI) getAreaProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	setAccessControl(w, http.MethodGet)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	logData := log.Data{"id": id}

	log.Event(ctx, "getAreaProfile endpoint: incoming request", log.INFO, logData)

	query := models.AreaProfileQuery{
		Query: models.Query{
			Term: map[string]string{
				"id": id,
			},
		},
	}

	response, status, err := api.elasticsearch.GetAreaProfile(ctx, api.areaProfileIndex, query)
	if err != nil {
		logData["elasticsearch_status"] = status
		log.Event(ctx, "getAreaProfile endpoint: failed to get all datat type search results", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	b, err := json.Marshal(response)
	if err != nil {
		log.Event(ctx, "getAreaProfile endpoint: failed to marshal search resource into bytes", log.ERROR, log.Error(err), logData)
		setErrorCode(w, errs.ErrInternalServer)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.Event(ctx, "getAreaProfile endpoint: error writing response", log.ERROR, log.Error(err), logData)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Event(ctx, "getAreaProfile endpoint: successfully searched index", log.INFO, logData)
}

func (api *SearchAPI) getAreaProfileSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	setAccessControl(w, http.MethodGet)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	q := r.FormValue("q")
	requestedLimit := r.FormValue("limit")
	requestedOffset := r.FormValue("offset")
	dimensions := r.FormValue("dimensions")
	topics := r.FormValue("topics")

	requestedRelation := r.FormValue("relation")

	logData := log.Data{
		"id":                 id,
		"query_term":         q,
		"requested_limit":    requestedLimit,
		"requested_offset":   requestedOffset,
		"dimensions":         dimensions,
		"topics":             topics,
		"requested_relation": requestedRelation,
	}

	log.Event(ctx, "getAreaProfileSearch endpoint: incoming request", log.INFO, logData)

	// Remove leading and/or trailing whitespace
	term := strings.TrimSpace(q)

	if term == "" {
		log.Event(ctx, "getAreaProfileSearch endpoint: query parameter \"q\" empty", log.ERROR, log.Error(errs.ErrEmptySearchTerm), logData)
		setErrorCode(w, errs.ErrEmptySearchTerm)
		return
	}

	var err error

	limit := defaultLimit
	if requestedLimit != "" {
		limit, err = strconv.Atoi(requestedLimit)
		if err != nil {
			log.Event(ctx, "getAreaProfileSearch endpoint: request limit parameter error", log.ERROR, log.Error(err), logData)
			setErrorCode(w, errs.ErrParsingQueryParameters)
			return
		}
	}

	offset := defaultOffset
	if requestedOffset != "" {
		offset, err = strconv.Atoi(requestedOffset)
		if err != nil {
			log.Event(ctx, "getAreaProfileSearch endpoint: request offset parameter error", log.ERROR, log.Error(err), logData)
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
		log.Event(ctx, "getAreaProfileSearch endpoint: validate pagination", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	logData["limit"] = page.Limit
	logData["offset"] = page.Offset

	relation, err := models.ValidateGeoShapeRelation(ctx, defaultRelation, requestedRelation)
	if err != nil {
		log.Event(ctx, "getAreaProfileSearch endpoint: validate geo shape relation", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	dimensionFilters, err := models.ValidateDimensions(dimensions)
	if err != nil {
		log.Event(ctx, "getAreaProfileSearch endpoint: validate dimensions filter", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	topicFilters, err := models.ValidateTopics(topics)
	if err != nil {
		log.Event(ctx, "getAreaProfileSearch endpoint: validate topics filter", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	query := models.AreaProfileQuery{
		Query: models.Query{
			Term: map[string]string{
				"id": id,
			},
		},
	}

	areaProfile, status, err := api.elasticsearch.GetAreaProfile(ctx, api.areaProfileIndex, query)
	if err != nil {
		logData["elasticsearch_status"] = status
		log.Event(ctx, "getAreaProfileSearch endpoint: failed to get all datat type search results", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	// build dataset search query
	datasetQuery := buildAreaProfileDatasetSearchQuery(&areaProfile.Location, term, dimensionFilters, topicFilters, relation, limit, offset)

	response, status, err := api.elasticsearch.QuerySearchIndex(ctx, api.datasetIndex, datasetQuery, limit, offset)
	if err != nil {
		logData["elasticsearch_status"] = status
		log.Event(ctx, "getAreaProfileSearch endpoint: failed to get dataset search results", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	datasets := models.DatasetSearchResults{
		Limit:      limit,
		Offset:     offset,
		TotalCount: response.Hits.Total,
		Items:      []models.SearchResult{},
	}

	for _, result := range response.Hits.HitList {

		doc := result.Source
		doc.Matches = models.NewMatches{
			Alias:          result.Matches.Alias,
			Description:    result.Matches.Description,
			DimensionLabel: result.Matches.DimensionLabel,
			DimensionName:  result.Matches.DimensionName,
			Title:          result.Matches.Title,
			Topic1:         result.Matches.Topic1,
			Topic2:         result.Matches.Topic2,
			Topic3:         result.Matches.Topic3,
		}

		datasets.Items = append(datasets.Items, doc)
	}

	datasets.Count = len(datasets.Items)

	b, err := json.Marshal(datasets)
	if err != nil {
		log.Event(ctx, "getAreaProfileSearch endpoint: failed to marshal search resource into bytes", log.ERROR, log.Error(err), logData)
		setErrorCode(w, errs.ErrInternalServer)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.Event(ctx, "getAreaProfileSearch endpoint: error writing response", log.ERROR, log.Error(err), logData)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Event(ctx, "getAreaProfileSearch endpoint: successfully searched index", log.INFO, logData)
}

func buildAreaProfileDatasetSearchQuery(geoLocation *models.GeoLocation, term string, dimensionFilters []models.Filter, topicFilters []models.Filter, relation string, limit, offset int) *models.Body {
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
			PreTags:  []string{"<b>"},
			PostTags: []string{"</b>"},
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
				Filter: []models.Filter{
					{
						Shape: &models.GeoShape{
							Location: models.GeoLocationObj{
								Shape:    *geoLocation,
								Relation: relation,
							},
						},
					},
				},
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
