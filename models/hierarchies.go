package models

// GeoHierarchiesDoc represents a list of geography hierarchies
type GeoHierarchiesDoc struct {
	Items      []GeographyObject `json:"items"`
	TotalCount int               `json:"total_count"`
}

// GeographyObject represents the structure of a dimension
type GeographyObject struct {
	Hierarchy           string `json:"hierarchy"`
	FilterableHierarchy string `json:"filterable_hierarchy"`
}
