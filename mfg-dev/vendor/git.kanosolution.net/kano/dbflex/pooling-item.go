package dbflex

import (
	"time"

	"github.com/sebarcode/logger"
)

// PoolItem is Item in the pool
type PoolItem struct {
	conn  IConnection
	inUse bool

	created  time.Time
	lastUsed time.Time

	AutoRelease time.Duration
	AutoClose   time.Duration

	_log *logger.LogEngine
	ID   string

	chanClose   chan<- string
	chanRelease chan<- string
}

func (h *PoolItem) Log() *logger.LogEngine {
	if h._log == nil {
		h._log = logger.NewLogEngine(true, false, "", "", "")
	}
	return h._log
}

func (h *PoolItem) SetLog(l *logger.LogEngine) *PoolItem {
	h._log = l
	return h
}

func (pi *PoolItem) fromDbPooling(p *DbPooling) {
	pi._log = p.Log()
	pi.AutoClose = p.AutoClose
	pi.AutoRelease = p.AutoRelease
	pi.chanClose = p.chanClose
	pi.chanRelease = p.chanRelease
}

// Release PoolItem
func (pi *PoolItem) Release() {
	lastUsedTime := pi.lastUsed
	pi.inUse = false
	pi.lastUsed = time.Now()
	pi.chanRelease <- pi.ID
	go pi.Log().Debugf("%s send release signal, lapse: %s", pi.ID, time.Since(lastUsedTime))
}

// IsFree check and return true if PoolItem is free
func (pi *PoolItem) IsFree() bool {
	return !pi.inUse
}

// Use mark that this PoolItem is used
func (pi *PoolItem) Use() {
	pi.inUse = true
	pi.lastUsed = time.Now()
}

// Connection return PoolItem connection
func (pi *PoolItem) Connection() IConnection {
	return pi.conn
}
