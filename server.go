package main

import (
	"fmt"
	"net"
)

// 定义server的结构体
type Server struct {
	Ip   string // ip地址
	Port int    // 端口
}

// 创建一个server的接口
func NewServer(ip string, port int) *Server {

	server := &Server{
		Ip:   ip,
		Port: port,
	}

	return server

}

func (this *Server) handler(conn net.Conn) {
	fmt.Println("链接建立成功")
}

func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))

	if err != nil {
		fmt.Println("net.Listen err: ", err)
		return
	}

	// close listen socket
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		// handler
		go this.handler(conn)

	}
}
