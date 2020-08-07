package models

type PostcodeDoc struct {
	Postcode    string `json:"postcode"`
	PostcodeRaw string `json:"postcode_raw"`
	Pin         PinObj `json:"pin"`
}

type PinObj struct {
	PinLocation CoordinatePoint `json:"location"`
}

type CoordinatePoint struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}
