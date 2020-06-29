package svc2

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"time"
)

//实现service
type svc2 struct {
}

func NewService() Service {
	return &svc2{}
}

func (s *svc2) Sum(ctx context.Context, a, b int64) (int64, error) {
	//解析context中的span
	span := opentracing.SpanFromContext(ctx)

	//key-value tag
	span.SetTag("service", "svc2")
	span.SetTag("string", "value1")
	span.SetTag("int", 123)
	span.SetTag("bool", true)

	//时间戳的事件，用于及时记录事件的存在
	//不推荐使用
	span.LogEvent("myEventAnnotation")

	s.fakeDBCall(span)
	//检查溢出的情况
	if b > 0 && a > (Int64Max-b) || (b < 0 && a < (Int64Min-b)) {
		span.SetTag("error", ErrIntOverFlow.Error())
		return 0, ErrIntOverFlow
	}
	return a + b, nil

}

//模拟访问DB
func (s *svc2) fakeDBCall(span opentracing.Span) {
	resourceSpan := opentracing.StartSpan(
		"query string",
		opentracing.ChildOf(span.Context()),
	)
	defer resourceSpan.Finish()

	//标识span的类型为resource
	ext.SpanKind.Set(resourceSpan, "resource")
	//命名resource
	ext.PeerService.Set(resourceSpan, "PostGreSQL")
	//resource的hostname
	ext.PeerHostname.Set(resourceSpan, "localhost")
	//resource的port
	ext.PeerPort.Set(resourceSpan, 3306)
	//标识执行的操作
	resourceSpan.SetTag("query", "SELECT recipes FROM cookbook WHERE topic='world domination'")

	//模拟SQL执行的耗时
	time.Sleep(20 * time.Millisecond)

}
