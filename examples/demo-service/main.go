package main

import (
	"GoLiteConfig/pkg/sdk"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	client, err := sdk.NewClient(sdk.Config{
		ServerAddr: "http://127.0.0.1:8080",
		App:        "order-service",
		Env:        "prod",
	})

	if err != nil {
		log.Fatalf("init sdk failed: %v", err)
	}
	defer client.Close()

	keys := []string{
		"database.host",
		"database.port",
		"redis.addr",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = client.Watch(ctx, func(event sdk.ConfigChangeEvent) {
		fmt.Printf("[hot-reload] revision=%d version=%s\n", event.Revision, event.Version)
		printConfigChanges(event, keys)
		fmt.Println()
	})
	if err != nil {
		log.Fatalf("start watch failed: %v", err)
	}

	printCurrentConfig(client, keys)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Println("demo-service is running, waiting for config changes...")
	<-sigCh
	log.Println("demo-service shutting down...")
}

func printCurrentConfig(client *sdk.Client, keys []string) {
	fmt.Println("[demo] current config:")
	for _, key := range keys {
		value, err := client.Get(key)
		if err != nil {
			fmt.Printf("  %s = <not found>\n", key)
			continue
		}
		fmt.Printf("  %s = %s\n", key, value)
	}
	fmt.Println()
}

func printConfigChanges(event sdk.ConfigChangeEvent, keys []string) {
	changed := false
	for _, key := range keys {
		oldValue := event.PrevConfig[key]
		newValue := event.Configs[key]

		if oldValue == newValue {
			continue
		}

		changed = true
		fmt.Printf("  %s: %s -> %s\n", key, oldValue, newValue)
	}

	if !changed {
		fmt.Println("  no tracked keys changed")
	}
}
