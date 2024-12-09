package main

import (
	"context"
	geerpc "gee-rpc"
	"log"
	"net"
	"net/http"
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
	geerpc.HandleHTTP()
	addr <- lis.Addr().String()
	//geerpc.Accept(lis)
	_ = http.Serve(lis, nil)

}

type Foo int

type Args struct {
	Num1, Num2 int
}

func (foo *Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func call(addrCh chan string) {
	client, _ := geerpc.DialHTTP("tcp", <-addrCh)
	defer func() { _ = client.Close() }()

	time.Sleep(time.Second)
	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{Num1: i, Num2: i * i}
			var reply int
			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)
	}
	wg.Wait()
}

func main() {
	log.SetFlags(0)
	ch := make(chan string)
	go call(ch)
	startServer(ch)
}
