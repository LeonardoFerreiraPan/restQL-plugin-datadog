package restqldatadog

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/b2wdigital/restQL-golang/v6/pkg/restql"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/ext"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
)


const datadogPluginName = "DatadogPlugin"

var once sync.Once

func init() {
	restql.RegisterPlugin(restql.PluginInfo{
		Name: datadogPluginName,
		Type: restql.LifecyclePluginType,
		New: func(logger restql.Logger) (restql.Plugin, error) {
			once.Do(func() {
				tracer.Start()
			})

			return &DatadogPlugin{log: logger}, nil
		},
	})
}

type DatadogPlugin struct {
	log restql.Logger
}

func (n *DatadogPlugin) Name() string {
	return datadogPluginName
}

func (n *DatadogPlugin) BeforeTransaction(ctx context.Context, tr restql.TransactionRequest) context.Context {
	resourceName := tr.Method + " " + tr.Url.Path

	span, ctx := tracer.StartSpanFromContext(ctx, "restql.transaction",
		tracer.ResourceName(resourceName),
		tracer.SpanType(ext.SpanTypeWeb),
		tracer.Tag(ext.HTTPMethod, tr.Method),
		tracer.Tag(ext.HTTPURL, tr.Url.String()),
	)

	span.SetTag("request.path", tr.Url.String())

	return ctx
}

func (n *DatadogPlugin) AfterTransaction(ctx context.Context, tr restql.TransactionResponse) context.Context {
	if span, ok := tracer.SpanFromContext(ctx); ok {
		span.SetTag(ext.HTTPCode, strconv.Itoa(tr.Status))
		if tr.Status >= 500 {
			span.SetTag(ext.Error, fmt.Errorf("http status %d", tr.Status))
		}
		span.Finish()
	}
	return ctx
}

func (n *DatadogPlugin) BeforeRequest(ctx context.Context, request restql.HTTPRequest) context.Context {
	operationName := "http.request"
	resourceName := request.Method + " " + request.Host + request.Path

	span, ctx := tracer.StartSpanFromContext(ctx, operationName,
		tracer.ResourceName(resourceName),
		tracer.SpanType(ext.SpanTypeHTTP),
		tracer.Tag(ext.HTTPMethod, request.Method),
		tracer.Tag(ext.HTTPURL, request.Path),
		tracer.Tag("http.host", request.Host),
	)

	span.SetTag("request.path", request.Path)

	return ctx
}

func (n *DatadogPlugin) AfterRequest(ctx context.Context, request restql.HTTPRequest, response restql.HTTPResponse, errordetail error) context.Context {
	if span, ok := tracer.SpanFromContext(ctx); ok {
		span.SetTag(ext.HTTPCode, strconv.Itoa(response.StatusCode))

		if errordetail != nil {
			span.SetTag(ext.Error, errordetail)
		} else if response.StatusCode >= 400 {
			span.SetTag(ext.Error, fmt.Errorf("response status %d", response.StatusCode))
		}

		span.Finish()
	}
	return ctx
}

func (n *DatadogPlugin) BeforeQuery(ctx context.Context, query string, queryCtx restql.QueryContext) context.Context {
	return ctx
}

func (n *DatadogPlugin) AfterQuery(ctx context.Context, query string, result map[string]interface{}) context.Context {
	return ctx
}