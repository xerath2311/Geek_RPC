package client

import (
	"GeekRPC/codec_day1"
	"fmt"
	"log"
	"net"
	"sync"
	"testing"
	"time"
)

func startServer(addr chan string) {
	l,err := net.Listen("tcp",":0")
	if err != nil {
		log.Fatal("network error",err)
	}
	log.Println("start rpc server on",l.Addr())
	addr <- l.Addr().String()
	codec_day1.Accept(l)
}

func TestClient_Call(t *testing.T) {
	log.SetFlags(0)
	addr := make(chan string)
	go startServer(addr)
	client,_ := Dial("tcp",<- addr)
	defer func() {
		_ = client.Close()
	}()

	time.Sleep(time.Second)

	var wg sync.WaitGroup
	for i:=0;i<5;i++{
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			args := fmt.Sprintf("geerpc req %d",i)

			var reply string
			if err := client.Call("Foo.Sum",args,&reply);err!= nil {
				log.Fatal("call Foo.Sum error:",err)
			}
			log.Println("reply:",reply)
		}(i)
	}
	wg.Wait()
}
