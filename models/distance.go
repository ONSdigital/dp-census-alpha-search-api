package models

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/ONSdigital/log.go/log"
)

// DistObj represents the parameters to determine a distance
type DistObj struct {
	Value float64
	Unit  string
}

var defaultDistance = "0.1,km"

var kilometres = map[string]bool{
	"km":         true,
	"kilometers": true,
	"kilometer":  true,
	"kilometre":  true,
	"kilometres": true,
}

var miles = map[string]bool{
	"m":     true,
	"mile":  true,
	"miles": true,
}

// ErrorInvalidDistance - return error
func ErrorInvalidDistance(m string) error {
	err := errors.New("invalid distance value: " + m + ". Should contain a number and unit of distance separated by a comma e.g. 40,km")
	return err
}

// ErrorInvalidRelation - return error
func ErrorInvalidRelation(m string) error {
	err := errors.New("invalid relation value: " + m + ". Should contain one of the following: intersects or within")
	return err
}

// ValidateDistance ...
func ValidateDistance(distance string) (*DistObj, error) {
	if distance == "" {
		distance = defaultDistance
	}

	lcDistance := strings.ToLower(distance)

	values := strings.SplitAfter(lcDistance, ",")

	if len(values) == 1 {
		return nil, ErrorInvalidDistance(distance)
	}

	if len(values) > 2 {
		return nil, ErrorInvalidDistance(distance)
	}

	dist := strings.Replace(values[0], ",", "", 1)
	unit := values[1]

	value, err := strconv.ParseFloat(dist, 64)
	if err != nil {
		return nil, ErrorInvalidDistance(distance)
	}

	if !kilometres[unit] && !miles[unit] {
		return nil, ErrorInvalidDistance(distance)
	}

	distObj := &DistObj{
		Value: value,
		Unit:  values[1],
	}

	return distObj, nil
}

func (dO *DistObj) CalculateDistanceInMetres(ctx context.Context) (distance float64) {

	switch {
	case kilometres[dO.Unit]:
		distance = dO.Value * 1000
	case miles[dO.Unit]:
		distance = dO.Value * 1609.34
	default:
		log.Event(ctx, "unrecognizable unit value: defaulting to kilometres", log.WARN)
		distance = dO.Value * 1000
	}

	return
}

var validRelations = map[string]bool{
	"intersects": true,
	"within":     true,
}

func ValidateGeoShapeRelation(ctx context.Context, defaultRelation, relation string) (string, error) {
	if relation == "" {
		return defaultRelation, nil
	}

	lcRelation := strings.ToLower(relation)

	if _, ok := validRelations[lcRelation]; !ok {
		return "", ErrorInvalidRelation(relation)
	}

	return lcRelation, nil
}
