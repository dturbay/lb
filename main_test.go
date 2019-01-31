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
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang/glog"
)

func init() {
}

type Handler struct{}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var numbers [100]byte
	var strNumbers [100]string
	rand.Read(numbers[0:])
	summ := 0
	for ind, number := range numbers {
		strNumbers[ind] = strconv.Itoa(int(number))
		summ += int(number)
	}
	w.Header().Set("Summ", strconv.Itoa(summ))
	fmt.Fprintf(w, strings.Join(strNumbers[:], ","))
}

func startWebServer() int {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		glog.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      &Handler{},
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}
	go func() {
		glog.Fatal(server.ListenAndServe())
	}()
	return port
}

func TestLB(t *testing.T) {
	webServerPort := startWebServer()
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

	const ClientCount = 1000   // Requests per goroutine
	const GoRoutineCount = 100 // Simulate 30 simultanious clients
	// https://stackoverflow.com/questions/39813587/go-client-program-generates-a-lot-a-sockets-in-time-wait-state
	// this property makes http client to use GoRoutineCount Keep-Alive connections
	// so in fact - LoadBalancer accept GoRoutineCount connections
	// TODO(dturbai): remove this prop and find out how to use SO_REUSE_ADDRESS option
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
				resp, err := httpClient.Get(lbURL)
				if err != nil {
					glog.Fatal(err)
				}
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				glog.V(1).Infof("Response size in bytes: %d", len(body))
				strSlice := strings.Split(string(body), ",")
				summ := 0
				for _, strNumber := range strSlice {
					number, _ := strconv.Atoi(strNumber)
					summ += number
				}
				headerSumm, _ := strconv.Atoi(resp.Header.Get("Summ"))
				if summ != headerSumm {
					t.Error("Summ of numbers received in http response: %d"+
						"is not equal to Sum from header: %d", summ, headerSumm)
				}
				glog.V(3).Infof("Calculated summ: %d", summ)
				glog.V(1).Infof("All headers: %v", resp.Header)
				glog.V(3).Infof("Summ: %s", resp.Header.Get("Summ"))
			}
		}()
	}
	wg.Wait()
	glog.Infof("LoadBalancer accepted %d connections", lb._acceptedConnCount)
	// time.Sleep(3e9)
	// glog.Infof("runtime.NumGoroutine: %d", runtime.NumGoroutine())
}

func printStatsForWebServer(url string) {
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
	webServerPort := startWebServer()
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

	printStatsForWebServer(webHost + "/")
	glog.Info("--------------------------------------------------")
	glog.Infof("runtime.NumGoroutine: %d", runtime.NumGoroutine())
	printStatsForWebServer(lbURL)
	glog.Infof("LoadBalancer accepted %d connections", lb._acceptedConnCount)
	// ok it seems I have goroutine leak
	glog.Infof("runtime.NumGoroutine: %d", runtime.NumGoroutine())
}
