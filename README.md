# Akgo

Akgo is used for rapid development of RESTful APIs, web apps and backend services in Go.

## Directory Structure

```
├── demo			// 根据Akgo编写的例程
|	└── filesharer	// demo1: 静态文件上传下载服务器
├── http			// HTTP框架
|	├── errors		// 错误类型封装
|	├── context.go	// HTTP上下文封装
|	└── server.go	// 框架核心
└── lib				// 模块、库
	└── crypto		// 加解密模块
```

## Quick Start

### download and install

```
go get github.com/akzk/akgo
```

### create hello.go

```go
package main

import "github.com/akzk/akgo/http"

func main() {
  server := http.NewServer()
  server.Serve(8080)
}
```

### run hello.go

```
go run hello.go
```

### Go to http://localhost:8080

The browser will show

```
hello, welcome to akgo
```

Congratulations！You've just built your first **akgo** app.

## Document

### Create a RESTful API (GET method)

1. register URL and link the handler function

   ```
   func main() {
     server := http.NewServer()
     server.Get("/hello", sayhello)
     server.Serve(8080)
   }
   ```

2. define handler function

   ```go
   func sayhello(context *Context) interface{} {
     
     params := struct {
   		UserName string `get:"username"`
     }{}
     
     err := context.ParseURL(&params)
     if err != nil {
       return err
     }
     
     return []byte("hello, "+params.UserName)
   }
   ```

3. run hello.go

   ```
   go run hello.go
   ```

4. go to http://localhost:8080/hello?username=akzk

   will show

   ```
   hello, akzk
   ```

### Create a RESTful API (POST method)

1. register URL and link the handler function

   ```
   server.Post("/login", login)
   ```

2. define handler function

   ```go
   func login(context *Context) interface{} {
     
     params := struct {
   		UserName string `json:"username"`
   		Passwd	 string `json:"passwd"`
     }{}
     
     err := context.ParseBody(&params)
     if er != nil {
       return err
     }
     
     return []byte("login successfully, " + pararms.Username)
   }
   ```

3. access the api with json body

   this is a example

   ```
   POST / HTTP/1.1
   Host: localhost; Content-Type: application/x-www-form-urlencoded
   {"username": "akzk", "passwd": "123456"}
   ```

### Download and Upload files

1. register URLs

   ```
   server.Down("/download", "/usr/local/project/downloadfiles")
   server.Up("/upload", "/usr/local/project/uploadfiles")
   ```

   then, you can download file by "http://localhost:8080/download/path/to/file" and upload file by "http://localhost:8080/upload"