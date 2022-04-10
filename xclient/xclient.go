package xclient

import (
	"GeekRPC/client"
	"GeekRPC/server"
	"context"
	"io"
	"reflect"
	"sync"
)

type XClient struct {
	d Discovery
	mode SelectMode
	opt *server.Option
	mu sync.Mutex
	clients map[string]*client.Client
}

var _ io.Closer = (*XClient)(nil)

func NewXClient(d Discovery, mode SelectMode, opt *server.Option) *XClient {
	return &XClient{
		d: d,
		mode: mode,
		opt: opt,
		clients: make(map[string]*client.Client),
	}
}

func (xc *XClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()

	for key,client := range xc.clients {
		_ = client.Close()
		delete(xc.clients,key)
	}
	
	return nil
}

func (xc *XClient) dial(rpcAddr string) (*client.Client, error) {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	
	clt,ok := xc.clients[rpcAddr]
	if ok && !clt.IsAvailable() {
		_ = clt.Close()
		delete(xc.clients,rpcAddr)
		clt = nil
	}
	if clt == nil {
		var err error
		clt,err = client.XDial(rpcAddr,xc.opt)
		if err != nil {
			return nil,err
		}
		xc.clients[rpcAddr] = clt
	}
	
	return clt,nil
}

func (xc *XClient) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	rpcAddr,err := xc.d.Get(xc.mode)
	if err != nil {
		return err
	}
	return xc.call(rpcAddr,ctx,serviceMethod,args,reply)
}

func (xc *XClient) call(rpcAddr string, ctx context.Context, serviceMethod string, args, reply interface{}) error {
	client ,err := xc.dial(rpcAddr)
	if err != nil {
		return err
	}
	return client.Call(ctx,serviceMethod,args,reply)
}

func (xc *XClient) Broadcast(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	servers,err := xc.d.GetAll()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var e error

	replyDone := reply==nil
	ctx,cancel := context.WithCancel(ctx)

	for _,rpcAddr := range servers {
		wg.Add(1)
		go func(rpcAddr string) {
			defer wg.Done()
			var cloneReply interface{}
			if reply != nil {
				cloneReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}
			err := xc.call(rpcAddr,ctx,serviceMethod,args,cloneReply)

			mu.Lock()
			if err != nil && e==nil {
				e = err
				cancel()
			}
			if err == nil && !replyDone {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(cloneReply).Elem())
				replyDone = true
			}

			mu.Unlock()
		}(rpcAddr)
	}

	wg.Wait()
	return e
}












