package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/akzk/akgo/http/errors"
)

// Server 框架核心
type Server struct {
	base   string              // 所有注册URL的前部增加部分，减少重复填写
	router map[string]HttpFunc // string: API(URL+method), HttpFunc: 处理函数
	urls   []string            // 已注册的URL
	upDirs map[string]string   // string: URL, string: upload dir
}

// HttpFunc 使用框架的所有请求处理函数必须按照此格式
type HttpFunc func(context *Context) interface{}

// NewServer 返回已初始化的Server类
func NewServer() *Server {
	s := &Server{}
	s.base = ""
	s.router = make(map[string]HttpFunc)
	s.urls = []string{}
	s.upDirs = make(map[string]string)
	return s
}

// Base 为所有注册URL的增加前部，减少重复填写
func (s *Server) Base(base string) {
	s.base = base
}

// Get 注册支持GET模式的URL
func (s *Server) Get(pattern string, handler HttpFunc) {
	s.handleFunc("GET", pattern, handler)
}

// Post 注册支持POST模式的URL
func (s *Server) Post(pattern string, handler HttpFunc) {
	s.handleFunc("POST", pattern, handler)
}

// Down 为pattern路径提供HTTP协议的文件下载功能
func (s *Server) Down(pattern, res string) {
	if pattern[:len(pattern)-1] != "/" {
		pattern += "/"
	}
	s.checkDupliAPI(pattern, "GET")
	http.Handle(pattern, http.StripPrefix(pattern, http.FileServer(http.Dir(res))))
}

// Up 为pattern路径提供HTTP POST上传文件功能
func (s *Server) Up(pattern, target string) {

	// 检查target文件夹

	info, err := os.Stat(target)
	if err != nil {
		panic(err)
	}
	if !info.IsDir() {
		panic("Target pats is not a dir")
	}

	if target[:len(target)-1] != "/" {
		target += "/"
	}

	// 注册上传文件夹
	s.upDirs[s.base+pattern] = target

	// 注册URL
	s.Post(pattern, s.receiveFile)
}

// Serve 开始服务，代码阻塞
func (s *Server) Serve(port int) error {

	// 首页默认返回hello
	length := len(s.urls)
	for index, url := range s.urls {
		if url == s.base {
			break
		}
		if index == length-1 {
			s.Get("", func(context *Context) interface{} {
				return []byte("hello, welcome to akgo\n")
			})
		}
	}

	return http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

// checkDupliAPI 便利检查API是否重复注册，重复则panic
func (s *Server) checkDupliAPI(pattern, method string) {
	api := pattern + method
	for _, url := range s.urls {
		if url == api {
			panic(fmt.Sprintf("Duplication of API registration: %s %s", pattern, method))
		}
	}
}

// handleFunc 完成API注册
func (s *Server) handleFunc(method string, pattern string, handler HttpFunc) {

	pattern = s.base + pattern
	s.checkDupliAPI(pattern, method) // 重复注册api则报panic
	s.router[pattern+method] = handler

	// 使同一注册URL可同时接受GET、POST模式

	isAppend := true
	for _, url := range s.urls {
		if url == pattern {
			isAppend = false
		}
	}

	if isAppend {
		http.HandleFunc(pattern, s.workFunc)
		s.urls = append(s.urls, pattern)
	}
}

func (s *Server) workFunc(w http.ResponseWriter, r *http.Request) {

	context := &Context{}
	context.R = r
	context.W = w

	// URL匹配但method不匹配时会触发，即该URL上的该method未注册
	handler, ok := s.router[r.URL.Path+r.Method]
	if !ok {
		context.RespondErr(errors.ErrMethodNotSupport)
		return
	}

	// 分析处理函数的返回采取不同的封装措施，并在最后返回给客户端
	result := handler(context)
	if result != nil {
		err, ok := result.(error)
		if ok {
			// 未注册的错误
			context.RespondErr(errors.Err(err.Error()))
			return
		}

		Err, ok := result.(*errors.Error)
		if ok {
			// 已注册错误类型
			context.RespondErr(Err)
			return
		}

		bs, ok := result.([]byte)
		if ok {
			// 字节数组输出，不做任何处理
			w.Write(bs)
			return
		}

		// 处理函数成功响应
		jbody, err := json.Marshal(result)
		if err != nil {
			context.RespondErr(errors.Err(err.Error()))
		}
		w.Write(jbody)
	}

	// result 为nil时默认响应码200，包体为空，处理函数成功响应
}

func (s *Server) receiveFile(context *Context) interface{} {

	formFile, handler, err := context.R.FormFile("uploadfile")
	if err != nil {
		return err
	}
	defer formFile.Close()

	dir := s.upDirs[context.R.URL.Path]
	tmpname := "." + handler.Filename + "-" + strconv.FormatInt(time.Now().Unix(), 10)
	file, err := os.Create(dir + tmpname)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, formFile)
	if err != nil {
		return err
	}

	err = os.Rename(dir+tmpname, dir+handler.Filename)
	if err != nil {
		return err
	}

	return []byte("ok")
}