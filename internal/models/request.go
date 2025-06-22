package models

import (
	"time"
)

type RssRequestParams struct {
	Limit *int
	Date  *time.Time
}
