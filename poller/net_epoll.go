// +build linux

package poller

import (
	"golang.org/x/sys/unix"
	"log"
	"net"
	"sync"
	"syscall"
)

type Poll struct {
	fd          int
	connections map[int]net.Conn
	lock        *sync.RWMutex		// connections的读写锁
}

func NewPoller() (poller *Poll, err error){
	// 创建 epoll 文件描述符
	fd, err := syscall.EpollCreate1(0)
	if err != nil {
		return nil, err
	} else if fd < 0{
		return nil, nil
	}

	syscall.CloseOnExec(fd)

	return &Poll{
		fd:          fd,
		connections: map[int]net.Conn{},
		lock:        &sync.RWMutex{},
	}, nil
}

func (epoll *Poll) Add(conn net.Conn) error{
	fd := SocketFD(conn)
	err := unix.EpollCtl(epoll.fd, syscall.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Events: unix.POLLIN | unix.POLLHUP, Fd: int32(fd)})
	if err != nil {
		return err
	}
	epoll.lock.Lock()
	defer epoll.lock.Unlock()
	epoll.connections[fd] = conn
	if len(epoll.connections)%100 == 0 {
		log.Printf("total number of connections: %v", len(epoll.connections))
	}
	return nil
}

func (epoll *Poll) Remove(conn net.Conn) error {
	fd := SocketFD(conn)
	err := unix.EpollCtl(epoll.fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return err
	}
	epoll.lock.Lock()
	defer epoll.lock.Unlock()
	delete(epoll.connections, fd)
	if len(epoll.connections)%100 == 0 {
		log.Printf("total number of connections: %v", len(epoll.connections))
	}
	return nil
}

func (epoll *Poll) Poll() (conn []net.Conn, err error){
	events := make([]unix.EpollEvent, 100)
	// 获取触发监听的事件
	n, err := unix.EpollWait(epoll.fd, events, 100)
	if err != nil {
		return nil, err
	}
	epoll.lock.RLock()
	defer epoll.lock.RUnlock()
	var connections []net.Conn
	for i := 0; i < n; i++ {
		conn := epoll.connections[int(events[i].Fd)]
		connections = append(connections, conn)
	}
	return connections, nil
}
