package models

// DimensionsDoc represents a list of dimensions
type DimensionsDoc struct {
	Dimensions []DimensionObject `json:"items"`
	TotalCount int               `json:"total_count"`
}

// DimensionObject represents the structure of a dimension
type DimensionObject struct {
	Label string `json:"label,omitempty"`
	Name  string `json:"name,omitempty"`
}
