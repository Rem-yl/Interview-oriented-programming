package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rem/load-balancer/pkg/algo"
	"github.com/rem/load-balancer/pkg/backend"
	"github.com/rem/load-balancer/pkg/errs"
)

var Version = "0.0"
var BuildTime = time.Now().String()

func buildBalancer(path string) algo.LoadBalanceAlgo {
	serverList := backend.NewSimpleBackEndFromYaml(path)
	backendList := make([]backend.BackEnd, len(serverList))
	for i, backend := range serverList {
		backendList[i] = backend
	}
	balancer := algo.NewRoundRobinLoadBalancer(backendList)

	return balancer
}

var balancer = buildBalancer("configs/simple_backend.yaml")

func loadBalanceHandler(c *gin.Context) {
	backend, err := balancer.GetBackEnd()
	if err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{
			"data": errs.ErrNoBackEnd.Error(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"name": backend.GetName(),
			"url":  backend.GetURL(),
		},
	})

}

func main() {
	fmt.Printf("Version: %s \n", Version)
	fmt.Printf("BuildTime: %s \n", BuildTime)
	fmt.Println("Welcome to load balancer project!")

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": "curl /balancer please",
		})
	})

	r.GET("/balancer", loadBalanceHandler)

	r.Run(":8187")
}
