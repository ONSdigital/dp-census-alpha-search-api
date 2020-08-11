package models

// AreaProfile represents the data structure for an area profile page
type AreaProfile struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Code           string         `json:"code"`
	Datasets       Datasets       `json:"datasets"`
	Hierarchy      string         `json:"hierarchy"`
	Links          Links          `json:"links"`
	Location       GeoLocation    `json:"location"`
	Statistics     []Statistic    `json:"statistics"`
	Visualisations Visualisations `json:"visualisation"`
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

// Visualisations represents the visualisations stored against area profile doc
type Visualisations struct {
	Count int    `json:"count"`
	Items []Item `json:"items"`
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

// AreaProfileQuery represents a query to retrieve an area profile
type AreaProfileQuery struct {
	Query Query `json:"query"`
}

// AreaProfileResponse represents a the data returned from querying the area profile index
type AreaProfileResponse struct {
	Hits AHits `json:"hits"`
}

type AHits struct {
	Total   int        `json:"total"`
	HitList []AHitList `json:"hits"`
}

type AHitList struct {
	Score   float64     `json:"_score"`
	Source  AreaProfile `json:"_source"`
	Matches Matches     `json:"highlight,omitempty"`
}
