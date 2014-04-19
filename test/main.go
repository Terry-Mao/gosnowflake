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
