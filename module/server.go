package module

import (
	module "github.com/ProjectAthenaa/sonic-core/protos"
	"github.com/ProjectAthenaa/sonic-core/sonic/face"
	"github.com/prometheus/common/log"
)

type Server struct {
	module.UnimplementedModuleServer
}

func (s Server) Task(server module.Module_TaskServer) error {
	dt, err := server.Recv() //recv first data
	if err != nil {
		log.Error("first recv:", err)
		return err
	}

	//first data must be a init command
	if dt.Command != module.COMMAND_START {
		return face.ErrTaskNotInit
	}
	//must set profile or some data
	if dt.Data == nil {
		return face.ErrFirstTaskDataError
	}
	task := Task{}

	task.Init(server)

	//1.init and start
	err = task.Start(dt.Data)
	if err != nil {
		return err
	}

	//2.listen commands
	return task.Listen()
}
