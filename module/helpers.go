package module

import (
	"fmt"
	creditcard "github.com/ProjectAthenaa/go-credit-card"
	http "github.com/ProjectAthenaa/sonic-core/fasttls"
	"github.com/ProjectAthenaa/sonic-core/protos/module"
	"github.com/prometheus/common/log"
	"regexp"
	"strings"
)

var (
	offerIdRe = regexp.MustCompile(`"offerId":"(\w+)"`)
	storeIdRe = regexp.MustCompile(`"storeId":"(\d+)"`)
)

func (tk *Task) OfferId() {
	req, err := tk.NewRequest("GET", tk.url, nil)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "could not fetch homepage")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders("https://walmart.com/")

	res, err := tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt read walmart page")
		tk.Stop()
		return
	}

	tk.offerid = string(offerIdRe.FindSubmatch(res.Body)[1])
	for _, submatch := range storeIdRe.FindAllSubmatch(res.Body, -1) {
		tk.storeids = append(tk.storeids, string(submatch[1]))
	}
}

func (tk *Task) FormatStores() string {
	var outstores []string
	for _, store := range tk.stores.Stores {
		var outstring strings.Builder
		outstring.WriteString(fmt.Sprintf(`{"id":%s,"address":{"postalCode":"%s","address1":"%s","city":"%s","state":"%s","country":"%s"},"storeType":{"id":%s,"name":"%s","displayName":"%s"},"customerInfo":{"distance":%s,"isPreferred":false,"isWithinRange":true}}`, store.ID, store.Address.PostalCode, store.Address.Address1, store.Address.City, store.Address.State, store.Address.Country, store.StoreType.ID, store.StoreType.Name, store.StoreType.DisplayName, store.Distance))
		outstores = append(outstores, outstring.String())
	}
	return strings.Join(outstores, ",")
}

func (tk *Task) GenerateDefaultHeaders(referrer string) http.Headers {
	return http.Headers{
		`user-agent`:       {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36"},
		`accept`:           {`application/json`},
		`accept-encoding`:  {`gzip, deflate, br`},
		`accept-language`:  {`en-us`},
		`content-type`:     {`application/x-www-form-urlencoded; charset=UTF-8`},
		`sec-ch-ua`:        {`"Chromium";v="91", " Not A;Brand";v="99", "Google Chrome";v="91"`},
		`sec-ch-ua-mobile`: {`?0`},
		`Sec-Fetch-Site`:   {`same-site`},
		`Sec-Fetch-Dest`:   {`empty`},
		`Sec-Fetch-Mode`:   {`cors`},
		`referer`:          {referrer},
		`X-Requested-With`: {`XMLHttpRequest`},
		`origin`:           {`https://www.newbalance.com`},
		`Pragma`:           {`no-cache`},
		`Cache-Control`:    {`no-cache`},
		`Connection`:       {`keep-alive`},
	}
}

func (tk *Task) cardType()string{
	card := creditcard.Card{Number: tk.Data.Profile.Billing.Number, Cvv: tk.Data.Profile.Billing.CVV, Month: tk.Data.Profile.Billing.ExpirationMonth, Year: "20" + tk.Data.Profile.Billing.ExpirationYear}
	card.Method()
	return card.Company.Short
}

func (tk *Task) SendPX(payload []byte) []byte {
	req, err := tk.NewRequest("POST", "https://collector-pxu6b0qd2s.px-cloud.net/api/v2/collector", payload)
	if err != nil {
		log.Info("couldnt create post px")
		return nil
	}
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com")

	res, err := tk.Do(req)
	if err != nil {
		log.Info("couldnt post px")
		return nil
	}

	return res.Body
}