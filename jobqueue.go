package main

import (
    "sync"
)

// JobQueue is our internal job queue
var JobQueue = queue{
    list: make(chan Doer, 100),
    wg: &sync.WaitGroup{},
}

// Doer interface describes an object that can be considered a unit of work
type Doer interface {
    Do() error
}

// PkgInstallJob handles the installation of a package
type PkgInstallJob struct {
    pkg *Pkg
}

// PkgDiscoverJob handles the discovery of dependencies for a package
type PkgDiscoverJob struct {
    pkg *Pkg
}

// Queue is our internal Doer list manager
type queue struct {
    list chan Doer
    wg *sync.WaitGroup
    stop chan bool
}

func (q *queue) Add(j Doer) {
    q.list <- j
}

func (q *queue) Start() {
    go func() {
        for {
    		select {
    		case <-q.stop:
                Log("Stopping queue execution")
    			return
    		case j := <- q.list:
    			j.Do()
                q.wg.Done()
    			return
    		}
    	}
    }()
}

func (q *queue) Stop() {
    q.stop <- true
}

func (q *queue) Wait() {
    q.wg.Wait()
}
