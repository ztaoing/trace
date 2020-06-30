package main

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"github.com/pborman/uuid"
	"os"
	"strconv"
)

//服务的注册
func Register(consulHost, consulPort, svcHost, svcPort string, logger log.Logger) (register sd.Registrar) {
	//创建Consul consul.client
	var client consul.Client
	{
		consulConfig := api.DefaultConfig()
		consulConfig.Address = consulHost + ":" + consulPort
		consulClient, err := api.NewClient(consulConfig)
		if err != nil {
			logger.Log("create consul client error:", err)
			os.Exit(-1)
		}
		client = consul.NewClient(consulClient)
	}

	//设置consul对服务健康检查的参数
	check := api.AgentServiceCheck{
		HTTP:     "http://" + svcHost + ":" + svcPort + "/health",
		Interval: "10s",
		Timeout:  "1s",
		Notes:    "consul check service health stauts",
	}
	port, _ := strconv.Atoi(svcPort)

	//设置微服务向consul注册信息
	reg := api.AgentServiceRegistration{
		ID:      "string-service" + uuid.New(),
		Name:    "string-service",
		Address: svcHost,
		Port:    port,
		Tags:    []string{"string-service"},
		Check:   &check,
	}

	//执行注册
	register = consul.NewRegistrar(client, &reg, logger)
	return
}
