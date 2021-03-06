package api

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo"
	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/search"
	"github.com/kevintavog/findaphoto/findaphotoserver/util"
)

type ReindexMediaFunction func(bool)

var ReindexMedia ReindexMediaFunction
var FindAPhotoVersionNumber string

type PathAndDate struct {
	Path        string     `json:"path,omitempty"`
	LastIndexed *time.Time `json:"lastIndexed,omitempty"`
}

var fieldsAggregateToStringFormat = map[string]string{
	"aperture":                     "%1.1f",
	"cachedlocationdistancemeters": "%1.1f",
	"dayofyear":                    "%1.f",
	"durationseconds":              "%1.3f",
	"exposuretime":                 "%1.3f",
	"fnumber":                      "%1.1f",
	"focallengthmm":                "%1.1f",
	"iso":                          "%1.f",
	"lengthinbytes":                "%1.f",
	"height":                       "%1.f",
	"width":                        "%1.f",
}

var fieldsAggregateDisallowed = map[string]bool{
	"location": true,
}

var fieldsNotExposed = map[string]bool{
	"location":            true,
	"originalcameramake":  true,
	"originalcameramodel": true,
}

func indexAPI(c echo.Context) error {
	fc := c.(*util.FpContext)
	propertiesFilter := getStatsPropertiesFilter(c.QueryParam("properties"))

	return fc.Time("index", func() error {
		props := make(map[string]interface{})
		client := common.CreateClient()

		for _, name := range propertiesFilter {
			v := getValue(name, client)
			if v != nil {
				props[name] = v
			}
		}

		return c.JSON(http.StatusOK, props)
	})
}

func reindexAPI(c echo.Context) error {
	fc := c.(*util.FpContext)
	return fc.Time("reindex", func() error {
		force := fc.BoolFromQuery("force", false)
		fc.LogBool("force", force)
		ReindexMedia(force)
		return c.NoContent(http.StatusNoContent)
	})
}

func indexFieldValuesAPI(c echo.Context) error {
	fc := c.(*util.FpContext)
	return fc.Time("", func() error {
		fieldNames := make([]string, 0)
		field := c.QueryParam("fields")
		if len(field) < 1 {
			panic(&util.InvalidRequest{Message: "'fields' query parameter missing"})
		} else {
			fieldNames = strings.Split(field, ",")
		}

		searchText := c.QueryParam("q")
		month := c.QueryParam("month")
		day := c.QueryParam("day")
		maxCount := fc.IntFromQuery("max", 20)

		fc.LogStringArray("fields", fieldNames)

		drilldownOptions := search.NewDrilldownOptions()
		populateDrilldownOptions(fc, drilldownOptions)

		fieldsAndValues := getTopFieldValues(fieldNames, maxCount, searchText, month, day, drilldownOptions)

		response := make(map[string]interface{})
		response["fields"] = fieldsAndValues
		return c.JSON(http.StatusOK, response)
	})
}

func getValue(name string, client *elastic.Client) interface{} {
	switch strings.ToLower(name) {

	case "dependencyinfo":
		return getDependencyInfo(client)

	case "duplicatecount":
		return getDuplicateCount(client)

	case "fields":
		return getMappedFields()

	case "imagecount":
		return getCountsSearch(client, "mimetype:image*")

	case "paths":
		return getAliasedPaths()

	case "versionnumber":
		return FindAPhotoVersionNumber

	case "videocount":
		return getCountsSearch(client, "mimetype:video*")

	case "warningcount":
		return getCountsSearch(client, "warnings:*")
	}

	panic(&util.InvalidRequest{Message: fmt.Sprintf("Unknown property: '%s'", name)})
}

func getAliasedPaths() []PathAndDate {
	allPaths := make([]PathAndDate, 0)

	common.VisitAllPaths(func(alias common.AliasDocument) {
		pd := &PathAndDate{
			Path: alias.Path,
		}
		if !alias.DateLastIndexed.IsZero() {
			pd.LastIndexed = &alias.DateLastIndexed
		}
		allPaths = append(allPaths, *pd)
	})
	return allPaths
}

func getCountsSearch(client *elastic.Client, query string) string {
	search := client.Search().
		Index(common.MediaIndexName).
		Type(common.MediaTypeName).
		Query(elastic.NewQueryStringQuery(query)).
		From(0).
		Size(1).
		Pretty(true)

	result, err := search.Do(context.TODO())
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d", result.TotalHits())
}

func getDuplicateCount(client *elastic.Client) string {
	count, err := client.Count().
		Index(common.MediaIndexName).
		Type(common.DuplicateTypeName).
		Do(context.TODO())
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d", count)
}

func getDependencyInfo(client *elastic.Client) map[string]interface{} {
	dependencies := make(map[string]interface{})

	dependencies["elasticSearch"] = getElasticSearchDependencyInfo(client)

	return dependencies
}

func getElasticSearchDependencyInfo(client *elastic.Client) map[string]interface{} {
	info := make(map[string]interface{})

	info["index"] = common.MediaIndexName

	pingResult, httpStatusCode, err := client.Ping(common.ElasticSearchServer).Do(context.TODO())
	info["httpStatusCode"] = httpStatusCode
	if err != nil {
		info["error"] = err.Error()
	} else {
		info["version"] = pingResult.Version.Number

		healthResult, err := elastic.NewClusterHealthService(client).Index(common.MediaIndexName).Do(context.TODO())
		if err != nil {
			info["indexStatus"] = "error"
			info["indexError"] = err.Error()
		} else {
			info["indexStatus"] = healthResult.Status
		}
	}

	return info
}

func getStatsPropertiesFilter(propertiesFilter string) []string {
	if propertiesFilter == "" {
		return []string{"versionNumber"}
	}
	return strings.Split(propertiesFilter, ",")
}

func getMappedFields() []string {
	client := common.CreateClient()
	results, err := client.GetMapping().
		Index(common.MediaIndexName).
		Type(common.MediaTypeName).
		Do(context.TODO())
	if err != nil {
		panic(&util.InvalidRequest{Message: "Failed searching for mappings", Err: err})
	}

	allFields := make([]string, 0)

	// We expect a single index...
	index := results[common.MediaIndexName].(map[string]interface{})
	mappings := index["mappings"].(map[string]interface{})
	mediaType := mappings[common.MediaTypeName].(map[string]interface{})
	properties := mediaType["properties"].(map[string]interface{})
	for k := range properties {
		if _, ignored := fieldsNotExposed[k]; !ignored {
			allFields = append(allFields, k)
		}
	}

	sort.Strings(allFields)
	return allFields
}

func getTopFieldValues(fieldNames []string, maxCount int, searchText string, monthString string,
	dayString string, drilldownOptions *search.DrilldownOptions) []interface{} {

	var query elastic.Query
	if len(monthString) > 0 || len(dayString) > 0 {
		if len(searchText) > 0 {
			panic(&util.InvalidRequest{Message: "Either 'q' OR 'month' & 'day' should be specified, not both"})
		}

		month := util.IntFromString("month", monthString)
		day := util.IntFromString("day", dayString)

		query = elastic.NewTermQuery("dayofyear", common.DayOfYear(month, day))
	} else {
		// This is gross - for reasons I don't understand, when using the match all query, the field
		// enumeration/aggregations come back empty when combined with drilldowns.
		// But using the wildcard works...
		if searchText == "" {
			searchText = "*"
		}
		//		query = elastic.NewMatchAllQuery()
		//	} else {
		query = elastic.NewQueryStringQuery(searchText).
			Field("path"). // Folder name
			Field("monthname").
			Field("dayname").
			Field("keywords").
			Field("placename"). // Full reverse location lookup
			Field("tags")
		//	}
	}

	fieldInfo := make([]interface{}, 0)

	client := common.CreateClient()
	searchService := client.Search().
		Index(common.MediaIndexName).
		Type(common.MediaTypeName).
		Size(0).
		Query(query)

	for _, name := range fieldNames {
		internalName, _ := common.GetIndexFieldName(name)
		if _, notSupported := fieldsAggregateDisallowed[internalName]; notSupported {
			fmt.Printf("Unsupported field '%s' ('%s')\n", internalName, name)
			return fieldInfo
		}

		searchService.Aggregation(name, elastic.NewTermsAggregation().Field(internalName).Size(maxCount))
	}

	search.AddDrilldown(searchService, &query, drilldownOptions)

	result, err := searchService.Do(context.TODO())

	if err != nil {
		panic(&util.InvalidRequest{Message: "Failed searching for field values", Err: err})
	}

	for _, name := range fieldNames {
		values := make([]interface{}, 0)
		fieldValues, found := result.Aggregations.Terms(name)
		if !found {
			continue
		}

		for _, bucket := range fieldValues.Buckets {
			// datetime needs to be converted to a Date
			internalName, _ := common.GetIndexFieldName(name)
			value := ""
			if internalName == "datetime" {
				msec := int64(bucket.Key.(float64))
				value = fmt.Sprintf("%s", time.Unix(msec/1000, 0))
			} else {
				format, isSet := fieldsAggregateToStringFormat[internalName]
				if !isSet {
					format = "%s"
				}

				value = fmt.Sprintf(format, bucket.Key)
			}

			values = append(values, map[string]interface{}{"value": value, "count": bucket.DocCount})
		}

		fv := make(map[string]interface{})
		fv["name"] = name
		fv["values"] = values

		fieldInfo = append(fieldInfo, fv)
	}

	return fieldInfo
}
