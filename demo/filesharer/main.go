package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"net"
	"os"

	"github.com/akzk/akgo/http"
)

const name = "filesharer"

var (
	localmode bool
	port      int
	shareDir  string
)

func main() {

	// 解析并检查命令行参数
	isexit := parseArgs()
	if isexit {
		return
	}

	// 启动服务
	server := http.NewServer()
	server.Base("/share")
	server.Get("", listFiles)
	server.Down("/files", shareDir)
	server.Up("/upload", shareDir)
	server.Serve(port)
}

func parseArgs() bool {

	// 注册命令行参数
	flag.BoolVar(&localmode, "l", false, "Transmission without router")
	flag.IntVar(&port, "p", 9999, "Port listened and served")
	flag.StringVar(&shareDir, "s", "/Users/leonardo/Desktop/share", "Directory shared")
	h := flag.Bool("h", false, "show helps")

	flag.Parse()

	// help
	if *h {
		fmt.Println(fmt.Sprintf("Usage: %s [-hl] [-p port] [-s sharedir]\n\nOptions:", name))
		flag.PrintDefaults()
		return true
	}

	// 检查路径是否存在
	_, err := os.Stat(shareDir)
	if err != nil {
		fmt.Println(fmt.Sprintf("[FATAL ERROR]: 分享目录路径(%s)不存在", shareDir))
		return true
	}

	return false
}

// 获取局域网IP地址，若失败则返回"127.0.0.1"
func getLANip() string {

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, address := range addrs {

		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}

		}
	}

	return "127.0.0.1"
}

func listFiles(c *http.Context) interface{} {

	// 参数结构体，get参数使用get标签，post参数使用json标签
	params := struct {
		Path string `get:"path"`
	}{}

	// 解析URL，填充参数结构体的get标签参数
	err := c.ParseURL(&params)
	if err != nil {
		return err
	}

	// 获取地址，用于a标签的资源链接
	ip := "127.0.0.1"
	if !localmode {
		ip = getLANip()
	}
	address := fmt.Sprintf("http://%s:%d", ip, port)

	// 获取目标文件夹路径
	path := shareDir + params.Path
	_, err = os.Stat(path)
	if err != nil {
		return err
	}

	// 读取目录下所有文件
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	// 整理首页、文件(夹)链接

	items := make(map[string]string)
	for _, file := range files {

		if file.Name()[:1] == "." {
			continue
		}

		if file.IsDir() {
			items[file.Name()] = address + "/share?path=" + path[len(shareDir):] + "/" + file.Name()
		} else {
			items[file.Name()] = address + "/files" + path[len(shareDir):] + "/" + file.Name()
		}
	}

	tplMap := make(map[string]interface{})
	tplMap["Items"] = items
	tplMap["Index"] = address

	// HTML模版渲染

	tpl := `<!DOCTYPE html>
	<html>
		<head>
			<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
			<title>file sharer</title>
		</head>
		<body>
			<a href="{{.Index}}">根目录</a>
			<br><br>
			<form action="/share/upload" method="post" enctype="multipart/form-data" >
			　　　<input type="file" id="upload" name="uploadfile"/><br>
			　　　<input type="submit"/>
			</form><br><br>
			{{range $key, $value := .Items}}
				<a href="{{$value}}">{{$key}}</a><br>
			{{end}}
		</body>
	</html>`

	t := template.New("index")
	t, err = t.Parse(tpl)
	if err != nil {
		return err
	}
	t.Execute(c.W, tplMap)

	return nil
}
