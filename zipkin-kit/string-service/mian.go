package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-kit/kit/log"
	kitzipkin "github.com/go-kit/kit/tracing/zipkin"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	endpoint2 "trace/zipkin-kit/string-service/endpoint"
	"trace/zipkin-kit/string-service/service"
	"trace/zipkin-kit/string-service/transport"
)

func main() {
	var (
		consulHost = flag.String("consul.host", "127.0.0.1", "consul ip address")
		consulPort = flag.String("consul.port", "8500", "consul port")

		serviceHost = flag.String("service.host", "127.0.0.1", "service host")
		servicePort = flag.String("service.port", "9009", "service port")

		zipkinURL = flag.String("zipkin.url", "http://127.0.0.1:/api/v2/spans", "zipkin server url")

		//grpcAddr = flag.String("grpc", ":9008", "grpc listen address")
	)
	flag.Parse()

	ctx := context.Background()
	errChan := make(chan error)

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "time", log.DefaultTimestampUTC)
		logger = log.With(logger, "calller", log.DefaultCaller)
	}

	var zipkinTracer *zipkin.Tracer
	{
		var (
			err           error
			hostPort      = *serviceHost + ":" + *servicePort
			serviceName   = "string-service"
			useNoopTracer = (*zipkinURL == "")
			reporter      = zipkinhttp.NewReporter(*zipkinURL)
		)
		defer reporter.Close()

		zEndpoint, _ := zipkin.NewEndpoint(serviceName, hostPort)

		zipkinTracer, err = zipkin.NewTracer(reporter,
			zipkin.WithLocalEndpoint(zEndpoint),
			zipkin.WithNoopTracer(useNoopTracer),
		)
		if err != nil {
			logger.Log("err", err.Error())
			os.Exit(-1)
		}
		if !useNoopTracer {
			logger.Log("tracer", "zipkin", "type", "native", "url", *zipkinURL)
		}
	}

	var svc service.Service
	svc = service.StringService{}

	svc = LoggingMiddleware(logger)(svc)

	stringEndpoint := endpoint2.MakeStringEndpoint(ctx, svc)
	stringEndpoint = kitzipkin.TraceEndpoint(zipkinTracer, "string-endpoint")(stringEndpoint)

	//创建健康检查的endpoint
	healthEndpoint := endpoint2.MakeHealthCheckEndpoint(svc)
	healthEndpoint = kitzipkin.TraceEndpoint(zipkinTracer, "health-endpoint")(healthEndpoint)

	//封装
	endpts := endpoint2.StringEndpoints{
		StringEndpoint:      stringEndpoint,
		HealthCheckEndpoint: healthEndpoint,
	}

	//创建http.handler
	r := transport.MakeHttpHandler(ctx, endpts, zipkinTracer, logger)

	//创建注册对象
	register := Register(*consulHost, *consulPort, *serviceHost, *servicePort, logger)

	go func() {
		fmt.Println("http server start at port:" + *servicePort)
		//注册服务
		register.Register()
		handler := r
		errChan <- http.ListenAndServe(":"+*servicePort, handler)

	}()

	//grpc
	/*go func() {
		fmt.Println("grpc server start at port " + *grpcAddr)
		listener, err := net.Listen("tcp", *grpcAddr)

		if err != nil {
			errChan <- err
			return
		}

	}()*/

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	error := <-errChan

	//服务退出取消注册
	register.Deregister()

	fmt.Println(error)

}
