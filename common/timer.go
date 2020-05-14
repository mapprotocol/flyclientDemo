package common

import (
	"sync"
	"time"
)

type property struct {
	ticker *time.Ticker
	timer  *time.Timer
	stop   chan struct{}
}

var recordID int
var recordLock sync.RWMutex
var record = make(map[int]property)

/*
 * SetPeriodAndRun create period timer and run handle periodically , return timer id
 * handle : user function
 * interval: time period
 */
func SetPeriodAndRun(handle func(), interval time.Duration) int {
	recordLock.Lock()
	defer recordLock.Unlock()

	recordID++

	ticker := time.NewTicker(interval)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				handle()

			case <-quit:
				return
			}
		}
	}()
	record[recordID] = property{ticker, nil, quit}
	return recordID
}

/*
 * StopWork stop periodic timer
 */
func StopWork(id int) {
	recordLock.Lock()
	defer recordLock.Unlock()

	if tm, ok := record[id]; ok {
		close(tm.stop)
		delete(record, id)
	}
}
