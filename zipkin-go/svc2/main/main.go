package main

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/openzipkin-contrib/zipkin-go-opentracing"
	"net/http"
	"os"
	"trace/zipkin-go/svc2"
)

const (
	serviceName = "svc2"
	hostPort    = "127.0.0.1:61002"

	zipkinHTTPEndpoint = "http://127.0.0.1:9411/api/v1/spans"
	debug              = false
	sameSpan           = true
	traceID128Bit      = true
)

//service2
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
	tracer, err := zipkintracer.NewTracer(recorder, zipkintracer.ClientServerSameSpan(sameSpan), zipkintracer.TraceID128Bit(traceID128Bit))

	if err != nil {
		fmt.Printf("unable to create zipkin tracer:%+v\n", err)
		os.Exit(-1)
	}

	//设置tracer
	opentracing.InitGlobalTracer(tracer)

	service := svc2.NewService()

	handler := svc2.NewHTTPHandler(tracer, service)

	fmt.Printf("starting %s on %s\n", serviceName, hostPort)
	http.ListenAndServe(hostPort, handler)

}
