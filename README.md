# dp-census-dataset-search-api

This is the census alpha search API application for census project. To provide a continuously improving working example of search features needed to answer questions posed during the 2021 census alpha project.

### Requirements

In order to run the service locally you will need the following:

- Go
- Git
- Java 8 or greater for elasticsearch
- ElasticSearch (version 6.7 or 6.8)

### Getting started

- Clone the repo go get github.com/ONSdigital/dp-census-alpha-search-api
- Run elasticsearch e.g. ./elasticsearch<version>/bin/elasticsearch
- Follow [setting up data](#setting-up-data)
- Run `make debug` to start search API service

Follow swagger documentation on how to interact with local api, some examples are below:

```
curl -XOPTIONS localhost:10300/search -vvv
curl -XGET localhost:10300/search?q={term} -vvv
curl -XGET "localhost:10300/search?q={term}&offset=5&limit=5" -vvv
curl -XGET "localhost:10300/search?q={term}&topics={topic}" -vvv (see taxonomy endpoint for topic filter options)
curl -XGET "localhost:10300/search?q={term}&dimensions={dimension}" -vvv (see dimensions endpoint for dimension filter options)
curl -XGET "localhost:10300/search?q={term}&hierarchies={geographical hierarchy}" -vvv (see hierarchies endpoint for hierarchy filter options)


curl -XGET localhost:10300/area-profiles/{id} -vvv
curl -XGET localhost:10300/area-profiles/{id}/search?q={term} -vvv (can use the dimensions and topics filter as well as offset and limit params to page through results)

curl -XGET localhost:10300/taxonomy -vvv
curl -XGET localhost:10300/taxonomy/{topic} -vvv
curl -XGET localhost:10300/dimensions -vvv
curl -XGET localhost:10300/hierarchies -vvv
```

#### Setting up data

Once elasticsearch is running and you can connect to your instance. Follow the instructions [here](scripts/README.md) to load in some prepared cmd datasets.

### Configuration

| Environment variable        | Default               | Description
| --------------------------- | --------------------- | -----------
| BIND_ADDR                   | :10300                | The host and port to bind to |
| AREA_PROFILE_SEARCH_INDEX   | area-profiles         | The index in which the area profile documents are stored in elasticsearch |
| DATASET_SEARCH_INDEX        | datasets              | The index in which the dataset documents are stored in elasticsearch |
| POSTCODE_SEARCH_INDEX       | postcodes             | The index in which the postcode documents are stored in elasticsearch |
| ELASTIC_SEARCH_URL          | http://localhost:9200 | The host name for elasticsearch |
| MAX_SEARCH_RESULTS_OFFSET   | 1000                  | The maximum offset for the number of results returned by search query |
| SIGN_ELASTICSEARCH_REQUESTS | false                 | Boolean flag to identify whether elasticsearch requests via elastic API need to be signed if elasticsearch cluster is running in aws |
| DIMENSIONS_FILENAME         | data/dimensions.json  | The json file that contains a list of dimensions that can be used to filter results from search endpoint |
| HIERARCHIES_FILENAME        | data/hierarchy.json   | The json file that contains a list of geographical hierarchies that can be used to filter results from search endpoint |
| TAXONOMY_FILENAME           | data/taxonomy.json    | The json file that contains a list of topics that can be used to filter results from search endpoint |


### Notes

See [command list](COMMANDS.md) for a list of helpful commands to run alongside setting up data, useful to check what search indexes exist and their individual mappings and number of documents etc..

One can run the unit tests with `make test`
