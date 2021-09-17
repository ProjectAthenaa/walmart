package module

import (
	"fmt"
	creditcard "github.com/ProjectAthenaa/go-credit-card"
	http "github.com/ProjectAthenaa/sonic-core/fasttls"
	"github.com/prometheus/common/log"
	"math/rand"
	"strings"
	"time"
)

var (
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func RandStringRunes(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (tk *Task) FormatPhone()string{
	return fmt.Sprintf(`(%s)+%s-%s`, tk.Data.Profile.Shipping.PhoneNumber[:3], tk.Data.Profile.Shipping.PhoneNumber[3:6], tk.Data.Profile.Shipping.PhoneNumber[6:])
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

func (tk *Task) AddGQLHeaders(req *http.Request, queryString string){
	req.Headers[`x-o-gql-query`] = []string{queryString}
	req.Headers[`x-apollo-operation-name`] = []string{strings.Split(queryString," ")[0]}
	req.Headers[`wm_qos.correlation_id`] = []string{}
	req.Headers[`x-o-correlation-id`] = []string{}

	req.Headers[`x-o-platform`] = []string{`rweb`}
	req.Headers[`x-latency-trace`] = []string{`1`}
	req.Headers[`x-o-platform-version`] = []string{`main-95-7de933`}
	req.Headers[`x-o-segment`] = []string{`oaoh`}
	req.Headers[`x-enable-server-timing`] = []string{`1`}
	req.Headers[`x-o-ccm`] = []string{`server`}
	req.Headers[`x-o-tp-phase`] = []string{`tp5`}
	//x-o-gql-query	mutation saveTenderPlanToPC
	//x-apollo-operation-name	saveTenderPlanToPC
	//wm_qos.correlation_id	2n6PoqoebFuCVajzWcwjrrkY82KU-Ep2VDxZ
	//x-o-correlation-id	TR7vFweqniBiYYSQthyvF5IZ-__R235qM9fN

	//x-o-platform	rweb
	//x-latency-trace	1
	//x-o-platform-version	main-95-7de933
	//x-o-segment	oaoh
	//x-enable-server-timing	1
	//x-o-ccm	server
	//x-o-tp-phase	tp5
}