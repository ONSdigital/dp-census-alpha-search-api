SHELL=bash

BUILD=build
BIN_DIR?=.

REFRESH=refresh
LSOA=2011-lsoa
MSOA=2011-msoa
OA=2011-oa
TCITY=2015-tcity
GEOJSON=geojson
LOAD_POSTCODES=load-postcodes

build:
	go generate ../...
	@mkdir -p ../$(BUILD)/$(BIN_DIR)
	
refreshgeojson: build
	go build -o ../$(BUILD)/$(BIN_DIR)/$(REFRESH) $(GEOJSON)/$(REFRESH)/main.go
	HUMAN_LOG=1 go run -race $(GEOJSON)/$(REFRESH)/main.go
	
lsoa: build
	go build -o ../$(BUILD)/$(BIN_DIR)/$(LSOA) $(GEOJSON)/$(LSOA)/main.go
	HUMAN_LOG=1 go run -race $(GEOJSON)/$(LSOA)/main.go

msoa: build
	go build -o ../$(BUILD)/$(BIN_DIR)/$(MSOA) $(GEOJSON)/$(MSOA)/main.go
	HUMAN_LOG=1 go run -race $(GEOJSON)/$(MSOA)/main.go

oa: build
	go build -o ../$(BUILD)/$(BIN_DIR)/$(OA) $(GEOJSON)/$(OA)/main.go
	HUMAN_LOG=1 go run -race $(GEOJSON)/$(OA)/main.go

tcity: build
	go build -o ../$(BUILD)/$(BIN_DIR)/$(TCITY) $(GEOJSON)/$(TCITY)/main.go
	HUMAN_LOG=1 go run -race $(GEOJSON)/$(TCITY)/main.go

geojson: refreshgeojson lsoa msoa tcity oa
	
postcode: build
	go build -o ../$(BUILD)/$(BIN_DIR)/$(LOAD_POSTCODES) $(LOAD_POSTCODES)/main.go
	HUMAN_LOG=1 go run -race $(LOAD_POSTCODES)/main.go

test:
	go test -cover -race ./...

.PHONY: build postcode geojson lsoa msoa tcity refresh test
