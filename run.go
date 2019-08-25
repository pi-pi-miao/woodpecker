package woodpecker

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"woodpecker/pkg"
	"woodpecker/pkg/logger"
	"woodpecker/safe_map"
	"woodpecker/session"
)

type Woodpecher struct {
	Stop   chan bool
	Fn     session.Woodpeck
	Once   *sync.Once
	file   string
	Config config
	Log    log
}

type config struct {
	Ip           string
	Port         int
	NetWork      string
	Route        string
	PprofAddr    string
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
	Lenth        int
	Env          string
}

type log struct {
	LogLevel string
	LogFile  string
	IsDebug  int
	DingUrl  string
}

func Run(f session.Woodpeck, file string) {
	if f == nil {
		panic("please input logic function")
	}
	w := &Woodpecher{
		Fn:   f,
		Stop: make(chan bool),
		Once: &sync.Once{},
	}

	go pkg.Wrapper(w.InitSessionID)
	w.file = file
	w.InitConfig().
	  InitLogger().
	  InitSession().
	  run()
}

func (w *Woodpecher) Close() {
	w.Once.Do(func() {
		pkg.Sessions.Manager.EachItem(func(item *safe_map.Item) {
			item.Value.(*session.Conn).Close()
		})
	})
}

func (w *Woodpecher) run() {
	go func(w *Woodpecher) {
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("framwork main goroutine panic err is %v", err)
			}
		}()
		fmt.Println(http.ListenAndServe(w.Config.PprofAddr, nil))
	}(w)

	if l, err := net.Listen(w.Config.NetWork, fmt.Sprintf("%v:%v", w.Config.Ip, w.Config.Port)); err == nil {
		fmt.Println("woodpecker is running...")
		for {
			select {
			case <-w.Stop:
				return
			default:
				if conn, err := l.Accept(); err == nil {
					go session.GetSessions(conn, <-pkg.SessionIdCh, w.Fn)
					continue
				}
				logger.Error(err)
				continue
			}
		}
	}
	panic("this service has An unexpected mistake")
}
