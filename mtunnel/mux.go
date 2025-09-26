package mtunnel

import (
	"context"
	"net"
)

type Muxer interface {
	net.Listener

	// Open 打开一个双向子流。
	Open(context.Context) (net.Conn, error)

	// RemoteAddr 远端节点地址。
	RemoteAddr() net.Addr

	// Protocol 返回通信协议类型。
	//	- protocol: 标准的底层通信协议，如：tcp udp
	//	- subprotocol: 子协议或具体的通信协议实现，一般为人工识别，如：github.com/quic-go/quic-go
	Protocol() (protocol, subprotocol string)
}
