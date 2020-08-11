package models

// Taxonomy represents the hierarchy of topics
type Taxonomy struct {
	Topics []Topic `json:"topics"`
}

// Topic represents the topic data and relates to child topics
type Topic struct {
	Title          string  `json:"title"`
	FormattedTitle string  `json:"filterable_title"`
	ChildTopics    []Topic `json:"child_topics,omitempty"`
}
