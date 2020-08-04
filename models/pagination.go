package models

import (
	"errors"
	"strconv"
)

// PageVariables are the necessary fields to determine paging
type PageVariables struct {
	DefaultMaxResults int
	Limit             int
	Offset            int
}

// ErrorMaximumOffsetReached - return error
func ErrorMaximumOffsetReached(m int) error {
	err := errors.New("the maximum offset has been reached, the offset cannot be more than " + strconv.Itoa(m))
	return err
}

// Validate represents a model for validating pagination variables
func (page *PageVariables) Validate() error {
	if page.Offset >= page.DefaultMaxResults {
		return ErrorMaximumOffsetReached(page.DefaultMaxResults)
	}

	if page.Offset+page.Limit > page.DefaultMaxResults {
		page.Limit = page.DefaultMaxResults - page.Offset
	}

	return nil
}
