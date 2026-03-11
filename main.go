package main

import (
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/b2wdigital/restQL-golang/v6/pkg/restql"
)

const (
	dataDogPluginName = "DataDogPlugin"
)

func init() {
	restql.RegisterPlugin(restql.PluginInfo{
		Name: dataDogPluginName,
		Type: restql.LifecyclePluginType,
		New: func(logger restql.Logger) (restql.Plugin, error) {

			err := tracer.Start()

			if err != nil {
				logger.Error("failed to initialize data dog", err)
				return nil, err
			}

			defer tracer.Stop()

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
