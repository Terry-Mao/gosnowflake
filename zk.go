package main

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
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
		glog.Errorf("zk.Connect(\"%v\", %d) error(%v)", MyConf.ZKAddr, MyConf.ZKTimeout, err)
		return err
	}
	zkConn = conn
	go func() {
		for {
			event := <-session
			glog.Infof("zookeeper get a event: %s", event.State.String())
		}
	}()
	return nil
}

// RegWorkerId register the workerid in zookeeper, check exists or not to avoid the duplicate workerid.
func RegWorkerId(workerId int64) error {
	glog.Infof("trying to claim workerId: %d", workerId)
	zkPath := fmt.Sprintf("%s/%d", MyConf.ZKPath, workerId)
	// retry
	for i := 0; i < regRetryTimes; i++ {
		_, err := zkConn.Create(zkPath, []byte(strings.Join(MyConf.RPCBind, ",")), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
		if err != nil {
			if err == zk.ErrNodeExists {
				glog.Warningf("zk.create(\"%s\") exists", zkPath)
				time.Sleep(regRetrySecond)
			} else {
				glog.Errorf("zk.create(\"%s\") error(%v)", zkPath, err)
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
			glog.Warningf("zk.create(\"%s\") exists", MyConf.ZKPath)
		} else {
			glog.Errorf("zk.create(\"%s\") error(%v)", MyConf.ZKPath, err)
			return nil, err
		}
	}
	// fetch all nodes
	workers, _, err := zkConn.Children(MyConf.ZKPath)
	if err != nil {
		glog.Errorf("zk.Get(\"%s\") error(%v)", MyConf.ZKPath, err)
		return nil, err
	}
	res := make(map[int][]string, len(workers))
	for _, worker := range workers {
		id, err := strconv.Atoi(worker)
		if err != nil {
			glog.Errorf("strconv.Atoi(\"%s\") error(%v)", worker, err)
			return nil, err
		}
		// get info
		zkPath := fmt.Sprintf("%s/%s", MyConf.ZKPath, worker)
		d, _, err := zkConn.Get(zkPath)
		if err != nil {
			glog.Errorf("zk.Get(\"%s\") error(%v)", zkPath, err)
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
		glog.Errorf("peers() error(%v)", err)
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
			glog.Warningf("peers: %d don't have any address", id)
			continue
		}
		// use first addr
		cli, err := rpc.Dial("tcp", addrs[0])
		if err != nil {
			glog.Errorf("rpc.Dial(\"tcp\", \"%s\") error(%v)", addrs[0])
			return err
		}
		defer cli.Close()
		if err = cli.Call("SnowflakeRPC.DatacenterId", 0, &datacenterId); err != nil {
			glog.Error("rpc.Call(\"SnowflakeRPC.DatacenterId\", 0) error(%v)", err)
			return err
		}
		// check datacenterid
		if datacenterId != MyConf.DatacenterId {
			glog.Errorf("worker at %s has datacenterId %d, but ours is %d", addrs[0], datacenterId, MyConf.DatacenterId)
			return errors.New("Datacenter id insanity")
		}
		if err = cli.Call("SnowflakeRPC.Timestamp", 0, &timestamp); err != nil {
			glog.Error("rpc.Call(\"SnowflakeRPC.Timestamp\", 0) error(%v)", err)
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
	glog.V(1).Infof("timestamps: %d, peer: %d, avg: %d, now - avg: %d, maxdelay: %d", timestamps, peerCount, avg, now-avg, timestampMaxDelay)
	if now-avg > timestampMaxDelay {
		glog.Errorf("timestamp sanity check failed. Mean timestamp is %d, but mine is %d so I'm more than 10s away from the mean", avg, now)
		return errors.New("timestamp sanity check failed")
	}
	return nil
}

// CloseZK close the zookeeper connection.
func CloseZK() {
	zkConn.Close()
}
