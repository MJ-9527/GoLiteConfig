package main

import (
	"GoLiteConfig/internal/handler"
	"GoLiteConfig/internal/service"
	"context"
	"log"

	"GoLiteConfig/internal/config"
	"GoLiteConfig/internal/etcd"
	"GoLiteConfig/internal/router"
)

func main() {
	cfg := config.Load()

	etcdClient, err := etcd.NewClient(cfg.EtcdEndpoints)
	if err != nil {
		log.Fatalf("init etcd client failed: %v", err)
	}
	defer etcdClient.Close()

	if err := etcdClient.Ping(context.Background()); err != nil {
		log.Fatalf("connect etcd failed: %v", err)
	}
	log.Printf("connected to etcd: %v", cfg.EtcdEndpoints)

	watchMgr := service.NewWatchManager()
	configService := service.NewConfigService(etcdClient, watchMgr)
	configHandler := handler.NewConfigHandler(configService)

	r := router.SetupRouter(configHandler)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("start http server failed: %v", err)
	}
}
