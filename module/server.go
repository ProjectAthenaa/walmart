package module

import (
	"context"
	"github.com/ProjectAthenaa/sonic-core/protos/module"
)

type Server struct {
	module.UnimplementedModuleServer
}

func (s Server) Task(_ context.Context, data *module.Data) (*module.StartResponse, error) {
	//v, _ := json.Marshal(data)
	//fmt.Println(string(v))
	task := NewTask(data)
	if err := task.Start(data); err != nil {
		return nil, err
	}

	return &module.StartResponse{Started: true}, nil
}
