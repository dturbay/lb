package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/golang/glog"
)

func retransmitRequestToResponse(w http.ResponseWriter, r *http.Request) {
	inputBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.Fatal(err)
	}
	w.Write(inputBytes)
}

// http handler that ingores input, output array of 100 random numbers
func randomNumbersResponse(w http.ResponseWriter, r *http.Request) {
	var responseBytes [100]byte
	if _, err := rand.Read(responseBytes[:]); err != nil {
		glog.Fatal(err)
	}
	if _, err := w.Write(responseBytes[:]); err != nil {
		glog.Fatal(err)
	}
}

func startWebServer(handler http.HandlerFunc) int {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		glog.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      handler,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}
	go func() {
		glog.Fatal(server.ListenAndServe())
	}()
	return port
}

// really? no way to compare slices???????
func compareSlices(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestLB(t *testing.T) {
	webServerPort := startWebServer(retransmitRequestToResponse)
	glog.Infof("Web Server port: %d", webServerPort)
	webHost := fmt.Sprintf("localhost:%d", webServerPort)
	lbStartedChan := make(chan int)
	backendTCPAddr, err := net.ResolveTCPAddr("tcp", webHost)
	if err != nil {
		glog.Fatal(err)
	}
	lb := LoadBalancer{port: 0, backends: []*net.TCPAddr{backendTCPAddr}, startedSignal: lbStartedChan}
	go lb.Start()
	lbPort := <-lbStartedChan
	glog.Infof("LoadBalancer port: %d", lbPort)
	lbURL := fmt.Sprintf("http://localhost:%d/", lbPort)

	const ClientCount = 50    // Requests per goroutine
	const GoRoutineCount = 50 // Simulate 30 simultanious clients
	// https://stackoverflow.com/questions/39813587/go-client-program-generates-a-lot-a-sockets-in-time-wait-state
	// this property makes http client to use GoRoutineCount Keep-Alive connections
	// so in fact - LoadBalancer accept GoRoutineCount connections
	// TODO(dturbai): remove this prop and find out how to turn SO_REUSE_ADDRESS option on
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = GoRoutineCount

	var wg sync.WaitGroup
	wg.Add(GoRoutineCount)

	for i := 0; i < GoRoutineCount; i++ {
		go func() {
			defer wg.Done()
			for i := 0; i < ClientCount; i++ {
				httpClient := &http.Client{
					Timeout: 1 * time.Second,
				}
				// 7000 is bigger than defaultBufSize in bufio (4096)
				const MaxRequestSize = 7000
				var numbers [MaxRequestSize]byte
				var payloadSize = rand.Intn(MaxRequestSize)
				postSlice := numbers[:payloadSize]
				_, err := rand.Read(postSlice)
				if err != nil {
					glog.Fatal(err)
				}
				resp, err := httpClient.Post(lbURL, "text/data", bytes.NewReader(postSlice))
				if err != nil {
					glog.Fatal(err)
				}
				respBytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					glog.Fatal(err)
				}
				resp.Body.Close()
				// glog.V(1).Infof("Response size in bytes: %d", len(body))
				if !compareSlices(postSlice, respBytes) {
					t.Errorf("Posted bytes %v are not equal to received bytes %v", postSlice, respBytes)
				}
			}
		}()
	}
	wg.Wait()
	glog.Infof("LoadBalancer accepted %d connections", lb._acceptedConnCount)
	// time.Sleep(3e9)
	// glog.Infof("runtime.NumGoroutine: %d", runtime.NumGoroutine())
}

func runABTool(url string) {
	// concurency - 100, total queries 5000
	cmd := exec.Command("ab", "-c", "100", "-n", "5000", url)
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	err := cmd.Run()
	if err != nil {
		glog.Infof("ab stderr: %s", stdErr.String())
		glog.Fatal(err)
	}
	glog.Infof("ab stdout: %s", stdOut.String())
}

/*
Load test / benchmark with ab tool.
It does 2 runs: against web server and LB
*/
func TestLB_With_ab(t *testing.T) {
	glog.Infof("runtime.NumGoroutine: %d", runtime.NumGoroutine())
	webServerPort := startWebServer(randomNumbersResponse)
	glog.Infof("Web Server port: %d", webServerPort)
	webHost := fmt.Sprintf("localhost:%d", webServerPort)
	lbStartedChan := make(chan int)
	backendTCPAddr, err := net.ResolveTCPAddr("tcp", webHost)
	if err != nil {
		glog.Fatal(err)
	}
	lb := LoadBalancer{port: 0, backends: []*net.TCPAddr{backendTCPAddr}, startedSignal: lbStartedChan}
	go lb.Start()
	lbPort := <-lbStartedChan
	glog.Infof("LoadBalancer port: %d", lbPort)
	lbURL := fmt.Sprintf("http://localhost:%d/", lbPort)

	runABTool(webHost + "/")
	glog.Info("--------------------------------------------------")
	// glog.Infof("runtime.NumGoroutine: %d", runtime.NumGoroutine())
	runABTool(lbURL)
	glog.Infof("LoadBalancer accepted %d connections", lb._acceptedConnCount)
	glog.Infof("runtime.NumGoroutine: %d", runtime.NumGoroutine())
}
