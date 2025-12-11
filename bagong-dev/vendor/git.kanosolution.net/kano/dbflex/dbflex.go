package dbflex

import (
	"errors"

	"github.com/sebarcode/logger"
)

var log *logger.LogEngine

// ErrEOF is the error returned by dbflex when document/data/row not found.
// This ErrEOF should be different with io.ErrEOF
// You can use for data not found
var ErrEOF = errors.New("EOF")

func Logger() *logger.LogEngine {
	if log == nil {
		log, _ = logger.NewLog(true, false, "", "", "")
	}
	return log
}

func SetLogger(l *logger.LogEngine) {
	log = l
}
