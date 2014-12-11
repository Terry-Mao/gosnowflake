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
	"github.com/samuel/go-zookeeper/zk"
	"net/rpc"
	"strconv"
	"strings"
	"time"
)

const (
	regRetryTimes     = 3
	regRetrySecond    = 1 * time.Second
	timestampMaxDelay = int64(10 * time.Second)
)

var (
	zkConn *zk.Conn
)

// InitZK init the zookeeper connection.
func InitZK() error {
	conn, session, err := zk.Connect(MyConf.ZKAddr, MyConf.ZKTimeout)
	if err != nil {
		log.Error("zk.Connect(\"%v\", %d) error(%v)", MyConf.ZKAddr, MyConf.ZKTimeout, err)
		return err
	}
	zkConn = conn
	go func() {
		for {
			event := <-session
			log.Info("zookeeper get a event: %s", event.State.String())
		}
	}()
	return nil
}

// RegWorkerId register the workerid in zookeeper, check exists or not to avoid the duplicate workerid.
func RegWorkerId(workerId int64) error {
	log.Info("trying to claim workerId: %d", workerId)
	zkPath := fmt.Sprintf("%s/%d", MyConf.ZKPath, workerId)
	// retry
	for i := 0; i < regRetryTimes; i++ {
		_, err := zkConn.Create(zkPath, []byte(strings.Join(MyConf.RPCBind, ",")), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
		if err != nil {
			if err == zk.ErrNodeExists {
				log.Warn("zk.create(\"%s\") exists", zkPath)
				time.Sleep(regRetrySecond)
			} else {
				log.Error("zk.create(\"%s\") error(%v)", zkPath, err)
				return err
			}
		} else {
			return nil
		}
	}
	return errors.New(fmt.Sprintf("workerId: %d register error", workerId))
}

// peers get workers all children in zookeeper.
func peers() (map[int][]string, error) {
	// try create ZKPath
	_, err := zkConn.Create(MyConf.ZKPath, []byte(""), 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		if err == zk.ErrNodeExists {
			log.Warn("zk.create(\"%s\") exists", MyConf.ZKPath)
		} else {
			log.Error("zk.create(\"%s\") error(%v)", MyConf.ZKPath, err)
			return nil, err
		}
	}
	// fetch all nodes
	workers, _, err := zkConn.Children(MyConf.ZKPath)
	if err != nil {
		log.Error("zk.Get(\"%s\") error(%v)", MyConf.ZKPath, err)
		return nil, err
	}
	res := make(map[int][]string, len(workers))
	for _, worker := range workers {
		id, err := strconv.Atoi(worker)
		if err != nil {
			log.Error("strconv.Atoi(\"%s\") error(%v)", worker, err)
			return nil, err
		}
		// get info
		zkPath := fmt.Sprintf("%s/%s", MyConf.ZKPath, worker)
		d, _, err := zkConn.Get(zkPath)
		if err != nil {
			log.Error("zk.Get(\"%s\") error(%v)", zkPath, err)
			return nil, err
		}
		res[id] = strings.Split(string(d), ",")
	}
	return res, nil
}

// sanityCheckPeers check the zookeeper datacenterId and all nodes time.
func SanityCheckPeers() error {
	allPeers, err := peers()
	if err != nil {
		log.Error("peers() error(%v)", err)
		return err
	}
	if len(allPeers) == 0 {
		return nil
	}
	timestamps := int64(0)
	timestamp := int64(0)
	datacenterId := int64(0)
	peerCount := int64(0)
	for id, addrs := range allPeers {
		if len(addrs) == 0 {
			log.Warn("peers: %d don't have any address", id)
			continue
		}
		// use first addr
		cli, err := rpc.Dial("tcp", addrs[0])
		if err != nil {
			log.Error("rpc.Dial(\"tcp\", \"%s\") error(%v)", addrs[0])
			return err
		}
		defer cli.Close()
		if err = cli.Call("SnowflakeRPC.DatacenterId", 0, &datacenterId); err != nil {
			log.Error("rpc.Call(\"SnowflakeRPC.DatacenterId\", 0) error(%v)", err)
			return err
		}
		// check datacenterid
		if datacenterId != MyConf.DatacenterId {
			log.Error("worker at %s has datacenterId %d, but ours is %d", addrs[0], datacenterId, MyConf.DatacenterId)
			return errors.New("Datacenter id insanity")
		}
		if err = cli.Call("SnowflakeRPC.Timestamp", 0, &timestamp); err != nil {
			log.Error("rpc.Call(\"SnowflakeRPC.Timestamp\", 0) error(%v)", err)
			return err
		}
		// add timestamps
		timestamps += timestamp
		peerCount++
	}
	// check 10s
	// calc avg timestamps
	now := time.Now().UnixNano()
	avg := int64(timestamps / peerCount)
	log.Debug("timestamps: %d, peer: %d, avg: %d, now - avg: %d, maxdelay: %d", timestamps, peerCount, avg, now-avg, timestampMaxDelay)
	if now-avg > timestampMaxDelay {
		log.Error("timestamp sanity check failed. Mean timestamp is %d, but mine is %d so I'm more than 10s away from the mean", avg, now)
		return errors.New("timestamp sanity check failed")
	}
	return nil
}

// CloseZK close the zookeeper connection.
func CloseZK() {
	zkConn.Close()
}
