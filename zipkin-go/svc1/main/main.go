package main

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	zipkintracer "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin-contrib/zipkin-go-opentracing/examples/cli_with_2_services/svc1"
	"net/http"
	"os"
	"trace/zipkin-go/svc2"
)

const (
	serviceName = "svc1"
	hostPort    = "127.0.0.1:61001"

	zipkinHTTPEndpoint = "http://127.0.0.1:9411/api/v1/spans"
	debug              = false
	//service2的地址
	svc2Endpoint = "http://localhost:61002"

	sameSpan      = true
	traceID128Bit = true
)

func main() {
	//创建collector
	collector, err := zipkintracer.NewHTTPCollector(zipkinHTTPEndpoint)
	if err != nil {
		fmt.Printf("unable to create zipkin http collector:%+v\n", err)
		os.Exit(-1)
	}

	//创建recorder
	recorder := zipkintracer.NewRecorder(collector, debug, hostPort, serviceName)

	//创建tracer
	tracer, err := zipkintracer.NewTracer(recorder, zipkintracer.TraceID128Bit(traceID128Bit))
	if err != nil {
		fmt.Printf("unable to create zipkin tracer:%+v\n", err)
		os.Exit(-1)
	}

	//设置tracer
	opentracing.InitGlobalTracer(tracer)

	svc2Client := svc2.NewHTTPClient(tracer, svc2Endpoint)
	service := svc1.NewService(svc2Client)

	handler := svc1.NewHTTPHandler(tracer, service)

	fmt.Printf("starting %s on %s\n", serviceName, hostPort)
	http.ListenAndServe(hostPort, handler)
}
