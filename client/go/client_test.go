package client

import (
	"fmt"
	"path"
	"testing"
	"time"
)

const (
	ROOT = "/gosnowflake-servers-test"
	ADDR = "10.33.21.151:2181"
)

func tFunc(t *testing.T) {
	client := NewClient([]string{ADDR}, ROOT)
	if err := client.initZK(); err != nil {
		t.Errorf("initZK error(%v)", err)
		t.FailNow()
	}
	nodes, _, err := zkConn.Children(ROOT)
	if err != nil {
		t.Errorf("getChildrenWatch error(%v)", err)
		t.FailNow()
	}
	t.Logf("nodes %v", nodes)
	for _, work := range nodes {
		wp := path.Join(ROOT, work)
		rpcs, _, err := zkConn.Children(wp)
		if err != nil {
			t.Errorf("zkConn.Children path(%s) error(%v)", wp, err)
			t.FailNow()
		}
		bs, _, err := zkConn.Get(path.Join(wp, rpcs[0]))
		if err != nil {
			t.Errorf("zkConn.Get(%s) error(%v)", path.Join(wp, nodes[0]), err)
			t.FailNow()
		}
		t.Logf("rpc %s", string(bs))
	}
}

func tId(t *testing.T) {
	client := NewClient([]string{ADDR}, ROOT)
	err := client.Dial()
	if err != nil {
		t.Errorf("client.Dial() error(%v)", err)
		t.FailNow()
	}
	time.Sleep(5 * time.Second)
	id, err := client.Id(0)
	if err != nil {
		t.Errorf("client.Id(0) error(%v)", err)
		t.FailNow()
	}
	fmt.Printf("id %d\n", id)
	t.Logf("id %d", id)
	time.Sleep(10 * time.Second)
	id, err = client.Id(0)
	if err != nil {
		t.Errorf("client.Id(0) error(%v)", err)
		t.FailNow()
	}
	fmt.Printf("id %d\n", id)
	t.Logf("id %d", id)
	time.Sleep(10 * time.Second)
}

func TestClient(t *testing.T) {
	// tFunc(t)
	tId(t)
}
