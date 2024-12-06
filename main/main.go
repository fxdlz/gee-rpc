package main

import (
	"fmt"
	geerpc "gee-rpc"
	"gee-rpc/client"
	"log"
	"net"
	"sync"
	"time"
)

func startServer(addr chan string) {
	lis, err := net.Listen("tcp", ":10024")
	if err != nil {
		panic(err)
	}
	log.Println("star rpc server on ", lis.Addr())
	addr <- lis.Addr().String()
	geerpc.Accept(lis)
}

func main() {
	log.SetFlags(0)
	addr := make(chan string)
	go startServer(addr)
	cl, _ := client.Dial("tcp", <-addr)
	defer func() { _ = cl.Close() }()
	time.Sleep(time.Second)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var reply string
			args := fmt.Sprintf("geerpc req %d", i)
			if err := cl.Call("Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Println("reply:", reply)
		}(i)

	}
	wg.Wait()
}
