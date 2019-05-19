package server

import (
	"Ming-RPC/codec"
	"Ming-RPC/network"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
)

// TODO 暂时把所有service放在map中，后续接入Etcd实现服务注册与发现
type Server struct {
	addr     string
	services map[string]reflect.Value
}

// New Server
func New(addr string) *Server {
	return &Server{addr: addr, services: make(map[string]reflect.Value)}
}

// 注册方法
func (s *Server) Register(name string, f interface{}) {
	if _, ok := s.services[name]; ok {
		return
	}
	s.services[name] = reflect.ValueOf(f)
}

// Run
func (s *Server) Run() {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Printf("listen on %s err: %v\n", s.addr, err)
		return
	}
	for {
		conn, err := l.Accept() //  Accept()方法会阻塞
		if err != nil {
			log.Printf("accept err: %v\n", err)
			continue
		}

		//  由于goroutine比线程简便得多，而且Golang中的IO底层实现方式和java NIO模型一致，都采用了EPOLL
		//  所以可以为每个连接都创建一个新的goroutine来处理
		go func(s *Server, conn net.Conn) {
			defer conn.Close()
			socket := network.New(&conn)
			for {
				data, err := socket.Receive() // 这里也会阻塞
				if err != nil {
					if err != io.EOF {
						log.Printf("read err: %v\n", err)
					}
					return
				}
				req, err := codec.Decode(data)
				checkError(err, "decode err")
				f, ok := s.services[req.Name]
				if !ok {
					e := fmt.Sprintf("func %s does not exist", req.Name)
					log.Println(e)
					response := codec.Data{Name: req.Name, Err: e}
					responseBytes, err := codec.Encode(response)
					checkError(err, "encode err")
					if err = socket.Send(&responseBytes); err != nil {
						log.Printf("transport write err: %v\n", err)
					}
					continue
				}
				log.Printf("func %s is called\n", req.Name)
				// 把请求列表（interface{}）转成reflect.Value类型
				inArgs := make([]reflect.Value, len(req.Args))
				for i := range req.Args {
					inArgs[i] = reflect.ValueOf(req.Args[i])
				}
				// call 对应 service
				out := f.Call(inArgs)
				// 把reflect.Value类型转回interface{}（不包括err，所以len - 1）
				outArgs := make([]interface{}, len(out)-1)
				for i := 0; i < len(out)-1; i++ {
					outArgs[i] = out[i].Interface()
				}
				// 处理error
				var e string
				if _, ok := out[len(out)-1].Interface().(error); !ok {
					e = ""
				} else {
					e = out[len(out)-1].Interface().(error).Error()
				}
				// send response to client
				response := codec.Data{Name: req.Name, Args: outArgs, Err: e}
				responseBytes, err := codec.Encode(response)
				checkError(err, "encode err")
				err = socket.Send(&responseBytes)
				checkError(err, "socket write err")
			}
		}(s, conn)
	}
}

func checkError(err error, logValue string) {
	if err != nil {
		log.Printf("%s: %v\n", logValue, err)
	}
}
