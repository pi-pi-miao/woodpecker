package pkg

import (
	"woodpecker/safe_map"
)

var (
	SessionId   int64
	SessionIdCh = make(chan string,100)
	Sessions   *Session
)

type Session struct {
	SessionId string
	Manager   *safe_map.SyncMap
	Route     *safe_map.SyncMap
}

type Message struct {
	Route string      `json:"route"`
	Id    int64		  `json:"id"`
	Method string	  `json:"method"`
	Data   []byte     `json:"data"`
	Hearbeat string   `json:"hearbeat"`
}



