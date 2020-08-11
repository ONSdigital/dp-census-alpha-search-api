package api

import (
	"context"

	"github.com/ONSdigital/dp-census-alpha-search-api/models"
)

// Elasticsearcher - An interface used to access elasticsearch
type Elasticsearcher interface {
	GetAreaProfile(ctx context.Context, indexName string, query interface{}) (*models.AreaProfile, int, error)
	QuerySearchIndex(ctx context.Context, indexName string, query interface{}, limit, offset int) (*models.SearchResponse, int, error)
	GetPostcodes(ctx context.Context, indexName, postcode string) (*models.PostcodeResponse, int, error)
}
