package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/golang/glog"
)

func handler(w http.ResponseWriter, r *http.Request) {
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
	http.HandleFunc("/", handler)
	go func() {
		glog.Fatal(http.Serve(listener, nil))
	}()
	return port
}

func TestLB(t *testing.T) {
	webServerPort := startWebServer()
	webHost := fmt.Sprintf("localhost:%d", webServerPort)
	webURL := fmt.Sprintf("http://%s/", webHost)
	fmt.Println(webURL)
	lbStartedChan := make(chan int)
	lb := LoadBalancer{port: 8888, backends: []string{webHost}, startedSignal: lbStartedChan}
	go lb.Start()
	lbPort := <-lbStartedChan
	fmt.Println("4", lbPort) // debug
	lbURL := fmt.Sprintf("http://localhost:%d/", lbPort)
	resp, err := http.Get(lbURL)
	if err != nil {
		glog.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	strSlice := strings.Split(string(body), ",")
	summ := 0
	for _, strNumber := range strSlice {
		number, _ := strconv.Atoi(strNumber)
		summ += number
	}
	fmt.Println("Calculated summ: ", summ)
	fmt.Println("Summ", resp.Header.Get("Summ"))

}
