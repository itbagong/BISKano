package orm

import "git.kanosolution.net/kano/dbflex"

type ReturnKind string

const (
	ReturnSingle ReturnKind = "Single"
	ReturnMulti  ReturnKind = "Multi"
	ReturnBoth   ReturnKind = ""
)

type Query struct {
	Name       string
	Param      *dbflex.QueryParam
	ReturnKind string
}
