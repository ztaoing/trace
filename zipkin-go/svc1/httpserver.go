package svc1

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/openzipkin-contrib/zipkin-go-opentracing/examples/middleware"
	"net/http"
	"strconv"
)

type httpService struct {
	service Service
}

//涉及到concatHandler和sumHandler两个接口的处理器，在处理器中会解析请求的参数，并调用每个接口对应的service层实现
func NewHTTPHandler(tracer opentracing.Tracer, service Service) http.Handler {
	//创建http服务
	svc := &httpService{service: service}

	//创建mux
	mux := http.NewServeMux()

	//创建concatHandler
	var concatHandler http.Handler
	concatHandler = http.HandlerFunc(svc.concatHandler)

	//封装concat handler
	concatHandler = middleware.FromHTTPRequest(tracer, "Concat")(concatHandler)

	//创建sumHandler
	var sumHandler http.Handler
	sumHandler = http.HandlerFunc(svc.sumHandler)

	//封装sumHandler
	sumHandler = middleware.FromHTTPRequest(tracer, "Sum")(sumHandler)

	//注册到mux
	mux.Handle("/concat/", concatHandler)
	mux.Handle("/sum/", sumHandler)
	return mux
}

func (s *httpService) concatHandler(w http.ResponseWriter, req *http.Request) {
	v := req.URL.Query()
	result, err := s.service.Concat(req.Context(), v.Get("a"), v.Get("b"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	//返回处理结果
	w.Write([]byte(result))
}

func (s *httpService) sumHandler(w http.ResponseWriter, req *http.Request) {
	v := req.URL.Query()

	a, err := strconv.ParseInt(v.Get("a"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b, err := strconv.ParseInt(v.Get("b"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := s.service.Sum(req.Context(), a, b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write([]byte(fmt.Sprintf("%d", result)))

}
