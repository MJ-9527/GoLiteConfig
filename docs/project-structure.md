# 项目目录设计说明

本文档说明 GoLiteConfig 的项目目录结构，作为 Day1 项目骨架设计依据。

## 1. 目录结构

```text
GoLiteConfig/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── handler/
│   ├── model/
│   ├── router/
│   ├── service/
│   └── etcd/
├── pkg/
│   └── sdk/
├── deploy/
│   └── docker-compose.yml
├── docs/
├── examples/
│   └── demo-service/
├── scripts/
├── go.mod
└── README.md
```

## 2. 目录职责

| 目录 | 职责 |
| --- | --- |
| `cmd/server` | 服务端启动入口，负责初始化路由、配置和依赖 |
| `internal/handler` | HTTP Handler 层，负责解析请求和返回响应 |
| `internal/model` | 请求、响应、元数据等结构体定义 |
| `internal/router` | Gin 路由注册 |
| `internal/service` | 业务逻辑层，例如发布配置、查询配置、回滚配置 |
| `internal/etcd` | Etcd clientv3 封装，例如 Get、Put、GetPrefix、Watch |
| `pkg/sdk` | 对外提供的 Go 客户端 SDK |
| `deploy` | Docker Compose 和部署相关文件 |
| `docs` | 需求、设计、API、任务拆分等文档 |
| `examples/demo-service` | SDK 热更新演示服务 |
| `scripts` | 本地启动、测试、演示脚本 |

## 3. 分层原则

1. `handler` 只处理 HTTP 入参和出参，不直接操作 Etcd。
2. `service` 负责核心业务规则，例如版本号生成、current 更新和回滚流程。
3. `etcd` 只封装 Etcd 读写细节，不关心 HTTP 和业务语义。
4. `model` 存放跨层共享的数据结构。
5. `pkg/sdk` 是给业务服务使用的客户端包，不能依赖 `internal` 目录。

## 4. Day1 结论

Day1 已确定项目采用 Go 标准工程结构：

1. 服务端入口放在 `cmd/server`。
2. 服务端内部实现放在 `internal`。
3. 对外 SDK 放在 `pkg/sdk`。
4. 部署文件放在 `deploy`。
5. 文档和设计放在 `docs`。
6. 演示服务放在 `examples/demo-service`。
