package main

import (
	"fmt"
	"github.com/danielh2942/markov_thingy/pkg/servsync"
)

func main() {
	data := servsync.New("1234")

	mp := servsync.SyncMap{}

	mp.Set("123", data)

	data1, ok := mp.Get("123")
	if !ok {
		fmt.Println("Failed to deref")
		return
	}

	data1.ChanId = "12345"

	if data.ChanId == data1.ChanId {
		fmt.Println("Great success!")
	} else {
		fmt.Println("data chanId", data.ChanId)
		fmt.Println("data1 chanId", data1.ChanId)
	}
}
