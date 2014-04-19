package main

import (
	"flag"
	"github.com/golang/glog"
	"runtime"
)

func main() {
	flag.Parse()
	glog.Errorf("gosnowflake service start")
	sc := InitSignal()
	defer glog.Flush()
	// config
	if err := InitConfig(); err != nil {
		glog.Errorf("InitConfig() error(%v)", err)
		return
	}
	runtime.GOMAXPROCS(MyConf.MaxProc)
	// process
	if err := InitProcess(); err != nil {
		glog.Errorf("InitProcess() error(%v)", err)
		return
	}
	// pprof
	InitPprof()
	if err := InitRPC(); err != nil {
		glog.Errorf("InitRPC() error(%v)", err)
		return
	}
	// init signals, block wait signals
	HandleSignal(sc)
	glog.Errorf("gosnowflake service stop")
}
