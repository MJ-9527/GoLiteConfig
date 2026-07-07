# Day1 API 草案

## 1. 设计目标

本草案用于定义进阶轻量级 Apollo 在引入 namespace 后的核心接口模型。Day1 先确定参数与响应结构，不急于一次性把所有接口代码都改完。

## 2. 通用约定

### 2.1 Base URL

```text
http://localhost:8080
```

### 2.2 统一响应结构

成功：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

失败：

```json
{
  "code": 40001,
  "message": "namespace is required",
  "data": null
}
```

## 3. 发布接口

### 3.1 路径

```text
POST /api/config
```

### 3.2 请求体

```json
{
  "app": "order-service",
  "env": "prod",
  "namespace": "application",
  "configs": {
    "server.port": "8080",
    "feature.enable_login": "true"
  },
  "publisher": "admin",
  "comment": "enable login"
}
```

### 3.3 响应体

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "app": "order-service",
    "env": "prod",
    "namespace": "application",
    "version": "v2",
    "revision": 18
  }
}
```

## 4. 查询当前配置接口

### 4.1 路径

```text
GET /api/config?app=order-service&env=prod&namespace=application
```

### 4.2 响应体

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "app": "order-service",
    "env": "prod",
    "namespace": "application",
    "version": "v2",
    "revision": 18,
    "configs": {
      "server.port": "8080",
      "feature.enable_login": "true"
    }
  }
}
```

## 5. 查询历史版本接口

### 5.1 路径

```text
GET /api/config/versions?app=order-service&env=prod&namespace=application
```

### 5.2 响应体

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "app": "order-service",
    "env": "prod",
    "namespace": "application",
    "current": "v2",
    "versions": [
      {
        "version": "v1",
        "namespace": "application",
        "revision": 12,
        "publisher": "admin",
        "comment": "init config",
        "created_at": 1782787200
      },
      {
        "version": "v2",
        "namespace": "application",
        "revision": 18,
        "publisher": "admin",
        "comment": "enable login",
        "created_at": 1782787300
      }
    ]
  }
}
```

## 6. 回滚接口

### 6.1 路径

```text
POST /api/config/rollback
```

### 6.2 请求体

```json
{
  "app": "order-service",
  "env": "prod",
  "namespace": "application",
  "target_version": "v1",
  "publisher": "admin",
  "comment": "rollback to v1"
}
```

### 6.3 响应体

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "app": "order-service",
    "env": "prod",
    "namespace": "application",
    "from_version": "v2",
    "target_version": "v1",
    "new_version": "v3",
    "revision": 25
  }
}
```

## 7. Watch 接口

### 7.1 路径

```text
GET /api/watch?app=order-service&env=prod&namespace=application&last_revision=18
```

### 7.2 说明

1. `namespace` 成为 watch 的必填参数。
2. 客户端仅监听指定 namespace 的变更。
3. 无变化时可返回 `304 Not Modified`。

## 8. Day1 结论

Day1 先统一以下接口规则：

1. 所有核心接口都增加 `namespace` 维度。
2. `namespace` 在本阶段视为必填字段。
3. 旧逻辑若需兼容，可在服务端默认填充 `application`。
