package models

type GeoDocs struct {
	Items []GeoDoc `json:"features"`
}

type GeoDoc struct {
	Name         string      `json:"name"`
	Code         string      `json:"code"`
	Hierarchy    string      `json:"hierarchy"`
	LSOA11NM     string      `json:"lsoa11nm,omitempty"`
	LSOA11NMW    string      `json:"lsoa11nmw,omitempty"`
	MSOA11NM     string      `json:"msoa11nm,omitempty"`
	MSOA11NMW    string      `json:"msoa11nmw,omitempty"`
	LAD11CD      string      `json:"lad11cd,omitempty"`
	OA11CD       string      `json:"oa11cd,omitempty"`
	ShapeArea    float64     `json:"shape_area,omitempty"`
	ShapeLength  float64     `json:"shape_length,omitempty"`
	StatedArea   float64     `json:"stated_area,omitempty"`
	StatedLength float64     `json:"stated_length,omitempty"`
	TCITY15NM    string      `json:"tcity15nm,omitempty"`
	Location     GeoLocation `json:"location"`
}

type GeoLocation struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"`
}
