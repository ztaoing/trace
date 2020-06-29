package svc1

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/openzipkin-contrib/zipkin-go-opentracing/examples/middleware"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

type client struct {
	baseURL      string
	httpClient   *http.Client
	tracer       opentracing.Tracer
	traceRequest middleware.RequestFunc
}

func (c *client) Concat(ctx context.Context, a, b string) (string, error) {
	//创建span，如果context有span，则作为父span，否则当前span为root span
	span, ctx := opentracing.StartSpanFromContext(ctx, "Concat")
	defer span.Finish()

	//组装请求的URL，两个请求参数
	//url.QueryEscape 对s进行转码使之可以安全的用在URL查询里
	url := fmt.Sprintf("%s/concat/?a=%s&b=%s", c.baseURL, url.QueryEscape(a), url.QueryEscape(b))

	//创建http请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	//传递trace的context
	req = c.traceRequest(req.WithContext(ctx))

	//执行请求
	resp, err := c.httpClient.Do(req)
	defer resp.Body.Close()

	//解析响应体
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//span 标注error
		span.SetTag("error", err.Error())
		return "", err
	}
	return string(data), nil
}

func (c *client) Sum(ctx context.Context, a, b int64) (int64, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Sum")
	defer span.Finish()

	//组装请求url，两个参数
	url := fmt.Sprintf("%s/sum/?a=%s&b=%s", c.baseURL, a, b)

	//创幻http请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	//传递trace的context
	req = c.traceRequest(req.WithContext(ctx))

	//执行请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	//解析响应体
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
