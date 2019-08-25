package client

import (
	"io"
	"net"
	"fmt"
	"encoding/binary"
	"woodpecker/pkg/json"
	"sync"
	"woodpecker/safe_map"
)

type (
	Client struct {
		Ip string
		Port string
		NetWork string
		c net.Conn
		m map[int64]struct{}
		readData *safe_map.SyncMap
		rwLock *sync.RWMutex
	}
	Message struct {
		Route string      `json:"route"`
		Id    int64	      `json:"id"`
		Method string	  `json:"method"`
		Data   []byte     `json:"data"`
		Hearbeat string   `json:"hearbeat"`
	}
)

func NewClient(network,ip,port string)*Client{
	c := &Client{
		NetWork:network,
		Ip:ip,
		Port:port,
		m:make(map[int64]struct{}),
		rwLock:&sync.RWMutex{},
		readData:safe_map.New(),
	}
	c.conn().read()
	return c
}

func (c *Client)conn()*Client{
	var err error
	if c.c,err = net.Dial(c.NetWork,fmt.Sprintf("%v:%v",c.Ip,c.Port));err != nil {
		panic(err)
	}
	return c
}

func (c *Client)Write(route,method string,id int64,data []byte)*Client{
	sendData,err := json.Marshal(&Message{
		Route:route,
		Method:method,
		Data:data,
		Id:id,
	})
	if err != nil {
		panic(err)
	}
	result := make([]byte,2)
	binary.LittleEndian.PutUint16(result,uint16(len(sendData)))
	result = append(result,sendData...)
	if _,err := c.c.Write(result);err != nil {
		c.c.Close()
	}
	return c
}

func (c *Client)read(){
	size := make([]byte,2)
	go func() {
		defer func() {
			if err := recover();err != nil {
				fmt.Println(err)
			}
		}()
		for {

			if _,err := io.ReadFull(c.c,size);err != nil {
				fmt.Println(err)
				return
			}
			data := make([]byte,uint16(binary.LittleEndian.Uint16(size)))
			if _,err := io.ReadFull(c.c,data);err != nil {
				fmt.Println(err)
				return
			}
			message := &Message{}
			// 如果序列化错误这里应该有三次机会
			err := json.Unmarshal(data,message)
			if err != nil {
				continue
			}

			if len(message.Hearbeat) != 0 {
				if message.Hearbeat == "syn_seq" {
					message.Hearbeat = "seq_ACK"
					d,err := json.Marshal(message)
					if err != nil {
						continue
					}
					c.c.Write(d)
				}
				continue
			}
			c.readData.Set(fmt.Sprintf("%v",message.Id),message)
		}
	}()
}

// sync
func (c *Client)GetData(d int64)[]byte{
	id := fmt.Sprintf("%v",d)
	for {
		if m,ok := c.readData.Get(id);ok {
			c.readData.Delete(id)
			return m.(*Message).Data
		}
		continue
	}
}
