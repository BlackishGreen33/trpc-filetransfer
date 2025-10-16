# FileTransfer

trpc-go 支持流式 RPC，通过流式 RPC，客户端和服务器可以建立连续连接，连续发送和接收数据，让服务器提供连续的响应。

这里是一个文件传输的例子，通过流式 RPC 进行通信。

## Usage

- 启动 server 服务端.

```shell
$ go run server/main.go -conf server/trpc_go.yaml
```

- 启动 client 客户端.

```shell
$ go run client/main.go -conf client/trpc_go.yaml
```
