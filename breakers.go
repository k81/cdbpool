package cdbpool

import (
	"fmt"
	"sync"
	"time"

	"github.com/k81/kate/log"
	"github.com/sony/gobreaker"
)

var (
	breakers     = make(map[string]*gobreaker.CircuitBreaker)
	breakersLock sync.Mutex
)

func getBreaker(addr string) *gobreaker.CircuitBreaker {
	breakersLock.Lock()
	defer breakersLock.Unlock()
	breaker, ok := breakers[addr]
	if !ok {
		breaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        fmt.Sprint("circuit breaker db-", addr),
			MaxRequests: 10,
			Interval:    time.Duration(5 * time.Second),
			Timeout:     time.Duration(10 * time.Second),
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.TotalFailures >= uint32(20)
			},
			OnStateChange: func(name string, from, to gobreaker.State) {
				log.Info(mctx, "circuit breaker state changed", "name", name, "from", from, "to", to)
			},
		})
		breakers[addr] = breaker
	}
	return breaker
}
