[运行方法]

1、先运行Server服务
`go run server/main.go`

2、运行pointer节点
`go run pointer/main.go`

3、运行client节点
`go run client/main.go`


[测试方法]

curl -x socks5://127.0.0.1:1080 http://www.baidu.com

[项目架构]

![img](./docs/架构.jpg)