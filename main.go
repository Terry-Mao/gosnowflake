package main

import (
	"flag"
	"github.com/golang/glog"
	"runtime"
)

func main() {
	flag.Parse()
	defer glog.Flush()
	// config
	if err := InitConfig(); err != nil {
		glog.Errorf("InitConfig() error(%v)", err)
		return
	}
	runtime.GOMAXPROCS(MyConf.MaxProc)
	glog.Infof("gosnowflake service start [datacenter: %d]", MyConf.DatacenterId)
	// process
	if err := InitProcess(); err != nil {
		glog.Errorf("InitProcess() error(%v)", err)
		return
	}
	// pprof
	InitPprof()
	// zookeeper
	if err := InitZK(); err != nil {
		glog.Errorf("InitZK() error(%v)", err)
		return
	}
	defer CloseZK()
	// rpc
	if err := InitRPC(); err != nil {
		glog.Errorf("InitRPC() error(%v)", err)
		return
	}
	// init signals, block wait signals
	sc := InitSignal()
	HandleSignal(sc)
	glog.Info("gosnowflake service stop")
}
