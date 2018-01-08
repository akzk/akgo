package main

import "github.com/akzk/akgo/http"

func main() {
	server := http.NewServer()
	server.Get("/hello", sayhello)
	server.Post("/login", login)
	server.Serve(8080)
}

func sayhello(context *http.Context) interface{} {

	params := struct {
		UserName string `get:"username"`
	}{}

	err := context.ParseURL(&params)
	if err != nil {
		return err
	}

	return []byte("hello, " + params.UserName)
}

func login(context *http.Context) interface{} {

	params := struct {
		UserName string `json:"username"`
		Passwd   string `json:"passwd"`
	}{}

	err := context.ParseBody(&params)
	if err != nil {
		return err
	}

	return []byte("login successfully, " + params.UserName)
}
