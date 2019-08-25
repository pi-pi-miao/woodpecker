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