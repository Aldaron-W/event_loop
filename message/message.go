package message

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
)

/**
 * 传递消息格式
 * 消息结构：
 * 	Id 消息ID
 * 	Payload 消息内容
 *
 * 序列化格式:
 *	len 长度 1字节
 * 	payload 长度 len长度
 */
type Message struct {
	Id uint32		// 消息ID
	Payload string	// 消息内容
}

func NewMessageByByte(data []byte) (*Message, error)  {
	message := &Message{}

	err := message.Unpack(data)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// 打包协议数据
func (m *Message) Pack()([]byte, error)  {
	jsonData, err := jsoniter.Marshal(m)
	if err != nil {
		fmt.Errorf(err.Error())
		return nil, err
	}
	data := make([]byte, len(jsonData)+1, len(jsonData)+1)
	data[0] = byte(len(jsonData))
	copy(data[1:], jsonData)
	return data, nil
}

// 解包协议数据
func (m *Message)Unpack(data []byte) error  {
	err := jsoniter.Unmarshal(data[1:], m)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	return nil
}