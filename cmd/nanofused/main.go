package main

import (
	"flag"
	"log"

	"github.com/daax-dev/nanofuse/internal/api"
)

func main() {
	configPath := flag.String("config", "/etc/nanofuse/nanofused.yaml", "Path to configuration file")
	tcpBind := flag.String("tcp", "", "Enable TCP mode and bind to address (e.g., 0.0.0.0:8080 or :8080)")
	unixSocket := flag.String("socket", "", "Unix socket path (default: from config or /var/run/nanofused.sock)")
	flag.Parse()

	// Override config with CLI flags if provided
	if *tcpBind != "" || *unixSocket != "" {
		if err := api.StartWithOverrides(*configPath, *tcpBind, *unixSocket); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := api.Start(*configPath); err != nil {
			log.Fatal(err)
		}
	}
}
