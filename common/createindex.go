package common

import (
	"errors"

	"github.com/ian-kent/go-log/log"
	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"
)

func CreateFindAPhotoIndex(client *elastic.Client) error {
	log.Warn("Creating index '%s'", MediaIndexName)

	mapping := `{
		"settings": {
			"max_result_window": 100000,
			"number_of_shards": 1,
	        "number_of_replicas": 0
		},
		"mappings": {
			"media": {
				"_all": {
					"enabled": false
			    },
				"properties" : {
					"date" : {
					  "type" : "keyword"
					},
					"datetime" : {
					  "type" : "date",
					  "format": "yyyy-MM-dd'T'HH:mm:ssZ"
					},
					"keywords" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"dayname" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"warnings" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"placename" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"monthname" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"mimetype" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"displayname" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"filename" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"hierarchicalname" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"location" : {
					  "type" : "geo_point"
					},
					"lengthinbytes" : {
					  "type" : "long"
					},
					"path" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"signature" : {
					  "type" : "keyword"
					},
					
					"aperture" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"exposureprogram" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"exposuretime" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"flash" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"fnumber" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"focallength" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"iso" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"whitebalance" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"lensinfo" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"lensmodel" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"cameramake" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"cameramodel" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},

					"cityname" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"sitename" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"countryname" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"countrycode" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					},
					"statename" : {
					  "type" : "text",
			          "fields": {
			            "value": { 
			              "type":  "keyword"
			            }
					  }
					}

				}
			}
		}
	}`

	response, err := client.CreateIndex(MediaIndexName).BodyString(mapping).Do(context.TODO())
	if err != nil {
		return err
	}

	if response.Acknowledged != true {
		return errors.New("Index creation not acknowledged")
	}
	return nil
}
