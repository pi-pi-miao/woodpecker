package main

import (
	"fmt"
	"time"
	"woodpecker/client"
	"woodpecker/pkg/json"
)

var (
	m = make(map[int64]bool)
)

type Message struct {
	Route    string `json:"route"`
	Id       int64  `json:"id"`
	Method   string `json:"method"`
	Data     []byte `json:"data"`
	Hearbeat string `json:"hearbeat"`
}

type data struct {
	D string `json:"d"`
}

func main() {
	c := client.NewClient("tcp", "127.0.0.1", "8080")
	d := &data{
		D: "hello 2312414",
	}
	result, _ := json.Marshal(d)
	c.Write("/", "get", 1, result)
	fmt.Println(string(c.GetData(1)))
	time.Sleep(10 * time.Second)
}
