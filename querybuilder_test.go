package main

import (
	"github.com/amundi/escheck/config"
	"github.com/amundi/escheck/eslog"
	"github.com/stretchr/testify/assert"
	"gopkg.in/olivere/elastic.v2"
	"testing"
)

func Test_getRangeFilter(t *testing.T) {
	eslog.InitSilent()
	var result elastic.Filter
	var err error

	test := []interface{}{"errorcode", "gte", "500"}
	real := elastic.NewRangeFilter("errorcode").Gte(500)
	result, err = getRangeFilter(test)
	assert.Equal(t, err, nil, "Should be nil")
	assert.Equal(t, real, result)

	test = []interface{}{"timestamp", "lt", "now-1h"}
	real = elastic.NewRangeFilter("timestamp").Lt("now-1h")
	result, err = getRangeFilter(test)
	assert.Equal(t, err, nil, "Should be nil")
	assert.Equal(t, real, result)

	test = []interface{}{"code", "lt", "800", "gte", "500"}
	real = elastic.NewRangeFilter("code").Lt(800).Gte(500)
	result, err = getRangeFilter(test)
	assert.Equal(t, err, nil, "Should be nil")
	assert.Equal(t, real, result)

	test = []interface{}{"timestamp", "lt", "now-1h", "gte", "2d"}
	real = elastic.NewRangeFilter("timestamp").Lt("now-1h").Gte("2d")
	result, err = getRangeFilter(test)
	assert.Equal(t, err, nil, "Should be nil")
	assert.Equal(t, real, result)

	test = []interface{}{"code", "lt", "800", "pouet", "500"}
	result, err = getRangeFilter(test)
	assert.NotNil(t, err)

	test = []interface{}{"code", "lt", "800", "gte"}
	result, err = getRangeFilter(test)
	assert.NotNil(t, err)

	test = []interface{}{"test", "lt", "800", "tge", "500"}
	result, err = getRangeFilter(test)
	assert.NotNil(t, err)

	test = []interface{}{"testshort"}
	result, err = getRangeFilter(test)
	assert.NotNil(t, err)

	test = []interface{}{}
	result, err = getRangeFilter(test)
	assert.NotNil(t, err)
}

func Test_getFilters(t *testing.T) {
	eslog.InitSilent()

	// term filters
	var filters []interface{} = []interface{}{
		map[interface{}]interface{}{"term": []interface{}{"test", "yes"}},
		map[interface{}]interface{}{"term": []interface{}{"required", true}},
	}

	realfilters := []elastic.Filter{
		elastic.NewTermFilter("test", "yes"),
		elastic.NewTermFilter("required", true),
	}

	testfilters, err := getFilters(filters)
	assert.Equal(t, err, nil, "Should be nil")
	assert.Equal(t, realfilters, testfilters)

	realfilters = []elastic.Filter{
		elastic.NewTermFilter("test", "yes"),
		elastic.NewTermFilter("required", false),
	}
	assert.NotEqual(t, realfilters, testfilters, "Should not be equal")

	//range filters
	filters = []interface{}{
		map[interface{}]interface{}{"term": []interface{}{"value", 146}},
		map[interface{}]interface{}{"term": []interface{}{"othervalue", "testTest"}},
		map[interface{}]interface{}{"range": []interface{}{"Timestamp", "gte", "now-1h"}},
	}

	realfilters = []elastic.Filter{
		elastic.NewTermFilter("value", 146),
		elastic.NewTermFilter("othervalue", "testTest"),
		elastic.NewRangeFilter("Timestamp").Gte("now-1h"),
	}

	testfilters, err = getFilters(filters)
	assert.Equal(t, err, nil, "Should be nil")
	assert.Equal(t, realfilters, testfilters)

	realfilters = []elastic.Filter{
		elastic.NewTermFilter("value", 146),
		elastic.NewTermFilter("othervalue", "testTest"),
		elastic.NewRangeFilter("Timestamp").Lt("now-1h"),
	}

	assert.NotEqual(t, realfilters, testfilters, "Should not be equal")

	filters = []interface{}{
		map[interface{}]interface{}{"term": []interface{}{"value", 146}},
		map[interface{}]interface{}{"term": []interface{}{"othervalue", "testTest"}},
		map[interface{}]interface{}{"range": []interface{}{"Timestamp", "lt", "now-1h"}},
	}
	testfilters, err = getFilters(filters)
	assert.Equal(t, err, nil, "Should be nil")
	assert.Equal(t, realfilters, testfilters)

	//non valid fields
	filters = []interface{}{
		map[interface{}]interface{}{"term": []interface{}{145, "yes"}},
		map[interface{}]interface{}{"term": []interface{}{"required", "yes"}},
	}
	testfilters, err = getFilters(filters)
	assert.NotNil(t, err, "Shoud be not nil")

	filters = []interface{}{
		map[interface{}]interface{}{"term": []interface{}{"yes"}},
		map[interface{}]interface{}{"term": []interface{}{"required", 28}},
	}
	testfilters, err = getFilters(filters)
	assert.NotNil(t, err, "Shoud be not nil")

	filters = []interface{}{
		map[interface{}]interface{}{"range": []interface{}{"Timestamp", "gte"}},
		map[interface{}]interface{}{"term": []interface{}{"required", 28}},
	}

	testfilters, err = getFilters(filters)
	assert.NotNil(t, err, "Shoud be not nil")

	filters = []interface{}{
		map[interface{}]interface{}{"range": []interface{}{42, "gte", "now-30m"}},
		map[interface{}]interface{}{"term": []interface{}{"required", 28}},
	}

	testfilters, err = getFilters(filters)
	assert.NotNil(t, err, "Shoud be not nil")
}

func Test_Boolfilter(t *testing.T) {
	eslog.InitSilent()
	info := &config.QueryInfo{}
	myQuery, err := computeQuery(info)
	assert.NotNil(t, err)

	queryInfo := &config.QueryInfo{
		Type: "boolfilter",
		Clauses: map[string]interface{}{
			"must": []interface{}{
				map[interface{}]interface{}{"term": []interface{}{"Value", 146.5}},
				map[interface{}]interface{}{"term": []interface{}{"othervalue", "testTest"}},
				map[interface{}]interface{}{"range": []interface{}{"Timestamp", "lt", "now-1h"}},
			},
			"must_not": []interface{}{
				map[interface{}]interface{}{"term": []interface{}{"status", "OK"}},
			},
		},
	}
	realQuery := elastic.NewBoolFilter().Must(
		elastic.NewTermFilter("Value", 146.5),
		elastic.NewTermFilter("othervalue", "testTest"),
		elastic.NewRangeFilter("Timestamp").Lt("now-1h"),
	).MustNot(
		elastic.NewTermFilter("status", "OK"),
	)
	myQuery, err = computeQuery(queryInfo)
	assert.Nil(t, err)
	assert.Equal(t, realQuery, myQuery, "Shoud be equal")

	queryInfo = &config.QueryInfo{
		Type: "boolfilter",
		Clauses: map[string]interface{}{
			"must_not": []interface{}{
				map[interface{}]interface{}{"term": []interface{}{"Status", "Error"}},
				map[interface{}]interface{}{"range": []interface{}{"Timestamp", "gt", "now-30m"}},
			},
		},
	}

	realQuery = elastic.NewBoolFilter().MustNot(
		elastic.NewTermFilter("Status", "Error"),
	).MustNot(
		elastic.NewRangeFilter("Timestamp").Gt("now-30m"),
	)

	myQuery, err = computeQuery(queryInfo)
	assert.Nil(t, err)
	assert.Equal(t, realQuery, myQuery, "Shoud be equal")

	queryInfo = &config.QueryInfo{
		Type: "boolfilter",
		Clauses: map[string]interface{}{
			"should": []interface{}{
				map[interface{}]interface{}{"term": []interface{}{"User", "Thomas"}},
			},
			"must": []interface{}{
				map[interface{}]interface{}{"term": []interface{}{"Status", "OK"}},
				map[interface{}]interface{}{"range": []interface{}{"Timestamp", "gt", "now-2h"}},
			},
		},
	}

	realQuery = elastic.NewBoolFilter().Should(
		elastic.NewTermFilter("User", "Thomas"),
	).Must(
		elastic.NewTermFilter("Status", "OK"),
		elastic.NewRangeFilter("Timestamp").Gt("now-2h"),
	)

	myQuery, err = computeQuery(queryInfo)
	assert.Nil(t, err)
	assert.Equal(t, realQuery, myQuery, "Shoud be equal")

	realQuery = elastic.NewBoolFilter().Should(
		elastic.NewTermFilter("User", "Tobias"),
	).Must(
		elastic.NewTermFilter("Status", "OK"),
		elastic.NewRangeFilter("Timestamp").Gt("now-2h"),
	)
	assert.NotEqual(t, realQuery, myQuery, "Shoud not be equal")

	queryInfo = &config.QueryInfo{
		Type: "boolfilter",
		Clauses: map[string]interface{}{
			"should": []interface{}{
				map[interface{}]interface{}{"plop": []interface{}{"Thomas"}},
			},
			"must": []interface{}{
				map[interface{}]interface{}{"term": []interface{}{"Status", "OK"}},
				map[interface{}]interface{}{"hihi": []interface{}{"Timestamp", "gt", "now-2h"}},
			},
		},
	}
	myQuery, err = computeQuery(queryInfo)
	assert.NotNil(t, err)
}

func TestQueryString(t *testing.T) {

	queryInfo := &config.QueryInfo{
		Type: "query_string",
		Clauses: map[string]interface{}{
			"query": "type:MySQL AND Timestamp [2012-01-01 TO 2012-12-31]",
		},
	}
	myQuery, err := computeQuery(queryInfo)
	assert.Equal(t, err, nil)
	realQuery := elastic.NewQueryStringQuery("type:MySQL AND Timestamp [2012-01-01 TO 2012-12-31]").AnalyzeWildcard(false)
	assert.Equal(t, realQuery, myQuery)

	queryInfo = &config.QueryInfo{
		Type: "querystring",
		Clauses: map[string]interface{}{
			"query":             "this OR (that OR thi*)",
			"analyze_wildcards": true,
		},
	}
	myQuery, err = computeQuery(queryInfo)
	assert.Equal(t, err, nil)
	realQuery = elastic.NewQueryStringQuery("this OR (that OR thi*)").AnalyzeWildcard(true)
	assert.Equal(t, realQuery, myQuery)

	queryInfo = &config.QueryInfo{
		Type: "querystring",
		Clauses: map[string]interface{}{
			"query":             "this OR (that OR this)",
			"analyze_wildcards": 42,
		},
	}
	myQuery, err = computeQuery(queryInfo)
	assert.Equal(t, err, nil)
	realQuery = elastic.NewQueryStringQuery("this OR (that OR this)").AnalyzeWildcard(false)
	assert.Equal(t, realQuery, myQuery)

	//ERRORS
	queryInfo = &config.QueryInfo{
		Type: "querystring",
		Clauses: map[string]interface{}{
			"quer": "type:Error AND method:GET",
		},
	}
	myQuery, err = computeQuery(queryInfo)
	assert.NotNil(t, err)

	queryInfo = &config.QueryInfo{
		Type: "plop",
		Clauses: map[string]interface{}{
			"query": "type:Error AND method:GET",
		},
	}
	myQuery, err = computeQuery(queryInfo)
	assert.NotNil(t, err)

}

func TestStringtoNb(t *testing.T) {
	t1 := "40.7"
	t2 := 3.3
	t3 := "100"
	t4 := "nothing to do here"
	t5 := "now-30m"
	t6 := "100mille"

	r1 := stringToNb(t1)
	r2 := stringToNb(t2)
	r3 := stringToNb(t3)
	r4 := stringToNb(t4)
	r5 := stringToNb(t5)
	r6 := stringToNb(t6)
	assert.Equal(t, 40.7, r1)
	assert.Equal(t, 3.3, r2)
	assert.Equal(t, 100, r3)
	assert.Equal(t, "nothing to do here", r4)
	assert.Equal(t, "now-30m", r5)
	assert.Equal(t, "100mille", r6)

	query1 := elastic.NewBoolFilter().Should(
		elastic.NewTermFilter("User", "Thomas"),
	).Must(
		elastic.NewTermFilter("Status", stringToNb("500")),
		elastic.NewRangeFilter("Code").Gt(stringToNb("42.42")),
	)

	queryWrong := elastic.NewBoolFilter().Should(
		elastic.NewTermFilter("User", "Thomas"),
	).Must(
		elastic.NewTermFilter("Status", "500"),
		elastic.NewRangeFilter("Code").Gt("42.42"),
	)

	query2 := elastic.NewBoolFilter().Should(
		elastic.NewTermFilter("User", "Thomas"),
	).Must(
		elastic.NewTermFilter("Status", 500),
		elastic.NewRangeFilter("Code").Gt(42.42),
	)
	assert.NotEqual(t, queryWrong, query2, "Should not be equal")
	assert.Equal(t, query1, query2)
}
