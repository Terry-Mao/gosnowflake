## Terry-Mao/gosnowflake

`Terry-Mao/gosnowflake` is a network service for generating unique ID numbers at high scale with some simple guarantees (golang).

## Requeriments

golang 1.2 is required.

zookeeper is required.

## Installation

Just pull `Terry-Mao/gosnowflake` from github using `go get`:

```sh
# install golang from https://code.google.com/p/go/downloads/list
# here golang 1.2(linux-amd64)
$ wget https://go.googlecode.com/files/go1.2.linux-amd64.tar.gz
$ tar -xvf go1.2.linux-amd64.tar.gz
$ cp -R go /usr/local/
$ vim /etc/profile.d/golang.sh
# add below to golang.sh
# export GOROOT=/usr/local/go
# export PATH=$PATH:$GOROOT/bin
# export GOPATH=/data/apps/go
$ source /etc/profile.d/gopush.sh
# download the code
$ go get -u github.com/Terry-Mao/gosnowflake
# find the dir
$ cd $GOPATH/src/github.com/Terry-Mao/gosnowflake
# compile
$ go install
$ cp ./gosnowflake-example.conf $GOPATH/bin/gosnowflake.conf
$ cd $GOPATH/bin
# run
$ ./gosnowflake -conf=./gosnowflake.conf
# for help
$ ./gosnowflake -h
# test
$ cd $GOPATH/src/github.com/Terry-Mao/gosnowflake/test
$ go build
$ ./test
```

## Document
```sh
[base]
# If the master process is run as root, then gosnowflake will setuid()/setgid() 
# to USER/GROUP. If GROUP is not specified, then gosnowflake uses the same name as 
# USER. By default it's nobody user and nobody or nogroup group.
# user maojian 

# When running daemonized, gosnowflake writes a pid file in 
# /tmp/gosnowflake.pid by default. You can specify a custom pid file 
# location here.
pidfile /tmp/gosnowflake.pid

# Sets the maximum number of CPUs that can be executing simultaneously.
# This call will go away when the scheduler improves. By default the number of 
# logical CPUs is set.
# 
# maxproc 4

# By default gosnowflake listens for connections from all the network interfaces
# available on the server on 8080 port. It is possible to listen to just one or 
# multiple interfaces using the "websocket.bind" configuration directive, 
# followed by one or more IP addresses and port.
#
# Examples:
#
# Note this directive is only support "websocket" protocol
# rpc.bind 192.168.1.100:8080,10.0.0.1:8080
# rpc.bind 127.0.0.1:8080
# rpc.bind :8080

# This is used by gosnowflake service profiling (pprof).
# By default gosnowflake pprof listens for connections from local interfaces on 6971
# port. It's not safty for listening internet IP addresses.
#
# Examples:
#
# pprof.bind 192.168.1.100:6971,10.0.0.1:6971
# pprof.bind 127.0.0.1:6971
# pprof.bind :6971

# This is used by gosnowflake service get stat info by http.
# By default gosnowflake pprof listens for connections from local interfaces on 6972
# port. It's not safty for listening internet IP addresses.
#
# Examples:
#
# stat.bind 192.168.1.100:6971,10.0.0.1:6971
# stat.bind 127.0.0.1:6971
# stat.bind :6971

# The working directory.
#
# The log will be written inside this directory, with the filename specified
# above using the 'logfile' configuration directive.
#  
# Note that you must specify a directory here, not a file name.
dir ./

################################## ZOOKEEPER ##################################
[zookeeper]
# The zookeeper cluster section. When gosnowflake start, it will register data 
# in the zookeeper cluster and create a ephemeral node. When gosnowflake died, 
# the node will drop by zookeeper cluster. 

# Zookeeper cluster addresses. Mutiple address split by a ",".
# Examples:
#
# addr 192.168.1.100:2181,10.0.0.1:2181
# addr 127.0.0.1:2181
# addr 10.20.216.122:2181

# Zookeeper cluster session idle timeout seconds. Zookeeper will close the 
# connection after a client is idle for N seconds.
# Examples:
#
# timeout 30s
# timeout 15s
timeout 30s

# gosnowflake zookeeper root path.
path /gosnowflake-servers

################################## GOSNOWFLAKE ################################
[snowflake]
# snowflake must set a datacenter [0, 31], must be unique in all datacenter.
# Examples:
#
# datacenter 0
# datacenter 1
datacenter 0

# register which worker, must be unique in one datacenter.
# Examples:
#
# worker 0
# worker 0,1,2
worker 0,1,2
```

## RPC API

`SnowflakeRPC.NextId`: generate a snowflake id.

`SnowflakeRPC.DatacenterId`: get gosnowflake service's datacenterId.

`SnowflakeRPC.Timestamp`: get gosnowflake service's current timestamp.

`SnowflakeRPC.Ping`: get gosnowflake service's status.

## Usage

```go
package main

import (
	"fmt"
	"net/rpc"
    "github.com/golang/glog"
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
```

## Highly Available

use `heartbeat` or `keepalived` apply a VIP for the client.
