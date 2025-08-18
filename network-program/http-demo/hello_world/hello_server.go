// go run hello_server.go
// curl 127.0.0.1:8081/hello
// curl -X POST http://127.0.0.1:8081/hello_post -d "name=rem&age=18"
package main

import (
	"fmt"
	"http-demo/logger"
	"io/ioutil"
	"net/http"
)

func HelloHandler(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprint(resp, "hello, world")
}

func HelloPostHandler(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(resp, "Only Post Method is allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(resp, "Error Reading body", http.StatusInternalServerError)
		return
	}
	defer req.Body.Close()

	logger.Info("Get Post body: ", string(body))

	fmt.Fprintf(resp, "Get: %s", body)
}

func main() {
	address := "127.0.0.1:8081"
	logger.Info("Listen and server on: ", address)

	http.HandleFunc("/hello", HelloHandler)
	http.HandleFunc("/hello_post", HelloPostHandler)

	err := http.ListenAndServe(address, nil)

	if err != nil {
		logger.Error("ListenAndServe error: ", err)
	}
}
