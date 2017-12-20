package iface

import (
	"fmt"
	"log"
	"reflect"

	"github.com/icexin/dueros/proto"
)

var (
	typeOfError   = reflect.TypeOf((*error)(nil)).Elem()
	typeOfMessage = reflect.TypeOf((*proto.Message)(nil)).Elem()
)

type service struct {
	name    string        // name of service
	rcvr    reflect.Value // receiver of methods for the service
	methods map[string]*reflect.Method
}

// Registry负责注册所有的用户接口对象，提供Dispatch方法来分发指令到具体的对象
// 同时也提供Context方法返回当前所有对象的状态
type Registry struct {
	services map[string]*service
}

// register adds a new service using reflection to extract its methods.
func (r *Registry) register(rcvr interface{}, name string) error {
	// Setup service.
	s := &service{
		name:    name,
		rcvr:    reflect.ValueOf(rcvr),
		methods: make(map[string]*reflect.Method),
	}
	rcvrType := reflect.TypeOf(rcvr)
	// Setup methods.
	for i := 0; i < rcvrType.NumMethod(); i++ {
		method := rcvrType.Method(i)
		mtype := method.Type
		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}
		// Method needs three ins: receiver, *Message
		if mtype.NumIn() != 2 {
			continue
		}
		// First argument must be a pointer and must be *Message.
		ctxType := mtype.In(1)
		if ctxType.Kind() != reflect.Ptr || ctxType.Elem() != typeOfMessage {
			continue
		}

		// Method needs one out: error.
		if mtype.NumOut() != 1 {
			continue
		}
		if returnType := mtype.Out(0); returnType != typeOfError {
			continue
		}
		s.methods[method.Name] = &method
	}
	if len(s.methods) == 0 {
		return fmt.Errorf("%q has no exported methods of suitable type",
			s.name)
	}
	if r.services == nil {
		r.services = make(map[string]*service)
	} else if _, ok := r.services[s.name]; ok {
		return fmt.Errorf("service already defined: %q", s.name)
	}
	r.services[s.name] = s
	return nil
}

// get returns a registered service given a method name.
//
// The method name uses a dotted notation as in "Service.Method".
func (r *Registry) get(namespace, name string) (*service, *reflect.Method, error) {
	service := r.services[namespace]
	if service == nil {
		err := fmt.Errorf("can't find service %q", namespace)
		return nil, nil, err
	}
	serviceMethod := service.methods[name]
	if serviceMethod == nil {
		err := fmt.Errorf("can't find method %q", name)
		return nil, nil, err
	}
	return service, serviceMethod, nil
}

func (r *Registry) getService(namespace string) *service {
	return r.services[namespace]
}

func (r *Registry) RegisterService(receiver interface{}, name string) error {
	return r.register(receiver, name)
}

func (r *Registry) Dispatch(m *proto.Message) error {
	serviceSpec, methodSpec, err := r.get(m.Header.Namespace, m.Header.Name)
	if err != nil {
		log.Printf("unhandled message: %s.%s", m.Header.Namespace, m.Header.Name)
		return err
	}

	retValue := methodSpec.Func.Call([]reflect.Value{
		serviceSpec.rcvr,
		reflect.ValueOf(m),
	})
	errInter := retValue[0].Interface()
	if errInter != nil {
		return errInter.(error)
	}
	return nil
}

type Contexter interface {
	Context() *proto.Message
}

func (r *Registry) Context() []*proto.Message {
	var ret []*proto.Message
	for _, s := range r.services {
		c, ok := s.rcvr.Interface().(Contexter)
		if ok {
			ret = append(ret, c.Context())
		}
	}
	return ret
}

func (r *Registry) GetService(namespace string) interface{} {
	service := r.getService(namespace)
	return service.rcvr.Interface()
}

var (
	DefaultRegistry = &Registry{
		services: make(map[string]*service),
	}
)

func RegisterService(receiver interface{}, name string) error {
	return DefaultRegistry.RegisterService(receiver, name)
}
