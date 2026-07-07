package main

import (
	"GoLiteConfig/internal/handler"
	"GoLiteConfig/internal/logging"
	"GoLiteConfig/internal/service"
	"context"
	"log"

	"GoLiteConfig/internal/config"
	"GoLiteConfig/internal/etcd"
	"GoLiteConfig/internal/router"
)

func main() {
	cfg := config.Load()

	loggerBundle, err := logging.New()
	if err != nil {
		log.Fatalf("init logger failed: %v", err)
	}
	defer loggerBundle.Sync()

	etcdClient, err := etcd.NewClient(cfg.EtcdEndpoints)
	if err != nil {
		log.Fatalf("init etcd client failed: %v", err)
	}
	defer etcdClient.Close()

	if err := etcdClient.Ping(context.Background()); err != nil {
		log.Fatalf("connect etcd failed: %v", err)
	}
	loggerBundle.App().Info("connected to etcd")

	watchMgr := service.NewWatchManager()
	auditLogger := service.NewAuditLogger(loggerBundle.Audit())
	configService := service.NewConfigService(etcdClient, watchMgr, loggerBundle.App(), auditLogger)
	configHandler := handler.NewConfigHandler(configService)

	r := router.SetupRouter(configHandler, loggerBundle.App())
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("start http server failed: %v", err)
	}
}
