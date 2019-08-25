 ![image](https://github.com/PyreneGitHub/woodpecker/blob/master/woodpecker.png)



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
	fmt.Println("get message is ",message)
	message.Data = []byte("1111")
	data ,err := json.Marshal(message)
	if err != nil {
		fmt.Println("server marshal err",err)
		return
	}
	fmt.Println("len is ",len(data))
	c.Write(data)
	fmt.Println("send client ok")
}

func main(){
	flag.Parse()
	woodpecker.Run(&server{},*f)
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

var (
	m = make(map[int64]bool)
)

type Message struct {
	Route string      `json:"route"`
	Id    int64	      `json:"id"`
	Method string	  `json:"method"`
	Data   []byte     `json:"data"`
	Hearbeat string   `json:"hearbeat"`
}

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

