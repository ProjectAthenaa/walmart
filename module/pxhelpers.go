package module

import (
	"github.com/ProjectAthenaa/sonic-core/protos/module"
	"github.com/ProjectAthenaa/sonic-core/sonic/antibots/perimeterx"
)

func (tk *Task) PXInit(){
	payload, err := pxClient.ConstructPayload(tk.Ctx, &perimeterx.Payload{
		Site:           perimeterx.SITE_WALMART,
		Type:           perimeterx.PXType_PX34,
		Cookie:         "",
		ResponseObject: nil,
		Token:          "",
		RSC:            tk.px.RSC,
	})

	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, err.Error())
		tk.Stop()
		return
	}

	req, err := tk.NewRequest("POST", "https://collector-pxu6b0qd2s.px-cloud.net/api/v2/collector", payload.Payload)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "could not get create px init post")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com")

	res, err := tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "could not post px init")
		tk.Stop()
		return
	}

	cookie, err := pxClient.GetCookie(tk.Ctx, &perimeterx.GetCookieRequest{PXResponse: res.Body})
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR)
		tk.Stop()
		return
	}
	tk.FastClient.Jar.Set("_px3", cookie.Value)
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
	})

	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, err.Error())
		tk.Stop()
		return
	}

	req, err := tk.NewRequest("POST", "https://collector-pxu6b0qd2s.px-cloud.net/api/v2/collector", payload.Payload)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "could not get create px init post")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com")

	res, err := tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "could not post px init")
		tk.Stop()
		return
	}

	cookie, err := pxClient.GetCookie(tk.Ctx, &perimeterx.GetCookieRequest{PXResponse: res.Body})
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR)
		tk.Stop()
		return
	}
	tk.FastClient.Jar.Set("_px3", cookie.Value)
	tk.px.Response = res.Body

	tk.px.RSC++
}

func (tk *Task) PXHoldCaptcha(blockedUrl string){
	payload, err := pxClient.ConstructPayload(tk.Ctx, &perimeterx.Payload{
		Site:           perimeterx.SITE_WALMART,
		Type:           perimeterx.PXType_PX34,
		Cookie:         "",
		ResponseObject: nil,
		Token:          "",
		RSC:            0,
	})

	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "cant get px init payload")
		tk.Stop()
		return
	}

	req, err := tk.NewRequest("POST", "https://collector-pxu6b0qd2s.px-cloud.net/api/v2/bundle", payload.Payload)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "cant create first bundle post")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders(blockedUrl)

	res, err := tk.Do(req)
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
	})

	req, err = tk.NewRequest("POST", "https://collector-pxu6b0qd2s.px-cloud.net/api/v2/bundle", payload.Payload)
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