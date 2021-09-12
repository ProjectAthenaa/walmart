package module

import (
	"fmt"
	"github.com/ProjectAthenaa/sonic-core/protos/module"
	"github.com/ProjectAthenaa/walmart/encryption"
	"github.com/json-iterator/go"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	json = jsoniter.ConfigFastest
	pieLRe = regexp.MustCompile(`PIE\.L = (\d+);`)
	pieERe = regexp.MustCompile(`PIE\.E = (\d+);`)
	pieKRe = regexp.MustCompile(`PIE\.K = "(\w+)";`)
	pieKeyIdRe = regexp.MustCompile(`PIE\.key_id = "(\d+)";`)
	phaseRe = regexp.MustCompile(`PIE\.phase = (\d+);`)
)

func (tk *Task) Homepage(){
	req, err := tk.NewRequest("GET", "https://www.walmart.com/", nil)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt create homepage request")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com")

	res, err := tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt make homepage request")
		tk.Stop()
		return
	}

	if res.StatusCode != 307{
		tk.PXHoldCaptcha(res.Headers["Location"][0])
	}
}

func (tk *Task) ATC() {
	req, err := tk.NewRequest("POST", "https://www.walmart.com/api/v3/cart/guest/:CID/items", []byte(fmt.Sprintf(`{"offerId":"%s","quantity":1,"location":{"postalCode":"%s","city":"%s","state":"%s","isZipLocated":true},"shipMethodDefaultRule":"SHIP_RULE_1","storeIds":[%s]}`, tk.offerid, tk.Data.Profile.Shipping.ShippingAddress.ZIP, tk.Data.Profile.Shipping.ShippingAddress.City, tk.Data.Profile.Shipping.ShippingAddress.ZIP, strings.Join(tk.storeids, ","))))
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt create atc request")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com")

	_, err = tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt make atc request")
		tk.Stop()
		return
	}
}

func (tk *Task) StoreLocator() {
	req, err := tk.NewRequest("PUT", "https://www.walmart.com/account/api/location", []byte(fmt.Sprintf(`{"postalCode":"%s","responseGroup":"STOREMETAPLUS","includePickUpLocation":true,"persistLocation":true,"clientName":"Web-Checkout-ShippingAddress","storeMeta":true,"plus":true}`, tk.Data.Profile.Shipping.ShippingAddress.ZIP)))
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt create store location request")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com")

	res, err := tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt make store location request")
		tk.Stop()
		return
	}

	var stores *Store
	json.Unmarshal(res.Body, &stores)
	tk.stores = *stores

}

func (tk *Task) SubmitShipping() {
	var form string
	if *tk.Data.Profile.Shipping.ShippingAddress.AddressLine2 != "" {
		form = fmt.Sprintf(`{"addressLineOne":"%s","addressLineTwo":"%s","city":"%s","firstName":"%s","lastName":"%s","phone":"%s","email":"%s","postalCode":"%s","state":"%s","addressType":"RESIDENTIAL","changedFields":[],"storeList":[%s]}`, tk.Data.Profile.Shipping.ShippingAddress.AddressLine, tk.Data.Profile.Shipping.ShippingAddress.AddressLine2, tk.Data.Profile.Shipping.ShippingAddress.City, tk.Data.Profile.Shipping.FirstName, tk.Data.Profile.Shipping.LastName, tk.Data.Profile.Shipping.PhoneNumber, tk.Data.Profile.Email, tk.Data.Profile.Shipping.ShippingAddress.ZIP, tk.Data.Profile.Shipping.ShippingAddress.StateCode, tk.FormatStores())
	} else {
		form = fmt.Sprintf(`{"addressLineOne":"%s","city":"%s","firstName":"%s","lastName":"%s","phone":"%s","email":"%s","postalCode":"%s","state":"%s","addressType":"RESIDENTIAL","changedFields":[],"storeList":[%s]}`, tk.Data.Profile.Shipping.ShippingAddress.AddressLine, tk.Data.Profile.Shipping.ShippingAddress.City, tk.Data.Profile.Shipping.FirstName, tk.Data.Profile.Shipping.LastName, tk.Data.Profile.Shipping.PhoneNumber, tk.Data.Profile.Email, tk.Data.Profile.Shipping.ShippingAddress.ZIP, tk.Data.Profile.Shipping.ShippingAddress.StateCode, tk.FormatStores())
	}
	req, err := tk.NewRequest("POST", "https://www.walmart.com/api/checkout/v3/contract/:PCID/shipping-address", []byte(form))
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt create store location request")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com")

	_, err = tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt make store location request")
		tk.Stop()
		return
	}
}

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

func (tk *Task) SubmitCard(){
	encarr, err := encryption.EncryptData(tk.cardType(), tk.Data.Profile.Billing.CVV, true, tk.PIE.L, tk.PIE.E, tk.PIE.K)
	var addrline2 string
	if tk.Data.Profile.Shipping.ShippingAddress.AddressLine2 != nil{
		addrline2 = *tk.Data.Profile.Shipping.ShippingAddress.AddressLine2
	}
	formdata, err := json.Marshal(CreditCardForm{
		EncryptedPan:   encarr[0],
		EncryptedCvv:   encarr[1],
		IntegrityCheck: encarr[2],
		KeyID:          tk.PIE.key_id,
		Phase:          strconv.Itoa(tk.PIE.phase),
		State:          tk.Data.Profile.Shipping.ShippingAddress.State,
		City:           tk.Data.Profile.Shipping.ShippingAddress.City,
		AddressType:    "RESIDENTIAL",
		PostalCode:     tk.Data.Profile.Shipping.ShippingAddress.ZIP,
		AddressLineOne: tk.Data.Profile.Shipping.ShippingAddress.AddressLine,
		AddressLineTwo: addrline2,
		FirstName:      tk.Data.Profile.Shipping.FirstName,
		LastName:       tk.Data.Profile.Shipping.LastName,
		ExpiryMonth:    tk.Data.Profile.Billing.ExpirationMonth,
		ExpiryYear:     tk.Data.Profile.Billing.ExpirationYear,
		Phone:          tk.Data.Profile.Shipping.PhoneNumber,
		CardType:       tk.cardType(),
		IsGuest:        true,
	})
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "couldnt serialize card form")
		tk.Stop()
		return
	}

	req, err := tk.NewRequest("POST", "https://www.walmart.com/api/checkout-customer/:CID/credit-card", formdata)
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "couldnt create credit card req")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders(`https://www.walmart.com/checkout/`)
	_, err = tk.Do(req)
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "couldnt make credit card req")
		tk.Stop()
		return
	}
}

func (tk *Task) Payment(){
	formdata, err := json.Marshal(Payment{
		Payments:     nil,
		CvvInSession: false,
	})
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "couldnt serialize payment form")
		tk.Stop()
		return
	}
	req, err := tk.NewRequest("POST", `https://www.walmart.com/api/checkout/v3/contract/:PCID/payment`, formdata)
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "coudldnt create payment post")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders(`https://www.walmart.com/checkout/`)
	_, err = tk.Do(req)
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "couldnt make payment post")
		tk.Stop()
		return
	}
}

func (tk *Task) OrderConfirm(){
	formdata, err := json.Marshal(OrderConfirm{
		CvvInSession:    false,
		VoltagePayments: []map[string]string{
			{"paymentType": "CREDITCARD",
				"encryptedCvv":   "", //todo
				"encryptedPan":   "", //todo
				"integrityCheck": "", //todo
				"keyId":          "",
				"phase":          "0"},
		},
	})
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "couldnt serialize order confirmation")
		tk.Stop()
		return
	}
	req, err := tk.NewRequest("PUT", `https://www.walmart.com/api/checkout/v3/contract/:PCID/order`, formdata)
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "couldnt create confirmation post")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com/checkout/")

	//{
	//	"statusCode": 400,
	//	"error": "Bad Request",
	//	"message": "Your payment couldn't be authorized. Please check the card number, CVV and expiration date and try again, or use a different payment method.",
	//	"validation": {
	//		"source": "PAYLOAD",
	//		"keys": ["PIH.ccdb.VISA.CREDITCARD.908962038.8972"]
	//	},
	//	"pangaeaErrors": ["400.CHECKOUT_SERVICE.513"],
	//	"code": "payment_service_invalid_account_no",
	//	"details": {
	//		"paymentId": "ce8c9e2b-e008-4aee-b601-91f8198dee8e",
	//		"paymentType": "CREDITCARD"
	//	}
	//}

	_, err = tk.Do(req)
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "could not read confirmation response")
		tk.Stop()
		return
	}
}
