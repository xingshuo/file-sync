package main

import "sync"

var (
	g_ChangeQueue = newChangeQueue()
)

type ChangeQueue struct {
	changedEvents sync.Map
}

func newChangeQueue() *ChangeQueue {
	return &ChangeQueue{
	}
}
