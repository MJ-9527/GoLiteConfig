# API 草案

本文档定义 GoLiteConfig 的 HTTP API 草案。当前阶段用于统一后续开发、测试和文档口径，具体实现时应尽量保持接口结构一致。

## 1. 通用约定

### 1.1 Base URL

```text
http://localhost:8080
```

### 1.2 统一响应结构

成功响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

失败响应：

```json
{
  "code": 40001,
  "message": "app is required",
  "data": null
}
```

### 1.3 常用错误码

| code | 含义 |
| --- | --- |
| 0 | 成功 |
| 40001 | 请求参数错误 |
| 40401 | 配置不存在 |
| 50001 | Etcd 读写失败 |
| 50002 | 服务内部错误 |

### 1.4 配置项格式

配置项使用扁平化 key：

```json
{
  "database.host": "127.0.0.1",
  "database.port": "3306",
  "redis.addr": "127.0.0.1:6379"
}
```

服务端写入 Etcd 时，将 `database.host` 拆成：

```text
group = database
key = host
```

对应 Etcd Key：

```text
/config/{app}/{env}/versions/{version}/database/host
```

## 2. GET /health

健康检查接口。

### 请求

```http
GET /health
```

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ok",
    "service": "golite-config",
    "version": "0.0.1"
  }
}
```

### curl

```bash
curl http://localhost:8080/health
```

## 3. POST /api/config

发布配置接口。每次发布都会生成一个新版本，并更新 current 指针。

### 请求

```http
POST /api/config
Content-Type: application/json
```

请求体：

```json
{
  "app": "order-service",
  "env": "prod",
  "configs": {
    "database.host": "127.0.0.1",
    "database.port": "3306",
    "redis.addr": "127.0.0.1:6379"
  },
  "publisher": "A",
  "comment": "init config"
}
```

字段说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| app | string | 是 | 应用名 |
| env | string | 是 | 环境，例如 dev / test / prod |
| configs | object | 是 | 配置项 |
| publisher | string | 否 | 发布人 |
| comment | string | 否 | 发布说明 |

### 成功响应

```json
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

### 错误响应

```json
{
  "code": 40001,
  "message": "configs is required",
  "data": null
}
```

### curl

```bash
curl -X POST http://localhost:8080/api/config \
  -H "Content-Type: application/json" \
  -d '{
    "app": "order-service",
    "env": "prod",
    "configs": {
      "database.host": "127.0.0.1",
      "database.port": "3306",
      "redis.addr": "127.0.0.1:6379"
    },
    "publisher": "A",
    "comment": "init config"
  }'
```

## 4. GET /api/config

查询当前配置接口。默认查询 current 指向的版本。

### 请求

```http
GET /api/config?app=order-service&env=prod
```

可选参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| app | string | 是 | 应用名 |
| env | string | 是 | 环境 |
| group | string | 否 | 配置分组 |

### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "app": "order-service",
    "env": "prod",
    "version": "v1",
    "revision": 12,
    "configs": {
      "database.host": "127.0.0.1",
      "database.port": "3306",
      "redis.addr": "127.0.0.1:6379"
    }
  }
}
```

### 错误响应

```json
{
  "code": 40401,
  "message": "config not found",
  "data": null
}
```

### curl

```bash
curl "http://localhost:8080/api/config?app=order-service&env=prod"
```

## 5. GET /api/config/versions

查询版本列表接口。

### 请求

```http
GET /api/config/versions?app=order-service&env=prod
```

参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| app | string | 是 | 应用名 |
| env | string | 是 | 环境 |

### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "app": "order-service",
    "env": "prod",
    "current": "v2",
    "versions": [
      {
        "version": "v1",
        "revision": 12,
        "publisher": "A",
        "comment": "init config",
        "created_at": 1782787200
      },
      {
        "version": "v2",
        "revision": 18,
        "publisher": "A",
        "comment": "update database host",
        "created_at": 1782787300
      }
    ]
  }
}
```

### curl

```bash
curl "http://localhost:8080/api/config/versions?app=order-service&env=prod"
```

## 6. POST /api/config/rollback

回滚接口。回滚时不会直接把 current 指回旧版本，而是复制目标版本内容并生成一个新版本。

### 请求

```http
POST /api/config/rollback
Content-Type: application/json
```

请求体：

```json
{
  "app": "order-service",
  "env": "prod",
  "target_version": "v1",
  "publisher": "A",
  "comment": "rollback to v1"
}
```

字段说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| app | string | 是 | 应用名 |
| env | string | 是 | 环境 |
| target_version | string | 是 | 要回滚到的目标版本 |
| publisher | string | 否 | 操作人 |
| comment | string | 否 | 回滚说明 |

### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "app": "order-service",
    "env": "prod",
    "from_version": "v2",
    "target_version": "v1",
    "new_version": "v3",
    "revision": 25
  }
}
```

### 错误响应

```json
{
  "code": 40401,
  "message": "target version not found",
  "data": null
}
```

### curl

```bash
curl -X POST http://localhost:8080/api/config/rollback \
  -H "Content-Type: application/json" \
  -d '{
    "app": "order-service",
    "env": "prod",
    "target_version": "v1",
    "publisher": "A",
    "comment": "rollback to v1"
  }'
```

## 7. GET /api/watch

监听配置变更接口。客户端传入自己已知的 `last_revision`，服务端判断是否有新配置。

### 请求

```http
GET /api/watch?app=order-service&env=prod&last_revision=12
```

参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| app | string | 是 | 应用名 |
| env | string | 是 | 环境 |
| last_revision | int64 | 是 | 客户端当前已知 revision |

### 有变更响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "changed": true,
    "app": "order-service",
    "env": "prod",
    "version": "v2",
    "revision": 18,
    "configs": {
      "database.host": "192.168.1.10"
    }
  }
}
```

### 无变更响应

```http
HTTP/1.1 304 Not Modified
```

### curl

```bash
curl "http://localhost:8080/api/watch?app=order-service&env=prod&last_revision=12"
```

## 8. P1 可选接口

以下接口作为扩展项，时间不足时可以不实现。

| Method | Path | 说明 |
| --- | --- | --- |
| DELETE | `/api/config` | 删除配置项，建议也生成新版本 |
| GET | `/api/config/diff` | 对比两个版本差异 |

## 9. Day1 结论

Day1 API 草案确认以下核心接口：

| Method | Path | 优先级 | 说明 |
| --- | --- | --- | --- |
| GET | `/health` | P0 | 健康检查 |
| POST | `/api/config` | P0 | 发布配置 |
| GET | `/api/config` | P0 | 查询当前配置 |
| GET | `/api/config/versions` | P0 | 查询版本列表 |
| POST | `/api/config/rollback` | P0 | 回滚配置 |
| GET | `/api/watch` | P0 | 监听配置变更 |
| DELETE | `/api/config` | P1 | 删除配置 |
| GET | `/api/config/diff` | P1 | 版本对比 |
