package main

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/samuel/go-zookeeper/zk"
	"time"
)

const (
	regRetryTimes  = 3
	regRetrySecond = 1 * time.Second
)

// InitZK init the zookeeper connection.
func InitZK() (*zk.Conn, error) {
	conn, session, err := zk.Connect(MyConf.ZKAddr, MyConf.ZKTimeout)
	if err != nil {
		glog.Errorf("zk.Connect(\"%v\", %d) error(%v)", MyConf.ZKAddr, MyConf.ZKTimeout, err)
		return nil, err
	}
	go func() {
		for {
			event := <-session
			glog.Infof("zookeeper get a event: %s", event.State.String())
		}
	}()

	return conn, nil
}

// RegWorkerId register the workerid in zookeeper, check exists or not to avoid the duplicate workerid.
func RegWorkerId(conn *zk.Conn, int workerId) error {
	glog.Infof("trying to claim workerId: %d", workerId)
	zkPath := fmt.Sprintf("%s/%d", MyConf.ZKPath)
	// retry
	for i := 0; i < regRetryTimes; i++ {
		_, err := conn.Create(zkPath, []byte(""), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
		if err != nil {
			if err == zk.ErrNodeExists {
				glog.Warningf("zk.create(\"%s\") exists", zkPath)
				time.Sleep(regRetrySecond)
			} else {
				glog.Errorf("zk.create(\"%s\") error(%v)", zkPath, err)
				return nil, err
			}
		} else {
			return nil
		}
	}
	return errors.New(fmt.Sprintf("workerId: %d register error", workerId))
}

// peers get workers all children in zookeeper.
func peers(conn *zk.Conn) (map[int]string, error) {
    // try create ZKPath
    _, err := conn.Create(MyConf.ZKPath, []byte(""), 0, zk.WorldACL(zk.PermAll))
    if err != nil {
        if err == zk.ErrNodeExists {
            glog.Warningf("zk.create(\"%s\") exists", MyConf.ZKPath)
        } else {
            glog.Errorf("zk.create(\"%s\") error(%v)", MyConf.ZKPath, err)
            return nil, err
        }
	}
    // fetch all nodes
	workers, _, err := conn.Children(MyConf.ZKPath)
	if err != nil {
        glog.Errorf("zk.Get(\"%s\") error(%v)", MyConf.ZKPath, err)
        return nil, err
	}
    res := make(map[int]string, len(workers))
    for _, worker := range workers {
        workerId, err := strconv.Atoi(worker)
        if err != nil {
            glog.Errorf("strconv.Atoi(\"%s\") error(%v)", worker, err)
            return nil, err
        }
        // get info
        zkPath := fmt.Sprintf("%s/%s", MyConf.ZKPath, worker)
        d, _, err := conn.Get(zkPath)
        if err != nil {
            glog.Errorf("zk.Get(\"%s\") error(%v)", zkPath, err)
            return nil, err
        }
        res[workerId] = string(d)
    }
    return res, nil
}

// sanityCheckPeers check the zookeeper datacenterId and all nodes time.
func sanityCheckPeers() error {
}
