package server

import (
	"encoding/json"
	"fmt"
	"github.com/mrpc/codec"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
)

var DefaultServer = NewServer()

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}


// 阻塞等待socket 链接 建立，起go routine 去serve
func (s *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept err:", err)
			return
		}
		go s.ServeConn(conn)
	}
}

func (s *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() { _ = conn.Close() }() // 处理完毕关闭链接
	var opt Option
	// 根据约定的json格式读option，从中拿到具体内容的解析规则
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options error: ", err)
		return
	}

	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}

	// 拿到具体的编解码器(gob)的构造函数，可以有多种配置
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}

	// 构造codec 进行处理
	s.serveCodec(f(conn))
}

var invalidRequest = struct{}{}

type request struct {
	h            *codec.Header
	argv, replyv reflect.Value
}

// 根据codec 进入处理流程，主要包括三个阶段 readRequest handleRequest, sendResponse
func (s *Server) serveCodec(cc codec.Codec) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)

	for {
		req, err := s.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			s.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}

		// 一个链接可以处理多个请求
		wg.Add(1)
		go s.handleRequest(cc, req, sending, wg)
	}
	wg.Wait()
	_ = cc.Close()
}

// decode requestHeader: service method seq
func (s *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

// readRequest: 解析header 和 body，组装 request
func (s *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := s.readRequestHeader(cc) // 解析header
	if err != nil {
		return nil, err
	}

	req := &request{h: h}
	// todo: 目前只支持string
	req.argv = reflect.New(reflect.TypeOf(""))
	if err = cc.ReadBody(req.argv.Interface()); err != nil {
		log.Println("rpc server: read argv err:", err)
	}
	return req, nil
}

// sendResponse 加锁避免sending时 乱序，导致客户端无法解析
func (s *Server) sendResponse(cc codec.Codec, header *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()

	if err := cc.Write(header, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

func (s *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	// todo: 这里应该去调用rpc method
	defer wg.Done()
	log.Println("server handle request:", req.h, req.argv.Elem())
	req.replyv = reflect.ValueOf(fmt.Sprintf("rps resp %d", req.h.Seq))
	s.sendResponse(cc, req.h, req.replyv.Interface(), sending)
}
