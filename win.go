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

var mu sync.Mutex
var Output []infoStrore

func handler(resp http.ResponseWriter, req *http.Request) {
	mu.Lock()
	//time.RFC3339
	//"2006-01-02T15:04:01Z"
	t := time.Now()
	newTime := t.Format("2006-01-02T15:04:05Z")
	Output = append(Output, infoStrore{currentTime: newTime, requestInfo: req.RemoteAddr})
	printOutput(resp)
	mu.Unlock()
}

func printOutput(resp http.ResponseWriter) {
	for key := range Output {
		fmt.Fprintf(resp, "%v \t%v \t%v\n", key, Output[key].currentTime, Output[key].requestInfo)
	}
}

func CleanOutput() {
	for len(Output) != 0 {
		Output = Output[0 : len(Output)-1]
	}
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
			mu.Lock()
			CleanOutput()
			mu.Unlock()
		}
	}()
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		return
	}
}
