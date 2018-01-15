package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/akzk/akgo/module/crypto"
)

// Context 供请求处理函数使用的输入参数
type Context struct {
	R *http.Request
	W http.ResponseWriter
}

// Response 响应报文
type Response struct {
	code    int               // http status code
	headers map[string]string // 响应头部
	body    []byte
}

// NewResponse 创建Response，一般在HttpFunc函数结尾使用
// code, http status code
// headers, http headers
// msg, 简述
// text, 详情
func (c *Context) NewResponse(code int, headers map[string]string, body []byte) *Response {
	return &Response{code, headers, body}
}

// ParseBody 定义了POST Body的解析方式
func (c *Context) ParseBody(dst interface{}) error {
	return c.parseBodyWithChoice(dst, false)
}

// ParseBodyAfterDeRSA 进行RSA解密后才进行json解析
func (c *Context) ParseBodyAfterDeRSA(dst interface{}) error {
	return c.parseBodyWithChoice(dst, true)
}

// 选择是否先进行RSA解密
func (c *Context) parseBodyWithChoice(dst interface{}, deRsa bool) error {
	body, err := ioutil.ReadAll(c.R.Body)
	if err != nil {
		return err
	}
	if deRsa {
		body, err = crypto.DeRSA(body)
		if err != nil {
			return err
		}
	}
	err = json.Unmarshal(body, dst)
	if err != nil {
		return err
	}
	return nil
}

// ParseURL 解析GET参数
func (c *Context) ParseURL(dst interface{}) error {

	params, err := url.ParseQuery(c.R.URL.RawQuery)
	if err != nil {
		return err
	}

	vs := reflect.ValueOf(dst).Elem()
	ts := vs.Type()

	for i := 0; i < ts.NumField(); i++ {

		t := ts.Field(i)
		v := vs.Field(i)

		key := t.Tag.Get("get")
		if len(params[key]) > 0 {
			value := params[key][0]
			if key != "" { // 存在get标签且不为空字符串
				switch v.Kind() {
				case reflect.String:
					v.SetString(value)
				case reflect.Int64:
					tmp, err := strconv.ParseInt(value, 10, 64)
					if err != nil {
						continue
					}
					v.SetInt(tmp)
				case reflect.Float64:
					tmp, err := strconv.ParseFloat(value, 64)
					if err != nil {
						continue
					}
					v.SetFloat(tmp)
				case reflect.Bool:
					if value == "1" {
						v.SetBool(true)
					}
				}
			}
		}
	}
	return nil
}

// SendErr 向客户端返回JSON格式的错误信息
func (c *Context) SendErr(err *Error) {
	c.W.WriteHeader(err.Code)

	// 返回包体
	body := struct {
		Msg  string `json:"msg"`
		Text string `json:"text"`
	}{err.Msg, err.Text}

	jbody, _ := json.Marshal(body)
	c.W.Write(jbody)
}

func (c *Context) sendResponse(response *Response) {
	c.W.WriteHeader(response.code)

	if response.headers != nil {
		for key, value := range response.headers {
			c.W.Header().Add(key, value)
		}
	}

	if response.body != nil {
		c.W.Write(response.body)
	}
}
