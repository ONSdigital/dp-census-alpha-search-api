//go:generate go get github.com/jteeuwen/go-bindata/go-bindata
//go:generate go-bindata -pkg elasticsearch ./postcode-mappings.json ./geography-mappings.json ./dataset-mappings.json

package elasticsearch
