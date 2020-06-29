package svc2

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

func NewHTTPHandler(tracer opentracing.Tracer, service Service) http.Handler {
	//创建http service
	svc := &httpService{
		service: service,
	}

	/*
		ServeMux是HTTP请求多路复用器。
		它根据注册模式列表将每个传入请求的URL匹配，并为与URL最匹配的模式调用处理程序。
	*/
	mux := http.NewServeMux()

	//创建sum handler
	var sumHandler http.Handler
	sumHandler = http.HandlerFunc(svc.sumHandler)

	sumHandler = middleware.FromHTTPRequest(tracer, "Sum")(sumHandler)

	mux.Handle("/sum/", sumHandler)

	return mux
}

func (hs *httpService) sumHandler(w http.ResponseWriter, req *http.Request) {
	v := req.URL.Query()
	a, err := strconv.ParseInt(v.Get("a"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b, err := strconv.ParseInt(v.Get("b"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	result, err := hs.service.Sum(req.Context(), a, b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte(fmt.Sprintf("%d", result)))
}
