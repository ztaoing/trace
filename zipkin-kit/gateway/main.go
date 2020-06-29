package main

import (
	"flag"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/hashicorp/consul/api"
	"github.com/openzipkin/zipkin-go"
	zipkinhttpsvr "github.com/openzipkin/zipkin-go/middleware/http"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

//gateway在接收请求后，会创建一个span，其中的traceID将作为本次请求的唯一编号，
//gateway必须把这个traceID传递给字符串服务，字符串服务才能为该请求持续记录追踪信息

func main() {
	//创建环境变量
	var (
		consulHost = flag.String("consul.host", "127.0.0.1", "consul host")
		consulPort = flag.String("consul.port", "8500", "consul port")

		zipkinURL = flag.String("zipkin.url", "http://127.0.0.1:9411/api/v2/spans", "zipkin server url")
	)
	flag.Parse()

	//创建日志组件
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "time", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)

	}

	var zipkinTracer *zipkin.Tracer
	{
		var (
			err           error
			hostPort      = "localhost:9090"
			serviceName   = "gate-service"
			useNoopTracer = (*zipkinURL == "")
			reporter      = zipkinhttp.NewReporter(*zipkinURL)
		)
		defer reporter.Close()

		zEndpoint, err := zipkin.NewEndpoint(serviceName, hostPort)
		zipkinTracer, err = zipkin.NewTracer(
			reporter, zipkin.WithLocalEndpoint(zEndpoint), zipkin.WithNoopTracer(useNoopTracer),
		)
		if err != nil {
			logger.Log("err", err)
			os.Exit(-1)
		}

		if !useNoopTracer {
			logger.Log("tracer", "zipkin", "type", "Native", "URL", *zipkinURL)
		}

	}

	//创建consul api 客户端
	consulConfig := api.DefaultConfig()
	consulConfig.Address = "http://" + *consulHost + ":" + *consulPort
	consulClient, err := api.NewClient(consulConfig)

	if err != nil {
		logger.Log("err", err)
		os.Exit(-1)
	}

	//创建反向代理
	proxy := NewReverseProxy(consulClient, zipkinTracer, logger)

	tags := map[string]string{
		"component": "gateway_server",
	}

	handler := zipkinhttpsvr.NewServerMiddleware(
		zipkinTracer,
		zipkinhttpsvr.SpanName("gateway"),
		zipkinhttpsvr.TagResponseSize(true),
		zipkinhttpsvr.ServerTags(tags),
	)(proxy)

	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)

	}()

	//开始监听
	go func() {
		logger.Log("tansport", "HTTP", "addr", "9090")
		errc <- http.ListenAndServe(":9090", handler)
	}()

}

//创建反向dialing处理方法
func NewReverseProxy(client *api.Client, tracer *zipkin.Tracer, logger log.Logger) *httputil.ReverseProxy {
	//创建director
	director := func(req *http.Request) {
		//查询原始的请求路径
		reqPath := req.URL.Path
		if reqPath == "" {
			return
		}

		//获取服务名称
		pathArray := strings.Split(reqPath, "/")
		serviceName := pathArray[1]

		//根据服务名称查询实例列表
		result, _, err := client.Catalog().Service(serviceName, "", nil)
		if err != nil {
			logger.Log("reverseProxy failed", "query service instance error:", err.Error())
			return
		}
		if len(result) == 0 {
			logger.Log("reverseProxy failed", "no such service instance exist", err.Error())
			return
		}

		//重新组成请求路径
		destPath := strings.Join(pathArray[2:], "/")

		//随机选择一个服务实例
		targetInstance := result[rand.Int()%len(result)]
		logger.Log("service id ", targetInstance.ID)

		//设置代理服务器地址
		req.URL.Scheme = "http"
		req.URL.Host = fmt.Sprintf("%s:%d", targetInstance.Address, targetInstance.ServicePort)
		req.URL.Path = "/" + destPath
	}

	//为反向代理增加追踪逻辑，使用如下RoundTrip替代默认的Transport
	roundTrip, _ := zipkinhttpsvr.NewTransport(tracer, zipkinhttpsvr.TransportTrace(true))

	return &httputil.ReverseProxy{
		Director:  director,
		Transport: roundTrip,
	}
}
