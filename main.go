package restqldatadog

import (
	"context"
	"fmt"
	"strconv"

	"github.com/b2wdigital/restQL-golang/v6/pkg/restql"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const datadogPluginName = "DatadogPlugin"

func init() {
	restql.RegisterPlugin(restql.PluginInfo{
		Name: datadogPluginName,
		Type: restql.LifecyclePluginType,
		New: func(logger restql.Logger) (restql.Plugin, error) {
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

// BeforeTransaction inicia o Trace principal (Root Span)
func (n *DatadogPlugin) BeforeTransaction(ctx context.Context, tr restql.TransactionRequest) context.Context {
	resourceName := tr.Method + " " + tr.Url.Path

	// Inicia o span e já o injeta no contexto retornado
	span, ctx := tracer.StartSpanFromContext(ctx, "restql.transaction",
		tracer.ResourceName(resourceName),
		tracer.Tag(ext.SpanType, ext.SpanTypeWeb),
		tracer.Tag(ext.HTTPMethod, tr.Method),
		tracer.Tag(ext.HTTPURL, tr.Url.String()),
	)

	span.SetTag("request.path", tr.Url.String())

	return ctx
}

// AfterTransaction finaliza o Trace principal
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

// BeforeRequest inicia um Sub-span para cada chamada externa
func (n *DatadogPlugin) BeforeRequest(ctx context.Context, request restql.HTTPRequest) context.Context {
	operationName := "http.request"
	resourceName := request.Method + " " + request.Host + request.Path

	// Cria um span filho do contexto atual
	span, ctx := tracer.StartSpanFromContext(ctx, operationName,
		tracer.ResourceName(resourceName),
		tracer.Tag(ext.SpanType, ext.SpanTypeHTTP),
		tracer.Tag(ext.HTTPMethod, request.Method),
		tracer.Tag(ext.HTTPURL, request.Path),
		tracer.Tag("http.host", request.Host),
	)

	span.SetTag("request.path", request.Path)

	return ctx
}

// AfterRequest finaliza o sub-span da chamada externa
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

// Métodos obrigatórios da interface que não precisam de lógica específica
func (n *DatadogPlugin) BeforeQuery(ctx context.Context, query string, queryCtx restql.QueryContext) context.Context {
	return ctx
}
func (n *DatadogPlugin) AfterQuery(ctx context.Context, query string, result map[string]interface{}) context.Context {
	return ctx
}
