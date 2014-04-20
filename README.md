## Terry-Mao/gosnowflake

`Terry-Mao/gosnowflake` is a network service for generating unique ID numbers at high scale with some simple guarantees (golang).

## Requeriments

golang 1.2 is required.

zookeeper is required.

## Installation

Just pull `Terry-Mao/gosnowflake` from github using `go get`:

```sh
# download the code
$ go get -u github.com/Terry-Mao/gosnowflake
# find the dir
$ cd $GOPATH/src/github.com/Terry-Mao/gosnowflake
# compile
$ go build
# run
$ ./gosnowflake -conf=./gosnowflake-example.conf
# for help
$ ./gosnowflake -h
```

## Usage

```go
package main

import (
    "fmt"
    "net/rpc"
)

func main() {
    cli, err := rpc.Dial("tcp", "localhost:8080")
    if err != nil {
        panic(err)
    }
    defer cli.Close()
    id := int64(0)
    workerId := 0
    if err = cli.Call("SnowflakeRPC.NextId", workerId, &id); err != nil {
        panic(err)
    }
    fmt.Printf("id: %d\n", id)
}
```

## Highly Available

use `heartbeat` or `keepalived` apply a VIP for the client.
