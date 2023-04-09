# my-go-redis
Learn to write redis using Golang

2023-04-04
为项目初始化mod文件
``` go
go mod init github.com/LynchQ/my-go-redis
```

新建main.go文件

2023-04-05
添加lib，config，tcp模块

2023-04-09
redis解析协议-RESP

什么是RESP？
Redis Serialization Protocol，Redis序列化协议，是Redis的通信协议，是一种文本协议，是一种简单的二进制协议，是一种基于行的协议。

它有五种类型的数据：
- 正常回复
  - 以“+”开头， 以“\r\n”结尾的字符串形式
  - eg：`+OK\r\n`
- 错误回复
  - 以“-”开头， 以“\r\n”结尾的字符串形式
  - eg: `-ERR unknown command 'foobar'\r\n`
- 整数
  - 以“:”开头， 以“\r\n”结尾的字符串形式
  - eg: `:1000\r\n`
- 多行字符串
  - 以“$”开头，后跟实际发送字节数， 以“\r\n”结尾的字符串形式
  - eg: `$6\r\nfoobar\r\n`，表示发送6个字节的字符串“foobar”
  - eg: `$0\r\n`，表示发送空字符串
  - eg: `$14\r\niiiii\r\niiiii\r\n`，表示发送14个字节的字符串
- 数组
  - 以“*”开头， 以“\r\n”结尾的字符串形式
  - eg: `*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n`
  - SET key value
  - `*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n`
实现自定义reply
实现ParseStrema方法，解析RESP协议
