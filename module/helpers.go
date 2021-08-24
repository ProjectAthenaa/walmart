package module

import (
	"fmt"
	module "github.com/ProjectAthenaa/sonic-core/protos"
	"io/ioutil"
	"regexp"
	"strings"
)

var (
	offerIdRe = regexp.MustCompile(`"offerId":"(\w+)"`)
	storeIdRe = regexp.MustCompile(`"storeId":"(\d+)"`)
)

func (tk *Task) OfferId(){
	res, err := tk.Client.Get(tk.url)
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "could not fetch homepage")
		tk.Stop()
		return
	}
	defer res.Body.Close()

	mainbody, err := ioutil.ReadAll(res.Body)
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "error reading homepage")
		tk.Stop()
		return
	}

	tk.offerid = offerIdRe.FindStringSubmatch(string(mainbody))[1]
	for _, submatch := range storeIdRe.FindAllStringSubmatch(string(mainbody), -1){
		tk.storeids = append(tk.storeids, submatch[1])
	}
}
//{"id":90262,"address":{"postalCode":"10023","address1":"221 W 72nd St","city":"New York","state":"NY","country":"US"},"storeType":{"id":8,"name":"FedEx Office","displayName":"New York FedEx Pickup Location"},"customerInfo":{"distance":1.48,"isPreferred":false,"isWithinRange":true}},
//{"id":3520,"address":{"postalCode":"07094","address1":"400 Park Pl","city":"Secaucus","state":"NJ","country":"US"},"storeType":{"id":1,"name":"Walmart Supercenter","displayName":"Secaucus Supercenter"},"customerInfo":{"distance":1.79,"isPreferred":false,"isWithinRange":true}},{"id":3795,"address":{"postalCode":"07047","address1":"2100 88th St","city":"North Bergen","state":"NJ","country":"US"},"storeType":{"id":1,"name":"Walmart Supercenter","displayName":"North Bergen Supercenter"},"customerInfo":{"distance":1.95,"isPreferred":false,"isWithinRange":true}},{"id":3159,"address":{"postalCode":"07608","address1":"1 Teterboro Landing Dr","city":"Teterboro","state":"NJ","country":"US"},"storeType":{"id":1,"name":"Walmart Supercenter","displayName":"Teterboro Supercenter"},"customerInfo":{"distance":5.98,"isPreferred":false,"isWithinRange":true}},{"id":5447,"address":{"postalCode":"07032","address1":"150 Harrison Ave","city":"Kearny","state":"NJ","country":"US"},"storeType":{"id":1,"name":"Walmart Supercenter","displayName":"Kearny Supercenter"},"customerInfo":{"distance":7.2,"isPreferred":false,"isWithinRange":true}},{"id":5752,"address":{"postalCode":"07026","address1":"174 Passaic St","city":"Garfield","state":"NJ","country":"US"},"storeType":{"id":1,"name":"Walmart Supercenter","displayName":"Garfield Supercenter"},"customerInfo":{"distance":7.34,"isPreferred":false,"isWithinRange":true}},{"id":3562,"address":{"postalCode":"07663","address1":"189 Us Highway 46","city":"Saddle Brook","state":"NJ","country":"US"},"storeType":{"id":2,"name":"Walmart","displayName":"Saddle Brook Store"},"customerInfo":{"distance":8.56,"isPreferred":false,"isWithinRange":true}},{"id":5867,"address":{"postalCode":"07002","address1":"500 Bayonne Crossing Way","city":"Bayonne","state":"NJ","country":"US"},"storeType":{"id":1,"name":"Walmart Supercenter","displayName":"Bayonne Supercenter"},"customerInfo":{"distance":10.09,"isPreferred":false,"isWithinRange":true}},{"id":3292,"address":{"postalCode":"07083","address1":"900 Springfield Rd","city":"Union","state":"NJ","country":"US"},"storeType":{"id":2,"name":"Walmart","displayName":"Union Store"},"customerInfo":{"distance":16.67,"isPreferred":false,"isWithinRange":true}},{"id":5293,"address":{"postalCode":"11581","address1":"77 Green Acres Rd","city":"Valley Stream","state":"NY","country":"US"},"storeType":{"id":1,"name":"Walmart Supercenter","displayName":"Valley Stream Supercenter"},"customerInfo":{"distance":17.14,"isPreferred":false,"isWithinRange":true}}

func (tk *Task) FormatStores()string{
	var outstores []string
	for _, store := range tk.stores.Stores{
		var outstring strings.Builder
		outstring.WriteString(fmt.Sprintf(`{"id":%s,"address":{"postalCode":"%s","address1":"%s","city":"%s","state":"%s","country":"%s"},"storeType":{"id":%s,"name":"%s","displayName":"%s"},"customerInfo":{"distance":%s,"isPreferred":false,"isWithinRange":true}}`, store.ID, store.Address.PostalCode, store.Address.Address1, store.Address.City, store.Address.State, store.Address.Country, store.StoreType.ID,store.StoreType.Name, store.StoreType.DisplayName, store.Distance))
		outstores = append(outstores, outstring.String())
	}
	return strings.Join(outstores, ",")
}