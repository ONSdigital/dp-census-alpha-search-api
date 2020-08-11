package models

import (
	"errors"
	"strings"

	errs "github.com/ONSdigital/dp-census-alpha-search-api/apierrors"
)

const (
	maximumDimensionFilters, maximumTopicFilters = 10, 10
	maximumHierarchyFilters                      = 5
	dimensionName                                = "dimensions.name"
	hierarchyName                                = "hierarchy"
	topic1                                       = "topic1"
	topic2                                       = "topic2"
	topic3                                       = "topic3"
)

// ErrorInvalidTopics - return error
func ErrorInvalidTopics(topicList []string) error {
	topics := strings.Join(topicList, ",")
	err := errors.New("invalid list of topics to filter by: " + topics)
	return err
}

// ErrorInvalidHierarchy - return error
func ErrorInvalidHierarchy(hierarchy string) error {
	err := errors.New("invalid hierarchy to filter by: " + hierarchy)
	return err
}

// ValidateDimensions checks the values in dimensions are valid for
// querying elasticsearch API
func ValidateDimensions(dimensions string) ([]Filter, error) {
	if dimensions == "" {
		return nil, nil
	}

	dimensionList := strings.Split(dimensions, ",")

	if len(dimensionList) > maximumDimensionFilters {
		return nil, errs.ErrTooManyDimensionFilters
	}

	var filters []Filter
	for _, dimension := range dimensionList {

		filters = append(filters, Filter{
			Nested: &Nested{
				Path: "dimensions",
				Query: []NestedQuery{
					{
						Term: map[string]string{
							dimensionName: dimension},
					},
				},
			},
		})
	}

	return filters, nil
}

// ValidateHierarchies checks the values in hierarchies are valid for
// querying elasticsearch API
func ValidateHierarchies(hierarchies string) ([]Filter, error) {

	// Lower case and remove all white space for hierarchies
	h := strings.ToLower(strings.ReplaceAll(hierarchies, " ", ""))

	if h == "" {
		return nil, nil
	}

	hierarchyList := strings.Split(h, ",")

	if len(hierarchyList) > maximumHierarchyFilters {
		return nil, errs.ErrTooManyHierarchyFilters
	}

	var newHierarchyList []string
	for _, hierarchy := range hierarchyList {
		// Check hierarchy is valid
		val, ok := validHierarchies[hierarchy]
		if !ok {
			return nil, ErrorInvalidHierarchy(hierarchy)
		}

		newHierarchyList = append(newHierarchyList, val)
	}

	var filters []Filter
	if len(newHierarchyList) > 0 {
		filters = append(filters, Filter{
			Terms: map[string]interface{}{hierarchyName: newHierarchyList},
		})
	}

	return filters, nil
}

// ValidateTopics checks the values in topics are valid
func ValidateTopics(topics string) ([]Filter, error) {
	if topics == "" {
		return nil, nil
	}

	topicList := strings.Split(topics, ",")

	if len(topicList) > maximumTopicFilters {
		return nil, errs.ErrTooManyTopicFilters
	}

	var invalidTopics, topic1List, topic2List, topic3List []string
	for _, topic := range topicList {
		if validTopics[topic] < 1 {
			invalidTopics = append(invalidTopics, topic)
		} else if validTopics[topic] == 1 {
			topic1List = append(topic1List, topic)
		} else if validTopics[topic] == 2 {
			topic2List = append(topic2List, topic)
		} else if validTopics[topic] == 3 {
			topic3List = append(topic3List, topic)
		}
	}

	if len(invalidTopics) > 0 {
		return nil, ErrorInvalidTopics(invalidTopics)
	}

	var filters []Filter
	if len(topic1List) > 0 {
		filters = append(filters, Filter{
			Terms: map[string]interface{}{topic1: topic1List},
		})
	}

	if len(topic2List) > 0 {
		filters = append(filters, Filter{
			Terms: map[string]interface{}{topic2: topic2List},
		})
	}

	if len(topic3List) > 0 {
		filters = append(filters, Filter{
			Terms: map[string]interface{}{topic3: topic3List},
		})
	}

	return filters, nil
}

var validHierarchies = map[string]string{
	"countries":                   "Countries",
	"lowerlayersuperoutputareas":  "Lower Layer Super Output Areas",
	"majortownsandcities":         "Major Towns and Cities",
	"middlelayersuperoutputareas": "Middle Layer Super Output Areas",
	"outputareas":                 "Output Areas",
}

var validTopics = map[string]int{
	"adoption":                                 3,
	"ageing":                                   3,
	"balanceofpayments":                        3,
	"birthsdeathsandmarriages":                 2,
	"causesofdeath":                            3,
	"childhealth":                              3,
	"conceptionandfertilityrates":              3,
	"conditionsanddiseases":                    3,
	"crimeandjustice":                          2,
	"culturalidentity":                         2,
	"deaths":                                   3,
	"debt":                                     3,
	"disability":                               3,
	"divorce":                                  3,
	"drugusealcoholandsmoking":                 3,
	"earningsandworkinghours":                  3,
	"economicinactivity":                       3,
	"economicoutputandproductivity":            2,
	"economy":                                  1,
	"educationandchildcare":                    2,
	"elections":                                2,
	"electoralregistration":                    3,
	"employmentandemployeetypes":               3,
	"employmentandlabourmarket":                1,
	"environmentalaccounts":                    2,
	"ethnicity":                                3,
	"expenditure":                              3,
	"families":                                 3,
	"generalelections":                         3,
	"governmentpublicsectorandtaxes":           2,
	"grossdisposablehouseholdincome":           3,
	"grossdomesticproductgdp":                  2,
	"grossvalueaddedgva":                       2,
	"healthandlifeexpectancies":                3,
	"healthandsocialcare":                      2,
	"healthandwellbeing":                       3,
	"healthcaresystem":                         3,
	"healthinequalities":                       3,
	"homeinternetandsocialmediausage":          3,
	"householdcharacteristics":                 2,
	"housing":                                  2,
	"incomeandwealth":                          3,
	"inflationandpriceindices":                 2,
	"internationalmigration":                   3,
	"investmentspensionsandtrusts":             2,
	"labourproductivity":                       3,
	"language":                                 3,
	"leisureandtourism":                        2,
	"lifeexpectancies":                         3,
	"livebirths":                               3,
	"localgovernmentelections":                 3,
	"localgovernmentfinance":                   3,
	"marriagecohabitationandcivilpartnerships": 3,
	"maternities":                              3,
	"mentalhealth":                             3,
	"migrationwithintheuk":                     3,
	"nationalaccounts":                         2,
	"outofworkbenefits":                        3,
	"output":                                   3,
	"pensionssavingsandinvestments":            3,
	"peopleinwork":                             2,
	"peoplenotinwork":                          2,
	"peoplepopulationandcommunity":             1,
	"personalandhouseholdfinances":             2,
	"populationandmigration":                   2,
	"populationestimates":                      3,
	"populationprojections":                    3,
	"productivitymeasures":                     3,
	"publicsectorfinance":                      3,
	"publicsectorpersonnel":                    3,
	"publicservicesproductivity":               3,
	"publicspending":                           3,
	"redundancies":                             3,
	"regionalaccounts":                         2,
	"religion":                                 3,
	"researchanddevelopmentexpenditure":        3,
	"satelliteaccounts":                        3,
	"sexuality":                                3,
	"socialcare":                               3,
	"stillbirths":                              3,
	"supplyandusetables":                       3,
	"taxesandrevenue":                          3,
	"uksectoraccounts":                         3,
	"unemployment":                             3,
	"wellbeing":                                2,
	"workplacedisputesandworkingconditions":    3,
	"workplacepensions":                        3,
}
