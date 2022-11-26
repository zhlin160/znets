package main

import "github.com/zhlin163/znets"

func main() {
	//srv := znets.NewServer()
	srv := znets.NewServerWithOptions(&znets.Options{
		IP:       "127.0.0.1",
		Port:     9898,
		WorkPool: 15,
	})

	srv.SetEventHandle(&Event{})
	srv.SetProtoPack(NewMessagePack())

	srv.Run()
}
