package helpers

import (
	"errors"
	"math"
)

// Shape ...
type Shape struct {
	Type        string
	Coordinates [][]float64
}

// Coordinate ...
type Coordinate struct {
	Lat float64
	Lon float64
}

const (
	twoPi         float64 = 2 * math.Pi
	radiusOfEarth float64 = 6378137 //(in metres defined by wgs84)
)

// List of errors
var (
	ErrTooManySegments          = errors.New("Too many segments")
	ErrInvalidLongitudinalPoint = errors.New("Longitude has to be between -180 and 180")
	ErrInvalidLatitudinalPoint  = errors.New("Latitude has to be between -90 and 90")
)

// CircleToPolygon ...
func CircleToPolygon(geoPoint Coordinate, radius float64, segments int) (*Shape, error) {

	// validate input
	if err := validateInput(geoPoint, radius, segments); err != nil {
		return nil, err
	}

	shape := &Shape{
		Type: "Polygon",
	}

	var coordinates [][]float64

	for i := 0; i < segments; i++ {
		segment := (twoPi * float64(-i)) / float64(segments)
		coordinate := generateCoordinate(geoPoint, radius, segment)
		coordinates = append(coordinates, coordinate)
	}

	// Push first coordinate to be last coordinate to complete polygon circle
	coordinates = append(coordinates, coordinates[0])

	shape.Coordinates = coordinates

	return shape, nil
}

func toRadians(angleInDegrees float64) float64 {
	return (angleInDegrees * math.Pi) / 180
}

func toDegrees(angleInRadians float64) float64 {
	return (angleInRadians * 180) / math.Pi
}

func generateCoordinate(geoPoint Coordinate, distance float64, segment float64) []float64 {
	lat1 := toRadians(geoPoint.Lat)
	lon1 := toRadians(geoPoint.Lon)

	// distance divided by radius of the earth
	dByR := distance / radiusOfEarth

	lat := math.Asin(
		math.Sin(lat1)*math.Cos(dByR) + math.Cos(lat1)*math.Sin(dByR)*math.Cos(segment),
	)

	lon := lon1 + math.Atan2(
		math.Sin(segment)*math.Sin(dByR)*math.Cos(lat1),
		math.Cos(dByR)-math.Sin(lat1)*math.Sin(lat),
	)

	return []float64{toDegrees(lon), toDegrees(lat)}
}

func validateInput(geoPoint Coordinate, radius float64, segments int) error {

	if err := validateSegments(segments); err != nil {
		return err
	}

	if err := validateCoordinates(geoPoint); err != nil {
		return err
	}

	return nil
}

func validateSegments(segments int) error {
	if segments > 180 {
		return ErrTooManySegments
	}

	return nil
}

func validateCoordinates(geoPoint Coordinate) error {
	if geoPoint.Lon > 180 || geoPoint.Lon < -180 {
		return ErrInvalidLongitudinalPoint
	}

	if geoPoint.Lat > 90 || geoPoint.Lat < -90 {
		return ErrInvalidLatitudinalPoint
	}

	return nil
}
