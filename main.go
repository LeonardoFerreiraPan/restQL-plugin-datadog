package restqdatadog

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/DataDog/dd-trace-go/v2/ddtrace/ext"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/b2wdigital/restQL-golang/v6/pkg/restql"
)

const dataDogPluginName = "DataDogPlugin"

func init() {
	restql.RegisterPlugin(restql.PluginInfo{
		Name: dataDogPluginName,
		Type: restql.LifecyclePluginType,
		New: func(logger restql.Logger) (restql.Plugin, error) {
			tracer.Start()

			span := tracer.StartSpan("plugin.init")
			span.Finish()
			return &DataDogPlugin{log: logger}, nil
		},
	})
}

type DataDogPlugin struct {
	log restql.Logger
}

func (n *DataDogPlugin) Name() string {
	return dataDogPluginName
}

func txnName(method string, reqUrl *url.URL) string {
	return method + " " + reqUrl.Path
}

func (n *DataDogPlugin) BeforeTransaction(ctx context.Context, tr restql.TransactionRequest) context.Context {

	operationName := txnName(tr.Method, tr.Url)

	options := []tracer.StartSpanOption{
		tracer.Tag("name", operationName),
		tracer.Tag(ext.SpanType, ext.SpanTypeWeb),
		tracer.Tag(ext.HTTPMethod, tr.Method),
		tracer.Tag(ext.HTTPURL, tr.Url),
		tracer.Tag("_dd.measured", 1),
	}

	span := tracer.StartSpan(operationName, options...)

	if tr.Url.Host != "" {
		options = append(options, tracer.Tag("http.host", tr.Url.Host))
	}

	ctx = tracer.ContextWithSpan(ctx, span)
	fmt.Println("BeforeTransaction", tr.Url, tr.Method)

	return ctx
}

func (n *DataDogPlugin) AfterTransaction(ctx context.Context, tr restql.TransactionResponse) context.Context {

	span, ok := tracer.SpanFromContext(ctx)
	if !ok {
		return ctx
	}

	var statusStr string

	// if status is 0, treat it like 200 unless 0 was called out in DD_TRACE_HTTP_SERVER_ERROR_STATUSES
	if tr.Status == 0 {
		statusStr = "200"
	} else {
		statusStr = strconv.Itoa(tr.Status)
		span.SetTag(ext.ErrorNoStackTrace, fmt.Errorf("%s: %s", statusStr, http.StatusText(tr.Status)))
	}

	options := []tracer.FinishOption{}

	fc := &tracer.FinishConfig{}
	for _, opt := range options {
		if opt == nil {
			continue
		}
		opt(fc)
	}

	span.SetTag(ext.HTTPCode, statusStr)
	span.Finish(tracer.WithFinishConfig(fc))

	fmt.Println("AfterTransaction", tr.Status, tr.Body)

	return ctx
}

func (n *DataDogPlugin) BeforeQuery(ctx context.Context, query string, queryCtx restql.QueryContext) context.Context {
	return ctx
}
func (n *DataDogPlugin) AfterQuery(ctx context.Context, query string, result map[string]interface{}) context.Context {
	return ctx
}

func (n *DataDogPlugin) BeforeRequest(ctx context.Context, request restql.HTTPRequest) context.Context {

	operationName := request.Method + " " + request.Path

	options := []tracer.StartSpanOption{
		tracer.Tag("name", operationName),
		tracer.Tag(ext.SpanType, ext.SpanTypeWeb),
		tracer.Tag(ext.HTTPMethod, request.Method),
		tracer.Tag(ext.HTTPURL, request.Path),
		tracer.Tag("_dd.measured", 1),
	}

	span := tracer.StartSpan(operationName, options...)

	if request.Host != "" {
		options = append(options, tracer.Tag("http.host", request.Host))
	}

	ctx = tracer.ContextWithSpan(ctx, span)

	fmt.Println("BeforeRequest", request.Path, request.Host, request.Body)

	return ctx
}

func (n *DataDogPlugin) AfterRequest(ctx context.Context, request restql.HTTPRequest, response restql.HTTPResponse, errordetail error) context.Context {

	operationName := request.Method + " " + request.Path

	options := []tracer.StartSpanOption{
		tracer.Tag("name", operationName),
		tracer.Tag(ext.SpanType, ext.SpanTypeWeb),
		tracer.Tag(ext.HTTPMethod, request.Method),
		tracer.Tag(ext.HTTPURL, request.Path),
		tracer.Tag("_dd.measured", 1),
	}

	span := tracer.StartSpan(operationName, options...)

	var statusStr string

	if response.StatusCode == 0 {
		statusStr = "200"
		span.SetTag(ext.HTTPCode, statusStr)
	} else {
		if response.StatusCode == 200 {
			statusStr = "200"
		} else {
			statusStr = strconv.Itoa(response.StatusCode)
			span.SetTag(ext.ErrorNoStackTrace, fmt.Errorf("%s: %s", statusStr, http.StatusText(response.StatusCode)))
		}
	}

	span.SetTag(ext.HTTPCode, statusStr)

	span.Finish()

	span, ok := tracer.SpanFromContext(ctx)
	if !ok {
		return ctx
	}

	fmt.Println("AfterRequest", request.Path, response.StatusCode, response.Body)

	return ctx
}
