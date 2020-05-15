# event_loop
[evet-loop networking] 通过 Epoll 和 Kqueue 实现网络I/O 多路复用

# 实现功能点
- 实现服务端与客户端通过TCP连接发送数据
    - 定制数据协议（很简单）
    - 消息的序列化与反序列化（JSON）
    - 全双工发送消息（TCP）
- 实现通过不同的网络 I/O模型
    - BIO 阻塞I/O模型
    - NIO 非阻塞I/O模型
    - I/O多路复用
        - Linux epoll
        - BSD kqueue
        
# TODO
- 消息发送
- 消息暂存
