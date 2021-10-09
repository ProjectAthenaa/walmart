package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ProjectAthenaa/sonic-core/protos/module"
	"github.com/ProjectAthenaa/sonic-core/protos/monitor"
	monitor_controller "github.com/ProjectAthenaa/sonic-core/protos/monitorController"
	"github.com/ProjectAthenaa/sonic-core/sonic/core"
	"github.com/ProjectAthenaa/sonic-core/sonic/database/ent/product"
	"github.com/ProjectAthenaa/walmart/config"
	module2 "github.com/ProjectAthenaa/walmart/module"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"testing"
	"time"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func init() {
	//go debug.StartShapeServer()
	lis = bufconn.Listen(bufSize)
	server := grpc.NewServer()
	module.RegisterModuleServer(server, module2.Server{})
	go func() {
		server.Serve(lis)
	}()
}

func TestModule(t *testing.T) {
	subToken, controlToken, monitorChannel := uuid.NewString(), uuid.NewString(), uuid.NewString()

	productlink := "570251169"

	//username := "LT7LF"
	//password := "H7Y0HUZP"
	//ip := "64.29.84.253"
	//port := "5564"

	ip := "localhost"
	port := "8866"

	core.Base.GetRedis("cache").Publish(context.Background(), fmt.Sprintf("proxies:%s", product.SiteWalmart), fmt.Sprintf(`%s:%s`, ip, port))

	tk := &module.Data{
		TaskID: uuid.NewString(),
		Profile: &module.Profile{
			Email: "terrydavis903@gmail.com",
			Shipping: &module.Shipping{
				FirstName:   "Omar",
				LastName:    "Hu",
				PhoneNumber: "6463222013",
				ShippingAddress: &module.Address{
					AddressLine:  "7004 JFK BLVD E",
					AddressLine2: nil,
					Country:      "US",
					State:        "NEW JERSEY",
					City:         "WEST NEW YORK",
					ZIP:          "07093",
					StateCode:    "NJ",
				},
				BillingAddress: &module.Address{
					AddressLine:  "7004 JFK BLVD E",
					AddressLine2: nil,
					Country:      "US",
					State:        "NEW JERSEY",
					City:         "WEST NEW YORK",
					ZIP:          "07093",
					StateCode:    "NJ",
				},
				BillingIsShipping: true,
			},
			Billing: &module.Billing{
				Number:          "4894537326576479",
				ExpirationMonth: "03",
				ExpirationYear:  "28",
				CVV:             "164",
			},
		},
		Proxy: &module.Proxy{
			//Username: &username,
			//Password: &password,
			IP:       ip,
			Port:     port,
		},
		TaskData: &module.TaskData{
			RandomSize:  false,
			RandomColor: false,
			Color:       []string{"1"},
			Size:        []string{"1"},
			Link:        &productlink,
		},
		Metadata: map[string]string{

			*config.Module.Fields[0].FieldKey: productlink,
		},
		Channels: &module.Channels{
			UpdatesChannel:  subToken,
			CommandsChannel: controlToken,
			MonitorChannel: monitorChannel,
		},
	}

	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	client := module.NewModuleClient(conn)
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Second*5))

	conn, err = grpc.Dial("localhost:4000", grpc.WithInsecure())
	monitorClient := monitor.NewMonitorClient(conn)
	monitorClient.Start(context.Background(), &monitor_controller.Task{
		Site:         string(product.SiteWalmart),
		Lookup:       &monitor_controller.Task_Other{Other: true},
		RedisChannel: monitorChannel,
		Metadata:     tk.Metadata,
	})

	t.Log("connecting to redis")
	pubsub := core.Base.GetRedis("cache").Subscribe(ctx, fmt.Sprintf("tasks:updates:%s", subToken))
	t.Log("connected to rediss")

	_, err = client.Task(context.Background(), tk)
	if err != nil {
		t.Fatal(err)
	}

	var start time.Time

	for msg := range pubsub.Channel() {
		var data module.Status
		_ = json.Unmarshal([]byte(msg.Payload), &data)
		t.Log(data.Status, data.Information["message"])

		if data.Status == module.STATUS_PRODUCT_FOUND {
			start = time.Now()
		}

		if data.Status == module.STATUS_CHECKED_OUT {
			t.Log(msg.Payload)
		}

		if data.Status == module.STATUS_STOPPED {
			t.Log(time.Since(start))
			return
		}
	}
}
