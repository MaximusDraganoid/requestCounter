// demonize
//simple example of demonize you programm

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

//structs to save information of request
type infoStrore struct {
	currentTime string //time.Time
	requestInfo string
}

type SliceOfInfoStore struct {
	slice []infoStrore
	mu    sync.Mutex
}

//funtions to processing some basicoperations
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

//master handler
func handler(resp http.ResponseWriter, req *http.Request) {
	//time.RFC3339
	//"2006-01-02T15:04:01Z"
	t := time.Now()
	newTime := t.Format("2006-01-02T15:04:05Z")
	Output.Add(infoStrore{currentTime: newTime, requestInfo: req.RemoteAddr})
	Output.printOutput(resp)
}

var PIDFile = "/tmp/daemonize.pid"

func savePID(pid int) {

	file, err := os.Create(PIDFile)
	if err != nil {
		log.Printf("Unable to create pid file : %v\n", err)
		os.Exit(1)
	}

	defer file.Close()

	_, err = file.WriteString(strconv.Itoa(pid))

	if err != nil {
		log.Printf("Unable to create pid file : %v\n", err)
		os.Exit(1)
	}

	file.Sync() // flush to disk

}

func SayHelloWorld(w http.ResponseWriter, r *http.Request) {
	html := "Hello World"

	w.Write([]byte(html))
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage : %s [start|stop] \n ", os.Args[0]) // return the program name back to %s
		os.Exit(0)
	}

	if strings.ToLower(os.Args[1]) == "main" {

		// Make arrangement to remove PID file upon receiving the SIGTERM from kill command
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)

		go func() {
			signalType := <-ch
			signal.Stop(ch)
			fmt.Println("Exit command received. Exiting...")

			// this is a good place to flush everything to disk
			// before terminating.
			fmt.Println("Received signal type : ", signalType)

			// remove PID file
			os.Remove(PIDFile)

			os.Exit(0)

		}()

		// request printer
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

	if strings.ToLower(os.Args[1]) == "start" {

		// check if daemon already running.
		if _, err := os.Stat(PIDFile); err == nil {
			fmt.Println("Already running or /tmp/daemonize.pid file exist.")
			os.Exit(1)
		}

		cmd := exec.Command(os.Args[0], "main")
		cmd.Start()
		fmt.Println("Daemon process ID is : ", cmd.Process.Pid)
		savePID(cmd.Process.Pid)
		os.Exit(0)

	}

	// upon receiving the stop command
	// read the Process ID stored in PIDfile
	// kill the process using the Process ID
	// and exit. If Process ID does not exist, prompt error and quit

	if strings.ToLower(os.Args[1]) == "stop" {
		if _, err := os.Stat(PIDFile); err == nil {
			data, err := ioutil.ReadFile(PIDFile)
			if err != nil {
				fmt.Println("Not running")
				os.Exit(1)
			}
			ProcessID, err := strconv.Atoi(string(data))

			if err != nil {
				fmt.Println("Unable to read and parse process id found in ", PIDFile)
				os.Exit(1)
			}

			process, err := os.FindProcess(ProcessID)

			if err != nil {
				fmt.Printf("Unable to find process ID [%v] with error %v \n", ProcessID, err)
				os.Exit(1)
			}
			// remove PID file
			os.Remove(PIDFile)

			fmt.Printf("Killing process ID [%v] now.\n", ProcessID)
			// kill process and exit immediately
			err = process.Kill()

			if err != nil {
				fmt.Printf("Unable to kill process ID [%v] with error %v \n", ProcessID, err)
				os.Exit(1)
			} else {
				fmt.Printf("Killed process ID [%v]\n", ProcessID)
				os.Exit(0)
			}

		} else {

			fmt.Println("Not running.")
			os.Exit(1)
		}
	} else {
		fmt.Printf("Unknown command : %v\n", os.Args[1])
		fmt.Printf("Usage : %s [start|stop]\n", os.Args[0]) // return the program name back to %s
		os.Exit(1)
	}

}
