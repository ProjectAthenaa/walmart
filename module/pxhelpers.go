package module

import (
	"fmt"
	"github.com/ProjectAthenaa/sonic-core/protos/module"
	"github.com/ProjectAthenaa/sonic-core/sonic/antibots/perimeterx"
	"github.com/google/uuid"
	"github.com/prometheus/common/log"
)

type PayloadOut struct{
	Payload string `json:"payload"`
	AppID	string `json:"appId"`
	Tag		string `json:"tag"`
	Uuid	string `json:"uuid"`
	Ft		string `json:"ft"`
	Seq		string `json:"seq"`
	En		string `json:"en"`
	Pc		string `json:"pc"`
	Sid		string `json:"sid,omitempty"`
	Vid		string `json:"vid,omitempty"`
	Cts		string `json:"cts,omitempty"`
	Rsc		string `json:"rsc"`
	Cs		string `json:"cs"`
	Ci		string `json:"ci"`
}

func (tk *Task) PXInit(){
	tk.pxuuid = uuid.NewString()

	payload, err := pxClient.ConstructPayload(tk.Ctx, &perimeterx.Payload{
		Site:           perimeterx.SITE_WALMART,
		Type:           perimeterx.PXType_PX2,
		RSC:            0,
		Uuid: 			tk.pxuuid,
	})
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "px error")
		tk.Stop()
		return
	}
	log.Info(payload.Payload)

 	var p2struct *PayloadOut
	json.Unmarshal(payload.Payload, &p2struct)

	req, err := tk.NewRequest("POST", "https://collector-pxu6b0qd2s.px-cloud.net/api/v2/collector", []byte(fmt.Sprintf(`payload=%s&appId=%s&tag=%s&uuid=%s&ft=%s&seq=%s&en=%s&pc=%s&pxhd=%s&rsc=%s`, p2struct.Payload, "PXu6b0qd2S", p2struct.Tag, tk.pxuuid, p2struct.Ft, "0", p2struct.En, p2struct.Pc, string(tk.FastClient.Jar.PeekValue("_pxhd")), "1")))
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "px error")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com/")

	log.Info("making px2 req")
	res, err := tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "px error")
		tk.Stop()
		return
	}
	log.Info("made px2 req")

	payload, err = pxClient.ConstructPayload(tk.Ctx, &perimeterx.Payload{
		Site:           perimeterx.SITE_WALMART,
		Type:           perimeterx.PXType_PX34,
		ResponseObject: res.Body,
		RSC:            1,
		Uuid: 			tk.pxuuid,
	})
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "px error")
		tk.Stop()
		return
	}
	var p3struct *PayloadOut
	json.Unmarshal(payload.Payload, &p3struct)

	req, err = tk.NewRequest("POST", "https://collector-pxu6b0qd2s.px-cloud.net/api/v2/collector",  []byte(fmt.Sprintf(`payload=%s&appId=%s&tag=%s&uuid=%s&ft=%s&seq=%s&en=%s&cs=%s&pc=%s&sid=%s&pxhd=%s&cts=%s&rsc=%s`, p3struct.Payload, "PXu6b0qd2S", p3struct.Tag, p3struct.Uuid, p3struct.Ft, "1", p3struct.En, p3struct.Cs, p3struct.Pc, p3struct.Sid, string(tk.FastClient.Jar.PeekValue("_pxhd")), p3struct.Cts, p3struct.Rsc)))
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "px error")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com/")

	res, err = tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "px error")
		tk.Stop()
		return
	}

	cookie, err := pxClient.GetCookie(tk.Ctx, &perimeterx.GetCookieRequest{PXResponse: res.Body})
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR)
		tk.Stop()
		return
	}

	log.Info("init px",  cookie.Value)
	tk.FastClient.Jar.Set(cookie.Name, cookie.Value)

	cookie, err = pxClient.GetPXde(tk.Ctx, &perimeterx.GetCookieRequest{PXResponse: res.Body})
	if err != nil {
		log.Info(err.Error())
		tk.SetStatus(module.STATUS_ERROR)
		tk.Stop()
		return
	}

	log.Info("init pxde",  cookie.Value)
	tk.FastClient.Jar.Set(cookie.Name, cookie.Value)
	tk.px.Response = res.Body

	tk.px.RSC++
}

func (tk *Task) PXEvent(){
	payload, err := pxClient.ConstructPayload(tk.Ctx, &perimeterx.Payload{
		Site:           perimeterx.SITE_WALMART,
		Type:           perimeterx.PXType_EVENT,
		Cookie:         string(tk.FastClient.Jar.PeekValue("_px3")),
		ResponseObject: tk.px.Response,
		Token:          "",
		RSC:            tk.px.RSC,
		Uuid: tk.pxuuid,
	})
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, err.Error())
		tk.Stop()
		return
	}
	var eventstruct *PayloadOut
	json.Unmarshal(payload.Payload, &eventstruct)

	//add event struct

	req, err := tk.NewRequest("POST", "https://collector-pxu6b0qd2s.px-cloud.net/api/v2/collector", []byte(fmt.Sprintf(`payload=%s&appId=%s&tag=%s&uuid=%s&ft=%s&seq=%s&en=%s&cs=%s&pc=%s&sid=%s&vid=%s&pxhd=%s&cts=%s&rsc=%s`, eventstruct.Payload, eventstruct.AppID, eventstruct.Tag, tk.pxuuid, eventstruct.Ft, "3", eventstruct.En, eventstruct.Cs, eventstruct.Pc, eventstruct.Sid, eventstruct.Vid,  string(tk.FastClient.Jar.PeekValue("_pxhd")), eventstruct.Cts, "4")))
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "could not get create px event post")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com")

	res, err := tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "could not post px event")
		tk.Stop()
		return
	}
	cookie, err := pxClient.GetCookie(tk.Ctx, &perimeterx.GetCookieRequest{PXResponse: res.Body})
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR)
		tk.Stop()
		return
	}

	log.Info("event px",  cookie.Value)
	tk.FastClient.Jar.Set("_px3", cookie.Value)
	tk.px.Response = res.Body

	tk.px.RSC++
}

func (tk *Task) PXHoldCaptcha(blockedUrl string){
	tk.pxuuid = uuid.NewString()

	payload, err := pxClient.ConstructPayload(tk.Ctx, &perimeterx.Payload{
		Site:           perimeterx.SITE_WALMART,
		Type:           perimeterx.PXType_PX2,
		RSC:            0,
		Uuid: 			tk.pxuuid,
	})
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "px error")
		tk.Stop()
		return
	}
	var p2struct *PayloadOut
	json.Unmarshal(payload.Payload, &p2struct)

	req, err := tk.NewRequest("POST", "https://collector-pxu6b0qd2s.px-cloud.net/api/v2/bundle", []byte(fmt.Sprintf(`payload=%s&appId=%s&tag=%s&uuid=%s&ft=%s&seq=%s&en=%s&pc=%s&pxhd=%s&rsc=%s`, p2struct.Payload, "PXu6b0qd2S", p2struct.Tag, tk.pxuuid, p2struct.Ft, "0", p2struct.En, p2struct.Pc, string(tk.FastClient.Jar.PeekValue("_pxhd")), "1")))
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "px error")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders(blockedUrl)

	res, err := tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "px error")
		tk.Stop()
		return
	}

	payload, err = pxClient.ConstructPayload(tk.Ctx, &perimeterx.Payload{
		Site:           perimeterx.SITE_WALMART,
		Type:           perimeterx.PXType_PX34,
		ResponseObject: res.Body,
		RSC:            1,
		Uuid: 			tk.pxuuid,
	})
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "px error")
		tk.Stop()
		return
	}
	var p3struct *PayloadOut
	json.Unmarshal(payload.Payload, &p3struct)

	req, err = tk.NewRequest("POST", "https://collector-pxu6b0qd2s.px-cloud.net/api/v2/bundle", []byte(fmt.Sprintf(`payload=%s&appId=%s&tag=%s&uuid=%s&ft=%s&seq=%s&en=%s&cs=%s&pc=%s&sid=%s&pxhd=%s&cts=%s&rsc=%s`, p3struct.Payload, "PXu6b0qd2S", p3struct.Tag, p3struct.Uuid, p3struct.Ft, "1", p3struct.En, p3struct.Cs, p3struct.Pc, p3struct.Sid, string(tk.FastClient.Jar.PeekValue("_pxhd")), p3struct.Cts, p3struct.Rsc)))
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "px error")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders(blockedUrl)

	res, err = tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "cant make first bundle post")
		tk.Stop()
		return
	}

	payload, err = pxClient.ConstructPayload(tk.Ctx, &perimeterx.Payload{
		Site:           perimeterx.SITE_WALMART,
		Type:           perimeterx.PXType_HCAPHIGH,
		ResponseObject: res.Body,
		RSC:            2,
		Uuid: tk.pxuuid,
	})
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "px error")
		tk.Stop()
		return
	}
	var hcapstruct *PayloadOut
	json.Unmarshal(payload.Payload, &hcapstruct)

	req, err = tk.NewRequest("POST", "https://collector-pxu6b0qd2s.px-cloud.net/api/v2/bundle", []byte(fmt.Sprintf(`payload=%s&appId=%s&tag=%s&uuid=%s&ft=%s&seq=%s&en=%s&cs=%s&pc=%s&sid=%s󠄶󠄳󠄱󠄹󠄴󠄵󠄳󠄶󠄷󠄶󠄷󠄳&vid=%s&ci=%s&pxhd=%s&cts=%s&rsc=%s`, hcapstruct.Payload, "PXu6b0qd2S", hcapstruct.Tag, hcapstruct.Uuid, hcapstruct.Ft, "5", hcapstruct.En, hcapstruct.Cs, hcapstruct.Pc, hcapstruct.Sid, hcapstruct.Vid, hcapstruct.Ci, string(tk.FastClient.Jar.PeekValue("_pxhd")), "4")))
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "cant create bundle second post")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders(blockedUrl)

	res, err = tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "cant do second bundle post")
		tk.Stop()
		return
	}
	cookie, err := pxClient.GetCookie(tk.Ctx, &perimeterx.GetCookieRequest{PXResponse: res.Body})
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "cant read cookie after solving captcha")
		tk.Stop()
		return
	}
	tk.FastClient.Jar.Set("_px3", cookie.Value)
}