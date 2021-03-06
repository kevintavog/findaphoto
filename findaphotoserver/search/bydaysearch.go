package search

import (
	"github.com/ian-kent/go-log/log"
	"gopkg.in/olivere/elastic.v5"

	"github.com/kevintavog/findaphoto/common"
)

type ByDayOptions struct {
	Month            int
	DayOfMonth       int
	Index            int
	Count            int
	Random           bool
	CategoryOptions  *CategoryOptions
	DrilldownOptions *DrilldownOptions
}

//-------------------------------------------------------------------------------------------------
func NewByDayOptions(month, dayOfMonth int) *ByDayOptions {
	return &ByDayOptions{
		Month:            month,
		DayOfMonth:       dayOfMonth,
		Index:            0,
		Count:            20,
		Random:           false,
		CategoryOptions:  NewCategoryOptions(),
		DrilldownOptions: NewDrilldownOptions(),
	}
}

func (bdo *ByDayOptions) Search() (*SearchResult, error) {
	client := common.CreateClient() // consider using elastic.NewSimpleClient
	search := client.Search().
		Index(common.MediaIndexName).
		Type(common.MediaTypeName).
		Pretty(true).
		From(bdo.Index).
		Size(bdo.Count)

	grouping := GroupByDate
	dayOfYear := common.DayOfYear(bdo.Month, bdo.DayOfMonth)
	dateQuery := elastic.NewTermQuery("dayofyear", dayOfYear)
	if bdo.Random {
		grouping = GroupByAll
		randomQuery := elastic.NewFunctionScoreQuery().
			Query(dateQuery).
			AddScoreFunc(elastic.NewRandomFunction())
		search.Query(randomQuery)
	} else {
		search.Query(dateQuery)
		search.Sort("datetime", false)
	}

	result, err := invokeSearch(search, nil, grouping, bdo.CategoryOptions, bdo.DrilldownOptions, nil)
	if err == nil {
		if dayOfYear == 1 {
			// Latest day of the year with matches
			result.PreviousAvailableByDay = getAvailableDay(client, elastic.NewRangeQuery("dayofyear").Gt(dayOfYear), false)
		} else {
			result.PreviousAvailableByDay = getAvailableDay(client, elastic.NewRangeQuery("dayofyear").Lt(dayOfYear), false)
		}

		if dayOfYear == 366 {
			// The first day of the year with matches
			result.NextAvailableByDay = getAvailableDay(client, elastic.NewRangeQuery("dayofyear").Lt(dayOfYear), true)
		} else {
			result.NextAvailableByDay = getAvailableDay(client, elastic.NewRangeQuery("dayofyear").Gt(dayOfYear), true)
		}
	}

	return result, err
}

func getAvailableDay(client *elastic.Client, query *elastic.RangeQuery, ascending bool) *ByDayResult {
	search := client.Search().
		Index(common.MediaIndexName).
		Type(common.MediaTypeName).
		Pretty(true).
		Size(1).
		Sort("dayofyear", ascending).
		Query(query)

	hit, err := returnFirstMatch(search)
	if err != nil {
		log.Warn("available day search failed: %s", err.Error())
		return nil
	}

	if hit != nil {
		return &ByDayResult{Day: hit.Media.DateTime.Day(), Month: int(hit.Media.DateTime.Month())}
	}

	return nil
}
