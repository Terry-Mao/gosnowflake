package main

import (
	"github.com/golang/glog"
	"testing"
)

func TestID(t *testing.T) {
	id, err := NewIdWorker(0, 0)
	if err != nil {
		glog.Errorf("NewIdWorker(0, 0) error(%v)", err)
		t.FailNow()
	}
	sid, err := id.NextId()
	if err != nil {
		glog.Errorf("id.NextId() error(%v)", err)
		t.FailNow()
	}
	glog.Infof("snowflake id: %d", sid)
}
