//+build darwin netbsd freebsd openbsd dragonfly

package poller

import (
	"log"
	"net"
	"sync"
	"syscall"
)

type Poll struct {
	fd          int
	connections map[int]net.Conn
	lock        *sync.RWMutex
}

func NewPoller() (poller *Poll, err error) {
	// 创建 kqueue 文件描述符
	fd, err := syscall.Kqueue()
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

func (kqueue *Poll) Add(conn net.Conn) error {
	fd := SocketFD(conn)
	// 用于读取数据的文件事件和写入文件的事件打包成 kevent 事件数组加入监听
	eventList := make([]syscall.Kevent_t, 2)
	// Read
	eventList[0] = syscall.Kevent_t{
		Ident:  uint64(fd),
		Filter: syscall.EVFILT_READ,
		Flags:  syscall.EV_ADD | syscall.EV_CLEAR,
	}
	// Write
	eventList[1] = syscall.Kevent_t{
		Ident:  uint64(fd),
		Filter: syscall.EVFILT_WRITE,
		Flags:  syscall.EV_ADD | syscall.EV_CLEAR,
	}

	/**
	 * Kevent(kq int, changes, events []Kevent_t, timeout *Timespec) (n int, err error)
	 * kq		kqueue 的文件描述符		int
	 * changes	要注册 / 反注册的事件数组	[]syscall.Kevent_t
	 * events	满足条件的通知事件数组		[]syscall.Kevent_t
	 * timeout	等待事件到来时的超时时间，0，立刻返回；NULL，一直等待；有一个具体值，等待 timespec 时间值
	 */
	n, err := syscall.Kevent(kqueue.fd, eventList, nil, nil)
	if n < 0 || err != nil {
		return err
	}

	kqueue.lock.Lock()
	defer kqueue.lock.Unlock()
	kqueue.connections[fd] = conn
	if len(kqueue.connections)%100 == 0 {
		log.Printf("total number of connections: %v", len(kqueue.connections))
	}

	return nil
}

func (kqueue *Poll) Remove(conn net.Conn) error {
	fd := SocketFD(conn)

	eventList := make([]syscall.Kevent_t, 2)
	// Read
	eventList[0] = syscall.Kevent_t{
		Ident:  uint64(fd),
		Filter: syscall.EVFILT_READ,
		Flags:  syscall.EV_DELETE,
	}
	// Write
	eventList[1] = syscall.Kevent_t{
		Ident:  uint64(fd),
		Filter: syscall.EVFILT_WRITE,
		Flags:  syscall.EV_DELETE,
	}

	n, err := syscall.Kevent(kqueue.fd, eventList, nil, nil)
	if n < 0 || err != nil {
		return err
	}

	kqueue.lock.Lock()
	defer kqueue.lock.Unlock()
	delete(kqueue.connections, fd)
	if len(kqueue.connections)%100 == 0 {
		log.Printf("total number of connections: %v", len(kqueue.connections))
	}

	return nil
}

func (kqueue *Poll) Poll() (conn []net.Conn, err error) {
	events := make([]syscall.Kevent_t, 100)
	delay := &syscall.Timespec{}
	// 获取触发监听的事件
	n, err := syscall.Kevent(kqueue.fd, nil, events, delay)
	if err != nil {
		return nil, err
	}
	kqueue.lock.RLock()
	defer kqueue.lock.RUnlock()
	var connections []net.Conn
	for i := 0; i < n; i++ {
		conn := kqueue.connections[int(events[i].Ident)]
		connections = append(connections, conn)
	}
	return connections, nil
}
