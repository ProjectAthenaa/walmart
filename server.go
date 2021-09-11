package main

import (
	"github.com/ProjectAthenaa/sonic-core/protos/module"
	"github.com/ProjectAthenaa/sonic-core/sonic"
	"github.com/ProjectAthenaa/walmart/config"
	moduleServer "github.com/ProjectAthenaa/walmart/module"
	"github.com/prometheus/common/log"
	"google.golang.org/grpc"
	"net"
)

func init() {
	if err := sonic.RegisterModule(config.Module); err != nil {
		panic(err)
	}
}

func main() {
	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatalln("start listener: ", err)
	}

	server := grpc.NewServer()

	module.RegisterModuleServer(server, moduleServer.Server{})

	log.Info("Walmart Module Initialized")
	if err = server.Serve(listener); err != nil {
		log.Fatalln("start server: ", err)
	}
}
