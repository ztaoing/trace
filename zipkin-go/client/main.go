package main

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin-contrib/zipkin-go-opentracing/examples/cli_with_2_services/svc1"
)

const (
	//服务名称
	serviceName = "client"
	//host
	hostPort = "0.0.0.0"
	//endpoint to send zipkin spans to
	zipkinHTTPendpoint = "http://127.0.0.1:9411/api/v1/spans"
	//debug 模式
	debug = false
	//service1的基础endpoint
	svc1Endpoint = "http://localhost:61001"

	sameSpan = true
	//traceid为128位
	traceID128Bit = true
)

func main() {
	/**
	首先创建collector，根据collector创建recorder，然后再使用recorder作为入参创建tracer
	*/
	//创建collector，参数为上报的zipkin endpoint
	collector, err := zipkin.NewHTTPCollector(zipkinHTTPendpoint)
	if err != nil {
		fmt.Printf("unable to create collector:%+v\n", err)
	}
	//创建recorder，入参为服务的相关信息和创建好的collector
	recorder := zipkin.NewRecorder(collector, debug, hostPort, serviceName)
	//创建tracer，并设置traceId为128位
	tracer, err := zipkin.NewTracer(recorder, zipkin.TraceID128Bit(true))

	//显示的设置zipkin为默认的tracer
	opentracing.InitGlobalTracer(tracer)

	//创建访问service1的client
	client := svc1.NewHTTPClient(tracer, svc1Endpoint)

	span := opentracing.StartSpan("Run")
	ctx := opentracing.ContextWithSpan(context.Background(), span)

	//调用concat方法
	span.LogEvent("call concat")
	res1, err := client.Concat(ctx, "Hello", "World")
	fmt.Printf("concat:%s err:%+v\n", res1, err)

	//调用sum方法
	span.LogEvent("call sum")
	res2, err := client.Sum(ctx, 10, 20)
	fmt.Printf("sum:%s err:+%v\n", res2, err)

	span.Finish()
	//关闭collector以确保spans在结束之前已经被发送
	collector.Close()

}
