package api

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	errs "github.com/ONSdigital/dp-census-alpha-search-api/apierrors"
	"github.com/ONSdigital/dp-census-alpha-search-api/helpers"
	"github.com/ONSdigital/dp-census-alpha-search-api/models"
	"github.com/ONSdigital/log.go/log"
)

const (
	defaultLimit    = 50
	defaultOffset   = 0
	defaultSegments = 20

	internalError         = "internal server error"
	exceedsDefaultMaximum = "the maximum offset has been reached, the offset cannot be more than"
	topicFilterError      = "invalid list of topics to filter by"
	hierarchyFilterError  = "invalid hierarchy to filter by"
)

var regPostcode = regexp.MustCompile(`(?i)[A-Z][A-HJ-Y]?\d[A-Z\d]? ?\d[A-Z]{2}|GIR ?0A{2}`)

func (api *SearchAPI) searchData(w http.ResponseWriter, r *http.Request) {
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
	hierarchies := r.FormValue("hierarchies")
	topics := r.FormValue("topics")

	requestedDistance := r.FormValue("distance")
	requestedRelation := r.FormValue("relation")

	logData := log.Data{
		"query_term":         q,
		"requested_limit":    requestedLimit,
		"requested_offset":   requestedOffset,
		"dimensions":         dimensions,
		"hierarchies":        hierarchies,
		"topics":             topics,
		"requested_distance": requestedDistance,
		"requested_relation": requestedRelation,
	}

	log.Event(ctx, "searchData endpoint: incoming request", log.INFO, logData)

	// Remove leading and/or trailing whitespace
	term := strings.TrimSpace(q)

	if term == "" {
		log.Event(ctx, "searchData endpoint: query parameter \"q\" empty", log.ERROR, log.Error(errs.ErrEmptySearchTerm), logData)
		setErrorCode(w, errs.ErrEmptySearchTerm)
		return
	}

	limit := defaultLimit
	if requestedLimit != "" {
		limit, err = strconv.Atoi(requestedLimit)
		if err != nil {
			log.Event(ctx, "searchData endpoint: request limit parameter error", log.ERROR, log.Error(err), logData)
			setErrorCode(w, errs.ErrParsingQueryParameters)
			return
		}
	}

	offset := defaultOffset
	if requestedOffset != "" {
		offset, err = strconv.Atoi(requestedOffset)
		if err != nil {
			log.Event(ctx, "searchData endpoint: request offset parameter error", log.ERROR, log.Error(err), logData)
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
		log.Event(ctx, "searchData endpoint: validate pagination", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	logData["limit"] = page.Limit
	logData["offset"] = page.Offset

	dimensionFilters, err := models.ValidateDimensions(dimensions)
	if err != nil {
		log.Event(ctx, "searchData endpoint: validate dimensions filter", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	hierarchyFilters, err := models.ValidateHierarchies(hierarchies)
	if err != nil {
		log.Event(ctx, "searchData endpoint: validate hierarchies filter", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	topicFilters, err := models.ValidateTopics(topics)
	if err != nil {
		log.Event(ctx, "searchData endpoint: validate topics filter", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	distObj, err := models.ValidateDistance(requestedDistance)
	if err != nil {
		log.Event(ctx, "searchData endpoint: validate query param, distance", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	log.Event(ctx, "searchData endpoint: just before querying search index", log.INFO, logData)

	var (
		allChan         = make(chan models.SearchResults, 1)
		datasetChan     = make(chan models.SearchResults, 1)
		areaProfileChan = make(chan models.SearchResults, 1)
		publicationChan = make(chan models.SearchResults, 1)
		reqError        error
	)

	// find all data
	go func() {
		geoLocation, err := api.getPostcodeLocation(ctx, term, distObj, logData)
		if err != nil {
			reqError = err
			allChan <- models.SearchResults{}
			return
		}

		// build all search query
		allDataQuery := api.buildAllSearchQuery(term, geoLocation, dimensionFilters, hierarchyFilters, topicFilters, limit, offset)

		response, status, err := api.elasticsearch.QuerySearchIndex(ctx, api.datasetIndex+","+api.areaProfileIndex, allDataQuery, limit, offset)
		if err != nil {
			logData["elasticsearch_status"] = status
			log.Event(ctx, "searchData endpoint: failed to get all datat type search results", log.ERROR, log.Error(err), logData)
			reqError = err
		}

		log.Event(ctx, "Got data", log.Data{"respose": response})

		allData := models.SearchResults{
			TotalCount: response.Hits.Total,
			Items:      []models.SearchResult{},
		}

		for _, result := range response.Hits.HitList {

			doc := result.Source
			doc.Matches = result.Matches

			allData.Items = append(allData.Items, doc)
		}

		allData.Count = len(allData.Items)

		allChan <- allData
	}()

	// find datasets
	go func() {
		// build dataset search query
		datasetQuery := buildDatasetSearchQuery(term, dimensionFilters, topicFilters, limit, offset)

		response, status, err := api.elasticsearch.QuerySearchIndex(ctx, api.datasetIndex, datasetQuery, limit, offset)
		if err != nil {
			logData["elasticsearch_status"] = status
			log.Event(ctx, "searchData endpoint: failed to get dataset search results", log.ERROR, log.Error(err), logData)
			reqError = err
			datasetChan <- models.SearchResults{}
			return
		}

		datasets := models.SearchResults{
			TotalCount: response.Hits.Total,
			Items:      []models.SearchResult{},
		}

		for _, result := range response.Hits.HitList {

			doc := result.Source
			doc.Matches = result.Matches

			datasets.Items = append(datasets.Items, doc)
		}

		datasets.Count = len(datasets.Items)

		datasetChan <- datasets
	}()

	// find area profiles
	go func() {
		geoLocation, err := api.getPostcodeLocation(ctx, term, distObj, logData)
		if err != nil {
			reqError = err
			areaProfileChan <- models.SearchResults{}
			return
		}

		areaProfileQuery := buildAreaSearchQuery(term, hierarchyFilters, geoLocation, limit, offset)

		response, status, err := api.elasticsearch.QuerySearchIndex(ctx, api.areaProfileIndex, areaProfileQuery, limit, offset)
		if err != nil {
			logData["elasticsearch_status"] = status
			log.Event(ctx, "searchData endpoint: failed to get area profile search results", log.ERROR, log.Error(err), logData)
			reqError = err
			areaProfileChan <- models.SearchResults{}
			return
		}

		areaProfiles := models.SearchResults{
			TotalCount: response.Hits.Total,
			Items:      []models.SearchResult{},
		}

		for _, result := range response.Hits.HitList {

			doc := result.Source
			doc.Matches = result.Matches

			areaProfiles.Items = append(areaProfiles.Items, doc)
		}

		areaProfiles.Count = len(areaProfiles.Items)

		areaProfileChan <- areaProfiles
	}()

	// find publications
	go func() {
		publicationChan <- models.SearchResults{
			Items: []models.SearchResult{},
		}
	}()

	// Wait till we have results from both search requests
	all := <-allChan
	datasets := <-datasetChan
	areaProfiles := <-areaProfileChan
	publications := <-publicationChan

	// handle any request errors from search queries
	if reqError != nil {
		setErrorCode(w, reqError)
		return
	}

	searchResults := models.AllSearchResults{
		Limit:  page.Limit,
		Offset: page.Offset,
		Counts: models.Counts{
			All:          all.TotalCount,
			Datasets:     datasets.TotalCount,
			AreaProfiles: areaProfiles.TotalCount,
			Publications: publications.TotalCount,
		},
		All:          all,
		Datasets:     datasets,
		AreaProfiles: areaProfiles,
		Publications: publications,
	}

	b, err := json.Marshal(searchResults)
	if err != nil {
		log.Event(ctx, "searchData endpoint: failed to marshal search resource into bytes", log.ERROR, log.Error(err), logData)
		setErrorCode(w, errs.ErrInternalServer)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.Event(ctx, "searchData endpoint: error writing response", log.ERROR, log.Error(err), logData)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Event(ctx, "searchData endpoint: successfully searched index", log.INFO, logData)
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
	case strings.Contains(err.Error(), hierarchyFilterError):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, internalError, http.StatusInternalServerError)
	}
}

func (api *SearchAPI) buildAllSearchQuery(term string, geoLocation *models.GeoLocation, dimensionFilters []models.Filter, hierarchyFilters []models.Filter, topicFilters []models.Filter, limit, offset int) *models.Body {
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

	highlight["code"] = object
	highlight["hierarchy"] = object
	highlight["name"] = object

	alias := make(map[string]string)
	description := make(map[string]string)
	title := make(map[string]string)
	topic1 := make(map[string]string)
	topic2 := make(map[string]string)
	topic3 := make(map[string]string)
	dimensionLabels := make(map[string]string)
	dimensionNames := make(map[string]string)

	code := make(map[string]string)
	hierarchy := make(map[string]string)
	name := make(map[string]string)

	alias["alias"] = term
	description["description"] = term
	title["title"] = term
	topic1["topic1"] = term
	topic2["topic2"] = term
	topic3["topic3"] = term
	dimensionLabels["dimensions.label"] = term
	dimensionNames["dimensions.name"] = term

	code["code"] = term
	hierarchy["hierarchy"] = term
	name["name"] = term

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

	codeMatch := models.Match{
		Match: code,
	}

	hierarchyMatch := models.Match{
		Match: hierarchy,
	}

	nameMatch := models.Match{
		Match: name,
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
		Query:     models.Query{},
		Sort:      listOfScores,
		TotalHits: true,
	}

	if geoLocation != nil {
		query.Query = models.Query{
			Bool: &models.Bool{
				Filter: []models.Filter{
					{
						Shape: &models.GeoShape{
							Location: models.GeoLocationObj{
								Shape:    *geoLocation,
								Relation: "intersects",
							},
						},
					},
					{
						Terms: map[string]interface{}{"_index": []string{api.areaProfileIndex}},
					},
				},
			},
		}

		if hierarchyFilters != nil && len(hierarchyFilters) > 0 {
			query.Query.Bool.Filter = append(query.Query.Bool.Filter, hierarchyFilters...)
		}
	} else {
		// datasetQuery := buildDatasetSearchQuery(term, dimensionFilters, topicFilters, limit, offset)
		// areaProfileQuery := buildAreaSearchQuery(term, hierarchyFilters, geoLocation, limit, offset)

		// query.Query.Bool = &models.Bool{
		// 	Should: []models.Match{
		// 		{
		// 			Bool: datasetQuery.Query.Bool,
		// 		},
		// 		{
		// 			Bool: areaProfileQuery.Query.Bool,
		// 		},
		// 	},
		// }

		query.Query = models.Query{
			Bool: &models.Bool{
				Should: []models.Match{
					{
						Bool: &models.Bool{
							Should: []models.Match{
								aliasMatch,
								descriptionMatch,
								titleMatch,
								topic1Match,
								topic2Match,
								topic3Match,
								codeMatch,
								hierarchyMatch,
								nameMatch,
							},
							MinimumShouldMatch: 1,
						},
					},
				},
			},
		}
	}

	return query
}

func buildDatasetSearchQuery(term string, dimensionFilters []models.Filter, topicFilters []models.Filter, limit, offset int) *models.Body {
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

func buildAreaSearchQuery(term string, hierarchyFilters []models.Filter, geoLocation *models.GeoLocation, limit, offset int) *models.Body {
	var object models.Object
	highlight := make(map[string]models.Object)

	highlight["code"] = object
	highlight["hierarchy"] = object
	highlight["name"] = object

	code := make(map[string]string)
	hierarchy := make(map[string]string)
	name := make(map[string]string)
	code["code"] = term
	hierarchy["hierarchy"] = term
	name["name"] = term

	codeMatch := models.Match{
		Match: code,
	}

	hierarchyMatch := models.Match{
		Match: hierarchy,
	}

	nameMatch := models.Match{
		Match: name,
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

		Sort:      listOfScores,
		TotalHits: true,
	}

	if geoLocation != nil {
		query.Query = models.Query{
			Bool: &models.Bool{
				Filter: []models.Filter{
					{
						Shape: &models.GeoShape{
							Location: models.GeoLocationObj{
								Shape:    *geoLocation,
								Relation: "intersects",
							},
						},
					},
				},
			},
		}
	} else {
		query.Query = models.Query{
			Bool: &models.Bool{
				Should: []models.Match{
					codeMatch,
					hierarchyMatch,
					nameMatch,
				},
				MinimumShouldMatch: 1,
			},
		}
	}

	if hierarchyFilters != nil && len(hierarchyFilters) > 0 {
		query.Query.Bool.Filter = append(query.Query.Bool.Filter, hierarchyFilters...)
	}

	return query
}

func (api *SearchAPI) getPostcodeLocation(ctx context.Context, term string, distObj *models.DistObj, logData log.Data) (*models.GeoLocation, error) {
	var geoLocation *models.GeoLocation
	postcodes := regPostcode.FindAllString(term, -1)
	if len(postcodes) > 0 {
		// Only use first postcode found
		p := strings.ReplaceAll(postcodes[0], " ", "")
		lcPostcode := strings.ToLower(p)

		postcodeResponse, _, err := api.elasticsearch.GetPostcodes(ctx, api.postcodeIndex, lcPostcode)
		if err != nil {
			log.Event(ctx, "getPostcodeSearch endpoint: failed to search for postcode", log.ERROR, log.Error(err), logData)

			return geoLocation, nil
		}

		if len(postcodeResponse.Hits.Hits) < 1 {
			log.Event(ctx, "getPostcodeSearch endpoint: failed to find postcode", log.WARN, log.Error(errs.ErrPostcodeNotFound), logData)
		}

		// calculate distance (in metres) based on distObj
		dist := distObj.CalculateDistanceInMetres(ctx)

		pcCoordinate := helpers.Coordinate{
			Lat: postcodeResponse.Hits.Hits[0].Source.Pin.Location.Lat,
			Lon: postcodeResponse.Hits.Hits[0].Source.Pin.Location.Lon,
		}

		// build polygon from circle using long/lat of postcod and distance
		polygonShape, err := helpers.CircleToPolygon(pcCoordinate, dist, defaultSegments)
		if err != nil {
			return geoLocation, nil
		}

		var coordinates [][][]float64
		geoLocation = &models.GeoLocation{
			Type:        "polygon",
			Coordinates: append(coordinates, polygonShape.Coordinates),
		}
	}

	return geoLocation, nil
}
