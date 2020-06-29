package svc2

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/openzipkin-contrib/zipkin-go-opentracing/examples/middleware"
	"io/ioutil"
	"net/http"
	"strconv"
)

type client struct {
	baseURL      string
	httpClient   *http.Client
	tracer       opentracing.Tracer
	traceRequest middleware.RequestFunc
}

func (c *client) Sum(ctx context.Context, a, b int64) (int64, error) {
	//创建span
	span, ctx := opentracing.StartSpanFromContext(ctx, "Sum")
	defer span.Finish()

	url := fmt.Sprintf("%s/sum/?a=%d&b=%d", c.baseURL, a, b)

	//创建http request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	//传递trace的context
	req = c.traceRequest(req.WithContext(ctx))

	//执行request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.SetTag("error", err.Error())
		return 0, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		span.SetTag("error", err.Error())
		return 0, err
	}
	result, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		span.SetTag("error", err.Error())
		return 0, err
	}
	return result, nil
}

func NewHTTPClient(tracer opentracing.Tracer, baseURL string) Service {
	return &client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
		tracer:     tracer,
		//ToHTTPRequest返回一个RequestFunc，它将在上下文中找到的OpenTracing Span注入HTTP标头。
		//如果找不到这样的跨度，则RequestFunc为noop。
		traceRequest: middleware.ToHTTPRequest(tracer),
	}
}
