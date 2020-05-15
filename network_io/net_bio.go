package network_io

import (
	"../message"
	"../poller"
	"fmt"
	"io"
	"net"
	"syscall"
)

type BIO struct {

}

func (bio *BIO) Read(conn net.Conn) error {
	// 设置文件描述符默认为阻塞式的
	fd := poller.SocketFD(conn)
	err := syscall.SetNonblock(fd, false)
	if err != nil {
		return err
	}
	size := make([]byte, 1, 1)
	// 阻塞式的等待文件描述符准备好
	fmt.Printf("【傻傻等待】就等着你开口 FD: %d",fd)
	if _, err := syscall.Read(fd, size); err != nil {
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
