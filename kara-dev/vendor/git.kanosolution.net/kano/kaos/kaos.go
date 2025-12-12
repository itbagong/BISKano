package kaos

import "github.com/ariefdarmawan/datahub"

type GetHubFn func(*Context) *datahub.Hub

type EndPointNamingEnum string

const (
	NamingAsIs    EndPointNamingEnum = "ASIS"
	NamingIsLower EndPointNamingEnum = "LOWER"
)

var NamingType EndPointNamingEnum = NamingAsIs
var NamingJoiner = "-"
