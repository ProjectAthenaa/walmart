package module

import (
	"github.com/ProjectAthenaa/sonic-core/protos/module"
	"github.com/ProjectAthenaa/sonic-core/sonic/base"
	"github.com/ProjectAthenaa/sonic-core/sonic/face"
	"github.com/ProjectAthenaa/walmart/encryption"
	"sync"
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

	pxuuid string

	px struct {
		Response []byte
		RSC      int32
	}

	accountlock *sync.Mutex
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
	tk.accountlock = &sync.Mutex{}
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
	funcarr := []func(){
		tk.Preload,
		tk.MonitorProd,
		tk.accountlock.Lock,
		tk.ATC,
		tk.CreateDelivery,
		tk.SetFulfillment,
		tk.PXInit,
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
