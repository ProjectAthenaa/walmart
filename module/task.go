package module

import (
	module "github.com/ProjectAthenaa/sonic-core/protos"
	"github.com/ProjectAthenaa/sonic-core/sonic/base"
	"github.com/ProjectAthenaa/sonic-core/sonic/face"
	"github.com/ProjectAthenaa/sonic-core/sonic/frame"
	"github.com/prometheus/common/log"
	http "github.com/ProjectAthenaa/fasttls"
	"github.com/useflyent/fhttp/cookiejar"
)

var _ face.ICallback = (*Task)(nil)

type Task struct {
	*base.BTask
	Monitor              *frame.PubSub
	url                  string
	offerid				 string
	storeids			 []string
	stores			     Store
}

func (tk *Task) OnInit() {
	roundtripper, err := flyent.NewHelloIDRoundtripper(tls.HelloChrome_91, tk.FormatProxy())
	if err != nil {
		log.Error("create roundtripper:", err)
		tk.SetStatus(module.STATUS_ERROR, "error creating roundtripper")
		tk.Stop()
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Error("create cookiejar:", err)
		tk.SetStatus(module.STATUS_ERROR, "error creating cookie jar")
		tk.Stop()
	}
	tk.Client = http.NewClient(tls.HelloChrome_91, "")

	pubsub, err := frame.SubscribeToChannel(tk.Data.MonitorChannel)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, err.Error())
		tk.Stop()
		return
	}

	tk.Monitor = pubsub
}
func (tk *Task) OnPreStart() error {
	return nil
}
func (tk *Task) OnStarting() {
	tk.SetStatus(module.STATUS_STARTING, "starting task")
	sizeinfo := <-tk.Monitor.Channel
	tk.pid = sizeinfo.Payload
}
func (tk *Task) OnPause() error {
	return nil
}
func (tk *Task) OnStopping() {

}
