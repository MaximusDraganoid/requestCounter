package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type infoStrore struct {
	currentTime string //time.Time
	requestInfo string
}

type SliceOfInfoStore struct {
	slice []infoStrore
	mu    sync.Mutex
}

func (p *SliceOfInfoStore) Add(toAdd infoStrore) {
	p.mu.Lock()
	p.slice = append(p.slice, toAdd)
	p.mu.Unlock()
}

func (p *SliceOfInfoStore) printOutput(resp http.ResponseWriter) {
	p.mu.Lock()
	for key := range p.slice {
		fmt.Fprintf(resp, "%v \t%v \t%v\n", key, Output.slice[key].currentTime, Output.slice[key].requestInfo)
	}
	p.mu.Unlock()
}

func (p *SliceOfInfoStore) CleanOutput() {
	p.mu.Lock()
	for len(Output.slice) != 0 {
		Output.slice = Output.slice[0 : len(Output.slice)-1]
	}
	p.mu.Unlock()
}

var Output SliceOfInfoStore

func handler(resp http.ResponseWriter, req *http.Request) {
	//time.RFC3339
	//"2006-01-02T15:04:01Z"
	t := time.Now()
	newTime := t.Format("2006-01-02T15:04:05Z")
	Output.Add(infoStrore{currentTime: newTime, requestInfo: req.RemoteAddr})
	Output.printOutput(resp)
}

func main() {
	go func() {
		for {
			m := time.Now().Round(time.Minute)
			waitTime := time.Now().Sub(m)
			if waitTime < 0 {
				waitTime = -waitTime
			} else {
				waitTime = 60*time.Second - waitTime
			}
			<-time.After(waitTime)
			Output.CleanOutput()
		}
	}()
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		return
	}
}
