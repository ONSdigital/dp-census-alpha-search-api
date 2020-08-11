## List of useful Commands

### Check health of cluster

`curl -XGET 'localhost:9200/_cluster/health/<index>?pretty'`

### Get mapping

`curl -XGET 'localhost:9200/<index>/_mapping?pretty'`

### Set mapping

`curl -XPUT localhost:9200/<index> -d@.<path to mappping schema file(json)>`
e.g. `curl -XPUT localhost:9200/new-index -d@./elasticsearch/mappings.json`

### Get index

`curl -XGET 'localhost:9200/<index>/*?pretty'`

### Delete index

`curl -XDELETE localhost:9200/<index>`

### Get all indexes

`curl -XGET 'localhost:9200/_aliases?pretty'`

### Get indexes which are aliased

`curl -XGET 'localhost:9200/_alias/*?pretty'`

### Alias index

Using multiple actions:

`curl -XPOST 'localhost/_aliases' -d ' { "actions" : [ {"remove" : { "index" : <index-name>,"alias" : <alias-name> } }, { "add" : { "index" : <index-name>, "alias" : <alias-name> } } ] }'

### Watch Index being built

Install watch command with `brew install watch`

`watch -n 2 "curl -s localhost:9200/<index>/_count?pretty"`

It's advisable using this command instead of building in log outputs of how many documents have been uploaded to elasticsearch index for the following reasons:

1) Script or process will be faster as it wont have to log outputs to terminal
2) Calling elasticsearch index to count the number of documents is a better indicator of documents that have properly indexed as elasticsearch responds to (bulk) requests with success before fully indexing the documents. The count is based on the documents actually being indexed in search.
