// Copyright © 2014 Terry Mao, LiuDing All rights reserved.
// This file is part of gosnowflake.

// gosnowflake is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// gosnowflake is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with gosnowflake.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	log "code.google.com/p/log4go"
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"time"
)

type SnowflakeRPC struct {
	idWorkers []*IdWorker
}

// StartRPC start rpc listen.
func InitRPC() error {
	if err := SanityCheckPeers(); err != nil {
		log.Error("SanityCheckPeers() error(%v)", err)
		return err
	}
	idWorkers := make([]*IdWorker, maxWorkerId)
	for _, workerId := range MyConf.WorkerId {
		if t := idWorkers[workerId]; t != nil {
			log.Error("init workerId: %d already exists", workerId)
			return fmt.Errorf("init workerId: %d exists", workerId)
		}
		idWorker, err := NewIdWorker(workerId, MyConf.DatacenterId)
		if err != nil {
			log.Error("NewIdWorker(%d, %d) error(%v)", MyConf.DatacenterId, workerId)
			return err
		}
		idWorkers[workerId] = idWorker
		if err := RegWorkerId(workerId); err != nil {
			log.Error("RegWorkerId(%d) error(%v)", workerId, err)
			return err
		}
	}
	s := &SnowflakeRPC{idWorkers: idWorkers}
	rpc.Register(s)
	for _, bind := range MyConf.RPCBind {
		log.Info("start listen rpc addr: \"%s\"", bind)
		go rpcListen(bind)
	}
	return nil
}

// rpcListen start rpc listen.
func rpcListen(bind string) {
	l, err := net.Listen("tcp", bind)
	if err != nil {
		log.Error("net.Listen(\"tcp\", \"%s\") error(%v)", bind, err)
		panic(err)
	}
	// if process exit, then close the rpc bind
	defer func() {
		log.Info("rpc addr: \"%s\" close", bind)
		if err := l.Close(); err != nil {
			log.Error("listener.Close() error(%v)", err)
		}
	}()
	rpc.Accept(l)
}

// NextId generate a id.
func (s *SnowflakeRPC) NextId(workerId int64, id *int64) error {
	if workerId > maxWorkerId || workerId < 0 {
		log.Error("worker Id can't be greater than %d or less than 0", maxWorkerId)
		return errors.New(fmt.Sprintf("worker Id: %d error", workerId))
	}
	if worker := s.idWorkers[workerId]; worker == nil {
		log.Warn("workerId: %d not register", workerId)
		return fmt.Errorf("snowflake workerId: %d don't register in this service", workerId)
	} else {
		if tid, err := worker.NextId(); err != nil {
			log.Error("worker.NextId() error(%v)", err)
			return err
		} else {
			*id = tid
		}
	}
	return nil
}

// DatacenterId return the services's datacenterId.
func (s *SnowflakeRPC) DatacenterId(ignore int, dataCenterId *int64) error {
	*dataCenterId = MyConf.DatacenterId
	return nil
}

// Timestamp return the service current unix seconds.
func (s *SnowflakeRPC) Timestamp(ignore int, timestamp *int64) error {
	*timestamp = time.Now().Unix()
	return nil
}

// Ping return the service status.
func (s *SnowflakeRPC) Ping(ignore int, status *int) error {
	*status = 0
	return nil
}
