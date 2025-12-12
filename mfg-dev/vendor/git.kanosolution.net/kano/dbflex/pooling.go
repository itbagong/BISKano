package dbflex

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sebarcode/logger"
)

var (
	DefaultPoolingTimeout = 30 * time.Second
)

// DbPooling is database pooling system in dbflex
type DbPooling struct {
	sync.RWMutex
	size  int
	items map[string]*PoolItem

	fnNew func() (IConnection, error)

	// Timeout max time required to obtain new connection
	Timeout time.Duration

	// AutoRelease max time for a connection to be auto released after it is being idle. 0 = no autorelease (default)
	AutoRelease time.Duration

	// AutoClose max time for a connection to be autoclosed after it is being idle. 0 = no auto close (default)
	AutoClose time.Duration

	_log *logger.LogEngine

	freeItemIndexes *SliceMtx[string]
	useItemIndexes  *SliceMtx[string]

	ChanSize    int
	chanClose   chan string
	chanRelease chan string
}

// NewDbPooling create new pooling with given size
func NewDbPooling(size int, fnNew func() (IConnection, error)) *DbPooling {
	dbp := new(DbPooling)
	dbp.size = size
	dbp.fnNew = fnNew
	dbp.Timeout = DefaultPoolingTimeout
	dbp.freeItemIndexes = &SliceMtx[string]{}
	dbp.freeItemIndexes.SetComparator(func(a, b string) bool {
		res := a == b
		return res
	})
	dbp.useItemIndexes = &SliceMtx[string]{}
	dbp.useItemIndexes.SetComparator(func(a, b string) bool { return a == b })

	if dbp.ChanSize == 0 {
		dbp.ChanSize = 100
	}
	dbp.chanClose = make(chan string, dbp.ChanSize)
	dbp.chanRelease = make(chan string, dbp.ChanSize)
	dbp.items = make(map[string]*PoolItem)

	go func() {
		for piID := range dbp.chanRelease {
			go func(piID string) {
				dbp.Lock()
				defer dbp.Unlock()

				pi, ok := dbp.items[piID]
				if ok {
					lastUsed := pi.lastUsed
					pi.inUse = false
					dbp.freeItemIndexes.Add(piID)
					go dbp.Log().Debugf("%s processing release signal, conn count: %d, free: %d, lapse: %s", piID, len(dbp.items), len(dbp.freeItemIndexes.slice), time.Since(lastUsed))
				}
			}(piID)
		}
	}()

	go func() {
		for piID := range dbp.chanClose {
			dbp.closeItem(dbp.items[piID])
			go dbp.Log().Debugf("%s processing close signal, conn count: %d, free: %d", piID, len(dbp.items), len(dbp.freeItemIndexes.slice))
		}
	}()

	return dbp
}

func (p *DbPooling) getFreeItem() *PoolItem {
	p.Lock()
	defer p.Unlock()

	freeItemsCount := p.freeItemIndexes.Count()
	if freeItemsCount == 0 {
		return nil
	}

	for i := 0; i < freeItemsCount; i++ {
		indexStr, err := p.freeItemIndexes.Get(i)
		if err == nil {
			p.freeItemIndexes.Remove(i)

			pi, ok := p.items[indexStr]
			if ok && pi.IsFree() {
				lastUsedTime := pi.lastUsed
				pi.Use()
				go p.Log().Debugf("%s re-use free connection, conn count: %d, free: %d, lapse: %s", pi.ID, len(p.items), len(p.freeItemIndexes.slice), time.Since(lastUsedTime))
				return pi
			}
		}
	}

	return nil
}

func (p *DbPooling) closeItem(pi *PoolItem) {
	p.Lock()
	defer p.Unlock()

	if pi != nil {
		pi.conn.Close()
	}
	delete(p.items, pi.ID)
	p.freeItemIndexes.RemoveByValue(pi.ID)
}

func (h *DbPooling) Log() *logger.LogEngine {
	if h._log == nil {
		h._log = logger.NewLogEngine(true, false, "", "", "")
	}
	return h._log
}

func (h *DbPooling) SetLog(l *logger.LogEngine) *DbPooling {
	h._log = l
	return h
}

// Get new connection. If all connection is being used and number of connection is less than
// pool capacity, new connection will be spin off. If capabity has been max out. It will waiting for
// any connection to be released before timeout reach
func (p *DbPooling) Get() (*PoolItem, error) {
	reqTime := time.Now()

	// since pool already full, we have to wait until one of the poolitem is released or timeout
	chPi := make(chan *PoolItem)
	chErr := make(chan error)
	chStopFindFreeItem := make(chan bool)

	go func() {
		for {
			select {
			case <-chStopFindFreeItem:
				return

			default:
				pi := p.getFreeItem()
				if pi != nil {
					chPi <- pi
					return
				}

				pi, err := p.createNewItemIfLessThanSize()
				if err != nil {
					chErr <- err
					return
				} else if pi != nil {
					pi.fromDbPooling(p)
					pi.Use()
					chPi <- pi
					return
				}
			}
		}
	}()

	select {
	case pi := <-chPi:
		return pi, nil

	case err := <-chErr:
		return nil, fmt.Errorf("unable to get pool item. %s, conn count: %d, free: %d", err.Error(), len(p.items), len(p.freeItemIndexes.slice))

	case <-time.After(p.Timeout):
		currentTime := time.Now()
		chStopFindFreeItem <- true
		return nil, fmt.Errorf("unable to get pool item. timeout exceeded %s, conn count: %d, free: %d, requested: %s, current: %s, diff: %s",
			p.Timeout.String(), len(p.items), len(p.freeItemIndexes.slice), reqTime, currentTime, currentTime.Sub(reqTime))
	}
}

// GetItems return pool items within connection pooling
func (p *DbPooling) GetItems() []*PoolItem {
	p.RLock()
	defer p.RUnlock()

	items := make([]*PoolItem, len(p.items))
	i := 0
	for _, pi := range p.items {
		items[i] = pi
		i++
	}

	return items
}

// Count number of connection within connection pooling
func (p *DbPooling) Count() int {
	p.RLock()
	defer p.RUnlock()

	c := len(p.items)
	return c
}

// FreeCount number of item has been released
func (p *DbPooling) FreeCount() int {
	p.RLock()
	defer p.RUnlock()
	return p.freeItemIndexes.Count()
}

// ClosedCount number of item has been closed
func (p *DbPooling) UsedCount() int {
	p.RLock()
	defer p.RUnlock()
	return len(p.items) - p.freeItemIndexes.Count()
}

// Size number of connection can be hold within the connection pooling
func (p *DbPooling) Size() int {
	return p.size
}

// Close all connection within connection pooling
func (p *DbPooling) Close() {
	p.Lock()
	defer p.Unlock()

	close(p.chanClose)
	close(p.chanRelease)

	for _, pi := range p.items {
		if pi != nil {
			pi.conn.Close()
		}
	}
	p.items = map[string]*PoolItem{}
}

func (p *DbPooling) createNewItemIfLessThanSize() (*PoolItem, error) {
	p.Lock()
	defer p.Unlock()

	poolCount := len(p.items)
	if poolCount < p.size {
		pi, err := p.newItem()
		if err != nil {
			return nil, fmt.Errorf("unable to get new item. %s", err.Error())
		}

		pi.fromDbPooling(p)
		pi.Use()
		return pi, nil
	}
	return nil, nil
}

func (p *DbPooling) newItem() (*PoolItem, error) {
	conn, err := p.fnNew()
	if err != nil {
		return nil, fmt.Errorf("unable to open connection for DB pool. %s", err.Error())
	}

	pi := &PoolItem{conn: conn, inUse: false}
	pi.SetLog(p.Log())
	pi.fromDbPooling(p)
	pi.created = time.Now()
	hashClose := uuid.New().String()
	pi.ID = hashClose
	p.items[hashClose] = pi
	go p.Log().Debugf("%s create new connection on %s, conn count: %d, free: %d, time: %s", pi.ID, time.Now(), len(p.items), len(p.freeItemIndexes.slice), time.Now())

	//-- auto release if it's enabled
	if pi.AutoRelease > 0 {
		go func(pi *PoolItem) {
			for {
				if pi == nil {
					return
				}

				<-time.After(100 * time.Millisecond)
				diff := time.Since(pi.lastUsed)
				if diff > pi.AutoRelease && !pi.IsFree() {
					pi.Release()
					return
				}
			}
		}(pi)
	}

	//-- auto close if it's enabled
	if pi.AutoClose > 0 {
		go func(pi *PoolItem) {
			for {
				if pi == nil {
					return
				}

				<-time.After(100 * time.Millisecond)
				diff := time.Since(pi.lastUsed)
				if diff > pi.AutoClose && pi.IsFree() {
					p.chanClose <- pi.ID
					return
				}
			}
		}(pi)
	}

	return pi, nil
}
