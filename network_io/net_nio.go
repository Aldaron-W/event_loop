package network_io

import (
	"../message"
	"../poller"
	"fmt"
	"io"
	"net"
	"syscall"
)

type NIO struct {

}

func (nio *NIO) Read(conn net.Conn) error {
	// 设置文件描述符的Flag为非阻塞
	fd := poller.SocketFD(conn)
	err := syscall.SetNonblock(fd, true)
	if err != nil {
		return err
	}
	for  {
		size := make([]byte, 1, 1)
		// 非阻塞的通过文件描述符读取数据
		if _, err := syscall.Read(fd, size); err != nil {
			// 若读取文件描述符还没有准备好，则继续轮训
			if err == syscall.EAGAIN {
				fmt.Printf("【絮絮叨叨】你说呀～你说呀～你说呀～ FD:%d", fd)
				continue
			}
			if err != io.EOF {
				fmt.Println("read error:", err)
			}
			return err
		}

		data := make([]byte, size[0]+1, size[0]+1)
		data[0] = size[0]
		if n, err := syscall.Read(fd, data[1:]); err != nil && n != int(size[0]) {
			if err != io.EOF {
				fmt.Println("read error:", err)
			}
			return err
		}

		tcpMessage, err := message.NewMessageByByte(data)
		if err != nil {
			_ = fmt.Errorf(err.Error())
		}

		// 显示结果
		fmt.Println("response:", tcpMessage)
		return nil
	}
}
