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

// ------------------------------------------------------------------------

type PostcodeRequest struct {
	Query PostcodeQuery `json:"query"`
}

type PostcodeQuery struct {
	Distance PostcodeTerm `json:"term"`
}

type PostcodeTerm struct {
	Postcode string `json:"postcode"`
}

// ------------------------------------------------------------------------

type PostcodeResponse struct {
	Hits EmbededHits `json:"hits"`
}

type EmbededHits struct {
	Hits []HitObj `json:"hits"`
}

type HitObj struct {
	Source Source `json:"_source"`
}

type Source struct {
	Postcode    string      `json:"postcode,omitempty"`
	RawPostcode string      `json:"postcode_raw,omitempty"`
	Pin         Pin         `json:"pin,omitempty"`
	ID          string      `json:"id,omitempty"`
	Location    GeoLocation `json:"location,omitempty"`
}

type Pin struct {
	Location PinLocation `json:"location"`
}

type PinLocation struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type GeoLocation struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"`
}
