package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/golang/glog"
)

// LoadBalancer configuration
type LoadBalancer struct {
	port               int
	backends           []*net.TCPAddr
	startedSignal      chan int
	_acceptedConnCount uint64 // stats
}

/**
  There are 2 connections here, communication happens between client and backend through LB
  1: client <-> LB and  2: LB <-> backend

  Any of specified connections may be closed by client or backend.

  "Everithing works fine" case:
	  connection is naturally closed (either by client|backend or OS on client|backend side)
	  Reader that reads from closed side gets notified.
	  Another side have to receive all data that was sent by ClosingInitiator

	  E.g. I make rest query to LB, LB transfer it to backend, backend receive my request, send result to LB, close connection,
	  on LB side I need to receive all data and close connection to client side.

  "Either client or server hangs up" case:
	TBD
*/
func (lb *LoadBalancer) handleIncomingConn(clientConn net.Conn) {
	defer clientConn.Close()
	backendConn, err := net.DialTCP("tcp", nil, lb.backends[0]) // TODO(dturbai): implement round-robin for backend selection
	if err != nil {
		glog.Error("Failed to connect to backend", err)
		return
	}
	defer backendConn.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		// transfer data from client to backend
		if _, err := io.Copy(backendConn, clientConn); err != nil {
			glog.V(1).Info(err) //  clientConn close trigger error
		}
		glog.V(5).Info("client data transfered")
		backendConn.Close() // Client initiated closing communication, all data that comes from backend makes no sense now
	}()

	go func() {
		defer wg.Done()
		// transfer data from backend to client
		if _, err := io.Copy(clientConn, backendConn); err != nil {
			glog.V(1).Info(err) // backendConn close trigger error
		}
		glog.V(5).Info("backend data transfered")
		clientConn.Close() // backend closed connection, everithing from client makes no sense now
	}()
	wg.Wait()
}

// Start load balancer
func (lb *LoadBalancer) Start() {
	listener, err := net.Listen("tcp6", fmt.Sprintf(":%d", lb.port))
	if err != nil {
		glog.Fatal(err)
	}
	port := lb.port
	if lb.port == 0 {
		port = listener.Addr().(*net.TCPAddr).Port
	}
	if lb.startedSignal != nil {
		lb.startedSignal <- port
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			glog.Error(err)
			continue
		}
		// tcpconn, _ := conn.(*net.TCPConn)
		// file, err := tcpconn.File()
		// err = syscall.SetsockoptInt(int(file.Fd()), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
		// file.Close()

		atomic.AddUint64(&lb._acceptedConnCount, 1)
		glog.V(3).Infof("Connection accepted from: %s", conn.RemoteAddr())
		go lb.handleIncomingConn(conn)
	}
}

var lbPort = flag.Int("port", 8888, "LoadBalancer tcp port")

func main() {
	flag.Parse()
	glog.Info("Starting lb ...")
	addr, err := net.ResolveTCPAddr("tcp", "localhost:7777")
	if err != nil {
		glog.Fatal(err)
	}
	loadBalancer := LoadBalancer{
		port:     *lbPort,
		backends: []*net.TCPAddr{addr}}
	loadBalancer.Start()
}
