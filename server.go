package main

import (
	"fmt"
	"io"
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

// NewServer 创建server的方法
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

// BroadCast 广播消息的方法，由user发起，广播内容为msg
// 在0.2版本中，user发给Message管道“我已上线”的消息，然后server广播给所有人
func (this *Server) BroadCast(user *User, msg string) {
	// 构建消息内容：用户地址+用户名+消息内容
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	// 将消息发送到server的广播管道中
	this.Message <- sendMsg
}

// Handler 业务处理方法
// 在定义的函数名之前加上(this *Server)的说明，表示为Server类添加一个方法
// 若无这样的说明，则只是普通的函数
// 本方法中，传入的参数conn表示当前和server连接的用户socket。可以有多个，因为开辟了多个go程
func (this *Server) Handler(conn net.Conn) {
	//...当前连接的业务
	//fmt.Println("[INFO] Connection Established!")

	user := NewUser(conn)

	//用户上线，首先加入用户表
	this.mapLock.Lock()
	this.OnLineMap[user.Name] = user
	this.mapLock.Unlock()

	//然后，广播当前用户上线的消息给到全部user
	this.BroadCast(user, "Now is online! ")

	//[v0.3更新] 为了实现广播，先接收客户端发送的消息
	//上线之后新开了一个go程，实时监控当前连接用户是否下线
	go func() {
		buf := make([]byte, 4096) //配置一个4K长度的缓冲区数组
		for {
			// 从conn中读取数据（对端发来的消息）存到buf里
			n, err := conn.Read(buf)
			//Println(buf)
			if n == 0 {
				this.BroadCast(user, "Now is Offline! ")
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("[ERROR] Conn Read Error: ", err)
				return
			}

			msg := string(buf[:n-1]) //提取用户的消息（去除'\n'）

			//将msg广播出去
			this.BroadCast(user, msg)
		}
	}()

	//先保持handler仍然运作（阻塞），否则有可能让前面的go程终止
	select {}

}

// Start 启动服务器的接口
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

	//启动ListenMessager
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
