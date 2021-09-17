package module

import (
	"github.com/ProjectAthenaa/sonic-core/protos/module"
	"github.com/ProjectAthenaa/sonic-core/sonic/base"
	"github.com/ProjectAthenaa/sonic-core/sonic/face"
	"github.com/ProjectAthenaa/sonic-core/sonic/frame"
	"github.com/ProjectAthenaa/walmart/encryption"
	"github.com/prometheus/common/log"
)

var _ face.ICallback = (*Task)(nil)

type Task struct {
	*base.BTask
	url      string
	offerid  string
	storeids []string
	cartid	string
	customerid string
	lineitemid string
	price string
	accesspoint string
	newaddress string
	contractid string
	tenderid string
	preferenceid string
	orderid string
	stores   Store

	encryptedPan string
	encryptedCVV string
	PIE          encryption.PIEStruct

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
	log.Info("onstarting")
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
		log.Info(err)
		tk.SetStatus(module.STATUS_ERROR, "monitor err")
		tk.Stop()
		return
	}

	tk.SetStatus(module.STATUS_MONITORING)
	monitorData := <-pubsub.Chan(tk.Ctx)
	tk.offerid = monitorData["offerid"].(string)
	pubsub.Close()

	tk.SetStatus(module.STATUS_PRODUCT_FOUND)

	funcarr := []func(){
		tk.Homepage,
		tk.CreateAcc,
		tk.GetCartIds,
		tk.ATC,
		tk.CreateDelivery,
		tk.SetFulfillment,
		tk.CreateContract,
		tk.CreateCreditCart,
		tk.UpdateTenderPlan,
		tk.PlaceOrder,
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
