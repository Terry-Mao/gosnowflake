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
	log "code.google.com/p/log4go"
	"testing"
)

func TestID(t *testing.T) {
	id, err := NewIdWorker(0, 0)
	if err != nil {
		log.Error("NewIdWorker(0, 0) error(%v)", err)
		t.FailNow()
	}
	sid, err := id.NextId()
	if err != nil {
		log.Error("id.NextId() error(%v)", err)
		t.FailNow()
	}
	log.Info("snowflake id: %d", sid)
}

func BenchmarkID(b *testing.B) {
	id, err := NewIdWorker(0, 0)
	if err != nil {
		log.Error("NewIdWorker(0, 0) error(%v)", err)
		b.FailNow()
	}
	for i := 0; i < b.N; i++ {
		if _, err := id.NextId(); err != nil {
			b.FailNow()
		}
	}
}
