package pkg

import "fmt"

func Wrapper(f func()){
	defer func() {
		if err := recover();err != nil {
			fmt.Println(err)
		}
	}()
	f()
}
