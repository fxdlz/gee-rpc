package main

import (
	"context"
	geerpc "gee-rpc"
	"gee-rpc/registry"
	"gee-rpc/xclient"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

//func startServer(addr chan string) {
//	var foo Foo
//	if err := geerpc.Register(&foo); err != nil {
//		log.Fatal("register error:", err)
//	}
//	lis, err := net.Listen("tcp", ":9999")
//	if err != nil {
//		panic(err)
//	}
//	log.Println("star rpc server on ", lis.Addr())
//	//geerpc.HandleHTTP()
//	addr <- lis.Addr().String()
//	geerpc.Accept(lis)
//	//_ = http.Serve(lis, nil)
//
//}

func startRegistry(wg *sync.WaitGroup) {
	l, _ := net.Listen("tcp", ":9999")
	registry.HandleHTTP()
	wg.Done()
	_ = http.Serve(l, nil)
}

func startServer(registryAddr string, wg *sync.WaitGroup) {
	var foo Foo
	l, _ := net.Listen("tcp", ":0")
	server := geerpc.NewServer()
	_ = server.Register(&foo)
	registry.Heartbeat(registryAddr, "tcp@"+l.Addr().String(), 0)
	wg.Done()
	server.Accept(l)
}

type Foo int

type Args struct {
	Num1, Num2 int
}

func (foo *Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f *Foo) Sleep(args Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.Num1))
	*reply = args.Num1 + args.Num2
	return nil
}

//func call(addrCh chan string) {
//	client, _ := geerpc.DialHTTP("tcp", <-addrCh)
//	defer func() { _ = client.Close() }()
//
//	time.Sleep(time.Second)
//	// send request & receive response
//	var wg sync.WaitGroup
//	for i := 0; i < 10; i++ {
//		wg.Add(1)
//		go func(i int) {
//			defer wg.Done()
//			args := &Args{Num1: i, Num2: i * i}
//			var reply int
//			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
//				log.Fatal("call Foo.Sum error:", err)
//			}
//			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
//		}(i)
//	}
//	wg.Wait()
//}

func foo(xc *xclient.XClient, ctx context.Context, typ, serviceMethod string, args *Args) {
	var reply int
	var err error
	switch typ {
	case "call":
		err = xc.Call(ctx, serviceMethod, args, &reply)
	case "broadcast":
		err = xc.Broadcast(ctx, serviceMethod, args, &reply)
	}
	if err != nil {
		log.Printf("%s %s error: %v", typ, serviceMethod, err)
	} else {
		log.Printf("%s %s success: %d + %d = %d", typ, serviceMethod, args.Num1, args.Num2, reply)
	}
}

func call(registry string) {
	d := xclient.NewGeeRegistryDiscovery(registry, 0)
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = xc.Close() }()
	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "call", "Foo.Sum", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

func broadcast(registry string) {
	d := xclient.NewGeeRegistryDiscovery(registry, 0)
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = xc.Close() }()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "broadcast", "Foo.Sum", &Args{Num1: i, Num2: i * i})
			// expect 2 - 5 timeout
			ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
			foo(xc, ctx, "broadcast", "Foo.Sleep", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

func main() {
	log.SetFlags(0)
	registryAddr := "http://localhost:9999/_geerpc_/registry"
	var wg sync.WaitGroup
	wg.Add(1)
	go startRegistry(&wg)
	wg.Wait()

	time.Sleep(time.Second)
	wg.Add(2)
	go startServer(registryAddr, &wg)
	go startServer(registryAddr, &wg)
	wg.Wait()

	time.Sleep(time.Second)
	call(registryAddr)
	broadcast(registryAddr)
}
