package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/daax-dev/nanofuse/internal/trayapp"
)

var (
	version   = "0.1.0"
	commit    = "dev"
	buildDate = "unknown"
)

func main() {
	cfg := trayapp.ConfigFromEnv()
	smoke := flag.Bool("smoke", false, "run API smoke check and exit without starting the tray UI")
	launchImage := flag.String("launch-image", "", "create and start a VM from this image through the tray launch path, then exit")
	showVersion := flag.Bool("version", false, "show version information")
	apiURL := flag.String("api-url", cfg.APIURL, "nanofused API URL, for example http://127.0.0.1:18080")
	apiSocket := flag.String("api-socket", cfg.APISocket, "nanofused Unix socket path")
	timeout := flag.Duration("timeout", cfg.Timeout, "API request timeout")
	debug := flag.Bool("debug", cfg.Debug, "enable debug API client logging")
	flag.Parse()

	if *showVersion {
		fmt.Printf("nanofuse-tray %s commit=%s built_at=%s\n", version, commit, buildDate)
		return
	}

	cfg = trayapp.Config{
		APIURL:    *apiURL,
		APISocket: *apiSocket,
		Timeout:   *timeout,
		Debug:     *debug,
	}.Normalize()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if *smoke {
		if err := runSmoke(ctx, cfg); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *launchImage != "" {
		if err := runLaunchImage(ctx, cfg, *launchImage); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	if err := runTray(ctx, cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runSmoke(ctx context.Context, cfg trayapp.Config) error {
	api := cfg.NewClient()
	checkCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	status, err := trayapp.CollectStatus(checkCtx, api, cfg.Endpoint())
	encoded, encodeErr := json.MarshalIndent(status, "", "  ")
	if encodeErr != nil {
		return fmt.Errorf("encode status: %w", encodeErr)
	}
	fmt.Println(string(encoded))
	if err != nil {
		return err
	}
	return nil
}

func runLaunchImage(ctx context.Context, cfg trayapp.Config, imageRef string) error {
	api := cfg.NewClient()
	launchCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	vm, err := trayapp.LaunchVMFromImage(launchCtx, api, imageRef)
	if err != nil {
		return err
	}

	encoded, err := json.MarshalIndent(vm, "", "  ")
	if err != nil {
		return fmt.Errorf("encode VM: %w", err)
	}
	fmt.Println(string(encoded))
	return nil
}
