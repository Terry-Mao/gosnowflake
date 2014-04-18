package main

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"sync"
	"time"
)

const (
	twepoch            = int64(1288834974657)
	workerIdBits       = uint(5)
	datacenterIdBits   = uint(5)
	maxWorkerId        = -1 ^ (-1 << workerIdBits)
	maxDatacenterId    = -1 ^ (-1 << datacenterIdBits)
	sequenceBits       = uint(12)
	workerIdShift      = sequenceBits
	datacenterIdShift  = sequenceBits + workerIdBits
	timestampLeftShift = sequenceBits + workerIdBits + datacenterIdBits
	sequenceMask       = -1 ^ (-1 << sequenceBits)
)

type IdWorker struct {
	sequence      int64
	workerId      int64
	datacenterId  int64
	lastTimestamp int64
	mutex         *sync.Mutex
}

// NewIdWorker new a snowflake id generator object.
func NewIdWorker(workerId, datacenterId int64) (*IdWorker, error) {
	idWorker := &IdWorker{}
	if workerId > maxWorkerId || workerId < 0 {
		glog.Errorf("worker Id can't be greater than %d or less than 0", maxWorkerId)
		return nil, errors.New(fmt.Sprintf("worker Id: %d error", workerId))
	}
	if datacenterId > maxDatacenterId || datacenterId < 0 {
		glog.Errorf("datacenter Id can't be greater than %d or less than 0", maxDatacenterId)
		return nil, errors.New(fmt.Sprintf("datacenter Id: %d error", datacenterId))
	}
	idWorker.workerId = workerId
	idWorker.datacenterId = datacenterId
	idWorker.lastTimestamp = -1
	idWorker.sequence = 0
	idWorker.mutex = &sync.Mutex{}
	glog.Infof("worker starting. timestamp left shift %d, datacenter id bits %d, worker id bits %d, sequence bits %d, workerid %d", timestampLeftShift, datacenterIdBits, workerIdBits, sequenceBits, workerId)
	return idWorker, nil
}

// timeGen generate a unix millisecond.
func timeGen() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// tilNextMillis spin wait till next millisecond.
func tilNextMillis(lastTimestamp int64) int64 {
	timestamp := timeGen()
	for timestamp <= lastTimestamp {
		timestamp = timeGen()
	}
	return timestamp
}

// NextId get a snowflake id.
func (id *IdWorker) NextId() (int64, error) {
	id.mutex.Lock()
	defer id.mutex.Unlock()
	timestamp := timeGen()
	if timestamp < id.lastTimestamp {
		glog.Errorf("clock is moving backwards.  Rejecting requests until %d.", id.lastTimestamp)
		return 0, errors.New(fmt.Sprintf("Clock moved backwards.  Refusing to generate id for %d milliseconds", id.lastTimestamp-timestamp))
	}
	if id.lastTimestamp == timestamp {
		id.sequence = (id.sequence + 1) & sequenceMask
		if id.sequence == 0 {
			timestamp = tilNextMillis(id.lastTimestamp)
		}
	} else {
		id.sequence = 0
	}
	id.lastTimestamp = timestamp
	return ((timestamp - twepoch) << timestampLeftShift) | (id.datacenterId << datacenterIdShift) | (id.workerId << workerIdShift) | id.sequence, nil
}
