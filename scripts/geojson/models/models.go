package models

type GeoDocs struct {
	Items []GeoDoc `json:"features"`
}

type GeoDoc struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Code           string         `json:"code"`
	Hierarchy      string         `json:"hierarchy"`
	LSOA11NM       string         `json:"lsoa11nm,omitempty"`
	LSOA11NMW      string         `json:"lsoa11nmw,omitempty"`
	MSOA11NM       string         `json:"msoa11nm,omitempty"`
	MSOA11NMW      string         `json:"msoa11nmw,omitempty"`
	LAD11CD        string         `json:"lad11cd,omitempty"`
	OA11CD         string         `json:"oa11cd,omitempty"`
	ShapeArea      float64        `json:"shape_area,omitempty"`
	ShapeLength    float64        `json:"shape_length,omitempty"`
	StatedArea     float64        `json:"stated_area,omitempty"`
	StatedLength   float64        `json:"stated_length,omitempty"`
	TCITY15NM      string         `json:"tcity15nm,omitempty"`
	Location       GeoLocation    `json:"location"`
	Statistics     []Statistic    `json:"statistics"`
	Datasets       Datasets       `json:"datasets"`
	Visualisations Visualisations `json:"visualisation"`
}

type GeoLocation struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"`
}

type Statistic struct {
	Header string  `json:"header"`
	Value  float64 `json:"value"`
	Units  string  `json:"units"`
}

type Datasets struct {
	Count int    `json:"count"`
	Items []Item `json:"items"`
}

type Item struct {
	Title string `json:"title"`
	Links Links  `json:"links"`
}

type Links struct {
	HRef string `json:"href"`
	ID   string `json:"id"`
}

type Visualisations struct {
	Count int    `json:"count"`
	Items []Item `json:"items"`
}
