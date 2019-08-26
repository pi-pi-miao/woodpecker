package main

import (
	"flag"
	"fmt"
	"woodpecker"
	"woodpecker/pkg"
	"woodpecker/pkg/json"
	"woodpecker/session"
	"time"
)

var (
	f = flag.String("f", "", "")
)

type server struct {
}

func (s *server) Do(c *session.Conn, message *pkg.Message) {
	message.Data = []byte("1111")
	data, err := json.Marshal(message)
	if err != nil {
		fmt.Println("server marshal err", err)
		return
	}
	c.Write(data)
}

func main() {
	flag.Parse()
	w := woodpecker.Run(&server{}, *f)
	time.Sleep(5*time.Second)
	w.Close()
}
