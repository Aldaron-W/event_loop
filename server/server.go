package main

import (
	"../network_io"
	"../poller"
	"flag"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
)

var poll *poller.Poll
var connections map[int]net.Conn
var useBIO bool
var useMultiplexing bool

func main() {
	flag.BoolVar(&useBIO, "bio", true, "是否采用阻塞IO模式(BIO)的方式处理网络I/O Default:true")
	flag.BoolVar(&useMultiplexing, "multiplexing", false, "是否采用IO多路复用方式操作文件描述符 Default:false")
	flag.Parse()

	ln, err := net.Listen("tcp", ":8972")
	if err != nil {
		panic(err)
	}

	// pprof
	go func() {
		if err := http.ListenAndServe(":6060", nil); err != nil {
			log.Fatalf("pprof failed: %v", err)
		}
	}()

	if !useMultiplexing {
		defer func() {
			for _, conn := range connections {
				_ = conn.Close()
			}
		}()
	} else {
		poll, err = poller.NewPoller()
		go startPoll()
	}

	for {
		conn, e := ln.Accept()
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				log.Printf("accept temp err: %v", ne)
				continue
			}
			log.Printf("accept err: %v", e)
			return
		}

		if !useMultiplexing {
			connections = map[int]net.Conn{}
			go handleConn(conn)
			fd := poller.SocketFD(conn)
			connections[fd] = conn
			if len(connections)%100 == 0 {
				log.Printf("total number of connections: %v", len(connections))
			}
		} else {
			if err := poll.Add(conn); err != nil {
				log.Printf("failed to add connection %v", err)
				conn.Close()
			}
		}
	}
}

func startPoll() {
	for {
		connections, err := poll.Poll()
		if err != nil {
			log.Printf("failed to epoll wait %v", err)
			continue
		}
		for _, conn := range connections {
			if conn == nil {
				break
			}
			if err := readConn(conn); err != nil {
				if err := poll.Remove(conn); err != nil {
					log.Printf("failed to remove %v", err)
				}
				_ = conn.Close()
			}
		}
	}
}

func handleConn(conn net.Conn) {
	for {
		err := readConn(conn)
		if err != nil {
			continue
		}
	}
}

func readConn(conn net.Conn) error {
	if useBIO {
		nio := network_io.BIO{}
		err := nio.Read(conn)
		if err != nil {
			return err
		}
	} else {
		nio := network_io.NIO{}
		err := nio.Read(conn)
		if err != nil {
			return err
		}
	}

	return nil
}