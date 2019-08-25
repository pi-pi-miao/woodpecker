package session

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
	"woodpecker/pkg"
	"woodpecker/pkg/json"
	"woodpecker/pkg/logger"
)

type (
	Conn struct {
		c        net.Conn
		Id       string
		m        *pkg.Session
		fn       Woodpeck
		stop     chan bool
		ReadConn chan []byte
		Response chan []byte
		sendTime time.Time
		tryNum   int
		Lock     *sync.RWMutex
		Once     *sync.Once
	}
)

type Woodpeck interface {
	Do(c *Conn, message *pkg.Message)
}

func NewConn(conn net.Conn, id string) *Conn {
	return &Conn{
		c:        conn,
		Id:       id,
		m:        pkg.Sessions,
		Lock:     &sync.RWMutex{},
		sendTime: time.Now(),
		stop:     make(chan bool),
		ReadConn: make(chan []byte, 1000),
		Response: make(chan []byte, 1000),
		Once:     &sync.Once{},
	}
}

func (c *Conn) Start() {
	go pkg.Wrapper(c.read)
	go pkg.Wrapper(c.do)
	go pkg.Wrapper(c.write)
	go c.write()
	go pkg.Wrapper(c.synSend)
	go pkg.Wrapper(c.check)
}

func (c *Conn) read() {
	sizeData := make([]byte, 2)
	for {
		select {
		case <-c.stop:
			return
		default:
			if _, err := io.ReadFull(c.c, sizeData); err != nil {
				logger.Errorf("read conn header err %v", err)
				c.m.Manager.Delete(c.Id)
				c.Close()
				return
			}

			data := make([]byte, uint16(binary.LittleEndian.Uint16(sizeData)))
			fmt.Println("read header is len ", len(data))
			if _, err := io.ReadFull(c.c, data); err != nil {
				logger.Errorf("read conn body err %v", err)
				fmt.Println(err)
				c.m.Manager.Delete(c.Id)
				c.Close()
				return
			}
			fmt.Println("read body is ", string(data))
			c.ReadConn <- data
		}
	}
}

func (c *Conn) do() {
	for data := range c.ReadConn {
		message := &pkg.Message{
			Data: make([]byte, 0, 1024),
		}
		err := json.Unmarshal(data, message)
		if err != nil {
			logger.Errorf("unmarshal data err %v",err)
			continue
		}

		if len(message.Hearbeat) != 0 {
			c.estabLished(message.Hearbeat)
			continue
		}

		go c.fn.Do(c, message)
	}
}

func (c *Conn) Write(data []byte) {
	c.Response <- data
}

func (c *Conn) write() {
	for data := range c.Response {
		result := make([]byte, 2)
		binary.LittleEndian.PutUint16(result, uint16(len(data)))
		result = append(result, data...)
		if _, err := c.c.Write(result); err != nil {
			c.m.Manager.Delete(c.Id)
			c.Close()
		}
	}
}

func (c *Conn) Close() {
	c.Once.Do(func() {
		c.m.Manager.Delete(c.Id)
		close(c.stop)
		close(c.Response)
		close(c.ReadConn)
	})
}

func (c *Conn) synSend() {
	c.seq()
	t := time.NewTicker(295 * time.Second)
	for {
		select {
		case <-t.C:
			if c.tryNum >= 3 {
				c.Close()
				return
			}
			c.seq()
		case <-c.stop:
			t.Stop()
			return
		}
	}
}

func (c *Conn) seq() {
	data, err := c.hearbeat("syn_seq")
	if err != nil {
		logger.Errorf("marshal harbeat err %v", err)
	}
	c.Response <- data
	c.Lock.Lock()
	c.sendTime = time.Now()
	c.Lock.Unlock()
}

func (c *Conn) hearbeat(message string) ([]byte, error) {
	return json.Marshal(&pkg.Message{
		Hearbeat: message,
	})
}

func (c *Conn) check() {
	t := time.NewTicker(325 * time.Second)
	for {
		select {
		case <-c.stop:
			return
		case <-t.C:
			if c.tryNum >= 3 {
				c.Close()
				t.Stop()
				return
			}
			if time.Now().Unix()-c.sendTime.Unix() > 30 {
				c.Lock.Lock()
				c.sendTime = time.Now()
				c.tryNum++
				c.Lock.Unlock()
			}
		}
	}
}

func (c *Conn) estabLished(ack string) {
	if ack == "seq_ACK" {
		data, err := c.hearbeat("ACK_seq")
		if err != nil {
			logger.Errorf("marshal harbeat err %v", err)
		}
		c.Response <- data

		c.Lock.Lock()
		c.sendTime = time.Now()
		c.Lock.Unlock()
	}
}

func GetSessions(conn net.Conn, id string, f Woodpeck) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	c := NewConn(conn, id)
	c.fn = f
	pkg.Sessions.Manager.Set(id, c)
	c.Start()
}
