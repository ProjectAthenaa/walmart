package module

import (
	"fmt"
	"github.com/ProjectAthenaa/sonic-core/protos/module"
	"regexp"
	"strconv"
	"time"
)

var(
	pieLRe = regexp.MustCompile(`PIE\.L = (\d+);`)
	pieERe = regexp.MustCompile(`PIE\.E = (\d+);`)
	pieKRe = regexp.MustCompile(`PIE\.K = "(\w+)";`)
	pieKeyIdRe = regexp.MustCompile(`PIE\.key_id = "(\d+)";`)
	phaseRe = regexp.MustCompile(`PIE\.phase = (\d+);`)
)

type PIEStruct struct{
	L int
	E int
	K string
	key_id string
	phase int
}

func (tk *Task) GetPIEVals(){
	req, err := tk.NewRequest("GET", fmt.Sprintf(`https://securedataweb.walmart.com/pie/v1/wmcom_us_vtg_pie/getkey.js?bust=%d`, time.Now().Unix()), nil)
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "couldnt create encryption request")
		tk.Stop()
		return
	}
	res, err := tk.Do(req)
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "couldnt get encryption variables")
		tk.Stop()
		return
	}

	lVal, err := strconv.Atoi(string(pieLRe.FindSubmatch(res.Body)[1]))
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "could not read lVal")
		tk.Stop()
		return
	}
	eVal, err := strconv.Atoi(string(pieERe.FindSubmatch(res.Body)[1]))
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "could not read eVal")
		tk.Stop()
		return
	}
	phaseVal, err := strconv.Atoi(string(phaseRe.FindSubmatch(res.Body)[1]))
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "could not read phaseVal")
		tk.Stop()
		return
	}

	tk.PIE = PIEStruct{
		L:      lVal,
		E:      eVal,
		K:      string(pieKRe.FindSubmatch(res.Body)[1]),
		key_id: string(pieKeyIdRe.FindSubmatch(res.Body)[1]),
		phase: phaseVal,
	}
}
//card, cvv, l, k, e
func (tk *Task) EncryptPANandCVV() []string{
	value, err := tk.ottoVM.Call(`ProtectPANandCVV`, nil, tk.Data.Profile.Billing.Number, tk.Data.Profile.Billing.CVV,tk.PIE.L, tk.PIE.K, tk.PIE.E)
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "could not encrypt data")
		tk.Stop()
		return nil
	}
	outarr, err := value.Export()
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "could not export encrypted array")
		tk.Stop()
		return nil
	}
	return outarr.([]string)
}