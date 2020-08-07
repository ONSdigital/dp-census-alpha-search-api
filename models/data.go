package models

type SearchResponse struct {
	Hits Hits `json:"hits"`
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
	All          SearchResults `json:"all"`
	Datasets     SearchResults `json:"datasets"`
	AreaProfiles SearchResults `json:"area_profiles"`
}

// SearchResults represents a structure for a list of returned objects
type SearchResults struct {
	Count      int            `json:"count"`
	Items      []SearchResult `json:"items"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
	TotalCount int            `json:"total_count"`
}

// SearchResult represents data on a single item of search results
type SearchResult struct {
	// dataset data
	Alias       string      `json:"alias,omitempty"`
	Description string      `json:"description,omitempty"`
	Dimensions  []Dimension `json:"dimensions,omitempty"`
	Title       string      `json:"title,omitempty"`
	Topic1      string      `json:"topic1,omitempty"`
	Topic2      string      `json:"topic2,omitempty"`
	Topic3      string      `json:"topic3,omitempty"`
	// area profile data
	ID             string          `json:"id,omitempty"`
	Code           string          `json:"code,omitempty"`
	Datasets       *Datasets       `json:"datasets,omitempty"`
	Hierarchy      string          `json:"hierarchy,omitempty"`
	Name           string          `json:"name,omitempty"`
	Statistics     []Statistic     `json:"statistics,omitempty"`
	Visualisations *Visualisations `json:"visualisation,omitempty"`
	// generic data
	Links   Links   `json:"links,omitempty"`
	Matches Matches `json:"matches,omitempty"`
}

// Dimension represents an object containing dimension data
type Dimension struct {
	Label string `json:"label,omitempty"`
	Name  string `json:"name,omitempty"`
}

// Statistic represents statistical data stored against area profile doc
type Statistic struct {
	Header string  `json:"header"`
	Value  float64 `json:"value"`
	Units  string  `json:"units"`
}

// Datasets represents the datasets stored against area profile doc
type Datasets struct {
	Count int    `json:"count"`
	Items []Item `json:"items"`
}

// Item represents a generic structure for a dataset or visualisation subdocument
type Item struct {
	Title string `json:"title"`
	Links Links  `json:"links"`
}

// Links represents a generic structure for storing links against an item
type Links struct {
	Self Self `json:"self"`
}

// Self represents a link to this resource
type Self struct {
	HRef string `json:"href"`
	ID   string `json:"id"`
}

// Visualisations represents the visualisations stored against area profile doc
type Visualisations struct {
	Count int    `json:"count"`
	Items []Item `json:"items"`
}

// Matches represents a list of members and their arrays of character offsets that matched the search term
type Matches struct {
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
