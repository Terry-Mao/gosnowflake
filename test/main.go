package main

import (
	"flag"
	"github.com/golang/glog"
	"net/rpc"
)

func main() {
	flag.Parse()
	if err := InitConfig(); err != nil {
		glog.Errorf("InitConfig() error(%v)", err)
		return
	}

	cli, err := rpc.Dial("tcp", MyConf.RPCAddr)
	if err != nil {
		glog.Errorf("rcp.Dial(\"tcp\", \"%s\") error(%v)", MyConf.RPCAddr, err)
		return
	}
	defer cli.Close()
	// get snowflake id by workerId
	id := int64(0)
	if err = cli.Call("SnowflakeRPC.NextId", MyConf.WorkerId, &id); err != nil {
		glog.Errorf("rpc.Call(\"SnowflakeRPC.NextId\", %d, &id) error(%v)", MyConf.WorkerId, err)
		return
	}
	glog.Infof("nextid: %d\n", id)
	// get datacenter id
	datacenterId := int64(0)
	if err = cli.Call("SnowflakeRPC.DatacenterId", 0, &datacenterId); err != nil {
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
	status := 0
	if err = cli.Call("SnowflakeRPC.Ping", 0, &status); err != nil {
		glog.Errorf("rpc.Call(\"SnowflakeRPC.Ping\", 0, &status) error(%v)", err)
		return
	}
	glog.Infof("status: %d\n", status)
}
