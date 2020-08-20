package models

// Body represents the request body to elasticsearch
type Body struct {
	Aggregations Aggs       `json:"aggs,omitempty"`
	From         int        `json:"from"`
	Size         int        `json:"size"`
	Highlight    *Highlight `json:"highlight,omitempty"`
	Query        Query      `json:"query"`
	Sort         []Scores   `json:"sort"`
	TotalHits    bool       `json:"track_total_hits"`
}

// Aggs represents the name in which an specific aggregation is returned as
type Aggs struct {
	Dimensions  Agg `json:"dimensions,omitempty"`
	Hierarchies Agg `json:"hierarchies,omitempty"`
	Topic1      Agg `json:"topic1,omitempty"`
	Topic2      Agg `json:"topic2,omitempty"`
	Topic3      Agg `json:"topic3,omitempty"`
}

// Agg represents a list of terms to aggregate results by
type Agg struct {
	// Nested *NestedPath `json:"nested,omitempty"`
	Terms AggTerm `json:"terms"`
	// Aggs   *NestedAgg  `json:"aggs,omitempty"`
}

// NestedPath represents the path to the aggregated field
type NestedPath struct {
	Path string `json:"path"`
}

// NestedAgg a list of terms to aggregate results by the nested term
type NestedAgg struct {
	Terms AggTerm `json:"terms"`
}

// AggTerm represents a term to aggregate results by
type AggTerm struct {
	Field string `json:"field"`
}

// Highlight represents parts of the fields that matched
type Highlight struct {
	PreTags  []string          `json:"pre_tags,omitempty"`
	PostTags []string          `json:"post_tags,omitempty"`
	Fields   map[string]Object `json:"fields,omitempty"`
	Order    string            `json:"score,omitempty"`
}

// Object represents an empty object (as expected by elasticsearch)
type Object struct{}

// Query represents the request query details
type Query struct {
	Bool *Bool             `json:"bool,omitempty"`
	Term map[string]string `json:"term,omitempty"`
}

// Bool represents the desirable goals for query
type Bool struct {
	Filter             []Filter `json:"filter,omitempty"`
	Must               []Match  `json:"must,omitempty"`
	Should             []Match  `json:"should,omitempty"`
	MinimumShouldMatch int      `json:"minimum_should_match,omitempty"`
}

// Filter represents the filtering object (can only contain eiter term or terms but not both)
type Filter struct {
	Term   map[string]string      `json:"term,omitempty"`
	Terms  map[string]interface{} `json:"terms,omitempty"`
	Nested *Nested                `json:"nested,omitempty"`
	Shape  *GeoShape              `json:"geo_shape,omitempty"`
}

// Match represents the fields that the term should or must match within query
type Match struct {
	Bool   *Bool             `json:"bool,omitempty"`
	Match  map[string]string `json:"match,omitempty"`
	Nested *Nested           `json:"nested,omitempty"`
}

// Nested represents a nested query object
type Nested struct {
	Path  string        `json:"path,omitempty"`
	Query []NestedQuery `json:"query,omitempty"`
}

// GeoShape represents the query object for a elasticsearch geography shape
type GeoShape struct {
	Location GeoLocationObj `json:"location"`
}

// GeoLocationObj represents the attributes of the geography elasticsearch query
type GeoLocationObj struct {
	Shape    GeoLocation `json:"shape"`
	Relation string      `json:"relation"`
}

// NestedQuery represents ...
type NestedQuery struct {
	Must  []Match                `json:"must,omitempty"`
	Term  map[string]string      `json:"term,omitempty"`
	Terms map[string]interface{} `json:"terms,omitempty"`
}

// Scores represents a list of scoring, e.g. scoring on relevance, but can add in secondary
// score such as alphabetical order if relevance is the same for two search results
type Scores struct {
	Score Score `json:"_score"`
}

// Score contains the ordering of the score (ascending or descending)
type Score struct {
	Order string `json:"order"`
}
