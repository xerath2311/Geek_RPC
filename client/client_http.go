package client

import (
	"GeekRPC"
	"GeekRPC/server"
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

func NewHTTPClient(conn net.Conn,opt *server.Option) (*Client,error) {
	_,_ = io.WriteString(conn,fmt.Sprintf("CONNECT %s HTTP/1.0\n\n",GeekRPC.DefaultRPCPath ))
	
	resp,err := http.ReadResponse(bufio.NewReader(conn),&http.Request{Method: "CONNECT",})
	if err == nil && resp.Status == GeekRPC.Connected {
		return NewClient(conn,opt)
	}
	if err == nil {
		err = errors.New("unexpected HTTP response: " + resp.Status)
	}
	return nil,err
}

func DialHTTP(network,address string,opts ...*server.Option)(*Client,error) {
	return dialTimeout(NewHTTPClient,network,address,opts...)
}

func XDial(rpcAddr string, opts ...*server.Option) (*Client, error) {
	parts := strings.Split(rpcAddr,"@")
	if len(parts) != 2 {
		return nil,fmt.Errorf("rpc client err: wrong format '%s' ,expect protocol@addr",rpcAddr)
	}

	protocol,addr := parts[0],parts[1]
	switch protocol {
	case "http":
		return DialHTTP("tcp",addr,opts...)
	default:
		return Dial(protocol,addr,opts...)
	}
}