package client

import (
	"fmt"
	"testing"
	"time"
)

func Test(t *testing.T) {
	if err := InitConfig(); err != nil {
		t.Error(err)
	}
	if err := Init(MyConf.ZKServers, MyConf.ZKPath, MyConf.ZKTimeout); err != nil {
		t.Error(err)
	}
	c := NewClient(MyConf.WorkerId)
	for i := 0; i < 60; i++ {
		time.Sleep(1 * time.Second)
		id, err := c.Id()
		if err != nil {
			t.Error(err)
		}
		fmt.Printf("gosnwoflake id: %d\n", id)
	}
	c.Close()
	// check global cache map
	if _, ok := workerIdMap[MyConf.WorkerId]; !ok {
		t.Error("no workerId")
	}
	c.Destroy()
	// check global cache map
	if _, ok := workerIdMap[MyConf.WorkerId]; ok {
		t.Error("workerId exists")
	}
}
