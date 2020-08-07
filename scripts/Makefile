SHELL=bash

BUILD=build
BIN_DIR?=.

MONGODB_BIND_ADDR=${mongodb_bind_addr}
FILENAME=${filename}
DATASET_INDEX=${dataset_index}
ELASTICSEARCH_URL=${elasticsearch_url}
DIMENSIONS_JSON=${dimensions_filename}
TAXONOMY_JSON=${taxonomy_filename}

RETRIEVE_CMD_DATASETS=retrieve-cmd-datasets
RETRIEVE_DATASET_TAXONOMY=retrieve-dataset-taxonomy
UPLOAD_DATASETS=upload-datasets
REFRESH=refresh
LSOA=2011-lsoa
MSOA=2011-msoa
OA=2011-oa
TCITY=2015-tcity
COUNTRY=2019-countries
GEOJSON=geojson
LOAD_POSTCODES=load-postcodes
HIERARCHIES=hierarchies

build:
	go generate ../...
	@mkdir -p ../$(BUILD)/$(BIN_DIR)

cmd-datasets-csv: build
	go build -o ../$(BUILD)/$(BIN_DIR)/$(RETRIEVE_CMD_DATASETS) $(RETRIEVE_CMD_DATASETS)/main.go
	HUMAN_LOG=1 go run -race $(RETRIEVE_CMD_DATASETS)/main.go -mongodb-bind-addr=$(MONGODB_BIND_ADDR) -filename=$(FILENAME)

taxonomy-json: build
	go build -o ../$(BUILD)/$(BIN_DIR)/$(RETRIEVE_DATASET_TAXONOMY) $(RETRIEVE_DATASET_TAXONOMY)/main.go
	HUMAN_LOG=1 go run -race $(RETRIEVE_DATASET_TAXONOMY)/main.go -filename=$(TAXONOMY_JSON)

upload-datasets: build
	go build -o ../$(BUILD)/$(BIN_DIR)/$(UPLOAD_DATASETS) $(UPLOAD_DATASETS)/main.go
	HUMAN_LOG=1 go run -race $(UPLOAD_DATASETS)/main.go -filename=$(FILENAME) -dimensions-filename=$(DIMENSIONS_JSON) -taxonomy-filename=$(TAXONOMY_JSON) -dataset-index=$(DATASET_INDEX) -elasticsearch-url=$(ELASTICSEARCH_URL)

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

country: build
	go build -o ../$(BUILD)/$(BIN_DIR)/$(COUNTRY) $(GEOJSON)/$(COUNTRY)/main.go
	HUMAN_LOG=1 go run -race $(GEOJSON)/$(COUNTRY)/main.go

hierarchies: build
	go build -o ../$(BUILD)/$(BIN_DIR)/$(HIERARCHIES) $(HIERARCHIES)/main.go
	HUMAN_LOG=1 go run -race $(HIERARCHIES)/main.go

geojson: hierarchies refreshgeojson country tcity lsoa msoa oa
	
postcode: build
	go build -o ../$(BUILD)/$(BIN_DIR)/$(LOAD_POSTCODES) $(LOAD_POSTCODES)/main.go
	HUMAN_LOG=1 go run -race $(LOAD_POSTCODES)/main.go

test:
	go test -cover -race ./...

.PHONY: cmd-datasets-csv taxonomy-json upload-datasets build postcode geojson lsoa msoa tcity country refresh test
