package main

import (
	"fmt"
	"net"
	"sync"
)

// 定义server的结构体
type Server struct {
	Ip   string // ip地址
	Port int    // 端口

	// 在线用户的列表
	OnlineMap map[string]*User
	MapLock   sync.RWMutex
	// 消息广播的channel
	Message chan string
}

// 创建一个server的接口
func NewServer(ip string, port int) *Server {

	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server

}

// 监听Message广播信息channel的gotoutine，一旦有消息就发送给全部的在线User
func (this *Server) ListenMessager() {

	for {
		msg := <-this.Message

		// 将msg发送给全部的在线User
		this.MapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.MapLock.Unlock()
	}

}

func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + "." + msg
	this.Message <- sendMsg

}

func (this *Server) handler(conn net.Conn) {
	fmt.Println("链接建立成功")

	user := NewUser(conn)

	// 用户上线，将用户加入到onlineMap中
	this.MapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.MapLock.Unlock()

	// 广播当前用户上线信息
	this.BroadCast(user, "已上线")

	// 当前handler阻塞
	select {}

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

	// 启动监听Message的goroutine
	go this.ListenMessager()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		// handler
		go this.handler(conn)

	}
}
