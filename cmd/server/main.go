package main

import (
	"github.com/DimKa163/keeper/app/server"
	"github.com/DimKa163/keeper/internal/server/shared/logging"
	"github.com/caarlos0/env"
)

func main() {
	var config server.Config
	var err error
	if err = env.Parse(&config); err != nil {
		panic(err)
	}
	srv, err := server.NewServer(&config)
	if err != nil {
		panic(err)
	}
	if err = srv.AddLogging(); err != nil {
		panic(err)
	}
	logger := logging.GetLogger()
	if err = srv.AddServices(); err != nil {
		logger.Fatal(err.Error())
	}
	srv.Map()
	if err = srv.Migrate(); err != nil {
		logger.Fatal(err.Error())
	}
	if err = srv.Run(); err != nil {
		logger.Fatal(err.Error())
	}
}
