package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn //user建立的连接

}

// 创建一个用户的方法
// 像C语言一样，需要声明返回值类型
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String() // 获取建立连接的用户的IP地址
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string), // 新建用户管道
		conn: conn,              //传入的参数

	}
	// 启动监听当前user channel消息的go程
	go user.ListenMessage()
	return user
}

// 监听当前USer chanel的方法
// 一旦有消息就立刻发给客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C                     //从当前用户管道中读取数据，C表示管道，(<-C)表示管道的头部数据
		this.conn.Write([]byte(msg + "\n")) //将msg写入User的操作

	}
}
