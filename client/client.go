package main

/**
 * 	TCP连接客户端
 */

import (
	"../message"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

var (
	ip          = flag.String("ip", "127.0.0.1", "server IP")
	connections = flag.Int("conn", 10, "number of tcp connections")
	isSlowMode	= flag.Bool("slow_mode", true, "Client slow mode")
)
func main() {
	flag.Parse()
	addr := *ip + ":8972"
	log.Printf("连接到 %s", addr)
	var conns []net.Conn
	for i := 0; i < *connections; i++ {
		c, err := net.DialTimeout("tcp", addr, 10*time.Second)
		if err != nil {
			fmt.Println("failed to connect", i, err)
			i--
			continue
		}
		conns = append(conns, c)
		time.Sleep(time.Millisecond)
	}
	defer func() {
		for _, c := range conns {
			c.Close()
		}
	}()
	log.Printf("完成初始化 %d 连接", len(conns))
	tts := time.Second
	if *isSlowMode {
		tts *= 10
	}
	if *connections > 100 {
		tts = time.Millisecond * 5
	}

	wg := new(sync.WaitGroup)

	for i := 0; i < len(conns); i++ {
		go func(id int) {
			wg.Add(1)
			defer wg.Done()
			for {
				time.Sleep(tts)
				conn := conns[id]

				mesage := &message.Message{
					Id:      uint32(id),
					Payload: fmt.Sprintf("发送消息 connect : %d", id),
				}

				buf, err := mesage.Pack()
				if err != nil {
					continue
				}
				n, err := conn.Write(buf)
				if n != len(buf) || err != nil {
					log.Println("服务端获取消息异常")
				}
				log.Println("已发送消息")
			}
		}(i)
	}

	wg.Wait()
}