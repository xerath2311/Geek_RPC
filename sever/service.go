package sever

import (
	"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)

type methodType struct {
	method reflect.Method
	ArgType reflect.Type
	ReplyType reflect.Type
	numCalls uint64
}

func (m *methodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value
	if m.ArgType.Kind() == reflect.Ptr {
		//如果是指针，则argv也是指针
		argv = reflect.New(m.ArgType.Elem())
	}else {
		//如果是值，则argv也是值
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

func (m *methodType) newReplyv() reflect.Value {
	replyv := reflect.New(m.ReplyType.Elem())
	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(),0,0))
	}
	return replyv
}

type service struct {
	name string  //映射的结构体的名称
	typ reflect.Type  //结构体的类型
	rcvr reflect.Value  //结构体的实例本身
	method map[string]*methodType  //存储映射的结构体的所有符合条件的方法
}

// 创建service实例
func newService(rcvr interface{}) *service {
	s := new(service)
	s.rcvr = reflect.ValueOf(rcvr)
	s.name = reflect.Indirect(s.rcvr).Type().Name() //如果rcvr是指针，s.rcvr.Type().Name()为空
	s.typ = reflect.TypeOf(rcvr)
	if !ast.IsExported(s.name) {
		log.Fatalf("rpc server: %s is not a valid service name",s.name)
	}
	s.registerMethods()
	return s
}

//遍历实例s的所有实现的方法，并将符合rpc规则的方法加入到s.method里面
func (s *service) registerMethods() {
	s.method = make(map[string]*methodType)
	for i:=0;i<s.typ.NumMethod();i++{
		method := s.typ.Method(i)
		mType := method.Type
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}

		argType,replyType := mType.In(1),mType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}

		s.method[method.Name] = &methodType{
			method: method,
			ArgType: argType,
			ReplyType: replyType,
		}
		log.Printf("rpc server: register %s.%s\n",s.name,method.Name)
	}
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath()== ""
}

func (s *service) call(m *methodType, argv, replyv reflect.Value) error {
	atomic.AddUint64(&m.numCalls,1)
	f := m.method.Func
	returnValues := f.Call([]reflect.Value{s.rcvr,argv,replyv})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}