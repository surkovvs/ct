package wgchan

import "sync"

type WgChan struct {
	wg *sync.WaitGroup
}

func NewWgChan(wg *sync.WaitGroup) <-chan struct{} {
	c := make(chan struct{})
	go func() {
		wg.Wait()
		close(c)
	}()
	return c
}
