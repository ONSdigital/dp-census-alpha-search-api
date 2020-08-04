package models

// Body represents the request body to elasticsearch
type Body struct {
	From      int        `json:"from"`
	Size      int        `json:"size"`
	Highlight *Highlight `json:"highlight,omitempty"`
	Query     Query      `json:"query"`
	Sort      []Scores   `json:"sort"`
	TotalHits bool       `json:"track_total_hits"`
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
	Bool *Bool `json:"bool,omitempty"`
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
}

// Match represents the fields that the term should or must match within query
type Match struct {
	Match  map[string]string `json:"match,omitempty"`
	Nested *Nested           `json:"nested,omitempty"`
}

// Nested represents a nested query object
type Nested struct {
	Path  string        `json:"path,omitempty"`
	Query []NestedQuery `json:"query,omitempty"`
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
