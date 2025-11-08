package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rem/load-balancer/pkg/algo"
	"github.com/rem/load-balancer/pkg/backend"
	"github.com/rem/load-balancer/pkg/config"
	"github.com/rem/load-balancer/pkg/errs"
)

var (
	Version   = "0.0"
	BuildTime = time.Now().String()
)

var availableBalancer = []string{"round_robin", "swrr_robin", "consistence_hash"}

// ---- 构建负载均衡器 ----
func buildBalancer(mode string, serverList []*config.Server) algo.LoadBalanceAlgo {
	var balancer algo.LoadBalanceAlgo
	backendList := make([]backend.BackEnd, len(serverList))
	for i, server := range serverList {
		backend := backend.NewSimpleBackEnd(server.URL, server.Name, server.Weight)
		backendList[i] = backend
	}

	switch mode {
	case "round_robin":
		balancer = algo.NewRoundRobinLoadBalancer(backendList)
	case "swrr_robin":
		balancer = algo.NewSwrrRobinLoadBalancer(backendList)
	case "consistence_hash":
		balancer = algo.NewConsistenceHashLoadBalancer(backendList)
	default:
		panic(fmt.Sprintf("Load balancer mode: %s not found! You can use balancer in %v \n", mode, availableBalancer))
	}

	fmt.Printf("Use mode: %s to build load balancer! \n", mode)

	return balancer
}

var balancer algo.LoadBalanceAlgo

// ---- handler ----
func loadBalanceHandler(c *gin.Context) {
	backend, err := balancer.GetBackEnd()
	if err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{
			"data": errs.ErrNoBackEnd.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"name": backend.GetName(),
			"url":  backend.GetURL(),
		},
	})
}

func main() {
	var configPath string
	// 命令行参数：支持指定配置文件路径
	flag.StringVar(&configPath, "config", "configs/simple_backend.yaml", "path to config yaml file")
	flag.Parse()

	cfg, err := config.LoadSimpleBackendConfig(configPath)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Version: %s \n", Version)
	fmt.Printf("BuildTime: %s \n", BuildTime)
	fmt.Printf("Using config file: %s\n", configPath)

	balancer = buildBalancer(cfg.LoadBalancer.Mode, cfg.Servers)

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "curl /balancer please"})
	})
	r.GET("/balancer", loadBalanceHandler)

	r.Run(cfg.LoadBalancer.URL)
}
