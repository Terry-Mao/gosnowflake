package main

import (
	"fmt"
	"github.com/golang/glog"
	"net/rpc"
)

func main() {
	addr := "localhost:8080"
	cli, err := rpc.Dial("tcp", addr)
	if err != nil {
		glog.Errorf("rcp.Dial(\"tcp\", \"%s\") error(%v)", addr, err)
		return
	}
	defer cli.Close()
	// get snowflake id by workerId
	id := int64(0)
	workerId := 0
	if err = cli.Call("SnowflakeRPC.NextId", workerId, &id); err != nil {
		glog.Errorf("rpc.Call(\"SnowflakeRPC.NextId\", %d, &id) error(%v)", workerId, err)
		return
	}
	glog.Infof("nextid: %d\n", id)
	// get datacenter id
	datacenterId := int64(0)
	if err = cli.Call("SnowflakeRPC.DatacenterId", 0, &dataCenterId); err != nil {
		glog.Errorf("rpc.Call(\"SnowflakeRPC.DatacenterId\", 0, &datacenterId) error(%v)", err)
		return
	}
	glog.Infof("datacenterid: %d\n", datacenterId)
	// get current timestamp
	timestamp := int64(0)
	if err = cli.Call("SnowflakeRPC.Timestamp", 0, &timestamp); err != nil {
		glog.Errorf("rpc.Call(\"SnowflakeRPC.Timestamp\", 0, &timestamp) error(%v)", err)
		return
	}
	glog.Infof("timestamp: %d\n", timestamp)
}
