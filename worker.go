package main

import "sync"

type worker struct {
	wg       sync.WaitGroup
	stopChan chan struct{}
}
