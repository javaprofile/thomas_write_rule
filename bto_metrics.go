package main

import (
	"fmt"
	"sync"
	"time"
)

type DataItem struct {
	value     int
	timestamp int64
}

type Transaction struct {
	id        int
	timestamp int64
}

var (
	data            = map[string]DataItem{}
	dataMutex       = sync.RWMutex{}
	transactionID   = 0
	abortReadCount  = 0
	abortWriteCount = 0
	metricsMutex    = sync.Mutex{}
)

func startTransaction() Transaction {
	transactionID++
	return Transaction{id: transactionID, timestamp: time.Now().UnixNano()}
}

func read(transaction Transaction, key string) (int, bool) {
	dataMutex.RLock()
	defer dataMutex.RUnlock()
	item, exists := data[key]
	if !exists || transaction.timestamp < item.timestamp {
		incrementAbort("read")
		return 0, false
	}
	return item.value, true
}

func write(transaction Transaction, key string, value int) bool {
	dataMutex.Lock()
	defer dataMutex.Unlock()
	item, exists := data[key]
	if exists && transaction.timestamp < item.timestamp {
		incrementAbort("write")
		return false
	}
	data[key] = DataItem{value: value, timestamp: transaction.timestamp}
	return true
}

func incrementAbort(action string) {
	metricsMutex.Lock()
	defer metricsMutex.Unlock()
	if action == "read" {
		abortReadCount++
	} else if action == "write" {
		abortWriteCount++
	}
}

func printAbortMetrics() {
	metricsMutex.Lock()
	defer metricsMutex.Unlock()
	fmt.Println("Abort Metrics:")
	fmt.Println("Read Aborts :", abortReadCount)
	fmt.Println("Write Aborts:", abortWriteCount)
}

func main() {
	t1 := startTransaction()
	write(t1, "x", 10)
	read(t1, "x")

	time.Sleep(time.Millisecond)
	t2 := startTransaction()
	write(t2, "x", 20)

	write(t1, "x", 30)
	read(t1, "x")

	printAbortMetrics()
}
