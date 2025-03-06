package main

import (
	"flag"
	"log"
	"os"

	"github.com/luraproject/lura/config"
	"github.com/luraproject/lura/logging"
	"github.com/luraproject/lura/proxy"
	"github.com/luraproject/lura/router/gin"
)

func main() {
	port := flag.Int("p", 0, "Port of the service")
	logLevel := flag.String("l", "ERROR", "Logging level")
	debug := flag.Bool("d", false, "Enable debug mode")
	configFile := flag.String("c", "../../krakend.json", "Path to the configuration file")
	flag.Parse()

	parser := config.NewParser()
	serviceConfig, err := parser.Parse(*configFile)
	if err != nil {
		log.Fatalf("ERROR parsing config: %s", err.Error())
	}
	serviceConfig.Debug = serviceConfig.Debug || *debug
	if *port != 0 {
		serviceConfig.Port = *port
	}

	logger, _ := logging.NewLogger(*logLevel, os.Stdout, "[LURA]")

	routerFactory := gin.DefaultFactory(proxy.DefaultFactory(logger), logger)

	routerFactory.New().Run(serviceConfig)
}
