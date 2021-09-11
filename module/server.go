package module

import (
	"context"
	"github.com/ProjectAthenaa/sonic-core/protos/module"
	"github.com/ProjectAthenaa/sonic-core/sonic"
)

var pxClient, _ = sonic.NewPerimeterXClient("localhost:3000")

type Server struct {
	module.UnimplementedModuleServer
}

func (s Server) Task(_ context.Context, data *module.Data) (*module.StartResponse, error) {
	task := NewTask(data)
	if err := task.Start(data); err != nil {
		return nil, err
	}

	return &module.StartResponse{Started: true}, nil
}
