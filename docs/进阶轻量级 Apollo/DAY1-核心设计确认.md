# Day1 核心设计确认

## 1. Day1 目标

Day1 的目标不是直接改代码，而是先把“进阶轻量级 Apollo”最关键的基础问题定下来：

1. 为什么要引入 namespace
2. 新的 Etcd Key 结构怎么设计
3. 发布、查询、版本、回滚怎么映射到新结构
4. 与当前 `GoLiteConfig` 现有结构如何衔接

## 2. 为什么引入 namespace

当前项目使用的是 `app + env` 二维模型，已经能满足单份配置的发布与查询，但当配置内容逐渐增多后，会出现几个问题：

1. 所有配置都混在一套版本里，不利于按模块管理。
2. 数据库、Redis、业务开关等配置难以分组演进。
3. SDK 无法只订阅某一类配置。
4. 后续控制台展示也不够清晰。

因此本阶段引入 `namespace`，将模型升级为：

```text
app + env + namespace
```

典型示例：

- `order-service / prod / application`
- `order-service / prod / database`
- `order-service / prod / redis`

## 3. Day1 设计结论

Day1 先确定以下结论：

1. `namespace` 是本阶段必须引入的核心概念。
2. 每个 namespace 独立维护自己的 `current`、`version_counter`、`versions`、`meta`。
3. 发布、查询、版本、回滚都以 namespace 为最小操作单元。
4. watch 机制后续也要升级到 namespace 维度。
5. SDK 未来支持按 namespace 拉取和监听。

## 4. 对现有项目的影响范围

本轮设计会直接影响以下模块：

- `internal/model`
- `internal/service/config_service.go`
- `internal/handler/config.go`
- `internal/service/watch_manager.go`
- `pkg/sdk`
- `docs/api.md`
- `docs/key-design.md`

## 5. Day1 之后的执行顺序

Day1 完成后，建议按以下顺序推进：

1. 先改 model 和 service 方法签名
2. 再改发布与查询接口
3. 然后改版本和回滚
4. 最后改 watch 与 SDK

## 6. Day1 验收标准

Day1 结束时应达到：

1. 新模型已文档化
2. Key 结构已定稿
3. API 草案已明确
4. 后续代码改造有清晰路径
