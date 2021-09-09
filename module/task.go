package module

import (
	"github.com/ProjectAthenaa/sonic-core/protos/module"
	"github.com/ProjectAthenaa/sonic-core/sonic/base"
	"github.com/ProjectAthenaa/sonic-core/sonic/face"
	"github.com/ProjectAthenaa/sonic-core/sonic/frame"
)

var _ face.ICallback = (*Task)(nil)

type Task struct {
	*base.BTask
	url      string
	pid      string
	offerid  string
	storeids []string
	stores   Store
}

func NewTask(data *module.Data) *Task {
	task := &Task{BTask: &base.BTask{Data: data}}
	task.Callback = task
	task.Init()
	return task
}

func (tk *Task) OnInit() {
	return
}
func (tk *Task) OnPreStart() error {
	return nil
}
func (tk *Task) OnStarting() {
	tk.FastClient.CreateCookieJar()
	tk.Flow()
}
func (tk *Task) OnPause() error {
	return nil
}
func (tk *Task) OnStopping() {
	tk.FastClient.Destroy()
	//panic("")
	return
}

func (tk *Task) Flow() {
	pubsub, err := frame.SubscribeToChannel(tk.Data.Channels.MonitorChannel)
	if err != nil{
		tk.Stop()
		return
	}

	tk.SetStatus(module.STATUS_MONITORING)
	monitorData := <- pubsub.Chan(tk.Ctx)
	tk.pid = monitorData["pid"].(string)
	pubsub.Close()

	tk.SetStatus(module.STATUS_PRODUCT_FOUND)

	funcarr := []func(){
	}

	for _, f := range funcarr {
		select {
		case <-tk.Ctx.Done():
			return
		default:
			f()
		}
	}
}
