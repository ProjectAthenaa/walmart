package module

import (
	"fmt"
	module "github.com/ProjectAthenaa/sonic-core/protos"
	http "github.com/ProjectAthenaa/fasttls"
	"io/ioutil"
	"strings"
	"github.com/json-iterator/go"
)

var(
	json = jsoniter.ConfigFastest
)

func (tk *Task) ATC(){
	req, err := http.NewRequest("POST", "https://www.walmart.com/api/v3/cart/guest/:CID/items", strings.NewReader(fmt.Sprintf(`{"offerId":"%s","quantity":1,"location":{"postalCode":"%s","city":"%s","state":"%s","isZipLocated":true},"shipMethodDefaultRule":"SHIP_RULE_1","storeIds":[%s]}`, tk.offerid, tk.Data.Profile.Shipping.ShippingAddress.ZIP, tk.Data.Profile.Shipping.ShippingAddress.City, tk.Data.Profile.Shipping.ShippingAddress.ZIP, strings.Join(tk.storeids, ","))))
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "couldnt create atc request")
		tk.Stop()
		return
	}
	_, err := tk.Client.Do(req)
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "couldnt make atc request")
		tk.Stop()
		return
	}
}

func (tk *Task) StoreLocator(){
	req, err := http.NewRequest("PUT", "https://www.walmart.com/account/api/location", strings.NewReader(fmt.Sprintf(`{"postalCode":"%s","responseGroup":"STOREMETAPLUS","includePickUpLocation":true,"persistLocation":true,"clientName":"Web-Checkout-ShippingAddress","storeMeta":true,"plus":true}`, tk.Data.Profile.Shipping.ShippingAddress.ZIP)))
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "couldnt create store location request")
		tk.Stop()
		return
	}
	res, err := tk.Client.Do(req)
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "couldnt make store location request")
		tk.Stop()
		return
	}
	defer res.Body.Close()

	storesbody, err := ioutil.ReadAll(res.Body)
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "couldnt read store location request")
		tk.Stop()
		return
	}

	var stores *Store
	json.Unmarshal(storesbody, &stores)
	tk.stores = *stores

}

func (tk *Task) SubmitShipping(){
	var form string
	if *tk.Data.Profile.Shipping.ShippingAddress.AddressLine2 != ""{
		form = fmt.Sprintf(`{"addressLineOne":"%s","addressLineTwo":"%s","city":"%s","firstName":"%s","lastName":"%s","phone":"%s","email":"%s","postalCode":"%s","state":"%s","addressType":"RESIDENTIAL","changedFields":[],"storeList":[%s]}`, tk.Data.Profile.Shipping.ShippingAddress.AddressLine, tk.Data.Profile.Shipping.ShippingAddress.AddressLine2, tk.Data.Profile.Shipping.ShippingAddress.City, tk.Data.Profile.Shipping.FirstName, tk.Data.Profile.Shipping.LastName, tk.Data.Profile.Shipping.PhoneNumber, tk.Data.Profile.Email, tk.Data.Profile.Shipping.ShippingAddress.ZIP, tk.Data.Profile.Shipping.ShippingAddress.StateCode ,tk.FormatStores())
	}else
	{
		form = fmt.Sprintf(`{"addressLineOne":"%s","city":"%s","firstName":"%s","lastName":"%s","phone":"%s","email":"%s","postalCode":"%s","state":"%s","addressType":"RESIDENTIAL","changedFields":[],"storeList":[%s]}`, tk.Data.Profile.Shipping.ShippingAddress.AddressLine, tk.Data.Profile.Shipping.ShippingAddress.City, tk.Data.Profile.Shipping.FirstName, tk.Data.Profile.Shipping.LastName, tk.Data.Profile.Shipping.PhoneNumber, tk.Data.Profile.Email, tk.Data.Profile.Shipping.ShippingAddress.ZIP, tk.Data.Profile.Shipping.ShippingAddress.StateCode ,tk.FormatStores())
	}
	req, err := http.NewRequest("POST", "https://www.walmart.com/api/checkout/v3/contract/:PCID/shipping-address", strings.NewReader(form))
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "couldnt create store location request")
		tk.Stop()
		return
	}
	res, err := tk.Client.Do(req)
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "couldnt make store location request")
		tk.Stop()
		return
	}
	defer res.Body.Close()

	shippingbody, err := ioutil.ReadAll(res.Body)
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "couldnt read store location request")
		tk.Stop()
		return
	}
}

//:CID/credit-card
//:PCID/payment
//:PCID/order