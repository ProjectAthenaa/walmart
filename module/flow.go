package module

import (
	"fmt"
	"github.com/ProjectAthenaa/sonic-core/fasttls"
	"github.com/ProjectAthenaa/sonic-core/protos/module"
	"github.com/ProjectAthenaa/sonic-core/sonic/frame"
	"github.com/ProjectAthenaa/walmart/config"
	"github.com/ProjectAthenaa/walmart/encryption"
	"github.com/json-iterator/go"
	"github.com/prometheus/common/log"
	"regexp"
	"strconv"
	"time"
)

var (
	leadZeroRe = regexp.MustCompile(`^0`)

	json = jsoniter.ConfigFastest
	pieLRe = regexp.MustCompile(`PIE\.L = (\d+);`)
	pieERe = regexp.MustCompile(`PIE\.E = (\d+);`)
	pieKRe = regexp.MustCompile(`PIE\.K = "(\w+)";`)
	pieKeyIdRe = regexp.MustCompile(`PIE\.Key_id = "(\d+)";`)
	PhaseRe = regexp.MustCompile(`PIE\.Phase = (\d+);`)

	//offerIdRe = regexp.MustCompile(`"offerId":"(\w+)"`)
	storeIdRe = regexp.MustCompile(`"storeId":"(\d+)"`)
	lineItemIdRe = regexp.MustCompile(`"id":"(\w+)","imageInfo"`)
	cartIdRe = regexp.MustCompile(`"mergeAndGetCart":\{"id":"([\w-]+)"`)
	customerIdRe = regexp.MustCompile(`"customer":\{"id":"([\w-]+)"`)

	priceRe = regexp.MustCompile(`"USD","price":([\d.]+),`)

	accessPointRe = regexp.MustCompile(`"accessPoint":\{"id":"([\w-]+)"`)
	newAddressRe = regexp.MustCompile(`"newAddress":\{"id":"([\w-]+)"`)
	preferenceIdRe = regexp.MustCompile(`"id":"([\w-]+)"`)

	contractIdRe = regexp.MustCompile(`"createPurchaseContract":\{"id":"([\w-]+)"`)
	tenderPlanRe = regexp.MustCompile(`"tenderPlanId":"([\w-]+)"`)
)

func (tk *Task) Preload(){
	go func(){
		tk.accountlock.Lock()
		defer tk.accountlock.Unlock()
		for _, sf := range []func(){
			tk.Homepage,
			tk.PXInit,
			tk.PXEvent,
			tk.CreateAcc,
			tk.GetCartIds,
		}{
			sf()
		}
	}()
}

func (tk *Task) MonitorProd(){
	pubsub, err := frame.SubscribeToChannel(tk.Data.Channels.MonitorChannel)
	if err != nil {
		log.Info(err)
		tk.SetStatus(module.STATUS_ERROR, "monitor err")
		tk.Stop()
		return
	}
	defer pubsub.Close()

	tk.SetStatus(module.STATUS_MONITORING)
	for monitorData := range pubsub.Chan(tk.Ctx){
		log.Info(monitorData)
		if monitorData == nil{
			continue
		}
		tk.offerid = monitorData["offerid"].(string)
		break
	}


	tk.SetStatus(module.STATUS_PRODUCT_FOUND)
}

func (tk *Task) Homepage(){
	req, err := tk.NewRequest("GET", fmt.Sprintf("https://www.walmart.com/ip/~/%s", tk.Data.Metadata[*config.Module.Fields[0].FieldKey]), nil)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt create homepage request")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com")

	res, err := tk.Do(req)
	if err != nil {
		log.Info(err)
		tk.SetStatus(module.STATUS_ERROR, "couldn't make homepage request")
		tk.Stop()
		return
	}

	if res.StatusCode == 307{
		tk.PXHoldCaptcha(res.Headers["Location"][0])
		tk.HomepageRedirect(res.Headers["Location"][0])
		return
	}
	if res.StatusCode == 301 || res.StatusCode == 302{
		tk.HomepageRedirect(res.Headers["Location"][0])
		return
	}

	//"/blocked?url=Lw==&uuid=3ec36ab1-17a8-11ec-a2d4-78524271664b&vid=&g=b"
	tk.price = string(priceRe.FindSubmatch(res.Body)[1])
	tk.lineitemid = string(lineItemIdRe.FindSubmatch(res.Body)[1])
	for _, submatch := range storeIdRe.FindAllSubmatch(res.Body, -1) {
		tk.storeids = append(tk.storeids, string(submatch[1]))
	}
}

func (tk *Task) HomepageRedirect(link string){
	req, err := tk.NewRequest("GET", link, nil)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt create homepage request")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com")

	res, err := tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldn't make homepage redirect request")
		tk.Stop()
		return
	}

	tk.price = string(priceRe.FindSubmatch(res.Body)[1])
	tk.lineitemid = string(lineItemIdRe.FindSubmatch(res.Body)[1])
	for _, submatch := range storeIdRe.FindAllSubmatch(res.Body, -1) {
		tk.storeids = append(tk.storeids, string(submatch[1]))
	}
}

func (tk *Task) CreateAcc(){
	tk.SetStatus(module.STATUS_GENERATING_COOKIES)
	//eb := strings.Split(tk.Data.Profile.Email, "@")
	//email := eb[0] + "." + RandStringRunes(10) + "@" + eb[1]

	req, err := tk.NewRequest("POST", `https://identity.walmart.com/account/electrode/api/signup?vid=oaoh&mode=frame&svp=true&sgc=true&hhf=true&dps=true`, []byte(`{"personName":{"firstName":"Omar","lastName":"Hu"},"email":"jninini.niunini@gmail.com","password":"0o0p0o0P.","rememberme":false,"emailNotificationAccepted":true,"captcha":{"sensorData":"2a25G2m84Vrp0o9c4235091.12-1,8,-36,-890,Mozilla/9.8 (Windows NT 52.1; Win95; x80) AppleWebKit/791.17 (KHTML, like Gecko) Chrome/35.1.3547.92 Safari/227.39,uaend,30216,90340069,en-US,Gecko,2,2,4,8,191212,6599772,4788,817,2764,067,0517,873,3102,,cpen:3,i0:2,dm:1,cwen:1,non:5,opc:0,fc:0,sc:4,wrc:1,isc:9,vib:9,bat:1,x10:2,x30:8,8989,0.66491677221,125890950371,loc:-1,8,-36,-891,do_en,dm_en,t_en-4,2,-10,-912,0,2,1,0,784,337,0;1,1,9,3,201,573,0;1,-7,6,2,-8,-8,3;1,-3,4,8,-0,-1,1;9,4,4,9,8198,247,6;6,-9,7,1,1485,447,2;1,1,7,4,783,162,1;0,7,3,1,7324,690,0;0,3,9,3,197,210,0;1,-7,6,2,-8,-8,3;1,-3,4,8,-0,-1,1;9,4,4,9,372,881,9;3,4,8,8,1857,879,6;2,-8,0,0,-1,-3,4;9,-0,7,3,-0,-7,2;1,1,7,4,1391,355,7;0,-4,0,7,7741,511,0;1,1,2,5,752,590,1;0,2,4,9,8205,193,6;6,4,1,0,8245,482,4;9,7,0,1,2563,425,1;0,7,3,1,7670,690,0;0,3,9,3,105,540,0;1,1,2,5,9240,553,0;6,8,2,2,0633,2141,6;2,-8,0,0,3471,162,1;0,7,3,0,8280,690,0;0,-1,6,6,4973,590,1;0,4,4,9,9187,193,6;7,-9,7,1,2581,409,2;2,9,7,4,1452,201,7;-8,5,-80,-520,7,2,0,2,482,948,7;1,0,1,0,047,690,0;0,-1,7,6,-9,-0,7;4,-0,2,4,-2,-1,0;1,1,3,5,9360,630,0;6,-5,9,8,1878,817,6;2,3,0,8,007,146,2;2,9,8,4,1022,201,7;0,2,1,0,933,337,0;0,-1,6,6,-9,-0,7;4,-0,2,4,-2,-1,0;1,1,2,5,499,820,1;0,2,4,9,8051,193,6;6,-9,7,0,-4,-0,2;5,-2,9,7,-2,-7,6;2,3,9,8,4538,191,8;7,-8,3,1,7587,638,0;0,3,9,3,363,210,0;2,9,2,5,9477,586,0;6,8,2,2,0284,629,2;5,8,7,1,1887,409,2;2,9,7,4,1378,201,7;0,2,1,0,941,667,0;0,3,9,3,5352,757,3;0,8,6,3,2353,1465,6;6,-9,7,0,2795,146,2;2,9,7,3,2988,201,7;0,-4,0,6,8584,210,0;2,1,2,5,0359,586,0;7,-5,8,8,2974,879,6;3,1,9,8,4699,047,8;-0,9,-04,-362,8,8,3081,undefined,9,7,247;7,8,5107,undefined,4,8,627;2,2,2918,undefined,0,6,579;4,1,0323,undefined,0,0,036;6,5,1462,undefined,2,1,834;8,2,9558,undefined,9,2,355;3,1,5634,-4,4,8,210;7,4,4940,-0,7,0,551;7,3,9514,-2,0,1,482;3,1,2230,-3,9,2,948;89,9,8499,-8,2,1,427;42,1,7008,-9,0,0,629;38,3,6157,-5,0,6,162;20,3,5778,-4,4,8,210;17,1,1708,-3,9,7,830;77,7,3613,-2,1,9,784;97,1,2581,-8,6,2,699;87,2,4240,-0,7,0,551;04,9,7431,-9,3,0,146;37,9,5348,-1,2,4,337;27,4,5674,-4,1,9,223;20,5,9993,-2,0,1,482;64,3,4791,-2,6,6,701;13,1,6779,-6,8,7,590;34,9,1101,-1,7,3,516;89,0,2732,-3,9,2,948;95,9,8967,-8,2,1,427;58,0,8110,-9,0,0,629;44,5,7268,-5,0,6,162;36,2,6975,-4,4,8,210;33,1,3235,-3,9,7,830;93,7,5140,-2,1,9,784;13,0,4025,-8,6,2,699;03,4,6610,-0,7,0,551;20,7,9925,-9,3,0,146;53,0,7811,-1,2,4,337;33,5,7022,-4,1,9,223;36,4,1359,-2,0,1,482;70,2,6156,-2,6,6,701;29,3,8172,-6,8,7,590;50,8,3512,-1,7,3,516;05,9,4170,-3,9,2,948;11,0,0372,-8,2,1,427;74,1,9824,-9,0,0,629;60,3,8015,-5,0,6,162;52,3,7635,-4,4,8,210;49,2,3596,-3,9,7,830;09,5,5505,-2,1,9,784;29,2,4467,-8,6,2,699;19,3,6092,-0,7,0,551;46,8,9340,-9,3,0,146;79,8,7422,15,2,2,337;59,4,7710,-4,1,7,223;52,5,1039,-2,0,9,482;96,3,6913,26,6,6,701;45,2,8974,-6,8,7,590;66,7,4039,-1,7,3,516;11,1,5518,-3,9,2,948;27,9,1834,-8,2,1,427;80,0,0472,-9,0,0,629;86,5,9520,-5,0,6,162;78,2,8247,-4,4,8,210;65,1,4153,-3,9,7,830;25,7,6068,-2,1,9,784;45,0,6023,-8,6,2,699;35,4,8618,-0,7,0,551;52,8,1908,-9,3,0,146;85,9,9917,-1,2,4,337;65,4,9141,-4,1,9,223;68,5,3460,-2,0,1,482;12,3,8337,-2,6,6,701;61,1,0673,-6,8,7,590;82,9,5005,-1,7,3,516;37,0,6691,-3,9,2,948;43,8,2019,-8,2,1,427;06,2,1552,-9,0,0,629;92,3,27909,-3,9,2,948;46,0,41098,-6,8,7,590;88,8,74233,-2,0,1,482;11,2,00782,-1,2,4,337;87,6,19605,-0,7,0,551;77,8,38292,-2,1,9,784;63,1,80513,-4,4,8,210;86,1,72927,0,9,7,830;46,6,91295,9,1,9,047;66,0,81917,-4,4,8,573;89,3,73249,-9,0,0,982;03,3,28550,-3,9,2,201;57,0,42649,-6,8,7,853;99,8,75922,-2,0,1,745;32,2,01511,-1,2,4,690;98,6,10434,-0,7,0,814;88,8,39962,-2,1,9,047;74,1,81223,-4,4,8,573;97,1,74677,-9,0,0,982;11,5,29932,-3,9,2,201;65,9,43100,-6,8,7,853;07,7,76825,06,0,9,745;30,3,02559,15,2,4,690;-8,5,-80,-538,7,1,444,0672,764;0,8,731,7218,385;2,1,528,3694,642;0,4,423,7466,254;4,2,307,5225,453;8,1,086,3857,517;7,0,621,9102,231;7,7,016,2717,550;7,3,809,8915,827;5,7,673,0082,619;36,3,564,8035,464;53,2,409,206,011;30,8,582,088,642;94,0,212,971,026;21,1,558,839,896;84,8,845,598,656;06,1,659,128,276;14,4,543,591,535;88,2,426,379,152;12,1,197,127,367;51,0,732,758,427;31,7,127,005,158;21,3,910,622,484;19,7,770,825,761;86,5,414,993,559;368,1,8719,008,140,-9;895,4,2880,330,693,-1;205,4,5077,695,834,-3;579,0,2581,712,701,006;857,4,9265,533,139,036;319,0,0909,036,785,118;090,6,4962,140,009,551;057,8,2209,425,305,162;265,9,7239,986,126,590;389,5,56676,658,641,-7;462,3,84207,204,723,-1;213,1,30316,378,573,-7;825,3,04692,674,184,2747;361,2,32645,470,750,-7;973,5,06040,776,361,-8;302,3,02514,597,799,-1;-3,6,-01,-810,-0,4,-12,-018,3,28,-7,-9,-0;-1,3,-56,-380,9,99,-1,-3,-3,-8,-8,-2,-7,-5,-2;-1,2,-93,-756,-8,2,-25,-729,-9,9,-64,-108,-5,0,-84,-425,NaN,489440,0,39,81,4,NaN,13383,0257294706187,2627030818818,96,16073,17,023,2219,72,4,97043,971016,7,f9k9iemotyji4sahhyzs_2719,2274,154,8882780729,94442383-1,3,-56,-387,-1,0-2,1,-58,-87,490527621;54,22,13,48,18,81,10,30,45,43,1;,9,9;true;true;true;930;true;86;66;true;false;-7-3,3,-91,-90,2376-9,9,-64,-102,2055285-8,5,-80,-536,938823-1,8,-36,-811,;9;2;7"}}`))
	//req, err := tk.NewRequest("POST", `https://identity.walmart.com/account/electrode/api/signup?vid=oaoh&mode=frame&svp=true&sgc=true&hhf=true&dps=true`, []byte(fmt.Sprintf(`{"personName":{"firstName":"%s","lastName":"%s"},"email":"%s","password":"%s.","rememberme":false,"emailNotificationAccepted":false,"captcha":{"sensorData":"2a25G2m84Vrp0o9c4235981.12-1,8,-36,-890,Mozilla/9.8 (Windows NT 52.1; Win95; x80) AppleWebKit/791.17 (KHTML, like Gecko) Chrome/35.1.3547.92 Safari/227.39,uaend,30216,90340069,en-US,Gecko,2,2,4,8,191213,110197,1167,589,3000,204,051,827,0189,,cpen:7,i2:9,dm:2,cwen:0,non:3,opc:7,fc:3,sc:2,wrc:8,isc:1,vib:5,bat:1,x42:9,x76:9,5183,3.10219692827,518029688884,loc:-4,2,-10,-918,do_en,dm_en,t_en-8,5,-80,-523,7,2,0,2,482,948,7;1,0,1,0,047,690,0;0,-1,6,6,-9,-0,7;4,-0,2,4,-2,-1,0;1,1,2,5,9360,630,0;6,-5,8,8,1878,817,6;2,3,9,8,007,146,2;2,9,7,4,1022,201,7;0,2,1,0,933,337,0;0,-1,6,6,-9,-0,7;4,-0,2,4,-2,-1,0;1,1,2,5,499,820,1;0,2,4,9,8051,193,6;6,-9,7,0,-4,-0,2;5,-2,9,7,-2,-7,6;2,3,9,8,4538,191,8;7,-8,3,1,7587,638,0;0,3,9,3,363,210,0;2,9,2,5,9477,586,0;6,8,2,2,0284,629,2;5,8,7,1,1887,409,2;2,9,7,4,1378,201,7;0,2,1,0,941,667,0;0,3,9,3,5352,757,3;0,8,6,3,2353,1465,6;6,-9,7,0,2795,146,2;2,9,7,3,2988,201,7;0,-4,0,6,8584,210,0;2,1,2,5,0359,586,0;7,-5,8,8,2974,879,6;3,1,9,8,4699,047,8;-0,9,-04,-366,8,9,0,1,629,784,8;8,0,0,2,745,201,7;0,-4,0,6,-5,-2,9;8,-2,9,2,-3,-8,0;0,3,0,3,5472,834,3;0,-3,5,9,8072,131,6;6,4,2,0,490,516,6;3,1,0,8,4269,047,8;7,2,0,2,631,948,7;0,-4,0,6,-5,-2,9;8,-2,9,2,-3,-8,0;0,3,9,3,000,540,0;2,9,2,5,9223,586,0;6,-5,8,7,-8,-2,9;3,-3,1,9,-1,-1,6;6,4,1,0,8460,899,4;8,-0,7,4,1285,249,7;0,2,1,0,109,337,0;1,1,9,3,5589,780,3;0,8,6,3,2904,551,9;3,4,8,8,1270,879,6;3,1,9,8,4515,047,8;7,2,0,2,649,278,7;0,2,1,0,3713,929,7;3,2,6,7,3470,1858,0;6,-5,8,7,2188,516,6;3,1,9,7,5125,047,8;7,-8,3,0,8320,337,0;1,3,9,3,6461,780,3;1,-3,4,9,9178,193,6;7,2,1,0,8521,745,4;-2,1,-97,-060,4,9,0253,-3,9,2,355;8,3,3666,-4,4,8,627;2,2,2058,-0,7,0,968;2,5,7614,-2,0,1,899;8,9,0367,-3,9,2,355;2,3,3771,-4,4,8,627;6,3,2003,-0,7,0,968;6,3,7792,-2,0,1,899;2,1,0419,-3,9,2,355;6,2,3868,-4,4,8,627;13,2,9746,-3,9,7,247;73,6,1790,-2,1,9,191;93,0,0741,8,2,4,744;10,4,3647,-4,1,9,524;13,5,7967,-2,0,1,783;57,3,2899,0,9,2,249;85,8,6944,-8,2,1,728;48,2,5488,-9,0,0,920;34,4,4616,-5,0,6,463;26,2,3273,-4,4,8,511;23,1,9136,7,7,0,852;17,8,6893,9,0,1,482;64,2,4729,-2,6,6,701;13,3,6745,-6,8,7,590;34,7,1292,-1,7,3,516;89,1,2771,-3,9,2,948;95,9,8989,-8,2,1,427;58,0,7509,-9,0,0,629;44,5,6657,-5,0,6,162;36,2,6334,-4,4,8,210;33,2,2224,-3,9,7,830;93,5,4290,-2,1,9,784;13,2,3152,-8,6,2,699;03,3,5842,-0,7,0,551;20,7,8035,-9,3,0,146;53,0,6921,-1,2,4,337;33,5,6294,-4,1,9,223;36,3,0534,-2,0,1,482;70,4,5323,-2,6,6,701;29,1,7428,-6,8,7,590;50,9,2850,-1,7,3,516;05,0,3391,-3,9,2,948;11,9,9644,-8,2,1,427;74,0,8269,-9,0,0,629;60,5,7317,-5,0,6,162;52,2,6047,-4,4,8,210;49,1,2922,-3,9,7,830;09,7,4837,-2,1,9,784;29,1,3802,-8,6,2,699;19,2,5411,-0,7,0,551;46,9,8602,-9,3,0,146;79,9,7715,-1,2,4,337;59,4,7486,-4,1,9,223;52,5,1705,-2,0,1,482;96,3,6690,-2,6,6,701;45,1,8681,-6,8,7,590;66,9,3013,-1,7,3,516;11,0,4681,-3,9,2,948;27,8,0955,-8,2,1,427;80,2,9498,-9,0,0,629;86,4,8629,-5,0,6,162;78,1,8559,4,6,2,699;32,3,7945,0,1,9,223;62,3,2692,7,7,3,516;28,0,5406,1,0,6,162;72,1,8993,79,4,2,210;69,1,4995,46,9,1,830;29,5,6862,5,3,4,146;86,9,8507,16,2,4,337;66,5,8703,-4,1,9,223;79,4,2075,7,7,3,516;35,9,5862,-3,9,2,948;41,0,1064,-8,2,1,427;04,0,1677,-9,0,0,629;90,5,0725,-5,0,6,162;82,2,9377,-4,4,8,210;79,2,5307,-3,9,7,830;39,5,7377,-2,1,9,784;59,2,6239,-8,6,2,699;49,2,8886,-0,7,0,551;76,9,1077,-9,3,0,146;09,8,9973,-1,2,4,337;89,6,9132,-4,1,9,223;82,4,3599,-2,0,1,482;26,2,8399,-2,6,6,701;75,3,0315,-6,8,7,590;96,8,5719,-1,7,3,516;41,0,6295,-3,9,2,948;57,9,2537,-8,2,1,427;10,0,1150,-9,0,0,629;16,5,0208,-5,0,6,162;08,1,9884,-4,4,8,210;95,3,5738,-3,9,7,830;55,6,7790,-2,1,9,784;75,1,6602,-8,6,2,699;65,2,8374,34,7,8,551;82,7,1503,-9,3,8,146;15,0,9499,-1,2,2,337;95,5,9754,-4,1,7,223;98,4,3076,06,0,1,482;528,8,9697,-1,2,4,337;171,4,8510,-0,7,0,551;064,5,91710,-5,0,6,162;272,0,41016,-6,8,7,590;203,4,52897,-9,3,0,146;386,1,80428,-4,4,8,210;137,0,36567,-1,7,3,516;749,4,00842,-1,2,4,337;178,2,06479,-3,9,7,830;725,5,27112,-3,9,2,948;800,2,20133,-4,1,9,223;107,7,38350,-2,1,9,784;929,3,13380,-8,2,1,427;423,8,74518,-2,0,1,482;532,9,17310,-8,6,2,699;818,1,72009,-9,0,0,629;370,1,89437,-2,6,6,701;014,4,19172,-0,7,0,551;070,7,91246,-5,0,6,162;288,9,41545,-6,8,7,590;229,4,52464,-9,3,0,146;302,0,80983,-4,4,8,210;153,2,36091,-1,7,3,516;765,2,00406,-1,2,4,337;194,4,06967,-3,9,7,830;741,4,27653,-3,9,2,948;816,2,20787,-4,1,9,223;113,7,38996,6,3,0,146;309,1,80249,5,6,2,952;822,1,73848,-9,0,0,982;394,1,80276,-2,6,6,064;038,4,10891,-0,7,0,814;094,7,92965,-5,0,6,425;202,9,42394,-6,8,7,853;233,3,53198,-9,3,0,409;316,2,81691,-4,4,8,573;167,1,37843,-1,7,3,879;779,3,01171,-1,2,4,690;108,2,07725,-3,9,7,193;755,5,28468,-3,9,2,201;830,2,21500,-4,1,9,586;137,7,30285,86,1,7,047;959,2,15490,32,2,1,780;453,7,76855,8,7,3,879;786,3,02488,9,6,6,-9;835,1,23074,3,8,7,-8;456,8,77498,8,7,3,1919;-3,3,-91,-219,2,5,352,339,453;3,5,369,334,447;4,5,377,331,443;5,5,375,328,441;6,5,383,327,440;7,5,307,325,438;8,5,304,324,438;9,5,411,323,438;0,5,428,322,438;1,5,465,320,438;36,3,735,041,325;75,9,371,376,202;54,2,585,648,961;31,8,665,432,501;95,0,404,312,994;22,1,739,282,774;85,8,025,938,543;07,1,821,568,174;15,4,726,838,443;89,2,604,607,070;23,1,362,466,289;52,0,918,084,357;32,7,304,338,094;22,3,198,950,326;10,7,966,154,606;52,7,9790,872,697,-5;72,3,8756,031,029,-0;62,3,0343,350,508,-2;087,7,0680,820,157,579;297,1,5739,381,978,907;228,4,6812,024,471,553;486,2,84750,720,732,-1;237,3,30954,894,582,-7;849,3,04239,190,193,-8;278,3,00793,911,521,2043;877,4,04549,93,876,-2;133,7,13568,12,302,-4;223,8,32011,64,711,-7;419,1,84254,56,030,8008;592,9,70306,058,058,-3;601,1,13235,013,166,-0;987,2,78848,220,441,-0;-7,4,-63,-148,-7,8,-75,-181,-1,8,-36,-899,-4,2,-10,-921,-8,5,-80,-521,0,1326;-0,4,-12,-019,-2,1,-58,-284,8345535,869602,0,0,1,9,3184251,13230,0257268293856,2627004305622,80,16072,328,182,5008,36,2,23118,4508216,3,a169dg2vbcayfxjhbnmu_1230,8473,156,-8651073161,11951396-1,8,-36,-896,-4,0-7,8,-75,-77,730179301;29,12,47,03,86,08,29,43,41,60,1;,4,8;true;true;true;943;true;66;31;true;false;-9-8,2,-25,-42,0393-0,9,-04,-370,431579-4,2,-10,-925,263940-7,8,-75,-191,;10;1;7"}}`, tk.Data.Profile.Shipping.FirstName, tk.Data.Profile.Shipping.LastName, email, RandStringRunes(20) + "aA.")))
	//first last email pass
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt create account request")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com")


	res, err := tk.Do(req)
	if err != nil || res.StatusCode != 200{
		tk.SetStatus(module.STATUS_ERROR, "couldn't make account request")
		tk.Stop()
		return
	}
}
func (tk *Task) GetCartIds(){
	tk.SetStatus(module.STATUS_GENERATING_COOKIES)
	req, err := tk.NewRequest("POST", "https://www.walmart.com/orchestra/cartxo/graphql", []byte(`{"query":"mutation MergeAndGetCart( $input:MergeAndGetCartInput! $detailed:Boolean! $includePartialFulfillmentSwitching:Boolean! = false ){mergeAndGetCart(input:$input){id checkoutable customer{id isGuest}addressMode lineItems{id quantity quantityString quantityLabel createdDateTime displayAddOnServices selectedAddOnServices{offerId quantity groupType error{code upstreamErrorCode errorMsg}}isPreOrder @include(if:$detailed) bundleComponents @include(if:$detailed){offerId quantity}selectedVariants @include(if:$detailed){name value}registryId registryInfo{registryId registryType}fulfillmentPreference priceInfo{priceDisplayCodes{showItemPrice priceDisplayCondition finalCostByWeight}itemPrice{...merge_lineItemPriceInfoFragment}wasPrice{...merge_lineItemPriceInfoFragment}unitPrice{...merge_lineItemPriceInfoFragment}linePrice{...merge_lineItemPriceInfoFragment}}product{itemType offerId isAlcohol name @include(if:$detailed) sellerType usItemId addOnServices{serviceType serviceTitle serviceSubTitle groups{groupType groupTitle assetUrl shortDescription services{displayName selectedDisplayName offerId currentPrice{priceString price}serviceMetaData}}}imageInfo @include(if:$detailed){thumbnailUrl}sellerId @include(if:$detailed) sellerName @include(if:$detailed) hasSellerBadge @include(if:$detailed) orderLimit @include(if:$detailed) orderMinLimit @include(if:$detailed) weightUnit @include(if:$detailed) weightIncrement @include(if:$detailed) salesUnit salesUnitType fulfillmentType @include(if:$detailed) fulfillmentSpeed @include(if:$detailed) fulfillmentTitle @include(if:$detailed) classType @include(if:$detailed) rhPath @include(if:$detailed) availabilityStatus @include(if:$detailed) brand @include(if:$detailed) category @include(if:$detailed){categoryPath}departmentName @include(if:$detailed) configuration @include(if:$detailed) snapEligible @include(if:$detailed) preOrder @include(if:$detailed){isPreOrder}}wirelessPlan @include(if:$detailed){planId mobileNumber postPaidPlan{...merge_postpaidPlanDetailsFragment}}fulfillmentSourcingDetails @include(if:$detailed){currentSelection requestedSelection fulfillmentBadge}}fulfillment{intent accessPoint{...merge_accessPointFragment}reservation{...merge_reservationFragment}storeId displayStoreSnackBarMessage homepageBookslotDetails{title subTitle expiryText expiryTime slotExpiryText}deliveryAddress{addressLineOne addressLineTwo city state postalCode firstName lastName id}fulfillmentItemGroups @include(if:$detailed){...on FCGroup{__typename defaultMode collapsedItemIds startDate endDate checkoutable checkoutableErrors{code shouldDisableCheckout itemIds upstreamErrors{offerId upstreamErrorCode}}priceDetails{subTotal{...merge_priceTotalFields}}fulfillmentSwitchInfo{fulfillmentType benefit{type price itemCount date isWalmartPlusProgram @include(if:$detailed)}partialItemIds @include(if:$includePartialFulfillmentSwitching)}shippingOptions{__typename itemIds availableShippingOptions{__typename id shippingMethod deliveryDate price{__typename displayValue value}label{prefix suffix}isSelected isDefault slaTier}}hasMadeShippingChanges slaGroups{__typename label sellerGroups{__typename id name isProSeller type catalogSellerId shipOptionGroup{__typename deliveryPrice{__typename displayValue value}itemIds shipMethod @include(if:$detailed)}}warningLabel}}...on SCGroup{__typename defaultMode collapsedItemIds checkoutable checkoutableErrors{code shouldDisableCheckout itemIds upstreamErrors{offerId upstreamErrorCode}}priceDetails{subTotal{...merge_priceTotalFields}}fulfillmentSwitchInfo{fulfillmentType benefit{type price itemCount date isWalmartPlusProgram @include(if:$detailed)}partialItemIds @include(if:$includePartialFulfillmentSwitching)}itemGroups{__typename label itemIds}accessPoint{...merge_accessPointFragment}reservation{...merge_reservationFragment}}...on DigitalDeliveryGroup{__typename defaultMode collapsedItemIds checkoutable checkoutableErrors{code shouldDisableCheckout itemIds upstreamErrors{offerId upstreamErrorCode}}priceDetails{subTotal{...merge_priceTotalFields}}itemGroups{__typename label itemIds}}...on Unscheduled{__typename defaultMode collapsedItemIds checkoutable checkoutableErrors{code shouldDisableCheckout itemIds upstreamErrors{offerId upstreamErrorCode}}priceDetails{subTotal{...merge_priceTotalFields}}itemGroups{__typename label itemIds}accessPoint{...merge_accessPointFragment}reservation{...merge_reservationFragment}fulfillmentSwitchInfo{fulfillmentType benefit{type price itemCount date isWalmartPlusProgram @include(if:$detailed)}partialItemIds @include(if:$includePartialFulfillmentSwitching)}}...on AutoCareCenter{__typename defaultMode collapsedItemIds startDate endDate accBasketType checkoutable checkoutableErrors{code shouldDisableCheckout itemIds upstreamErrors{offerId upstreamErrorCode}}priceDetails{subTotal{...merge_priceTotalFields}}itemGroups{__typename label itemIds}accessPoint{...merge_accessPointFragment}reservation{...merge_reservationFragment}fulfillmentSwitchInfo{fulfillmentType benefit{type price itemCount date isWalmartPlusProgram @include(if:$detailed)}partialItemIds @include(if:$includePartialFulfillmentSwitching)}}}suggestedSlotAvailability @include(if:$detailed){isPickupAvailable isDeliveryAvailable nextPickupSlot{startTime endTime slaInMins}nextDeliverySlot{startTime endTime slaInMins}nextUnscheduledPickupSlot{startTime endTime slaInMins}nextSlot{__typename...on RegularSlot{fulfillmentOption fulfillmentType startTime}...on DynamicExpressSlot{fulfillmentOption fulfillmentType startTime slaInMins}...on UnscheduledSlot{fulfillmentOption fulfillmentType startTime unscheduledHoldInDays}...on InHomeSlot{fulfillmentOption fulfillmentType startTime}}}}priceDetails{subTotal{value displayValue label @include(if:$detailed) key @include(if:$detailed) strikeOutDisplayValue @include(if:$detailed) strikeOutValue @include(if:$detailed)}fees @include(if:$detailed){...merge_priceTotalFields}taxTotal @include(if:$detailed){...merge_priceTotalFields}grandTotal @include(if:$detailed){...merge_priceTotalFields}belowMinimumFee @include(if:$detailed){...merge_priceTotalFields}minimumThreshold @include(if:$detailed){value displayValue}ebtSnapMaxEligible @include(if:$detailed){displayValue value}}affirm @include(if:$detailed){isMixedPromotionCart message{description termsUrl imageUrl monthlyPayment termLength isZeroAPR}nonAffirmGroup{...nonAffirmGroupFields}affirmGroups{...on AffirmItemGroup{__typename message{description termsUrl imageUrl monthlyPayment termLength isZeroAPR}flags{type displayLabel}name label itemCount itemIds defaultMode}}}migrationLineItems @include(if:$detailed){quantity quantityLabel quantityString accessibilityQuantityLabel offerId usItemId productName thumbnailUrl addOnService priceInfo{linePrice{value displayValue}}selectedVariants{name value}}checkoutableErrors{code shouldDisableCheckout itemIds}checkoutableWarnings @include(if:$detailed){code itemIds}operationalErrors{offerId itemId requestedQuantity adjustedQuantity code upstreamErrorCode}cartCustomerContext{...cartCustomerContextFragment}}}fragment merge_lineItemPriceInfoFragment on Price{displayValue value}fragment merge_postpaidPlanDetailsFragment on PostPaidPlan{espOrderSummaryId espOrderId espOrderLineId warpOrderId warpSessionId devicePayment{...merge_postpaidPlanPriceFragment}devicePlan{price{...merge_postpaidPlanPriceFragment}frequency duration annualPercentageRate}deviceDataPlan{...merge_deviceDataPlanFragment}}fragment merge_deviceDataPlanFragment on DeviceDataPlan{carrierName planType expiryTime activationFee{...merge_postpaidPlanPriceFragment}planDetails{price{...merge_postpaidPlanPriceFragment}frequency name}agreements{...merge_agreementFragment}}fragment merge_postpaidPlanPriceFragment on PriceDetailRow{key label displayValue value strikeOutDisplayValue strikeOutValue info{title message}}fragment merge_agreementFragment on CarrierAgreement{name type format value docTitle label}fragment merge_priceTotalFields on PriceDetailRow{label displayValue value key strikeOutDisplayValue strikeOutValue}fragment merge_accessPointFragment on AccessPoint{id assortmentStoreId name nodeAccessType fulfillmentType fulfillmentOption displayName timeZone address{addressLineOne addressLineTwo city postalCode state phone}}fragment merge_reservationFragment on Reservation{expiryTime isUnscheduled expired showSlotExpiredError reservedSlot{__typename...on RegularSlot{id price{total{displayValue}expressFee{displayValue}baseFee{displayValue}memberBaseFee{displayValue}}accessPointId fulfillmentOption startTime fulfillmentType slotMetadata endTime available supportedTimeZone isAlcoholRestricted}...on DynamicExpressSlot{id price{total{displayValue}expressFee{displayValue}baseFee{displayValue}memberBaseFee{displayValue}}accessPointId fulfillmentOption startTime fulfillmentType slotMetadata available slaInMins maxItemAllowed supportedTimeZone isAlcoholRestricted}...on UnscheduledSlot{price{total{displayValue}expressFee{displayValue}baseFee{displayValue}memberBaseFee{displayValue}}accessPointId fulfillmentOption startTime fulfillmentType slotMetadata unscheduledHoldInDays supportedTimeZone}...on InHomeSlot{id price{total{displayValue}expressFee{displayValue}baseFee{displayValue}memberBaseFee{displayValue}}accessPointId fulfillmentOption startTime fulfillmentType slotMetadata endTime available supportedTimeZone isAlcoholRestricted}}}fragment nonAffirmGroupFields on NonAffirmGroup{label itemCount itemIds collapsedItemIds}fragment cartCustomerContextFragment on CartCustomerContext{isMembershipOptedIn isEligibleForFreeTrial membershipData{isActiveMember}paymentData{hasCreditCard hasCapOne hasDSCard hasEBT isCapOneLinked showCapOneBanner}}","variables":{"input":{"cartId":null,"strategy":"MERGE"},"detailed":false}}`))
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt create atc request")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com")
	tk.AddGQLHeaders(req, `mutation MergeAndGetCart`)

	res, err := tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt make atc request")
		tk.Stop()
		return
	}

	tk.cartid = string(cartIdRe.FindSubmatch(res.Body)[1])
	tk.customerid = string(customerIdRe.FindSubmatch(res.Body)[1])
}
func (tk *Task) ATC() {
	tk.SetStatus(module.STATUS_ADDING_TO_CART)
	req, err := tk.NewRequest("POST", "https://www.walmart.com/orchestra/home/graphql", []byte(fmt.Sprintf(`{"query":"mutation updateItems( $input:UpdateItemsInput! $detailed:Boolean! = false $includePartialFulfillmentSwitching:Boolean! = false ){updateItems(input:$input){id checkoutable customer @include(if:$detailed){id isGuest}addressMode migrationLineItems @include(if:$detailed){quantity quantityLabel quantityString accessibilityQuantityLabel offerId usItemId productName thumbnailUrl addOnService priceInfo{linePrice{value displayValue}}selectedVariants{name value}}lineItems{id quantity quantityString quantityLabel createdDateTime displayAddOnServices selectedAddOnServices{offerId quantity groupType error{code upstreamErrorCode errorMsg}}isPreOrder @include(if:$detailed) bundleComponents{offerId quantity}registryId registryInfo{registryId registryType}fulfillmentPreference selectedVariants @include(if:$detailed){name value}priceInfo{priceDisplayCodes{showItemPrice priceDisplayCondition finalCostByWeight}itemPrice{...merge_lineItemPriceInfoFragment}wasPrice{...merge_lineItemPriceInfoFragment}unitPrice{...merge_lineItemPriceInfoFragment}linePrice{...merge_lineItemPriceInfoFragment}}product{name @include(if:$detailed) usItemId addOnServices{serviceType serviceTitle serviceSubTitle groups{groupType groupTitle assetUrl shortDescription services{displayName selectedDisplayName offerId currentPrice{priceString price}serviceMetaData}}}imageInfo @include(if:$detailed){thumbnailUrl}itemType offerId sellerId @include(if:$detailed) sellerName @include(if:$detailed) hasSellerBadge @include(if:$detailed) orderLimit orderMinLimit weightUnit @include(if:$detailed) weightIncrement @include(if:$detailed) salesUnit salesUnitType sellerType isAlcohol fulfillmentType @include(if:$detailed) fulfillmentSpeed @include(if:$detailed) fulfillmentTitle @include(if:$detailed) classType @include(if:$detailed) rhPath @include(if:$detailed) availabilityStatus @include(if:$detailed) brand @include(if:$detailed) category @include(if:$detailed){categoryPath}departmentName @include(if:$detailed) configuration @include(if:$detailed) snapEligible @include(if:$detailed) preOrder @include(if:$detailed){isPreOrder}}wirelessPlan @include(if:$detailed){planId mobileNumber postPaidPlan{...merge_postpaidPlanDetailsFragment}}fulfillmentSourcingDetails @include(if:$detailed){currentSelection requestedSelection fulfillmentBadge}}fulfillment{intent @include(if:$detailed) accessPoint @include(if:$detailed){...merge_accessPointFragment}reservation @include(if:$detailed){...merge_reservationFragment}storeId @include(if:$detailed) displayStoreSnackBarMessage homepageBookslotDetails @include(if:$detailed){title subTitle expiryText expiryTime slotExpiryText}deliveryAddress @include(if:$detailed){addressLineOne addressLineTwo city state postalCode firstName lastName id}fulfillmentItemGroups @include(if:$detailed){...on FCGroup{__typename defaultMode collapsedItemIds startDate endDate checkoutable priceDetails{subTotal{...merge_priceTotalFields}}fulfillmentSwitchInfo{fulfillmentType benefit{type price itemCount date isWalmartPlusProgram}partialItemIds @include(if:$includePartialFulfillmentSwitching)}shippingOptions{__typename itemIds availableShippingOptions{__typename id shippingMethod deliveryDate price{__typename displayValue value}label{prefix suffix}isSelected isDefault slaTier}}hasMadeShippingChanges slaGroups{__typename label sellerGroups{__typename id name isProSeller type catalogSellerId shipOptionGroup{__typename deliveryPrice{__typename displayValue value}itemIds shipMethod @include(if:$detailed)}}warningLabel}}...on SCGroup{__typename defaultMode collapsedItemIds checkoutable priceDetails{subTotal{...merge_priceTotalFields}}fulfillmentSwitchInfo{fulfillmentType benefit{type price itemCount date isWalmartPlusProgram}partialItemIds @include(if:$includePartialFulfillmentSwitching)}itemGroups{__typename label itemIds}accessPoint{...merge_accessPointFragment}reservation{...merge_reservationFragment}}...on DigitalDeliveryGroup{__typename defaultMode collapsedItemIds checkoutable priceDetails{subTotal{...merge_priceTotalFields}}itemGroups{__typename label itemIds}}...on Unscheduled{__typename defaultMode collapsedItemIds checkoutable priceDetails{subTotal{...merge_priceTotalFields}}itemGroups{__typename label itemIds}accessPoint{...merge_accessPointFragment}reservation{...merge_reservationFragment}fulfillmentSwitchInfo{fulfillmentType benefit{type price itemCount date isWalmartPlusProgram}partialItemIds @include(if:$includePartialFulfillmentSwitching)}}...on AutoCareCenter{__typename defaultMode collapsedItemIds startDate endDate accBasketType checkoutable priceDetails{subTotal{...merge_priceTotalFields}}itemGroups{__typename label itemIds}accessPoint{...merge_accessPointFragment}reservation{...merge_reservationFragment}fulfillmentSwitchInfo{fulfillmentType benefit{type price itemCount date isWalmartPlusProgram}partialItemIds @include(if:$includePartialFulfillmentSwitching)}}}suggestedSlotAvailability @include(if:$detailed){isPickupAvailable isDeliveryAvailable nextPickupSlot{startTime endTime slaInMins}nextDeliverySlot{startTime endTime slaInMins}nextUnscheduledPickupSlot{startTime endTime slaInMins}nextSlot{__typename...on RegularSlot{fulfillmentOption fulfillmentType startTime}...on DynamicExpressSlot{fulfillmentOption fulfillmentType startTime slaInMins}...on UnscheduledSlot{fulfillmentOption fulfillmentType startTime unscheduledHoldInDays}...on InHomeSlot{fulfillmentOption fulfillmentType startTime}}}}priceDetails{subTotal{value displayValue label @include(if:$detailed) key @include(if:$detailed) strikeOutDisplayValue @include(if:$detailed) strikeOutValue @include(if:$detailed)}fees @include(if:$detailed){...merge_priceTotalFields}taxTotal @include(if:$detailed){...merge_priceTotalFields}grandTotal @include(if:$detailed){...merge_priceTotalFields}belowMinimumFee @include(if:$detailed){...merge_priceTotalFields}minimumThreshold @include(if:$detailed){value displayValue}ebtSnapMaxEligible @include(if:$detailed){displayValue value}balanceToMinimumThreshold @include(if:$detailed){value displayValue}}affirm @include(if:$detailed){isMixedPromotionCart message{description termsUrl imageUrl monthlyPayment termLength isZeroAPR}nonAffirmGroup{...nonAffirmGroupFields}affirmGroups{...on AffirmItemGroup{__typename message{description termsUrl imageUrl monthlyPayment termLength isZeroAPR}flags{type displayLabel}name label itemCount itemIds defaultMode}}}checkoutableErrors{code shouldDisableCheckout itemIds}checkoutableWarnings @include(if:$detailed){code itemIds}operationalErrors{offerId itemId requestedQuantity adjustedQuantity code upstreamErrorCode}cartCustomerContext{...cartCustomerContextFragment}}}fragment merge_postpaidPlanDetailsFragment on PostPaidPlan{espOrderSummaryId espOrderId espOrderLineId warpOrderId warpSessionId devicePayment{...merge_postpaidPlanPriceFragment}devicePlan{price{...merge_postpaidPlanPriceFragment}frequency duration annualPercentageRate}deviceDataPlan{...merge_deviceDataPlanFragment}}fragment merge_deviceDataPlanFragment on DeviceDataPlan{carrierName planType expiryTime activationFee{...merge_postpaidPlanPriceFragment}planDetails{price{...merge_postpaidPlanPriceFragment}frequency name}agreements{...merge_agreementFragment}}fragment merge_postpaidPlanPriceFragment on PriceDetailRow{key label displayValue value strikeOutDisplayValue strikeOutValue info{title message}}fragment merge_agreementFragment on CarrierAgreement{name type format value docTitle label}fragment merge_priceTotalFields on PriceDetailRow{label displayValue value key strikeOutDisplayValue strikeOutValue}fragment merge_lineItemPriceInfoFragment on Price{displayValue value}fragment merge_accessPointFragment on AccessPoint{id assortmentStoreId name nodeAccessType fulfillmentType fulfillmentOption displayName timeZone address{addressLineOne addressLineTwo city postalCode state phone}}fragment merge_reservationFragment on Reservation{expiryTime isUnscheduled expired showSlotExpiredError reservedSlot{__typename...on RegularSlot{id price{total{displayValue}expressFee{displayValue}baseFee{displayValue}memberBaseFee{displayValue}}accessPointId fulfillmentOption startTime fulfillmentType slotMetadata endTime available supportedTimeZone isAlcoholRestricted}...on DynamicExpressSlot{id price{total{displayValue}expressFee{displayValue}baseFee{displayValue}memberBaseFee{displayValue}}accessPointId fulfillmentOption startTime fulfillmentType slotMetadata available slaInMins maxItemAllowed supportedTimeZone isAlcoholRestricted}...on UnscheduledSlot{price{total{displayValue}expressFee{displayValue}baseFee{displayValue}memberBaseFee{displayValue}}accessPointId fulfillmentOption startTime fulfillmentType slotMetadata unscheduledHoldInDays supportedTimeZone}...on InHomeSlot{id price{total{displayValue}expressFee{displayValue}baseFee{displayValue}memberBaseFee{displayValue}}accessPointId fulfillmentOption startTime fulfillmentType slotMetadata endTime available supportedTimeZone isAlcoholRestricted}}}fragment nonAffirmGroupFields on NonAffirmGroup{label itemCount itemIds collapsedItemIds}fragment cartCustomerContextFragment on CartCustomerContext{isMembershipOptedIn isEligibleForFreeTrial membershipData{isActiveMember}paymentData{hasCreditCard hasCapOne hasDSCard hasEBT isCapOneLinked showCapOneBanner}}","variables":{"input":{"cartId":"%s","items":[{"offerId":"%s","quantity":1,"lineItemId":"%s"}],"semStoreId":"","semPostalCode":"","isGiftOrder":null},"detailed":false,"includePartialFulfillmentSwitching":true}}`, tk.cartid, tk.offerid, tk.lineitemid)))
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt create atc request")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com")
	tk.AddGQLHeaders(req, `mutation updateItems`)
	_, err = tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt make atc request")
		tk.Stop()
		return
	}
	tk.SetStatus(module.STATUS_ADDED_TO_CART)
}

func (tk *Task) CreateDelivery() {
	tk.SetStatus(module.STATUS_SUBMITTING_SHIPPING)
	var addrline2 string
	if tk.Data.Profile.Shipping.ShippingAddress.AddressLine2 != nil{
		addrline2 = *tk.Data.Profile.Shipping.ShippingAddress.AddressLine2
	}
	req, err := tk.NewRequest("POST", "https://www.walmart.com/orchestra/cartxo/graphql", []byte(fmt.Sprintf(`{"query":"mutation CreateDeliveryAddress($input:AccountAddressesInput!){createAccountAddress(input:$input){...DeliveryAddressMutationResponse}}fragment DeliveryAddressMutationResponse on MutateAccountAddressResponse{...AddressMutationResponse newAddress{id accessPoint{...AccessPoint}...BaseAddressResponse}}fragment AccessPoint on AccessPointRovr{id assortmentStoreId fulfillmentType accountFulfillmentOption accountAccessType}fragment AddressMutationResponse on MutateAccountAddressResponse{errors{code}enteredAddress{...BasicAddress}suggestedAddresses{...BasicAddress sealedAddress}newAddress{id...BaseAddressResponse}allowAvsOverride}fragment BasicAddress on AccountAddressBase{addressLineOne addressLineTwo city state postalCode}fragment BaseAddressResponse on AccountAddress{...BasicAddress firstName lastName phone isDefault deliveryInstructions serviceStatus capabilities allowEditOrRemove}","variables":{"input":{"address":{"addressLineOne":"%s","addressLineTwo":"%s","city":"%s","postalCode":"%s","state":"%s","addressType":null,"businessName":null,"isApoFpo":null,"isLoadingDockAvailable":null,"isPoBox":null,"sealedAddress":null},"firstName":"%s","lastName":"%s","deliveryInstructions":null,"displayLabel":null,"isDefault":false,"phone":"%s","overrideAvs":false}}}`, tk.Data.Profile.Shipping.ShippingAddress.AddressLine, addrline2, tk.Data.Profile.Shipping.ShippingAddress.City, tk.Data.Profile.Shipping.ShippingAddress.ZIP, tk.Data.Profile.Shipping.ShippingAddress.StateCode, tk.Data.Profile.Shipping.FirstName, tk.Data.Profile.Shipping.LastName, tk.Data.Profile.Shipping.PhoneNumber)))
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt create delivery request")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com")
	tk.AddGQLHeaders(req, `mutation CreateDeliveryAddress`)

	res, err := tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt make delivery request")
		tk.Stop()
		return
	}

	tk.accesspoint = string(accessPointRe.FindSubmatch(res.Body)[1])
	tk.newaddress = string(newAddressRe.FindSubmatch(res.Body)[1])
}
func (tk *Task) SetFulfillment() {
	tk.SetStatus(module.STATUS_SUBMITTING_SHIPPING)
	req, err := tk.NewRequest("POST", "https://www.walmart.com/orchestra/cartxo/graphql", []byte(fmt.Sprintf(`{"query":"mutation setFulfillment( $input:SetFulfillmentInput! $includePartialFulfillmentSwitching:Boolean! = false ){setFulfillment(input:$input){...CartFragment}}fragment CartFragment on Cart{id checkoutable customer{id isGuest}cartGiftingDetails{isGiftOrder hasGiftEligibleItem isAddOnServiceAdjustmentNeeded isWalmartProtectionPlanPresent isAppleCarePresent}addressMode migrationLineItems{quantity quantityLabel quantityString accessibilityQuantityLabel offerId usItemId productName thumbnailUrl addOnService priceInfo{linePrice{value displayValue}}selectedVariants{name value}}lineItems{id quantity quantityString quantityLabel isPreOrder isGiftEligible displayAddOnServices createdDateTime selectedAddOnServices{offerId quantity groupType isGiftEligible error{code upstreamErrorCode errorMsg}}bundleComponents{offerId quantity}registryId fulfillmentPreference selectedVariants{name value}priceInfo{priceDisplayCodes{showItemPrice priceDisplayCondition finalCostByWeight}itemPrice{...lineItemPriceInfoFragment}wasPrice{...lineItemPriceInfoFragment}unitPrice{...lineItemPriceInfoFragment}linePrice{...lineItemPriceInfoFragment}}product{name usItemId imageInfo{thumbnailUrl}addOnServices{serviceType serviceTitle serviceSubTitle groups{groupType groupTitle assetUrl shortDescription services{displayName selectedDisplayName offerId currentPrice{priceString price}serviceMetaData}}}itemType offerId sellerId sellerName hasSellerBadge orderLimit orderMinLimit weightUnit weightIncrement salesUnit salesUnitType sellerType isAlcohol fulfillmentType fulfillmentSpeed fulfillmentTitle classType rhPath availabilityStatus brand category{categoryPath}departmentName configuration snapEligible preOrder{isPreOrder}}registryInfo{registryId registryType}wirelessPlan{planId mobileNumber postPaidPlan{...postpaidPlanDetailsFragment}}fulfillmentSourcingDetails{currentSelection requestedSelection fulfillmentBadge}availableQty}fulfillment{intent accessPoint{...accessPointFragment}reservation{...reservationFragment}storeId displayStoreSnackBarMessage homepageBookslotDetails{title subTitle expiryText expiryTime slotExpiryText}deliveryAddress{addressLineOne addressLineTwo city state postalCode firstName lastName id}fulfillmentItemGroups{...on FCGroup{__typename defaultMode collapsedItemIds startDate endDate checkoutable checkoutableErrors{code shouldDisableCheckout itemIds upstreamErrors{offerId upstreamErrorCode}}priceDetails{subTotal{...priceTotalFields}}fulfillmentSwitchInfo{fulfillmentType benefit{type price itemCount date isWalmartPlusProgram}partialItemIds @include(if:$includePartialFulfillmentSwitching)}shippingOptions{__typename itemIds availableShippingOptions{__typename id shippingMethod deliveryDate price{__typename displayValue value}label{prefix suffix}isSelected isDefault slaTier}}hasMadeShippingChanges slaGroups{__typename label deliveryDate sellerGroups{__typename id name isProSeller type catalogSellerId shipOptionGroup{__typename deliveryPrice{__typename displayValue value}itemIds shipMethod}}warningLabel}}...on SCGroup{__typename defaultMode collapsedItemIds checkoutable checkoutableErrors{code shouldDisableCheckout itemIds upstreamErrors{offerId upstreamErrorCode}}priceDetails{subTotal{...priceTotalFields}}fulfillmentSwitchInfo{fulfillmentType benefit{type price itemCount date isWalmartPlusProgram}partialItemIds @include(if:$includePartialFulfillmentSwitching)}itemGroups{__typename label itemIds}accessPoint{...accessPointFragment}reservation{...reservationFragment}}...on DigitalDeliveryGroup{__typename defaultMode collapsedItemIds checkoutable checkoutableErrors{code shouldDisableCheckout itemIds upstreamErrors{offerId upstreamErrorCode}}priceDetails{subTotal{...priceTotalFields}}itemGroups{__typename label itemIds}}...on Unscheduled{__typename defaultMode collapsedItemIds checkoutable checkoutableErrors{code shouldDisableCheckout itemIds upstreamErrors{offerId upstreamErrorCode}}priceDetails{subTotal{...priceTotalFields}}itemGroups{__typename label itemIds}accessPoint{...accessPointFragment}reservation{...reservationFragment}fulfillmentSwitchInfo{fulfillmentType benefit{type price itemCount date isWalmartPlusProgram}partialItemIds @include(if:$includePartialFulfillmentSwitching)}}...on AutoCareCenter{__typename defaultMode collapsedItemIds startDate endDate accBasketType checkoutable checkoutableErrors{code shouldDisableCheckout itemIds upstreamErrors{offerId upstreamErrorCode}}priceDetails{subTotal{...priceTotalFields}}itemGroups{__typename label itemIds}accessPoint{...accessPointFragment}reservation{...reservationFragment}fulfillmentSwitchInfo{fulfillmentType benefit{type price itemCount date isWalmartPlusProgram}partialItemIds @include(if:$includePartialFulfillmentSwitching)}}}suggestedSlotAvailability{isPickupAvailable isDeliveryAvailable nextPickupSlot{startTime endTime slaInMins}nextDeliverySlot{startTime endTime slaInMins}nextUnscheduledPickupSlot{startTime endTime slaInMins}nextSlot{__typename...on RegularSlot{fulfillmentOption fulfillmentType startTime}...on DynamicExpressSlot{fulfillmentOption fulfillmentType startTime slaInMins}...on UnscheduledSlot{fulfillmentOption fulfillmentType startTime unscheduledHoldInDays}...on InHomeSlot{fulfillmentOption fulfillmentType startTime}}}}priceDetails{subTotal{...priceTotalFields}fees{...priceTotalFields}taxTotal{...priceTotalFields}grandTotal{...priceTotalFields}belowMinimumFee{...priceTotalFields}minimumThreshold{value displayValue}ebtSnapMaxEligible{displayValue value}balanceToMinimumThreshold{value displayValue}}affirm{isMixedPromotionCart message{description termsUrl imageUrl monthlyPayment termLength isZeroAPR}nonAffirmGroup{...nonAffirmGroupFields}affirmGroups{...on AffirmItemGroup{__typename message{description termsUrl imageUrl monthlyPayment termLength isZeroAPR}flags{type displayLabel}name label itemCount itemIds defaultMode}}}checkoutableErrors{code shouldDisableCheckout itemIds upstreamErrors{offerId upstreamErrorCode}}checkoutableWarnings{code itemIds}operationalErrors{offerId itemId requestedQuantity adjustedQuantity code upstreamErrorCode}cartCustomerContext{...cartCustomerContextFragment}}fragment postpaidPlanDetailsFragment on PostPaidPlan{espOrderSummaryId espOrderId espOrderLineId warpOrderId warpSessionId devicePayment{...postpaidPlanPriceFragment}devicePlan{price{...postpaidPlanPriceFragment}frequency duration annualPercentageRate}deviceDataPlan{...deviceDataPlanFragment}}fragment deviceDataPlanFragment on DeviceDataPlan{carrierName planType expiryTime activationFee{...postpaidPlanPriceFragment}planDetails{price{...postpaidPlanPriceFragment}frequency name}agreements{...agreementFragment}}fragment postpaidPlanPriceFragment on PriceDetailRow{key label displayValue value strikeOutDisplayValue strikeOutValue info{title message}}fragment agreementFragment on CarrierAgreement{name type format value docTitle label}fragment priceTotalFields on PriceDetailRow{label displayValue value key strikeOutDisplayValue strikeOutValue}fragment lineItemPriceInfoFragment on Price{displayValue value}fragment accessPointFragment on AccessPoint{id assortmentStoreId name nodeAccessType fulfillmentType fulfillmentOption displayName timeZone address{addressLineOne addressLineTwo city postalCode state phone}}fragment reservationFragment on Reservation{expiryTime isUnscheduled expired showSlotExpiredError reservedSlot{__typename...on RegularSlot{id price{total{displayValue}expressFee{displayValue}baseFee{displayValue}memberBaseFee{displayValue}}nodeAccessType accessPointId fulfillmentOption startTime fulfillmentType slotMetadata endTime available supportedTimeZone isAlcoholRestricted}...on DynamicExpressSlot{id price{total{displayValue}expressFee{displayValue}baseFee{displayValue}memberBaseFee{displayValue}}accessPointId fulfillmentOption startTime fulfillmentType slotMetadata available slaInMins maxItemAllowed supportedTimeZone isAlcoholRestricted}...on UnscheduledSlot{price{total{displayValue}expressFee{displayValue}baseFee{displayValue}memberBaseFee{displayValue}}accessPointId fulfillmentOption startTime fulfillmentType slotMetadata unscheduledHoldInDays supportedTimeZone}...on InHomeSlot{id price{total{displayValue}expressFee{displayValue}baseFee{displayValue}memberBaseFee{displayValue}}accessPointId fulfillmentOption startTime fulfillmentType slotMetadata endTime available supportedTimeZone isAlcoholRestricted}}}fragment nonAffirmGroupFields on NonAffirmGroup{label itemCount itemIds collapsedItemIds}fragment cartCustomerContextFragment on CartCustomerContext{isMembershipOptedIn isEligibleForFreeTrial membershipData{isActiveMember}paymentData{hasCreditCard hasCapOne hasDSCard hasEBT isCapOneLinked showCapOneBanner}}","variables":{"input":{"accessPointId":"%s","addressId":"%s","cartId":"%s","registry":null,"fulfillmentOption":"SHIPPING","postalCode":null,"storeId":%s,"isGiftAddress":null},"includePartialFulfillmentSwitching":true}}`, tk.accesspoint, tk.newaddress, tk.cartid, tk.storeids[0])))
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt create setfulfillment request")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com")
	tk.AddGQLHeaders(req, `mutation setFulfillment`)

	_, err = tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt make setfulfillment request")
		tk.Stop()
		return
	}
}
func (tk *Task) CreateContract() {
	tk.SetStatus(module.STATUS_SUBMITTING_PAYMENT)
	req, err := tk.NewRequest("POST", "https://www.walmart.com/orchestra/cartxo/graphql", []byte(fmt.Sprintf(`{"query":"mutation CreateContract( $createContractInput:CreatePurchaseContractInput! $promosEnable:Boolean! $wplusEnabled:Boolean! ){createPurchaseContract(input:$createContractInput){...ContractFragment}}fragment ContractFragment on PurchaseContract{id associateDiscountStatus addressMode tenderPlanId papEbtAllowed allowedPaymentGroupTypes cartCustomerContext @include(if:$wplusEnabled){isMembershipOptedIn isEligibleForFreeTrial paymentData{hasCreditCard}}checkoutError{code errorData{__typename...on OutOfStock{offerId}__typename...on UnavailableOffer{offerId}__typename...on ItemExpired{offerId}__typename...on ItemQuantityAdjusted{offerId requestedQuantity adjustedQuantity}}operationalErrorCode message}checkoutableWarnings{code itemIds}allocationStatus payments{id paymentType cardType lastFour isDefault cvvRequired preferenceId paymentPreferenceId paymentHandle expiryMonth expiryYear firstName lastName email amountPaid cardImage cardImageAlt isLinkedCard capOneReward{credentialId redemptionUrl redemptionRate redemptionMethod rewardPointsBalance rewardPointsSelected rewardAmountSelected}remainingBalance{displayValue value}}order{id status orderVersion mobileNumber}terms{alcoholAccepted bagFeeAccepted smsOptInAccepted marketingEmailPrefOptIn}donationDetails{charityEIN charityName amount{displayValue value}acceptDonation}lineItems{...LineItemFields}tippingDetails{suggestedAmounts{value displayValue}maxAmount{value displayValue}selectedTippingAmount{value displayValue}}customer{id firstName lastName isGuest email phone}fulfillment{deliveryDetails{deliveryInstructions deliveryOption}pickupChoices{isSelected fulfillmentType accessType accessMode accessPointId}deliveryAddress{...AddressFields}alternatePickupPerson{...PickupPersonFields}primaryPickupPerson{...PickupPersonFields}fulfillmentItemGroups{...FulfillmentItemGroupsFields}accessPoint{...AccessPointFields}reservation{...ReservationFields}}priceDetails{subTotal{...PriceDetailRowFields}totalItemQuantity fees{...PriceDetailRowFields}taxTotal{...PriceDetailRowFields}grandTotal{...PriceDetailRowFields}belowMinimumFee{...PriceDetailRowFields}authorizationAmount{...PriceDetailRowFields}weightDebitTotal{...PriceDetailRowFields}discounts{...PriceDetailRowFields}otcDeliveryBenefit{...PriceDetailRowFields}ebtSnapMaxEligible{...PriceDetailRowFields}ebtCashMaxEligible{...PriceDetailRowFields}hasAmountUnallocated affirm{__typename message{...AffirmMessageFields}}}checkoutGiftingDetails{isCheckoutGiftingOptin isWalmartProtectionPlanPresent isAppleCarePresent isRestrictedPaymentPresent giftMessageDetails{giftingMessage recipientEmail recipientName senderName}}promotions @include(if:$promosEnable){displayValue promoId terms}showPromotions @include(if:$promosEnable) errors{code message lineItems{...LineItemFields}}}fragment LineItemFields on LineItem{id quantity quantityString quantityLabel accessibilityQuantityLabel isPreOrder fulfillmentSourcingDetails{currentSelection requestedSelection}packageQuantity priceInfo{priceDisplayCodes{showItemPrice priceDisplayCondition finalCostByWeight}itemPrice{displayValue value}linePrice{displayValue value}preDiscountedLinePrice{displayValue value}wasPrice{displayValue value}unitPrice{displayValue value}}isSubstitutionSelected isGiftEligible selectedVariants{name value}product{id name usItemId itemType imageInfo{thumbnailUrl}offerId orderLimit orderMinLimit weightIncrement weightUnit averageWeight salesUnitType availabilityStatus isSubstitutionEligible isAlcohol configuration hasSellerBadge sellerId sellerName sellerType preOrder{...preOrderFragment}addOnServices{serviceType groups{groupType services{selectedDisplayName offerId currentPrice{priceString}}}}}discounts{key label displayValue @include(if:$promosEnable) displayLabel @include(if:$promosEnable)}wirelessPlan{planId mobileNumber __typename postPaidPlan{...postpaidPlanDetailsFragment}}selectedAddOnServices{offerId quantity groupType}registryInfo{registryId registryType}}fragment postpaidPlanDetailsFragment on PostPaidPlan{__typename espOrderSummaryId espOrderId espOrderLineId warpOrderId warpSessionId devicePayment{...postpaidPlanPriceFragment}devicePlan{__typename price{...postpaidPlanPriceFragment}frequency duration annualPercentageRate}deviceDataPlan{...deviceDataPlanFragment}}fragment deviceDataPlanFragment on DeviceDataPlan{__typename carrierName planType expiryTime activationFee{...postpaidPlanPriceFragment}planDetails{__typename price{...postpaidPlanPriceFragment}frequency name}agreements{...agreementFragment}}fragment postpaidPlanPriceFragment on PriceDetailRow{__typename key label displayValue value strikeOutDisplayValue strikeOutValue info{__typename title message}}fragment agreementFragment on CarrierAgreement{__typename name type format value docTitle label}fragment preOrderFragment on PreOrder{streetDate streetDateDisplayable streetDateType isPreOrder preOrderMessage preOrderStreetDateMessage}fragment AddressFields on Address{id addressLineOne addressLineTwo city state postalCode firstName lastName phone}fragment PickupPersonFields on PickupPerson{id firstName lastName email}fragment PriceDetailRowFields on PriceDetailRow{__typename key label displayValue value strikeOutValue strikeOutDisplayValue info{__typename title message}}fragment AccessPointFields on AccessPoint{id name assortmentStoreId displayName timeZone address{id addressLineOne addressLineTwo city state postalCode firstName lastName phone}isTest allowBagFee bagFeeValue isExpressEligible fulfillmentOption instructions nodeAccessType}fragment ReservationFields on Reservation{id expiryTime isUnscheduled expired showSlotExpiredError reservedSlot{__typename...on RegularSlot{id price{total{displayValue}expressFee{displayValue}baseFee{displayValue}memberBaseFee{displayValue}}accessPointId fulfillmentOption startTime fulfillmentType slotMetadata slotExpiryTime endTime available supportedTimeZone}...on DynamicExpressSlot{id price{total{displayValue}expressFee{displayValue}baseFee{displayValue}memberBaseFee{displayValue}}accessPointId fulfillmentOption startTime endTime fulfillmentType slotMetadata slotExpiryTime available slaInMins maxItemAllowed supportedTimeZone}...on UnscheduledSlot{price{total{displayValue}expressFee{displayValue}baseFee{displayValue}memberBaseFee{displayValue}}accessPointId fulfillmentOption startTime fulfillmentType slotMetadata unscheduledHoldInDays supportedTimeZone}...on InHomeSlot{id price{total{displayValue}expressFee{displayValue}baseFee{displayValue}memberBaseFee{displayValue}}accessPointId fulfillmentOption startTime fulfillmentType slotMetadata slotExpiryTime endTime available supportedTimeZone}}}fragment AffirmMessageFields on AffirmMessage{__typename description termsUrl imageUrl monthlyPayment termLength isZeroAPR}fragment FulfillmentItemGroupsFields on FulfillmentItemGroup{...on SCGroup{__typename defaultMode collapsedItemIds itemGroups{__typename label itemIds}accessPoint{...AccessPointFields}reservation{...ReservationFields}}...on DigitalDeliveryGroup{__typename defaultMode collapsedItemIds itemGroups{__typename label itemIds}}...on Unscheduled{__typename defaultMode collapsedItemIds itemGroups{__typename label itemIds}accessPoint{...AccessPointFields}reservation{...ReservationFields}}...on FCGroup{__typename defaultMode collapsedItemIds startDate endDate isUnscheduledDeliveryEligible shippingOptions{__typename itemIds availableShippingOptions{__typename id shippingMethod deliveryDate price{__typename displayValue value}label{prefix suffix}isSelected isDefault}}hasMadeShippingChanges slaGroups{__typename label deliveryDate warningLabel sellerGroups{__typename id name isProSeller type shipOptionGroup{__typename deliveryPrice{__typename displayValue value}itemIds shipMethod}}}}...on AutoCareCenter{__typename defaultMode startDate endDate accBasketType collapsedItemIds itemGroups{__typename label itemIds}accessPoint{...AccessPointFields}reservation{...ReservationFields}}}","variables":{"createContractInput":{"cartId":"%s"},"promosEnable":true,"wplusEnabled":true}}`, tk.cartid)))
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt create contract request")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com")
	tk.AddGQLHeaders(req, `mutation CreateContract`)

	res, err := tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt make contract request")
		tk.Stop()
		return
	}

	tk.contractid = string(contractIdRe.FindSubmatch(res.Body)[1])
	tk.tenderid = string(tenderPlanRe.FindSubmatch(res.Body)[1])
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
	PhaseVal, err := strconv.Atoi(string(PhaseRe.FindSubmatch(res.Body)[1]))
	if err != nil{
		tk.SetStatus(module.STATUS_ERROR, "could not read PhaseVal")
		tk.Stop()
		return
	}

	tk.PIE = encryption.PIEStruct{
		L:      lVal,
		E:      eVal,
		K:      string(pieKRe.FindSubmatch(res.Body)[1]),
		Key_id: string(pieKeyIdRe.FindSubmatch(res.Body)[1]),
		Phase: PhaseVal,
	}
}

func (tk *Task) CreateCreditCart(){
	tk.SetStatus(module.STATUS_SUBMITTING_PAYMENT)
	encarr := encryption.ProtectPANandCVV(tk.cardType(), tk.Data.Profile.Billing.CVV, 1, tk.PIE)
	var addrline2 string
	var req *fasttls.Request
	var err error
	
	if tk.Data.Profile.Shipping.BillingIsShipping{
		if tk.Data.Profile.Shipping.ShippingAddress.AddressLine2 != nil{
			addrline2 = *tk.Data.Profile.Shipping.ShippingAddress.AddressLine2
		}
		req, err = tk.NewRequest("POST", "https://www.walmart.com/orchestra/cartxo/graphql", []byte(fmt.Sprintf(`{"query":"mutation CreateCreditCard($input:AccountCreditCardInput!){createAccountCreditCard(input:$input){errors{code message}creditCard{...CreditCardFragment}}}fragment CreditCardFragment on CreditCard{__typename firstName lastName phone addressLineOne addressLineTwo city state postalCode cardType expiryYear expiryMonth lastFour id isDefault isExpired needVerifyCVV isEditable capOneProperties{shouldPromptForLink}linkedCard{availableCredit currentCreditBalance currentMinimumAmountDue minimumPaymentDueDate statementBalance statementDate rewards{rewardsBalance rewardsCurrency cashValue cashDisplayValue canRedeem}links{linkMethod linkHref linkType}}}","variables":{"input":{"firstName":"%s","lastName":"%s","phone":"%s","address":{"addressLineOne":"%s","addressLineTwo":"%s","postalCode":"%s","city":"%s","state":"%s","isApoFpo":null,"isLoadingDockAvailable":null,"isPoBox":null,"businessName":null,"addressType":null,"sealedAddress":null},"expiryYear":%s,"expiryMonth":%s,"isDefault":true,"cardType":"%s","integrityCheck":"%s","keyId":"%s","phase":"%s","encryptedPan":"%s","encryptedCVV":"%s","sourceFeature":"ACCOUNT_PAGE","cartId":null,"checkoutSessionId":null}}}`, tk.Data.Profile.Shipping.FirstName, tk.Data.Profile.Shipping.LastName, tk.FormatPhone(), tk.Data.Profile.Shipping.ShippingAddress.AddressLine, addrline2, tk.Data.Profile.Shipping.ShippingAddress.ZIP, tk.Data.Profile.Shipping.ShippingAddress.City, tk.Data.Profile.Shipping.ShippingAddress.StateCode, "20"+tk.Data.Profile.Billing.ExpirationYear, leadZeroRe.ReplaceAllString(tk.Data.Profile.Billing.ExpirationMonth, ""), tk.cardType(), encarr[3], tk.PIE.Key_id, tk.PIE.Phase, encarr[0], encarr[1])))
		if err != nil {
			tk.SetStatus(module.STATUS_ERROR, "couldnt create creditcard request")
			tk.Stop()
			return
		}	
	}else{
		if tk.Data.Profile.Shipping.BillingAddress.AddressLine2 != nil{
			addrline2 = *tk.Data.Profile.Shipping.BillingAddress.AddressLine2
		}
		req, err = tk.NewRequest("POST", "https://www.walmart.com/orchestra/cartxo/graphql", []byte(fmt.Sprintf(`{"query":"mutation CreateCreditCard($input:AccountCreditCardInput!){createAccountCreditCard(input:$input){errors{code message}creditCard{...CreditCardFragment}}}fragment CreditCardFragment on CreditCard{__typename firstName lastName phone addressLineOne addressLineTwo city state postalCode cardType expiryYear expiryMonth lastFour id isDefault isExpired needVerifyCVV isEditable capOneProperties{shouldPromptForLink}linkedCard{availableCredit currentCreditBalance currentMinimumAmountDue minimumPaymentDueDate statementBalance statementDate rewards{rewardsBalance rewardsCurrency cashValue cashDisplayValue canRedeem}links{linkMethod linkHref linkType}}}","variables":{"input":{"firstName":"%s","lastName":"%s","phone":"%s","address":{"addressLineOne":"%s","addressLineTwo":"%s","postalCode":"%s","city":"%s","state":"%s","isApoFpo":null,"isLoadingDockAvailable":null,"isPoBox":null,"businessName":null,"addressType":null,"sealedAddress":null},"expiryYear":%s,"expiryMonth":%s,"isDefault":true,"cardType":"%s","integrityCheck":"%s","keyId":"%s","phase":"%s","encryptedPan":"%s","encryptedCVV":"%s","sourceFeature":"ACCOUNT_PAGE","cartId":null,"checkoutSessionId":null}}}`, tk.Data.Profile.Shipping.FirstName, tk.Data.Profile.Shipping.LastName, tk.FormatPhone(), tk.Data.Profile.Shipping.BillingAddress.AddressLine, addrline2, tk.Data.Profile.Shipping.BillingAddress.ZIP, tk.Data.Profile.Shipping.BillingAddress.City, tk.Data.Profile.Shipping.BillingAddress.StateCode, "20"+tk.Data.Profile.Billing.ExpirationYear, leadZeroRe.ReplaceAllString(tk.Data.Profile.Billing.ExpirationMonth, ""), tk.cardType(), encarr[3], tk.PIE.Key_id, tk.PIE.Phase, encarr[0], encarr[1])))
		if err != nil {
			tk.SetStatus(module.STATUS_ERROR, "couldnt create creditcard request")
			tk.Stop()
			return
		}
	}
	
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com")
	tk.AddGQLHeaders(req, `mutation CreateCreditCard`)

	res, err := tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt make creditcard request")
		tk.Stop()
		return
	}

	tk.preferenceid = string(preferenceIdRe.FindSubmatch(res.Body)[1])
}
func (tk *Task) UpdateTenderPlan(){
	tk.SetStatus(module.STATUS_SUBMITTING_PAYMENT)
	req, err := tk.NewRequest("POST", "https://www.walmart.com/orchestra/cartxo/graphql", []byte(fmt.Sprintf(`{"query":"mutation updateTenderPlan($input:UpdateTenderPlanInput!){updateTenderPlan(input:$input){__typename tenderPlan{...TenderPlanFields}errors{...ErrorFields}}}fragment TenderPlanFields on TenderPlan{__typename id contractId grandTotal{...PriceDetailRowFields}authorizationAmount{...PriceDetailRowFields}allocationStatus paymentGroups{...PaymentGroupFields}otcDeliveryBenefit{...PriceDetailRowFields}otherAllowedPayments{type status}addPaymentType hasAmountUnallocated weightDebitTotal{...PriceDetailRowFields}}fragment PriceDetailRowFields on PriceDetailRow{__typename key label displayValue value info{__typename title message}}fragment PaymentGroupFields on TenderPlanPaymentGroup{__typename type subTotal{__typename key label displayValue value info{__typename title message}}selectedCount allocations{...CreditCardAllocationFragment...GiftCardAllocationFragment...EbtCardAllocationFragment...DsCardAllocationFragment...PayPalAllocationFragment...AffirmAllocationFragment}coversOrderTotal statusMessage}fragment CreditCardAllocationFragment on CreditCardAllocation{__typename card{...CreditCardFragment}canEditOrDelete canDeselect isEligible isSelected allocationAmount{__typename displayValue value}capOneReward{...CapOneFields}statusMessage{__typename messageStatus messageType}paymentType}fragment CapOneFields on CapOneReward{credentialId redemptionRate redemptionUrl redemptionMethod rewardPointsBalance rewardPointsSelected rewardAmountSelected}fragment CreditCardFragment on CreditCard{__typename id isDefault cardAccountLinked needVerifyCVV cardType expiryMonth expiryYear isExpired firstName lastName lastFour isEditable phone}fragment GiftCardAllocationFragment on GiftCardAllocation{__typename card{...GiftCardFields}canEditOrDelete canDeselect isEligible isSelected allocationAmount{__typename displayValue value}statusMessage{__typename messageStatus messageType}paymentType remainingBalance{__typename displayValue value}}fragment GiftCardFields on GiftCard{__typename id balance{cardBalance}lastFour displayLabel}fragment EbtCardAllocationFragment on EbtCardAllocation{__typename card{__typename id lastFour firstName lastName}canEditOrDelete canDeselect isEligible isSelected allocationAmount{__typename displayValue value}statusMessage{__typename messageStatus messageType}paymentType ebtMaxEligibleAmount{__typename displayValue value}cardBalance{__typename displayValue value}}fragment DsCardAllocationFragment on DsCardAllocation{__typename card{...DsCardFields}canEditOrDelete canDeselect isEligible isSelected allocationAmount{__typename displayValue value}statusMessage{__typename messageStatus messageType}paymentType canApplyAmount{__typename displayValue value}remainingBalance{__typename displayValue value}paymentPromotions{__typename programName canApplyAmount{__typename displayValue value}allocationAmount{__typename displayValue value}remainingBalance{__typename displayValue value}balance{__typename displayValue value}termsLink isInvalid}otcShippingBenefit termsLink}fragment DsCardFields on DsCard{__typename id displayLabel lastFour fundingProgram balance{cardBalance}dsCardType cardName}fragment PayPalAllocationFragment on PayPalAllocation{__typename allocationAmount{__typename displayValue value}paymentHandle paymentType email}fragment AffirmAllocationFragment on AffirmAllocation{__typename allocationAmount{__typename displayValue value}paymentHandle paymentType cardType firstName lastName}fragment ErrorFields on TenderPlanError{__typename code message}","variables":{"input":{"contractId":"%s","tenderPlanId":"%s","payments":[{"paymentType":"CREDITCARD","preferenceId":"%s","amount":null,"capOneReward":null,"cardType":null,"paymentHandle":null}],"accountRefresh":true,"isAmendFlow":false}}}`, tk.contractid, tk.tenderid, tk.preferenceid)))
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt create creditcard request")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com")
	tk.AddGQLHeaders(req, `mutation updateTenderPlan`)

	_, err = tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt make creditcard request")
		tk.Stop()
		return
	}
}
func (tk *Task) PlaceOrder(){
	tk.SetStatus(module.STATUS_CHECKING_OUT)
	req, err := tk.NewRequest("POST", "https://www.walmart.com/orchestra/cartxo/graphql", []byte(fmt.Sprintf(`{"query":"mutation PlaceOrder( $placeOrderInput:PlaceOrderInput! $promosEnable:Boolean! $wplusEnabled:Boolean! ){placeOrder(input:$placeOrderInput){...ContractFragment}}fragment ContractFragment on PurchaseContract{id associateDiscountStatus addressMode tenderPlanId papEbtAllowed allowedPaymentGroupTypes cartCustomerContext @include(if:$wplusEnabled){isMembershipOptedIn isEligibleForFreeTrial paymentData{hasCreditCard}}checkoutError{code errorData{__typename...on OutOfStock{offerId}__typename...on UnavailableOffer{offerId}__typename...on ItemExpired{offerId}__typename...on ItemQuantityAdjusted{offerId requestedQuantity adjustedQuantity}}operationalErrorCode message}checkoutableWarnings{code itemIds}allocationStatus payments{id paymentType cardType lastFour isDefault cvvRequired preferenceId paymentPreferenceId paymentHandle expiryMonth expiryYear firstName lastName email amountPaid cardImage cardImageAlt isLinkedCard capOneReward{credentialId redemptionUrl redemptionRate redemptionMethod rewardPointsBalance rewardPointsSelected rewardAmountSelected}remainingBalance{displayValue value}}order{id status orderVersion mobileNumber}terms{alcoholAccepted bagFeeAccepted smsOptInAccepted marketingEmailPrefOptIn}donationDetails{charityEIN charityName amount{displayValue value}acceptDonation}lineItems{...LineItemFields}tippingDetails{suggestedAmounts{value displayValue}maxAmount{value displayValue}selectedTippingAmount{value displayValue}}customer{id firstName lastName isGuest email phone}fulfillment{deliveryDetails{deliveryInstructions deliveryOption}pickupChoices{isSelected fulfillmentType accessType accessMode accessPointId}deliveryAddress{...AddressFields}alternatePickupPerson{...PickupPersonFields}primaryPickupPerson{...PickupPersonFields}fulfillmentItemGroups{...FulfillmentItemGroupsFields}accessPoint{...AccessPointFields}reservation{...ReservationFields}}priceDetails{subTotal{...PriceDetailRowFields}totalItemQuantity fees{...PriceDetailRowFields}taxTotal{...PriceDetailRowFields}grandTotal{...PriceDetailRowFields}belowMinimumFee{...PriceDetailRowFields}authorizationAmount{...PriceDetailRowFields}weightDebitTotal{...PriceDetailRowFields}discounts{...PriceDetailRowFields}otcDeliveryBenefit{...PriceDetailRowFields}ebtSnapMaxEligible{...PriceDetailRowFields}ebtCashMaxEligible{...PriceDetailRowFields}hasAmountUnallocated affirm{__typename message{...AffirmMessageFields}}}checkoutGiftingDetails{isCheckoutGiftingOptin isWalmartProtectionPlanPresent isAppleCarePresent isRestrictedPaymentPresent giftMessageDetails{giftingMessage recipientEmail recipientName senderName}}promotions @include(if:$promosEnable){displayValue promoId terms}showPromotions @include(if:$promosEnable) errors{code message lineItems{...LineItemFields}}}fragment LineItemFields on LineItem{id quantity quantityString quantityLabel accessibilityQuantityLabel isPreOrder fulfillmentSourcingDetails{currentSelection requestedSelection}packageQuantity priceInfo{priceDisplayCodes{showItemPrice priceDisplayCondition finalCostByWeight}itemPrice{displayValue value}linePrice{displayValue value}preDiscountedLinePrice{displayValue value}wasPrice{displayValue value}unitPrice{displayValue value}}isSubstitutionSelected isGiftEligible selectedVariants{name value}product{id name usItemId itemType imageInfo{thumbnailUrl}offerId orderLimit orderMinLimit weightIncrement weightUnit averageWeight salesUnitType availabilityStatus isSubstitutionEligible isAlcohol configuration hasSellerBadge sellerId sellerName sellerType preOrder{...preOrderFragment}addOnServices{serviceType groups{groupType services{selectedDisplayName offerId currentPrice{priceString}}}}}discounts{key label displayValue @include(if:$promosEnable) displayLabel @include(if:$promosEnable)}wirelessPlan{planId mobileNumber __typename postPaidPlan{...postpaidPlanDetailsFragment}}selectedAddOnServices{offerId quantity groupType}registryInfo{registryId registryType}}fragment postpaidPlanDetailsFragment on PostPaidPlan{__typename espOrderSummaryId espOrderId espOrderLineId warpOrderId warpSessionId devicePayment{...postpaidPlanPriceFragment}devicePlan{__typename price{...postpaidPlanPriceFragment}frequency duration annualPercentageRate}deviceDataPlan{...deviceDataPlanFragment}}fragment deviceDataPlanFragment on DeviceDataPlan{__typename carrierName planType expiryTime activationFee{...postpaidPlanPriceFragment}planDetails{__typename price{...postpaidPlanPriceFragment}frequency name}agreements{...agreementFragment}}fragment postpaidPlanPriceFragment on PriceDetailRow{__typename key label displayValue value strikeOutDisplayValue strikeOutValue info{__typename title message}}fragment agreementFragment on CarrierAgreement{__typename name type format value docTitle label}fragment preOrderFragment on PreOrder{streetDate streetDateDisplayable streetDateType isPreOrder preOrderMessage preOrderStreetDateMessage}fragment AddressFields on Address{id addressLineOne addressLineTwo city state postalCode firstName lastName phone}fragment PickupPersonFields on PickupPerson{id firstName lastName email}fragment PriceDetailRowFields on PriceDetailRow{__typename key label displayValue value strikeOutValue strikeOutDisplayValue info{__typename title message}}fragment AccessPointFields on AccessPoint{id name assortmentStoreId displayName timeZone address{id addressLineOne addressLineTwo city state postalCode firstName lastName phone}isTest allowBagFee bagFeeValue isExpressEligible fulfillmentOption instructions nodeAccessType}fragment ReservationFields on Reservation{id expiryTime isUnscheduled expired showSlotExpiredError reservedSlot{__typename...on RegularSlot{id price{total{displayValue}expressFee{displayValue}baseFee{displayValue}memberBaseFee{displayValue}}accessPointId fulfillmentOption startTime fulfillmentType slotMetadata slotExpiryTime endTime available supportedTimeZone}...on DynamicExpressSlot{id price{total{displayValue}expressFee{displayValue}baseFee{displayValue}memberBaseFee{displayValue}}accessPointId fulfillmentOption startTime endTime fulfillmentType slotMetadata slotExpiryTime available slaInMins maxItemAllowed supportedTimeZone}...on UnscheduledSlot{price{total{displayValue}expressFee{displayValue}baseFee{displayValue}memberBaseFee{displayValue}}accessPointId fulfillmentOption startTime fulfillmentType slotMetadata unscheduledHoldInDays supportedTimeZone}...on InHomeSlot{id price{total{displayValue}expressFee{displayValue}baseFee{displayValue}memberBaseFee{displayValue}}accessPointId fulfillmentOption startTime fulfillmentType slotMetadata slotExpiryTime endTime available supportedTimeZone}}}fragment AffirmMessageFields on AffirmMessage{__typename description termsUrl imageUrl monthlyPayment termLength isZeroAPR}fragment FulfillmentItemGroupsFields on FulfillmentItemGroup{...on SCGroup{__typename defaultMode collapsedItemIds itemGroups{__typename label itemIds}accessPoint{...AccessPointFields}reservation{...ReservationFields}}...on DigitalDeliveryGroup{__typename defaultMode collapsedItemIds itemGroups{__typename label itemIds}}...on Unscheduled{__typename defaultMode collapsedItemIds itemGroups{__typename label itemIds}accessPoint{...AccessPointFields}reservation{...ReservationFields}}...on FCGroup{__typename defaultMode collapsedItemIds startDate endDate isUnscheduledDeliveryEligible shippingOptions{__typename itemIds availableShippingOptions{__typename id shippingMethod deliveryDate price{__typename displayValue value}label{prefix suffix}isSelected isDefault}}hasMadeShippingChanges slaGroups{__typename label deliveryDate warningLabel sellerGroups{__typename id name isProSeller type shipOptionGroup{__typename deliveryPrice{__typename displayValue value}itemIds shipMethod}}}}...on AutoCareCenter{__typename defaultMode startDate endDate accBasketType collapsedItemIds itemGroups{__typename label itemIds}accessPoint{...AccessPointFields}reservation{...ReservationFields}}}","variables":{"placeOrderInput":{"contractId":"%s","substitutions":[],"acceptBagFee":null,"acceptAlcoholDisclosure":null,"acceptSMSOptInDisclosure":null,"marketingEmailPref":null,"deliveryDetails":{"deliveryInstructions":null,"deliveryOption":"LEAVE_AT_DOOR"},"mobileNumber":"%s","paymentCvvInfos":null,"paymentHandle":null,"acceptDonation":false,"emailAddress":"%s","fulfillmentOptions":null,"acceptedAgreements":[]},"promosEnable":true,"wplusEnabled":true}}`, tk.contractid, tk.Data.Profile.Shipping.PhoneNumber, tk.Data.Profile.Email)))
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt create creditcard request")
		tk.Stop()
		return
	}
	req.Headers = tk.GenerateDefaultHeaders("https://www.walmart.com")
	tk.AddGQLHeaders(req, `mutation PlaceOrder`)

	_, err = tk.Do(req)
	if err != nil {
		tk.SetStatus(module.STATUS_ERROR, "couldnt make creditcard request")
		tk.Stop()
		return
	}
}
