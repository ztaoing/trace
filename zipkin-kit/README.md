# go-kit 集成zipkin
* go-kit服务架构的tracing包为服务提供了dapper样式的请求追踪。
* go-kit支持opentracing Api，并使用opentracing-go包为其服务器和客户端提供追踪中间。
* go-kit在tracing中默认添加了zipkin的支持

# HTTP调用方式的链路追踪
* 在网关gateway中增加链路追踪的采集逻辑，在反向代理中追加tracer设置
* gateway作为链路追中的第一站和最后一站，我们需要截获到达gateway的所有请求，记录追踪信息。
* gateway在接收请求后，会创建一个span，其中的traceID将作为本次请求的唯一编号，gateway必须把这个traceID传递给字符串服务，字符串服务才能为该请求持续记录追踪信息


