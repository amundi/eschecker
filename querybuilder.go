package main

import (
	"errors"
	"github.com/amundi/escheck/config"
	"gopkg.in/olivere/elastic.v2"
	"strconv"
)

/*
** The querybuilder parse Elastic queries from a map of interfaces.
 */

func getRangeFilter(v []interface{}) (elastic.RangeFilter, error) {
	var ret elastic.RangeFilter

	if size := len(v); size < 3 {
		return ret, errors.New("not enough values for rangefilter")
	}

	//get name of the field
	if name, ok := v[0].(string); ok {
		ret = elastic.NewRangeFilter(name)
	} else {
		return ret, errors.New("Range filter : first parameter must be a string")
	}
	//get methods to apply
	methods := v[1:]
	if len(methods)%2 != 0 {
		return ret, errors.New("Range filter : Wrong number of parameters")
	}

	for len(methods) > 0 {
		method, ok := methods[0].(string)
		if !ok {
			return ret, errors.New("Range filter : parameter must be a string")
		}
		val := stringToNb(methods[1])
		switch method {
		case "gt":
			ret = ret.Gt(val)
		case "gte":
			ret = ret.Gte(val)
		case "lt":
			ret = ret.Lt(val)
		case "lte":
			ret = ret.Lte(val)
		default:
			return ret, errors.New("method not (yet) supported, only: gt, gte, lt, lte")
		}
		methods = methods[2:]
	}
	return ret, nil
}

func getFilters(filters []interface{}) ([]elastic.Filter, error) {
	var ret []elastic.Filter

	for i := 0; i < len(filters); i++ {
		term, ok := filters[i].(map[interface{}]interface{})
		if ok {
			for k, values := range term {
				v, ok := values.([]interface{})
				typ, ok2 := k.(string)
				if ok && ok2 {
					switch typ {
					case "term":
						if len(v) < 2 {
							return nil, errors.New("not enough values for termfilter")
						}
						if name, ok := v[0].(string); ok {
							ret = append(ret, elastic.NewTermFilter(name, v[1]))
						} else {
							return nil, errors.New("termfilter: first value of array must be a string")
						}
					case "range":
						rangeFilter, err := getRangeFilter(v)
						if err != nil {
							return nil, err
						}
						ret = append(ret, rangeFilter)
					default:
						return nil, errors.New("filter not (yet) supported")
					}
				} else {
					return nil, errors.New("wrong types for query")
				}
			}
		} else {
			return nil, errors.New("filter badly formatted")
		}
	}
	return ret, nil
}

func boolFilter(clauses map[string]interface{}) (elastic.Query, error) {
	var mustFilters, mustNotFilters, shouldFilters []elastic.Filter
	var err error

	//get must, must not, should clauses, if presents
	if mustClauses, ok := clauses["must"].([]interface{}); ok {
		mustFilters, err = getFilters(mustClauses)
		if err != nil {
			return nil, err
		}
	}

	if mustNotClauses, ok := clauses["must_not"].([]interface{}); ok {
		mustNotFilters, err = getFilters(mustNotClauses)
		if err != nil {
			return nil, err
		}
	}

	if shouldClauses, ok := clauses["should"].([]interface{}); ok {
		shouldFilters, err = getFilters(shouldClauses)
		if err != nil {
			return nil, err
		}
	}

	if len(mustFilters) == 0 && len(mustNotFilters) == 0 && len(shouldFilters) == 0 {
		// empty query, might be an Error
		return nil, errors.New("no filters specified, query building failed")
	}

	return elastic.NewBoolFilter().Must(mustFilters...).
		MustNot(mustNotFilters...).
		Should(shouldFilters...), nil
}

func queryString(clauses map[string]interface{}) (elastic.Query, error) {
	var ok bool
	analyzeWildcard := false
	var query string

	// do not allow emty strings
	if query, ok = clauses["query"].(string); !ok || query == "" {
		return nil, errors.New("missing query parameter in query string")
	}
	if wildcard, ok := clauses["analyze_wildcards"].(bool); ok {
		analyzeWildcard = wildcard
	}
	return elastic.NewQueryStringQuery(query).AnalyzeWildcard(analyzeWildcard), nil
}

func computeQuery(queryInfo *config.QueryInfo) (elastic.Query, error) {
	if queryInfo == nil || queryInfo.Type == "manual" {
		return nil, errors.New("no query info or query is not an autoquery")
	}
	switch queryInfo.Type {
	case "boolfilter":
		return boolFilter(queryInfo.Clauses)
	case "query_string", "querystring":
		return queryString(queryInfo.Clauses)
	}
	return nil, errors.New("type of query unknown")
}

func stringToNb(value interface{}) interface{} {
	switch t := value.(type) {
	case string:
		if ret, err := strconv.Atoi(t); err == nil {
			return ret
		} else if ret, err := strconv.ParseFloat(t, 64); err == nil {
			return ret
		} else {
			return value
		}
	default:
		return value
	}
}
