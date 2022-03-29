package sever

import (
	"fmt"
	"reflect"
	"testing"
)

func TestMethodType_NumCalls(t *testing.T) {
	mt := &methodType{numCalls: 23}
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
	_assert(len(s.method)==1,"wrong service Method,expect 1,but got %d",len(s.method))
	mType := s.method["Sum"]
	_assert(mType != nil,"wrong Method,Sum shouldn't nil")
}