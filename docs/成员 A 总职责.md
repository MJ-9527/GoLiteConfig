下面只讲 **成员 A**，也就是这个项目的**后端核心负责人**。A 的目标不是“每天写点代码”，而是每天都要推进一条主链路：
**配置发布 -> Etcd 存储 -> 配置查询 -> 版本管理 -> Watch -> SDK -> Demo 热更新**。

------

## 成员 A 总职责

A 负责项目最核心的代码：

```
cmd/server/main.go
internal/config/
internal/model/
internal/etcd/
internal/service/
internal/handler/
internal/router/
internal/watch/
pkg/sdk/
examples/demo-service/
```

A 的最终成果是：

```
1. 服务端能启动
2. 配置能发布到 Etcd
3. 配置能查询
4. 每次发布都有版本
5. 能回滚到历史版本
6. 客户端能长轮询监听变化
7. SDK 能缓存配置并自动热更新
8. demo-service 能展示“不重启更新配置”
```

------

# 第 1 天：确定技术骨架和核心设计

## A 要干什么

A 今天不急着写业务代码，重点是把项目骨架和技术决策定下来。

要完成：

```
1. 确认 Go module 名称
2. 确认服务端目录结构
3. 确认 Etcd Key 设计
4. 确认 API 请求/响应格式
5. 确认版本生成规则
6. 确认配置数据模型
```

## A 要写什么代码

今天可以只写少量骨架代码，甚至只写空结构体和注释也行。

建议先规划这些包：

```
internal/model
internal/etcd
internal/service
internal/handler
internal/router
internal/watch
```

其中 `model` 里后续要放这些结构：

```
type PublishConfigRequest struct {
    App       string            `json:"app"`
    Env       string            `json:"env"`
    Configs   map[string]string `json:"configs"`
    Publisher string            `json:"publisher"`
    Comment   string            `json:"comment"`
}

type ConfigMeta struct {
    Version   string `json:"version"`
    Revision  int64  `json:"revision"`
    Publisher string `json:"publisher"`
    Comment   string `json:"comment"`
    CreatedAt string `json:"created_at"`
}

type APIResponse struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}
```

## A 要确定的关键设计

推荐 Etcd Key 设计：

```
/config/{app}/{env}/current
/config/{app}/{env}/version_counter
/config/{app}/{env}/versions/{version}/{group}/{key}
/config/{app}/{env}/meta/{version}
```

比如：

```
/config/order-service/prod/current = v2
/config/order-service/prod/version_counter = 2
/config/order-service/prod/versions/v2/database/host = 127.0.0.1
/config/order-service/prod/meta/v2 = {"version":"v2",...}
```

## 当天成果

第 1 天结束时，A 应该能说清楚：

```
1. 配置如何存进 Etcd
2. 当前版本如何定位
3. 历史版本如何保留
4. 回滚为什么要生成新版本
5. handler/service/etcd/watch 各自负责什么
```

------

# 第 2 天：搭建服务端基础骨架

## A 要干什么

今天目标是让服务端真正跑起来。

要完成：

```
1. 初始化 Gin
2. 实现 GET /health
3. 初始化 Etcd client
4. 建立 router
5. 建立统一响应函数
6. 服务端能监听 8080 端口
```

## A 要写什么代码

### 1. `cmd/server/main.go`

负责启动服务：

```
func main() {
    // 1. 读取配置
    // 2. 初始化 Etcd client
    // 3. 初始化 service
    // 4. 初始化 handler
    // 5. 注册 router
    // 6. 启动 HTTP server
}
```

### 2. `internal/config`

定义服务自身配置：

```
type Config struct {
    Port          string
    EtcdEndpoints []string
}
```

默认值：

```
Port = 8080
EtcdEndpoints = ["localhost:2379"]
```

### 3. `internal/etcd`

封装 Etcd client：

```
type Client struct {
    cli *clientv3.Client
}

func NewClient(endpoints []string) (*Client, error)
func (c *Client) Close() error
func (c *Client) Ping(ctx context.Context) error
```

### 4. `internal/router`

注册路由：

```
func NewRouter(h *handler.Handler) *gin.Engine {
    r := gin.Default()
    r.GET("/health", h.Health)
    return r
}
```

### 5. `internal/handler`

实现健康检查：

```
func (h *Handler) Health(c *gin.Context) {
    c.JSON(200, model.APIResponse{
        Code: 0,
        Message: "success",
        Data: gin.H{"status": "ok"},
    })
}
```

## 当天成果

A 今天必须做到：

```
go run cmd/server/main.go
```

然后：

```
curl http://localhost:8080/health
```

返回：

```
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ok"
  }
}
```

## 审核标准

```
1. 服务能启动。
2. /health 返回 200。
3. Etcd 连接失败时，日志能看出原因。
4. 代码已经拆到 config/router/handler/etcd，不是全堆 main.go。
```

------

# 第 3 天：实现配置发布 POST /api/config

## A 要干什么

今天开始做核心功能：发布配置。

要完成：

```
1. 接收 POST /api/config 请求
2. 校验 app、env、configs
3. 生成新版本号
4. 把 configs 写入 Etcd
5. 写入 meta
6. 更新 current
7. 返回 version 和 revision
```

## A 要写什么代码

### 1. `internal/model`

新增请求结构：

```
type PublishConfigRequest struct {
    App       string            `json:"app"`
    Env       string            `json:"env"`
    Configs   map[string]string `json:"configs"`
    Publisher string            `json:"publisher"`
    Comment   string            `json:"comment"`
}
```

新增响应结构：

```
type PublishConfigResponse struct {
    App      string `json:"app"`
    Env      string `json:"env"`
    Version  string `json:"version"`
    Revision int64  `json:"revision"`
}
```

### 2. `internal/etcd`

新增基础方法：

```
func (c *Client) Get(ctx context.Context, key string) (string, int64, error)
func (c *Client) Put(ctx context.Context, key, value string) (int64, error)
func (c *Client) GetPrefix(ctx context.Context, prefix string) (map[string]string, int64, error)
```

### 3. `internal/service`

新增核心方法：

```
func (s *ConfigService) Publish(ctx context.Context, req model.PublishConfigRequest) (*model.PublishConfigResponse, error)
```

内部逻辑：

```
1. 校验 app/env/configs
2. 读取 version_counter
3. 生成 nextVersion，比如 v1
4. 遍历 configs
5. database.host 拆成 group=database, key=host
6. 写入 /versions/v1/database/host
7. 写入 /meta/v1
8. 更新 /current = v1
9. 返回 revision
```

### 4. `internal/handler`

新增：

```
func (h *Handler) PublishConfig(c *gin.Context)
```

负责：

```
1. BindJSON
2. 调用 service.Publish
3. 返回统一 JSON
```

### 5. `internal/router`

注册：

```
api := r.Group("/api")
api.POST("/config", h.PublishConfig)
```

## 当天成果

A 要能通过 curl 发布配置：

```
curl -X POST http://localhost:8080/api/config ^
  -H "Content-Type: application/json" ^
  -d "{\"app\":\"order-service\",\"env\":\"prod\",\"configs\":{\"database.host\":\"127.0.0.1\",\"database.port\":\"3306\"},\"publisher\":\"admin\",\"comment\":\"init\"}"
```

返回：

```
{
  "code": 0,
  "message": "success",
  "data": {
    "app": "order-service",
    "env": "prod",
    "version": "v1",
    "revision": 12
  }
}
```

## 审核标准

```
1. 缺 app 返回 400。
2. 缺 env 返回 400。
3. configs 为空返回 400。
4. 正常发布能写入 Etcd。
5. current 被更新为新版本。
6. meta 被写入。
```

------

# 第 4 天：实现配置查询 GET /api/config

## A 要干什么

今天目标是能查当前配置。

要完成：

```
1. 根据 app/env 查询 current
2. 根据 current 拼接版本前缀
3. 从 Etcd 获取该版本所有配置
4. 转换成 database.host 这种平铺格式
5. 返回 configs
```

## A 要写什么代码

### 1. `internal/model`

新增：

```
type GetConfigResponse struct {
    App      string            `json:"app"`
    Env      string            `json:"env"`
    Version  string            `json:"version"`
    Revision int64             `json:"revision"`
    Configs  map[string]string `json:"configs"`
}
```

### 2. `internal/service`

新增：

```
func (s *ConfigService) GetCurrent(ctx context.Context, app, env string) (*model.GetConfigResponse, error)
```

逻辑：

```
1. 读 /config/{app}/{env}/current
2. 得到 version
3. 读 /config/{app}/{env}/versions/{version}/ 前缀
4. 解析 key，转成 group.key
5. 返回结果
```

### 3. `internal/handler`

新增：

```
func (h *Handler) GetConfig(c *gin.Context)
```

从 query 获取：

```
app
env
group 可选
version 可选
```

MVP 先支持 app/env 即可。

### 4. `internal/router`

注册：

```
api.GET("/config", h.GetConfig)
```

## 当天成果

发布后能查询：

```
curl "http://localhost:8080/api/config?app=order-service&env=prod"
```

返回：

```
{
  "code": 0,
  "message": "success",
  "data": {
    "app": "order-service",
    "env": "prod",
    "version": "v1",
    "configs": {
      "database.host": "127.0.0.1",
      "database.port": "3306"
    }
  }
}
```

## 审核标准

```
1. 发布什么，查询就能查到什么。
2. 发布 v2 后，查询返回 v2。
3. 查询不存在的 app/env 返回 404 或业务错误码。
4. 返回结果不暴露 Etcd 内部完整路径。
```

------

# 第 5 天：实现版本列表 GET /api/config/versions

## A 要干什么

今天做版本元数据查询。

要完成：

```
1. 查询 current
2. 查询 meta 前缀
3. 解析每个版本的元数据
4. 返回版本列表
```

## A 要写什么代码

### 1. `internal/model`

新增：

```
type VersionInfo struct {
    Version   string `json:"version"`
    Revision  int64  `json:"revision"`
    Publisher string `json:"publisher"`
    Comment   string `json:"comment"`
    CreatedAt string `json:"created_at"`
}

type ListVersionsResponse struct {
    App      string        `json:"app"`
    Env      string        `json:"env"`
    Current  string        `json:"current"`
    Versions []VersionInfo `json:"versions"`
}
```

### 2. `internal/service`

新增：

```
func (s *ConfigService) ListVersions(ctx context.Context, app, env string) (*model.ListVersionsResponse, error)
```

逻辑：

```
1. 读 current
2. GetPrefix /config/{app}/{env}/meta/
3. JSON 反序列化 meta
4. 排序
5. 返回
```

### 3. `internal/handler`

新增：

```
func (h *Handler) ListVersions(c *gin.Context)
```

### 4. `internal/router`

注册：

```
api.GET("/config/versions", h.ListVersions)
```

## 当天成果

连续发布 v1、v2、v3 后：

```
curl "http://localhost:8080/api/config/versions?app=order-service&env=prod"
```

返回：

```
{
  "code": 0,
  "message": "success",
  "data": {
    "current": "v3",
    "versions": [
      {"version": "v1", "comment": "init"},
      {"version": "v2", "comment": "change db host"},
      {"version": "v3", "comment": "change redis"}
    ]
  }
}
```

## 审核标准

```
1. current 与最新发布版本一致。
2. 每个版本都有 meta。
3. meta 包含 version、revision、publisher、comment、created_at。
4. 列表顺序稳定。
```

------

# 第 6 天：实现回滚 POST /api/config/rollback

## A 要干什么

今天完成项目非常重要的演示点：版本回滚。

要完成：

```
1. 接收 target_version
2. 查询目标版本配置
3. 基于目标版本生成一个新版本
4. 写入新版本配置
5. 写入 rollback meta
6. 更新 current
```

## A 要写什么代码

### 1. `internal/model`

新增：

```
type RollbackRequest struct {
    App           string `json:"app"`
    Env           string `json:"env"`
    TargetVersion string `json:"target_version"`
    Publisher     string `json:"publisher"`
    Comment       string `json:"comment"`
}

type RollbackResponse struct {
    FromVersion   string `json:"from_version"`
    TargetVersion string `json:"target_version"`
    NewVersion    string `json:"new_version"`
    Revision      int64  `json:"revision"`
}
```

### 2. `internal/service`

新增：

```
func (s *ConfigService) Rollback(ctx context.Context, req model.RollbackRequest) (*model.RollbackResponse, error)
```

逻辑：

```
1. 读 current，记为 fromVersion
2. 读 target_version 下所有 configs
3. 如果 target_version 不存在，返回错误
4. 生成 nextVersion
5. 把 target configs 写入 nextVersion
6. meta comment 标记 rollback
7. current = nextVersion
8. 返回 from/target/new
```

### 3. `internal/handler`

新增：

```
func (h *Handler) Rollback(c *gin.Context)
```

### 4. `internal/router`

注册：

```
api.POST("/config/rollback", h.Rollback)
```

## 当天成果

演示：

```
发布 v1：database.host=127.0.0.1
发布 v2：database.host=192.168.1.10
回滚到 v1
查询当前配置：database.host=127.0.0.1
版本列表：current=v3
```

## 审核标准

```
1. 回滚不会删除 v1/v2。
2. 回滚不是简单 current=v1，而是生成新版本 v3。
3. 回滚后的当前配置与目标版本完全一致。
4. 目标版本不存在时返回错误。
```

------

# 第 7 天：实现长轮询 GET /api/watch

## A 要干什么

今天做客户端热更新的基础。

要完成：

```
1. 客户端传 last_revision
2. 如果服务端当前 revision 更大，立即返回最新配置
3. 如果没有变化，挂起等待
4. 发布新配置后唤醒等待请求
5. 超时无变化返回 304
```

## A 要写什么代码

### 1. `internal/watch`

定义等待管理器：

```
type Manager struct {
    mu      sync.Mutex
    waiters map[string][]chan struct{}
}
```

key 可以是：

```
app + "/" + env
```

方法：

```
func (m *Manager) Wait(ctx context.Context, app, env string) <-chan struct{}
func (m *Manager) Notify(app, env string)
```

### 2. `internal/service`

发布成功后调用：

```
watchManager.Notify(app, env)
```

或者 A 也可以用 Etcd Watch 监听 `/config/` 前缀，再通知等待者。MVP 推荐先在发布成功后主动 Notify，简单稳定。

### 3. `internal/handler`

新增：

```
func (h *Handler) WatchConfig(c *gin.Context)
```

逻辑：

```
1. 读取 app/env/last_revision
2. 获取当前配置 revision
3. 如果 currentRevision > lastRevision，立即返回
4. 否则等待 notify 或 60s timeout
5. notify 后再次查询并返回最新配置
6. timeout 返回 304
```

### 4. `internal/router`

注册：

```
api.GET("/watch", h.WatchConfig)
```

## 当天成果

终端 1：

```
curl "http://localhost:8080/api/watch?app=order-service&env=prod&last_revision=12"
```

终端 2 发布新配置后，终端 1 立即返回。

## 审核标准

```
1. last_revision 旧时立即返回 200。
2. last_revision 当前时挂起等待。
3. 发布新配置后能唤醒。
4. 60 秒无变化返回 304。
5. 多个客户端同时 watch 都能被唤醒。
```

------

# 第 8 天：实现 Go SDK

## A 要干什么

今天把服务端能力封装给业务服务使用。

要完成：

```
1. SDK 初始化时拉取配置
2. 本地缓存配置
3. 提供 Get 方法
4. 后台循环调用 /api/watch
5. 配置变化时更新缓存
6. 触发回调
```

## A 要写什么代码

位置：

```
pkg/sdk/
```

### 1. SDK 配置结构

```
type Options struct {
    ServerAddr string
    App        string
    Env        string
}
```

### 2. Client 结构

```
type Client struct {
    serverAddr string
    app        string
    env        string

    mu       sync.RWMutex
    cache    map[string]string
    revision int64

    httpClient *http.Client
}
```

### 3. 构造函数

```
func NewClient(opts Options) (*Client, error)
```

逻辑：

```
1. 参数校验
2. 调 GET /api/config
3. 初始化 cache
4. 保存 revision
```

### 4. 读取方法

```
func (c *Client) Get(key string) (string, error)
```

只读本地缓存。

### 5. Watch 方法

```
func (c *Client) Watch(ctx context.Context, onChange func(ConfigChangeEvent)) error
```

循环：

```
1. 请求 /api/watch
2. 200：更新 cache，触发回调
3. 304：继续下一轮
4. 错误：sleep 后重试
```

### 6. 事件结构

```
type ConfigChangeEvent struct {
    Version string
    Changes map[string]ConfigValueChange
}

type ConfigValueChange struct {
    OldValue string
    NewValue string
}
```

## 当天成果

A 可以写一个最小临时测试 main：

```
client, _ := sdk.NewClient(sdk.Options{
    ServerAddr: "http://localhost:8080",
    App: "order-service",
    Env: "prod",
})

v, _ := client.Get("database.host")
fmt.Println(v)
```

## 审核标准

```
1. SDK 初始化能成功拉取当前配置。
2. Get 不依赖实时请求，只读缓存。
3. Watch 收到变化能更新缓存。
4. 配置中心短暂不可用时，Get 仍然可用。
5. Watch 错误不会导致业务进程退出。
```

------

# 第 9 天：实现 demo-service 热更新演示

## A 要干什么

今天让项目有最终展示效果。

要完成：

```
1. 在 examples/demo-service 中接入 SDK
2. 启动时打印当前配置
3. 注册 Watch 回调
4. 配置变化时打印 old -> new
5. 主进程保持运行
```

## A 要写什么代码

位置：

```
examples/demo-service/main.go
```

逻辑：

```
func main() {
    client, err := sdk.NewClient(...)
    if err != nil {
        panic(err)
    }

    printCurrentConfig(client)

    ctx := context.Background()
    client.Watch(ctx, func(event sdk.ConfigChangeEvent) {
        fmt.Println("[hot-reload]", event)
    })

    select {}
}
```

可以先重点监听：

```
database.host
database.port
redis.addr
```

输出建议：

```
[demo] loaded config:
database.host = 127.0.0.1
database.port = 3306

[hot-reload] database.host: 127.0.0.1 -> 192.168.1.10
```

## 当天成果

完整演示链路：

```
1. 启动 Etcd
2. 启动 server
3. 发布 v1
4. 启动 demo-service，看到 v1
5. 发布 v2
6. demo-service 自动打印变化
7. rollback v1
8. demo-service 自动打印恢复
```

## 审核标准

```
1. demo-service 不需要重启。
2. 配置变化 5 秒内可见。
3. 回滚也能触发热更新。
4. 日志清晰，适合答辩展示。
```

------

# 第 10 天：补齐删除、异常处理和稳定性

## A 要干什么

今天不是猛加新功能，而是让主链路稳定。

优先做：

```
1. 修复发布/查询/回滚/watch 的 bug
2. 补充 DELETE /api/config
3. 统一错误处理
4. 检查并发 watch
5. 检查服务关闭资源释放
```

## A 要写什么代码

### 1. 删除接口

建议删除也生成新版本：

```
type DeleteConfigRequest struct {
    App       string   `json:"app"`
    Env       string   `json:"env"`
    Keys      []string `json:"keys"`
    Publisher string   `json:"publisher"`
    Comment   string   `json:"comment"`
}
```

逻辑：

```
1. 读取当前版本所有配置
2. 删除指定 keys
3. 发布成新版本
4. 更新 current
```

### 2. 统一错误码

例如：

```
40001 参数错误
40401 配置不存在
50001 Etcd 错误
50002 内部错误
```

### 3. Watch 稳定性

检查：

```
1. 请求取消后 waiter 是否移除
2. 超时后 waiter 是否移除
3. Notify 后 channel 是否关闭
4. 是否可能重复 close channel
```

## 当天成果

完整流程至少连续跑 3 次：

```
发布 -> 查询 -> watch -> 发布新版本 -> demo 热更新 -> 回滚 -> demo 恢复
```

## 审核标准

```
1. 主流程连续 3 次不失败。
2. 错误响应格式统一。
3. 不会因为异常参数 panic。
4. Watch 不会明显泄露 goroutine。
```

------

# 第 11 天：最终技术收尾和答辩准备

## A 要干什么

今天 A 主要准备交付和答辩技术说明。

要完成：

```
1. 最终检查核心代码
2. 整理架构讲解
3. 整理核心流程图
4. 准备可能被问到的问题
5. 协助 B/C 完成最终演示
```

## A 要准备讲清楚的问题

### 1. 为什么用 Etcd？

回答要点：

```
Etcd 是分布式 KV 存储，支持 Watch、revision、前缀查询，适合做配置中心底层存储。
```

### 2. 配置如何存储？

说明：

```
current 保存当前版本
versions 保存历史版本配置
meta 保存版本元数据
```

### 3. 为什么回滚要生成新版本？

回答：

```
为了保留完整审计链路。回滚也是一次新的发布行为，不能直接把 current 指回旧版本就结束。
```

### 4. 热更新怎么实现？

回答：

```
SDK 启动时先拉取配置并缓存，然后后台长轮询 /api/watch。
服务端配置变化后唤醒 watch 请求，SDK 收到新配置后更新本地缓存并触发回调。
```

### 5. 配置中心挂了业务会怎样？

回答：

```
SDK Get 读取的是本地缓存。配置中心短暂不可用时，业务服务仍然可以使用最后一次成功拉取的配置。
```

## 当天成果

A 最终应保证：

```
1. 所有核心接口能跑。
2. demo 热更新能展示。
3. 回滚能展示。
4. 代码结构能讲清楚。
5. 被问到核心设计时能回答。
```

------

# A 每日最小验收表

| 天数     | A 当天必须交付              | 最小通过标准                |
| -------- | --------------------------- | --------------------------- |
| 第 1 天  | 架构、Key、API 基础设计     | 能讲清楚存储和版本规则      |
| 第 2 天  | 服务端骨架 + `/health`      | 服务能启动，health 返回 200 |
| 第 3 天  | `POST /api/config`          | 能发布配置到 Etcd           |
| 第 4 天  | `GET /api/config`           | 能查询当前配置              |
| 第 5 天  | `GET /api/config/versions`  | 能看到版本列表              |
| 第 6 天  | `POST /api/config/rollback` | 能回滚并生成新版本          |
| 第 7 天  | `GET /api/watch`            | 配置变化能唤醒客户端        |
| 第 8 天  | Go SDK                      | SDK 能拉取、缓存、监听      |
| 第 9 天  | demo-service                | 不重启完成热更新            |
| 第 10 天 | 稳定性和 DELETE             | 主流程连续跑通              |
| 第 11 天 | 技术收尾                    | 能演示、能讲清楚            |

------

# A 的代码优先级

如果时间不够，A 必须保这些：

```
P0 必做：
1. /health
2. POST /api/config
3. GET /api/config
4. /api/watch
5. rollback
6. SDK Get + Watch
7. demo-service 热更新

P1 有时间做：
1. DELETE /api/config
2. diff
3. YAML
4. 更漂亮的错误码
5. 单元测试
```

判断标准很简单：
**只要会影响最终“配置热更新 + 回滚”演示，就是 P0。不会影响演示的，都是 P1。**