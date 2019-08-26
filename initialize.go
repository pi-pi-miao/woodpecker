package woodpecker

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"sync/atomic"
	"woodpecker/pkg"
	"woodpecker/pkg/logger"
	"woodpecker/safe_map"
)

func (w *Woodpecher) InitSessionID() {
	for {
		select {
		case <-w.Stop:
			w.Close()
			return
		default:
			pkg.SessionIdCh <- fmt.Sprintf("%v", atomic.AddInt64(&pkg.SessionId, 1))
		}
	}
}

func (w *Woodpecher) InitConfig() *Woodpecher {
	if len(w.file) == 0 {
		w.file = "D:/project/src/woodpecker/conf/config.toml"
		//panic("config file is null")
	}
	configBytes, err := ioutil.ReadFile(w.file)
	if err != nil {
		fmt.Printf("read config err %v", err)
		panic(err)
	}
	if _, err := toml.Decode(string(configBytes), w); err != nil {
		fmt.Printf("toml decode err %v", err)
		panic(err)
	}
	return w
}

func (w *Woodpecher) InitLogger() *Woodpecher {
	if len(w.Log.DingUrl) == 0 && w.Log.IsDebug < 0 && len(w.Log.DingUrl) == 0 && len(w.Log.LogLevel) == 0 {
		panic(errors.New("please read config file and do InitConfig() function"))
	}
	if err := logger.InitLogger(&logger.LogConfig{
		LogLevel: w.Log.LogLevel,
		LogFile:  w.Log.LogFile,
		IsDebug:  w.Log.IsDebug,
	}, w.Config.Env, w.Log.DingUrl); err != nil {
		fmt.Printf("set logger file and level err %v", err)
		panic(err)
	}
	logger.Debug("this log init succ")
	return w
}

func (w *Woodpecher) InitSession() *Woodpecher {
	pkg.Sessions = &pkg.Session{
		Manager: safe_map.New(),
		Route:   safe_map.New(),
	}

	return w
}
