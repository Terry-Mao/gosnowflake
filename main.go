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
