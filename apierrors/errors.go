package apierrors

import "errors"

// A list of error messages for Search API
var (
	ErrAreaProfileNotFound = errors.New("area profile not found")
	ErrBadSearchQuery      = errors.New("bad query sent to elasticsearch index")
	// ErrBoundaryFileNotFound    = errors.New("invalid id, boundary file does not exist")
	// ErrEmptyCoordinates        = errors.New("missing coordinates in array")
	// ErrEmptyDistanceTerm       = errors.New("empty query term: distance")
	ErrEmptySearchTerm = errors.New("empty search term")
	// ErrEmptyShape              = errors.New("empty shape")
	ErrIndexNotFound  = errors.New("search index not found")
	ErrInternalServer = errors.New("internal server error")
	// ErrInvalidCoordinates      = errors.New("should contain two coordinates, representing [latitude, longitude]")
	// ErrInvalidShape            = errors.New("invalid list of coordinates, the first and last coordinates should be the same to complete boundary line")
	// ErrLessThanFourCoordinates = errors.New("invalid number of coordinates, need a minimum of 4 values")
	// ErrLessThanTwoPolygons     = errors.New("invalid number of polygons, needs a minimum of 2 values if the geometry type is set to multipolygon")
	ErrMarshallingQuery = errors.New("failed to marshal query to bytes for request body to send to elastic")
	// ErrMissingShapeFile        = errors.New("missing shapefile value in request")
	// ErrMissingType             = errors.New("missing type value in request")
	ErrNegativeLimit           = errors.New("limit needs to be a positive number, limit cannot be lower than 0")
	ErrNegativeOffset          = errors.New("offset needs to be a positive number, offset cannot be lower than 0")
	ErrParsingQueryParameters  = errors.New("failed to parse query parameters, values must be an integer")
	ErrPostcodeNotFound        = errors.New("postcode not found")
	ErrTooManyDimensionFilters = errors.New("Too many dimension filters, limited to a maximum of 10")
	ErrTooManyHierarchyFilters = errors.New("Too many hierarchy filters, limited to a maximum of 5")
	ErrTooManyTopicFilters     = errors.New("Too many topic filters, limited to a maximum of 10")
	ErrTopicNotFound           = errors.New("Topic not found")
	ErrUnableToParseJSON       = errors.New("failed to parse json body")
	ErrUnableToReadMessage     = errors.New("failed to read message body")
	// ErrUnexpectedStatusCode    = errors.New("unexpected status code from elastic api")
	ErrUnmarshallingJSON = errors.New("failed to unmarshal data")

	NotFoundMap = map[error]bool{
		// ErrBoundaryFileNotFound: true,
		ErrAreaProfileNotFound: true,
		ErrTopicNotFound:       true,
	}

	BadRequestMap = map[error]bool{
		// ErrEmptyCoordinates:        true,
		// ErrEmptyDistanceTerm:       true,
		ErrEmptySearchTerm: true,
		// ErrEmptyShape:              true,
		// ErrInvalidCoordinates:      true,
		// ErrInvalidShape:            true,
		// ErrLessThanFourCoordinates: true,
		// ErrLessThanTwoPolygons:     true,
		// ErrMissingType:             true,
		ErrNegativeLimit:           true,
		ErrNegativeOffset:          true,
		ErrParsingQueryParameters:  true,
		ErrTooManyDimensionFilters: true,
		ErrTooManyHierarchyFilters: true,
		ErrTooManyTopicFilters:     true,
		ErrUnableToParseJSON:       true,
		ErrUnableToReadMessage:     true,
	}
)
