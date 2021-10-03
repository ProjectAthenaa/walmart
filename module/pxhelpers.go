package module

import (
	"fmt"
	"github.com/ProjectAthenaa/pxutils"
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
	tk.pxuuid = pxutils.UUID()

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
	//req, err := tk.NewRequest("POST", "https://collector-pxu6b0qd2s.px-cloud.net/api/v2/collector", []byte(`payload=aUkQRhAIEGJqABAeEFYQCEkQYmoLBBAIEFpGRkJBCB0dRUVFHEVTXl9TQEYcUV1fHVtCHWBTSF1AH2BbQkBbVldAHwEEAh92QFtURltcVR9mQFtZVx9wXkdXHwEfZVpXV15XVh92QFtURltcVR9zUUZbXVwdBQUACwIACgEEDVNGWlFCW1YPBQUACwIACgEEFFNGWkJVW1YPc0ZaV1xTel1fV0JTVVd2V0FZRl1CFFNGWlFVW1YPXEdeXhRTRlpIXFtWD3tGV19xU0BdR0FXXm0BVAZRBgAHBR8CVAQGHwYLBQIfUwEFCh8DClcGUwoGVlYBBVZtW0ZXX0EUU0ZaW1dbVg9cR15eFFNGWkFGW1YPcWECAAIUU0ZaVUdbVg9IfQdzV11wdgVhZncFQVRhdAtcVHZwXm1LQVdrXmVKfnR8QlgUU0ZaU1xRW1YPXEdeXhRTRlpXXFMPRkBHVxAeEGJqBAEQCBBlW1wBABAeEGJqAwsDEAgCHhBiagoHAhAIAh4QYmoKBwMQCAMAAwoeEGJqAwICChAIAQQCAh4QYmoDAgcHEAgDBAEACwoDAAMAAwMHHhBiagMCBwQQCAMEAQALCgMAAwADAAEeEGJqAwIBChAIEFBXAVBWBgMCHwADUAAfAwNXUR9QV1ELH1MDUwNTUwsKBQYEChAeEGJqAQUDEAhGQEdXT09v&appId=PXu6b0qd2S&tag=v6.7.9&uuid=be3bd410-21b2-11ec-bec9-a1a1aa987468&ft=221&seq=0&en=NTA&pc=7644296484018228&pxhd=pK-lPU8/JUkormbN5KeaCOs0RQOu1vc4V9tS7cvhFaC4fg2czrfyVMtwOBSxL3OqCaTeIeDWs-sGqcKnjUMTxQ==:wL23mjTaam6iNGhGdIlSg9X/IKhVTaKv3B753W057hvYi5dDEm/bGOE-QxA9Uv7jeDxkYYwutV8vcWPcVtVGrhmhsJ49SDN8sl6ae8YFE7M=`))
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

	log.Info(string(res.Body))

	cookie, err := pxClient.GetPXde(tk.Ctx, &perimeterx.GetCookieRequest{PXResponse: res.Body})
	if err != nil {
		log.Info(err.Error())
		tk.SetStatus(module.STATUS_ERROR)
		tk.Stop()
		return
	}

	log.Info("init pxde",  cookie.Value)
	tk.FastClient.Jar.Set(cookie.Name, cookie.Value)

	// todo: STARTS PX 3
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

	log.Info("init pxvid", p3struct.Vid)
	tk.FastClient.Jar.Set("_pxvid", p3struct.Vid)

	tk.px.Sid = p3struct.Sid
	tk.px.Vid = p3struct.Vid
	tk.px.Cts = p3struct.Cts
	tk.px.Cs = p3struct.Cs

	req, err = tk.NewRequest("POST", "https://collector-pxu6b0qd2s.px-cloud.net/api/v2/collector",  []byte(fmt.Sprintf(`payload=%s&appId=%s&tag=%s&uuid=%s&ft=%s&seq=%s&en=%s&cs=%s&pc=%s&sid=%s&vid=%s&pxhd=%s&cts=%s&rsc=%s`, p3struct.Payload, "PXu6b0qd2S", p3struct.Tag, p3struct.Uuid, p3struct.Ft, "1", p3struct.En, p3struct.Cs, p3struct.Pc, p3struct.Sid, p3struct.Vid,string(tk.FastClient.Jar.PeekValue("_pxhd")), p3struct.Cts, p3struct.Rsc)))
	//req, err = tk.NewRequest("POST", "https://collector-pxu6b0qd2s.px-cloud.net/api/v2/collector",  []byte(fmt.Sprintf(`payload=%s&appId=%s&tag=%s&uuid=%s&ft=%s&seq=%s&en=%s&cs=%s&pc=%s&sid=%s&vid=%s&pxhd=%s&cts=%s&rsc=%s`, p3struct.Payload, "PXu6b0qd2S", p3struct.Tag, tk.pxuuid, p3struct.Ft, "1", p3struct.En, "e9e03ab15dbc40ed6a7220660e2020b19a73c68abaee1e164ff5505c2f49d1a5", p3struct.Pc, "4efc1fa0-2019-11ec-955a-175002b6eef9󠄱󠄶󠄳󠄲󠄸󠄰󠄵󠄳󠄶󠄱󠄳󠄳󠄱󠄳󠄲󠄸󠄰󠄴󠄸󠄹󠄸󠄳󠄰󠄶󠄱󠄶󠄳󠄲󠄸󠄰󠄳󠄲󠄸󠄵󠄴󠄳󠄵󠄱󠄶󠄳󠄲󠄸󠄰󠄳󠄰󠄷󠄹󠄷󠄱󠄷", "4b824d10-2019-11ec-897a-5841624c5578", `/S12fRFh-gWIhRnARym7Oj2s81LOQ0vQZ45WDKRlH03dUOnxn3OD5z7bugnE3kISWSw4jh8tTm7JXLNsR9fFtw==:jyJEm5/2XJPLAp1OlYlyvgDJd0rP3/2uN-hMXd7v2vv0I555rI-RgCQ0cCnbmSG/GeyQwawCkQoGNkgXM-MW0xYJ1Q5GtEz2BR6QWyb8/6s=`, "4efc6dc0-2019-11ec-955a-175002b6eef9", p3struct.Rsc)))

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

	cookie, err = pxClient.GetCookie(tk.Ctx, &perimeterx.GetCookieRequest{PXResponse: res.Body})
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

	tk.px.RSC++
	//panic("")
}

func (tk *Task) PXEvent(){
	payload, err := pxClient.ConstructPayload(tk.Ctx, &perimeterx.Payload{
		Site:           perimeterx.SITE_WALMART,
		Type:           perimeterx.PXType_EVENT,
		Cookie:         string(tk.FastClient.Jar.PeekValue("_px3")),
		ResponseObject: nil,
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

	req, err := tk.NewRequest("POST", "https://collector-pxu6b0qd2s.px-cloud.net/api/v2/collector", []byte(fmt.Sprintf(`payload=%s&appId=%s&tag=%s&uuid=%s&ft=%s&seq=%s&en=%s&cs=%s&pc=%s&sid=%s&vid=%s&pxhd=%s&cts=%s&rsc=%s`, eventstruct.Payload, eventstruct.AppID, eventstruct.Tag, tk.pxuuid, eventstruct.Ft, "3", eventstruct.En, tk.px.Cs, eventstruct.Pc, tk.px.Sid, tk.px.Vid,  string(tk.FastClient.Jar.PeekValue("_pxhd")), tk.px.Cts, "4")))
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
	if cookie.Value != ""{
		tk.FastClient.Jar.Set("_px3", cookie.Value)
	}

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