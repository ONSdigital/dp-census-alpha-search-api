package models

type SearchResponse struct {
	Hits         Hits         `json:"hits"`
	Aggregations Aggregations `json:"aggregations,omitempty"`
}

type Hits struct {
	Total   int       `json:"total"`
	HitList []HitList `json:"hits"`
}

type HitList struct {
	Score   float64      `json:"_score"`
	Source  SearchResult `json:"_source"`
	Matches Matches      `json:"highlight,omitempty"`
}

type DimensionHits struct {
	Hits []DimensionHitList `json:"hits,omitempty"`
}

type DimensionHitList struct {
	Matches Matches `json:"highlight,omitempty"`
}

// AllSearchResults represents a structure capturing a list of all returned data type objects
type AllSearchResults struct {
	Counts       Counts        `json:"counts"`
	Limit        int           `json:"limit"`
	Offset       int           `json:"offset"`
	All          SearchResults `json:"all"`
	Datasets     SearchResults `json:"datasets"`
	AreaProfiles SearchResults `json:"area_profiles"`
	Publications SearchResults `json:"publications"`
}

// Counts represent a list of counts for each data type
type Counts struct {
	All          int `json:"all"`
	Datasets     int `json:"datasets"`
	AreaProfiles int `json:"area_profiles"`
	Publications int `json:"publications"`
}

// SearchResults represents a structure for a list of returned objects
type SearchResults struct {
	Aggregations *Aggregations  `json:"aggregations,omitempty"`
	Count        int            `json:"count"`
	Items        []SearchResult `json:"items"`
	TotalCount   int            `json:"total_count"`
}

// DatasetSearchResults represents a structure for a list of returned dataset resources
type DatasetSearchResults struct {
	Count      int            `json:"count"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
	TotalCount int            `json:"total_count"`
	Items      []SearchResult `json:"items"`
}

// SearchResult represents data on a single item of search results
type SearchResult struct {
	// dataset data
	Alias       string      `json:"alias,omitempty"`
	Description string      `json:"description,omitempty"`
	DocType     string      `json:"doc_type,omitempty"`
	Dimensions  []Dimension `json:"dimensions,omitempty"`
	Title       string      `json:"title,omitempty"`
	Topic1      string      `json:"topic1,omitempty"`
	Topic2      string      `json:"topic2,omitempty"`
	Topic3      string      `json:"topic3,omitempty"`
	// area profile data
	ID        string `json:"id,omitempty"`
	Code      string `json:"code,omitempty"`
	Hierarchy string `json:"hierarchy,omitempty"`
	Name      string `json:"name,omitempty"`
	// generic data
	Links   Links      `json:"links,omitempty"`
	Matches NewMatches `json:"matches,omitempty"`
}

// Dimension represents an object containing dimension data
type Dimension struct {
	Label string `json:"label,omitempty"`
	Name  string `json:"name,omitempty"`
}

// Matches represents a list of members and their arrays of character offsets that matched the search term
type Matches struct {
	// Dataset Matches
	Alias          []string `json:"alias.raw,omitempty"`
	Description    []string `json:"description.raw,omitempty"`
	DimensionLabel []string `json:"dimensions.label,omitempty"`
	DimensionName  []string `json:"dimensions.name,omitempty"`
	Title          []string `json:"title.raw,omitempty"`
	Topic1         []string `json:"topic1,omitempty"`
	Topic2         []string `json:"topic2,omitempty"`
	Topic3         []string `json:"topic3,omitempty"`
	// Area Profile Matches
	Code      []string `json:"code,omitempty"`
	Hierarchy []string `json:"hierarchy,omitempty"`
	Name      []string `json:"name,omitempty"`
}

// NewMatches represents a list of members and their arrays of character offsets that matched the search term
type NewMatches struct {
	// Dataset Matches
	Alias          []string `json:"alias,omitempty"`
	Description    []string `json:"description,omitempty"`
	DimensionLabel []string `json:"dimensions.label,omitempty"`
	DimensionName  []string `json:"dimensions.name,omitempty"`
	Title          []string `json:"title,omitempty"`
	Topic1         []string `json:"topic1,omitempty"`
	Topic2         []string `json:"topic2,omitempty"`
	Topic3         []string `json:"topic3,omitempty"`
	// Area Profile Matches
	Code      []string `json:"code,omitempty"`
	Hierarchy []string `json:"hierarchy,omitempty"`
	Name      []string `json:"name,omitempty"`
}

// Aggregations is a list of aggregated fields with the number of
// documents that are returned for unique values of an aggregated field
type Aggregations struct {
	Dimensions  *AggItems `json:"dimensions,omitempty"`
	Hierarchies *AggItems `json:"hierarchies,omitempty"`
	Topic1      *AggItems `json:"topic1,omitempty"`
	Topic2      *AggItems `json:"topic2,omitempty"`
	Topic3      *AggItems `json:"topic3,omitempty"`
}

// AggItems represents the a list of items/buckets for aggregation
type AggItems struct {
	Items []Bucket `json:"buckets,omitempty"`
}

// Bucket represents a single value of the hierarchy and how often it appears across the returned result
type Bucket struct {
	Key      string `json:"key,omitempty"`
	DocCount int    `json:"doc_count,omitempty"`
}
