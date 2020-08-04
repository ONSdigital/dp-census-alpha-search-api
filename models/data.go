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
	Alias       string      `json:"alias,omitempty"`
	Description string      `json:"description,omitempty"`
	Dimensions  []Dimension `json:"dimensions,omitempty"`
	Title       string      `json:"title,omitempty"`
	Topic1      string      `json:"topic1,omitempty"`
	Topic2      string      `json:"topic2,omitempty"`
	Topic3      string      `json:"topic3,omitempty"`
	Link        string      `json:"link,omitempty"`
	Matches     Matches     `json:"matches,omitempty"`
}

// Dimension represents an object containing dimension data
type Dimension struct {
	Label string `json:"label,omitempty"`
	Name  string `json:"name,omitempty"`
}

// Matches represents a list of members and their arrays of character offsets that matched the search term
type Matches struct {
	Alias          []string `json:"alias,omitempty"`
	Description    []string `json:"description,omitempty"`
	DimensionLabel []string `json:"dimensions.label,omitempty"`
	DimensionName  []string `json:"dimensions.name,omitempty"`
	Title          []string `json:"title,omitempty"`
	Topic1         []string `json:"topic1,omitempty"`
	Topic2         []string `json:"topic2,omitempty"`
	Topic3         []string `json:"topic3,omitempty"`
}
