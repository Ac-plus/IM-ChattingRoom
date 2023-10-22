package main

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	//在线用户列表
	OnLineMap map[string]*User
	mapLock   sync.RWMutex //sync pkg保管互斥锁

	//用于广播消息的管道
	Message chan string
}

// 创建server的方法
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnLineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// ListenMessager 监听Message广播消息的go程，一旦有消息就发送给全部的在线USer
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message
		// 将msg发给全部的在线user
		// 需要遍历userMap
		this.mapLock.Lock()
		for _, cli := range this.OnLineMap {
			cli.C <- msg //将消息发给cli的管道
		}
		this.mapLock.Unlock()
	}
}

// 广播消息的方法，由user发起，广播内容为msg
func (this *Server) BroadCast(user *User, msg string) {
	// 构建消息内容：用户地址+用户名+消息内容
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	// 将消息发送到server的广播管道中
	this.Message <- sendMsg
}

// 业务处理方法
// 在定义的函数名之前加上(this *Server)的说明，表示为Server类添加一个方法
// 若无这样的说明，则只是普通的函数
func (this *Server) Handler(conn net.Conn) {
	//...当前连接的业务
	//fmt.Println("[INFO] Connection Established!")

	user := NewUser(conn)

	//用户上线，首先加入用户表
	this.mapLock.Lock()
	this.OnLineMap[user.Name] = user
	this.mapLock.Unlock()

	//然后，广播当前用户上线的消息给到全部user
	this.BroadCast(user, "已上线")

	//先保持handler仍然运作（阻塞），否则有可能让前面的go程终止
	select {}

}

// 启动服务器的接口
func (this *Server) Start() {
	// SOCKET LISTEN
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("[ERROR] Network Error, Error Info:", err)
		return
	}

	// CLOSE SOCKET
	// 全部执行完后，关闭Socket
	defer listener.Close()

	//启动LitenMessager
	go this.ListenMessager()

	for {
		// SOCKET ACCEPT
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("[ERROR] Listener Accept Error, Error Info:", err)
			continue
		}
		// DO HANDLER
		go this.Handler(conn)
	}

}
