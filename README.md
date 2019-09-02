

![image](https://github.com/PyreneGitHub/woodpecker/blob/master/woodpecker.png?raw=true)

# woodpecker 是一个长连接框架

## 使用场景
```go

1、推送系统
2、长连接基础库
3、im系统
4、长连接网关
5、服务之间通信原生RPC
6、可以和quick结合使用
7、不同语言之间也可进行调用不过需要写相应的client
8、可以和app浏览器pc等不同端进行交互

```
## 微服务使用说明
```go
1、如果想要基于go做微服务可以使用如下
quick作为api网关，confcenter作为配置中心，woodpecker作为rpc进行调用

```

## 注意点
```go
1、关于心跳，这里设置为4分55秒，原因在于国内基站5分钟如果没有通信就会自动断开连接
2、关于连接数，这个只需要设置文件描述符的大小即可 https://jingyan.baidu.com/article/e4d08ffd64fd360fd2f60d27.html
3、关于一台机器支撑多少链接最好？
   这里建议如果是及时性比较高的服务，比如长连接网关等，8c16g可以控制在10w连接数以内，核心数量越多，每秒处理的链接也是越大延迟越小，64c128g服务器每秒处理能力在100w左右。如果只做推送系统来使用的话，4c8g支撑100w链接完全没问题
   
```

## 如何支撑高并发
```go
下面是一些使用本库如何解决高并发问题，以及后续版本会处理下面的问题3和使用者需要注意的问题

1、内部服务放弃短链接方案，短链接只有一个用处，那就是和浏览器交互，不过在未来，短链接技术是必然会被淘汰掉的
2、内部服务之间使用长连接池的方案，每个服务之间建议维护核心数相等的链接数量的连接池，比如一台机器或者一个容器服务只需要使用4c，那么建议在调用此服务的客户端在启动之前开启4个链接的连接池，这样可以服务端可以并行的进行处理，而且服务端只需要维护这四个链接即可，这样就把更多的性能放在了处理业务能力上面。如果做了连接池，那么需要在客户端做一个短暂的心跳
   连接池的方案在物理机部署的机器上面非常方便，但是如果在容器中部署并且使用hpa的方式，那么建议连接池的里面的链接数量越少越好，这样在频繁开启连接池的时候不会消耗太多的性能
3、关于消息丢失的问题
   为了防止消息的丢失，所以本库已经做好了防范措施，即做了消息确认机制，即下面示例中的c.Write("/","get",1,result)这里的1就是消息id，这个消息id不能够重复，建议在使用的时候做一个自增的id
   关于消息的超时，建议在客户端发送数据的时候做一个异步的消息检查机制，就是每次发送的时候把id和发送的时间放到一个新的容器里面，返回的时候就把这个消息id删除掉，定时遍历这个容器，根据当前时间减去id时间如果相隔太多，就说明服务端延迟比较高，就需要考虑对服务端进行扩容。这个功能会在未来版本中增加
4、关于消息的乱序
   因为服务端会并行的处理，所以返回来的结果并不能够保证按照顺序到达，这里已经做了处理，即通过消息id的方式，怎么设置消息id看第三条
5、服务端接收消息如何方式丢失
   客户端发送到服务端消息的时候，服务端可能接收到消息之后立刻这个容器就挂掉了，所以建议，在处理完毕这条数据的时候返回客户端的时候再把这个消息id返回，或者是这个消息确定已经存储成功之后再返回给客户端消息id
```



## 使用示例
server 端使用

```go
package main

import (
	"woodpecker"
	"woodpecker/session"
	"woodpecker/pkg"
	"fmt"
	"woodpecker/pkg/json"
	"flag"
)


var (
	f = flag.String("f","","")
)

type server struct {

}

func (s *server)Do(c *session.Conn,message *pkg.Message){
	message.Data = []byte("1111")
	data ,err := json.Marshal(message)
	if err != nil {
		return
	}
	c.Write(data)
}

func main(){
	flag.Parse()
	woodpecker.Run(&server{},*f)
	select{}
}
```

client 端使用

``` go
package main

import (
	"time"
	"woodpecker/client"
	"woodpecker/pkg/json"
	"fmt"
)

type data struct {
	D string  `json:"d"`
}

func main(){

	c := client.NewClient("tcp","127.0.0.1","8080")
	d := &data{
		D:"hello 2312414",
	}
	result,_:= json.Marshal(d)
	c.Write("/","get",1,result)
	fmt.Println(string(c.GetData(1)))

	time.Sleep(10*time.Second)
}
```

