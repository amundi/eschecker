package queries

import (
	"github.com/amundi/escheck/config"
	"gopkg.in/olivere/elastic.v2"
)

// The query interface's noble goal is to set a query, analyse the results, and
// do something.
type Query interface {
	SetQueryConfig(config.ManualQueryList) bool
	BuildQuery() (elastic.Query, error)
	CheckCondition(*elastic.SearchResult) bool
	DoAction(*elastic.SearchResult) error
	OnAlertEnd() error
}
