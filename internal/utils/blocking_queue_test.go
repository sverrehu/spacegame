package utils

import (
	"testing"
	"time"
)

func TestBlockingQueue(t *testing.T) {
	queue := NewBlockingQueue()
	gotAll := false
	go func() {
		for q := range 3 {
			v := queue.Dequeue()
			if v != q {
				t.Errorf("got %v; want %v", v, q)
			}
		}
		gotAll = true
	}()
	for q := range 3 {
		queue.Enqueue(q)
	}
	time.Sleep(1 * time.Second)
	if !gotAll {
		t.Errorf("got %v; want %v", gotAll, true)
	}
}
