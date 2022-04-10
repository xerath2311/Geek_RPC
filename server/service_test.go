package server

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestMethodType_NumCalls(t *testing.T) {
	mt := &MethodType{numCalls: 23}
	fmt.Println(mt.NumCalls())
}

type Foo int

type Args struct {
	Num1 int
	Num2 int
}

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func _assert(condition bool,msg string,v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertion failed: "+msg,v...))
	}
}

func TestNewService(t *testing.T) {
	var foo Foo
	T := reflect.TypeOf(&foo)

	fmt.Println(T.Method(0).Name)
	s := newService(&foo)
	_assert(len(s.method)==1,"wrong Service Method,expect 1,but got %d",len(s.method))
	mType := s.method["Sum"]
	_assert(mType != nil,"wrong Method,Sum shouldn't nil")
}

func TestOther(t *testing.T) {
	arg := Args{
		Num1: 1,
		Num2: 2,
	}
	fmt.Println(reflect.ValueOf(arg).Type().Name())
	fmt.Println(reflect.Indirect(reflect.ValueOf(arg)).Type().Name())
	fmt.Println(reflect.TypeOf(arg).Name())

}

func TestSelect(t *testing.T) {
	A := make(chan struct{})
	B := make(chan struct{})
	ctx,cancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		defer func() {
			fmt.Println("gorutine exit")
		}()

		time.Sleep(time.Second*2)

		select {
		case A <- struct{}{}:
		case <-ctx.Done():
			fmt.Println("ctx done A")
			return
		}

		fmt.Println("A")
		B <- struct{}{}
		fmt.Println("B")
	}(ctx)
	select {
	case <-time.After(time.Second):
		fmt.Println("time after")
		cancel()
		time.Sleep(time.Second*2)
	case <-A:
		<-B
		fmt.Println("action AB")
	}
}