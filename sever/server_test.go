package sever

import (
	"GeekRPC/codec"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"testing"
	"time"
)

func startServer(addr chan string) {
	l,err := net.Listen("tcp",":0")
	if err != nil {
		log.Fatal("network error: ",err)
	}
	log.Println("start rpc server on ",l.Addr())
	addr <- l.Addr().String()
	Accept(l)
}

func TestNewServer(t *testing.T) {
	addr := make(chan string)
	go startServer(addr)

	conn,_ := net.Dial("tcp",<- addr)
	defer func(){
		_ = conn.Close()
	}()

	time.Sleep(time.Second)

	_ = json.NewEncoder(conn).Encode(DefaultOption)
	cc := codec.NewGobCodec(conn)

	for i:=0;i<5;i++{
		h := &codec.Header{
			ServiceMethod: "Foo.Sum",
			Seq: uint64(i),
		}
		_ = cc.Write(h,fmt.Sprintf("geerpc req %d",h.Seq))
		_ = cc.ReadHeader(h)
		var body string
		_ = cc.ReadBody(&body)
		log.Println("header:",h)
		log.Println("reply:",body)
	}
}
