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
	data           = make(map[string]DataItem)
	dataMutex      = sync.RWMutex{}
	transactionID  = 0
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
		return 0, false
	}
	return item.value, true
}

func write(transaction Transaction, key string, value int) bool {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	item, exists := data[key]
	if exists && transaction.timestamp < item.timestamp {
		return false
	}

	data[key] = DataItem{value: value, timestamp: transaction.timestamp}
	return true
}

func main() {
	t1 := startTransaction()
	write(t1, "x", 10)
	v, ok := read(t1, "x")
	fmt.Println("Transaction 1 Read x:", v, ok)

	t2 := startTransaction()
	ok = write(t2, "x", 20)
	fmt.Println("Transaction 2 Write x:", ok)

	t3 := startTransaction()
	ok = write(t3, "x", 30)
	fmt.Println("Transaction 3 Write x (outdated):", ok)
}
