package main

import (
	geerpc "gee-rpc"
	"log"
	"net"
	"sync"
	"time"
)

func startServer(addr chan string) {
	var foo Foo
	if err := geerpc.Register(&foo); err != nil {
		log.Fatal("register error:", err)
	}
	lis, err := net.Listen("tcp", ":10024")
	if err != nil {
		panic(err)
	}
	log.Println("star rpc server on ", lis.Addr())
	addr <- lis.Addr().String()
	geerpc.Accept(lis)
}

type Foo int

type Args struct {
	Num1, Num2 int
}

func (foo *Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func main() {
	log.SetFlags(0)
	addr := make(chan string)
	go startServer(addr)
	cl, _ := geerpc.Dial("tcp", <-addr)
	defer func() { _ = cl.Close() }()
	time.Sleep(time.Second)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var reply int
			args := Args{
				Num1: i,
				Num2: i,
			}
			if err := cl.Call("Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)

	}
	wg.Wait()
}
