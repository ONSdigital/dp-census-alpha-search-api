# Scripts

A list of scripts which load data into elasticsearch for use in the Search API.

## A list of scripts

- [load postcodes](#load-postcode)
- [geojson](#load-data-from-geojson-files)
    - 2011 Lower Layer Super Output Areas (LSOA)
    - 2011 Middle Layer Super Output Areas (MSOA)
    - 2011 Output Areas (OA)
    - 2015 Towns and Cities (TCITY)

### Load Postcode

This script loads postcode data for all postcodes across the UK as of Febraury 2020 from a csv file downloaded from the geo portal [here](https://geoportal.statistics.gov.uk/datasets/national-statistics-postcode-lookup-february-2020).

Once file is downloaded (from above link), unzip file. The data layout to postcode data should look like this:
    - NSPL_FEB_2020_UK
      - Data
        - NSPL_FEB_2020_UK.csv

Upload postcode data to elasticsearch index with:
`make postcode`
This will take approximately 4 minutes and 20 seconds and documents will be stored in `test_postcode` index.

### Load data from GEOJSON files

This script loads geographical boundaries for 2011 census data. This includes lower and middle layer output areas, as well as other output areas, towns and cities across England and Wales only.

Files can be downloaded from the geoportal -> boundaries -> census boundaries -> select geography layer. This will tend to open up a search of all relevant boundaries, select the data you would like to view/import. The new screen will have a drop down list to the right of webpage titled `APIs`, click the drop down and copy the GEOJSON url. Paste the url into the browser and it will automatically download the data, be patient this may take some time; below is a list of urls used for the geojson scripts (these might break if geoportal decides to move the geojson file location):

- [Major Towns And Cities 2015](https://opendata.arcgis.com/datasets/58b0dfa605d5459b80bf08082999b27c_0.geojson)
- [Middle layer super output areas december 2011 ew-bfc](https://opendata.arcgis.com/datasets/02aa733fc3414b0ea4179899e499918d_0.geojson)
- [Middle layer super output areas december 2011 ew-bfe](https://opendata.arcgis.com/datasets/f185143921f445cda15d37e2b9d61c3e_1.geojson)
- [Middle layer super output areas december 2011 ew-bgc](https://opendata.arcgis.com/datasets/29fdaa2efced40378ce8173b411aeb0e_2.geojson)
- [Middle layer super output areas december 2011 ew-bsc](https://opendata.arcgis.com/datasets/c661a8377e2647b0bae68c4911df868b_3.geojson)
- [Lower layer super output areas december 2011 ew-bfc](https://opendata.arcgis.com/datasets/e886f1cd40654e6b94d970ecf437b7b5_0.geojson)
- [Lower layer super output areas december 2011 ew-bfe](https://opendata.arcgis.com/datasets/763196a293304551958fffdaa87cc6d9_0.geojson)
- [Lower layer super output areas december 2011 ew-bgc](https://opendata.arcgis.com/datasets/e993add3f1944437bc91ec7c76100c63_0.geojson)
- [Lower layer super output areas december 2011 ew-bsc](https://opendata.arcgis.com/datasets/007577eeb8e34c62a1844df090a93128_0.geojson)
- [Output areas december 2011 ew-bfc](https://opendata.arcgis.com/datasets/ff8151d927974f349de240e7c8f6c140_0.geojson)
- [Output areas december 2011 ew-bfe](https://opendata.arcgis.com/datasets/d74074ae6dec4de59fdcd2744fefc1f9_0.geojson)
- [Output areas december 2011 ew-bgc](https://opendata.arcgis.com/datasets/f79fc19485704ce68523d8d70d84a913_0.geojson)

Once the above files have downloaded, move the files to root of this repository and store under geojson folder.

Upload all the data to elasticsearch index with:
`make geojson`
This will take a long time as it i populates 700,000+ records with full polygon boundaries into elasticsearch `test_geo` index.

There are actually four separate scripts which handle generating data for LSOA, MSOA, OA and TCITY files. These can be run separately using `make lsoa`, `make msoa`, `make oa`, `make tcity` respectively. Be aware that if you are running this for the first time you will need to create the `test_geo` index, you can do this by running `make refreshgeojson`.

The refresh script deletes the index and recreates it with 0 data.
