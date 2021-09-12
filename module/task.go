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
	offerid  string
	storeids []string
	stores   Store

	encryptedPan string
	encryptedCVV string
	PIE          PIEStruct

	px struct {
		Response []byte
		RSC      int32
	}
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
	return
}

func (tk *Task) Flow() {
	pubsub, err := frame.SubscribeToChannel(tk.Data.Channels.MonitorChannel)
	if err != nil {
		tk.Stop()
		return
	}

	tk.SetStatus(module.STATUS_MONITORING)
	monitorData := <-pubsub.Chan(tk.Ctx)
	tk.offerid = monitorData["offerid"].(string)
	pubsub.Close()

	tk.SetStatus(module.STATUS_PRODUCT_FOUND)

	funcarr := []func(){
		tk.ATC,
		tk.StoreLocator,
		tk.SubmitShipping,
		tk.GetPIEVals,
		tk.SubmitCard,
		tk.Payment,
		tk.OrderConfirm,
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
