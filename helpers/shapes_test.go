package helpers_test

import (
	"testing"

	"github.com/ONSdigital/dp-census-alpha-search-api/helpers"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCircleToPolygon(t *testing.T) {
	Convey("Given the number of segments does not exceed 180", t, func() {
		geopoint := helpers.Coordinate{
			Lat: 51.486090,
			Lon: -3.227882,
		}

		expectedCoordinates := [][]float64{
			{-3.2278819999999997, 51.495073152841194},
			{-3.232340722565203, 51.494633401528986},
			{-3.2363627568939477, 51.49335721753477},
			{-3.239554251574248, 51.49136958551781},
			{-3.2416028009712456, 51.488865146452405},
			{-3.24230802558419, 51.48608911512211},
			{-3.241601130457992, 51.48331325278844},
			{-3.2395515486269555, 51.48080925616197},
			{-3.236360053946569, 51.47882217102961},
			{-3.2323390520518105, 51.47754642947436},
			{-3.2278819999999997, 51.4771068471588},
			{-3.223424947948189, 51.47754642947436},
			{-3.2194039460534305, 51.47882217102961},
			{-3.216212451373044, 51.48080925616197},
			{-3.2141628695420073, 51.48331325278844},
			{-3.2134559744158095, 51.48608911512211},
			{-3.2141611990287533, 51.488865146452405},
			{-3.2162097484257512, 51.49136958551781},
			{-3.219401243106052, 51.49335721753477},
			{-3.223423277434797, 51.494633401528986},
			{-3.2278819999999997, 51.495073152841194},
		}

		shape, err := helpers.CircleToPolygon(geopoint, 1000, 20)
		So(err, ShouldBeNil)
		So(shape, ShouldNotBeNil)
		So(shape.Type, ShouldEqual, "Polygon")
		So(len(shape.Coordinates), ShouldEqual, 21)
		So(shape.Coordinates, ShouldResemble, expectedCoordinates)
		So(shape.Coordinates[0], ShouldResemble, shape.Coordinates[20])
	})

	Convey("Given the number of segments does exceed 180", t, func() {
		geopoint := helpers.Coordinate{
			Lat: 51.486090,
			Lon: -3.227882,
		}

		shape, err := helpers.CircleToPolygon(geopoint, 1000, 181)
		So(shape, ShouldBeNil)
		So(err, ShouldResemble, helpers.ErrTooManySegments)
	})

	Convey("Given the coordinates of centre point are incorrect", t, func() {

		Convey("When the lattitudinal coordinate of centre point is greater than 90", func() {
			geopoint := helpers.Coordinate{
				Lat: 90.1,
				Lon: -3.227882,
			}

			shape, err := helpers.CircleToPolygon(geopoint, 1000, 10)
			So(shape, ShouldBeNil)
			So(err, ShouldResemble, helpers.ErrInvalidLatitudinalPoint)
		})

		Convey("When the lattitudinal coordinate of centre point is less than -90", func() {
			geopoint := helpers.Coordinate{
				Lat: -90.1,
				Lon: -3.227882,
			}

			shape, err := helpers.CircleToPolygon(geopoint, 1000, 10)
			So(shape, ShouldBeNil)
			So(err, ShouldResemble, helpers.ErrInvalidLatitudinalPoint)
		})

		Convey("When the longitudinal coordinate of centre point is greater than 180", func() {
			geopoint := helpers.Coordinate{
				Lat: 51.486090,
				Lon: 180.227882,
			}

			shape, err := helpers.CircleToPolygon(geopoint, 1000, 10)
			So(shape, ShouldBeNil)
			So(err, ShouldResemble, helpers.ErrInvalidLongitudinalPoint)
		})

		Convey("When the longitudinal coordinate of centre point is less than -180", func() {
			geopoint := helpers.Coordinate{
				Lat: 51.486090,
				Lon: -180.227882,
			}

			shape, err := helpers.CircleToPolygon(geopoint, 1000, 10)
			So(shape, ShouldBeNil)
			So(err, ShouldResemble, helpers.ErrInvalidLongitudinalPoint)
		})
	})
}
