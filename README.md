# IM-ChattingRoom
使用Go语言完成的一个Socket网络聊天室（跟随教程）

## 1. 版本详情
---
### V0.1 实现了服务器端
---
能够通过本地IP和8888端口登录server

### V0.2 实现了用户上线
---
新用户上线后更新server的用户表，并将该消息广播给全部用户

## 2.运行方法
---
Compile:
```bash
go build -o server.exe main.go server.go user.go
```
Run:

- 启动server:
```
./server
```

- 启动user(任意新建多个PS):
```
curl **--http0.9** 127.0.0.1:8888 
```
