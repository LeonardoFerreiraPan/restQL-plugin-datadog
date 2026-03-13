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

	return ctx
}

func (n *DataDogPlugin) BeforeQuery(ctx context.Context, query string, queryCtx restql.QueryContext) context.Context {
	return ctx
}
func (n *DataDogPlugin) AfterQuery(ctx context.Context, query string, result map[string]interface{}) context.Context {
	return ctx
}

func (n *DataDogPlugin) BeforeRequest(ctx context.Context, query string, queryCtx restql.QueryContext) context.Context {
	return ctx
}
func (n *DataDogPlugin) AfterRequest(ctx context.Context, query string, result map[string]interface{}) context.Context {
	return ctx
}
