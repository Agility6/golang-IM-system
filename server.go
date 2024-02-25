package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
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

func (this *Server) Handler(conn net.Conn) {
	// 	fmt.Println("链接建立成功")

	user := NewUser(conn, this)

	user.Online()

	isLive := make(chan bool)

	// 接收客户端发送到消息
	go func() {
		buf := make([]byte, 4096)

		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err: ", err)
				return
			}

			// 提取用户的消息(去除'\n')
			msg := string(buf[:n-1])

			// 用户针对msg进行消息处理
			user.DoMessage(msg)

			// 用户的任意消息，代表当前用户是一个活跃的
			isLive <- true
		}
	}()

	// 当前handler阻塞
	for {
		select {
		case <-isLive:
			// 当前用户是活跃的，应该重置定时器
			// 不做任何事情，为了激活select，更新下面的定时器
		case <-time.After(time.Second * 10):
			// 已经超时
			// 将当前的User强制的关闭

			user.SendMSg("您被强制下线")

			// 销毁用过的资源
			close(user.C)

			// 关闭连接
			conn.Close()

			// 退出当前的Handler
			return
		}
	}

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
		go this.Handler(conn)

	}
}
