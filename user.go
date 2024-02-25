package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {

	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}

	// 启动监听当前user channel消息的goroutine
	go user.ListenMessage()

	return user
}

// 用户的上线业务
func (this *User) Online() {

	// 用户上线，将用户加入到onlineMap中
	this.server.MapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.MapLock.Unlock()

	// 广播当前用户上线消息
	this.server.BroadCast(this, "已上线")
}

// 用户的下线业务
func (this *User) offline() {

	// 用户下线，将用户从onlineMap中删除
	this.server.MapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.MapLock.Unlock()

	this.server.BroadCast(this, "下线")
}

func (this *User) DoMessage(msg string) {

	// 查询当前在线用户
	if msg == "who" {

		this.server.MapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线..\n"
			this.SendMSg(onlineMsg)
		}
		this.server.MapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 消息格式rename|xxx

		newName := strings.Split(msg, "|")[1]

		// 判断name是否存在
		_, ok := this.server.OnlineMap[newName]

		if ok {
			this.SendMSg("当前用户名被使用\n")
		} else {
			this.server.MapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.MapLock.Unlock()

			this.Name = newName
			this.SendMSg("您已经更新用户名:" + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 消息格式 to|用户名|消息内容

		// 1. 获取对方的用户名
		remoteName := strings.Split(msg, "|")[1]

		if remoteName == "" {
			this.SendMSg("消息格式不正确，请使用 \"to|用户名｜消息内容\"格式。\n")
			return
		}

		// 2. 根据用户名 得到对方User对象
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMSg("用户名不存在\n")
			return
		}

		// 3. 获取消息内容，通过对方的User对象将消息内容发送过去
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMSg("无消息内容，请重试")
			return
		}

		remoteUser.SendMSg(this.Name + "发送一条消息：" + content)

	} else {
		this.server.BroadCast(this, msg)
	}
}

func (this *User) SendMSg(msg string) {
	this.conn.Write([]byte(msg))
}

// 监听当前User channel的方法，一旦有消息，就直接发送给对端的客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}
