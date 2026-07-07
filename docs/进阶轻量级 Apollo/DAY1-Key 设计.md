# Day1 Key 设计

## 1. 设计目标

为了支撑轻量级 Apollo 的 namespace 模型，Etcd Key 结构从当前：

```text
/config/{app}/{env}/...
```

升级为：

```text
/config/{app}/{env}/{namespace}/...
```

## 2. 新版 Key 结构

```text
/config/{app}/{env}/{namespace}/current
/config/{app}/{env}/{namespace}/version_counter
/config/{app}/{env}/{namespace}/versions/{version}/{group}/{key}
/config/{app}/{env}/{namespace}/meta/{version}
```

## 3. 字段含义

| 字段 | 含义 | 示例 |
| --- | --- | --- |
| app | 应用名 | `order-service` |
| env | 环境 | `dev` / `test` / `prod` |
| namespace | 命名空间 | `application` / `database` / `redis` |
| version | 版本号 | `v1` / `v2` / `v3` |
| group | 配置分组 | `db` / `pool` |
| key | 配置项名称 | `host` / `port` |

## 4. Key 作用说明

| Key | 作用 |
| --- | --- |
| `/current` | 保存当前生效版本 |
| `/version_counter` | 保存当前最大版本序号 |
| `/versions/{version}/...` | 保存某个版本下的配置内容 |
| `/meta/{version}` | 保存某个版本的元数据 |

## 5. 示例

```text
/config/order-service/prod/application/current = v2
/config/order-service/prod/application/version_counter = 2
/config/order-service/prod/application/versions/v2/server/port = 8080
/config/order-service/prod/application/versions/v2/feature/enable_login = true
/config/order-service/prod/application/meta/v2 = {
  "version": "v2",
  "namespace": "application",
  "revision": 18,
  "publisher": "admin",
  "comment": "enable login feature",
  "created_at": 1782787300
}
```

## 6. current 设计规则

1. `current` 只保存当前生效版本号。
2. 每个 namespace 独立维护自己的 `current`。
3. 客户端查询某个 namespace 时，只需要先读取当前 namespace 的 `current`。

## 7. version_counter 设计规则

1. 每个 namespace 独立维护自己的版本计数器。
2. 第一次发布时若不存在，则生成 `v1`。
3. 后续版本按序号递增。

## 8. 发布流程映射

发布某个 namespace 时：

1. 读取 `version_counter`
2. 生成新版本号
3. 写入 `versions/{version}/...`
4. 写入 `meta/{version}`
5. 更新 `current`
6. 更新 `version_counter`

## 9. 回滚流程映射

回滚某个 namespace 时：

1. 读取目标版本配置
2. 基于目标版本内容生成一个新版本
3. 写入新版本配置与 meta
4. 更新 `current`
5. 更新 `version_counter`

## 10. 兼容建议

为了兼容当前项目，可以先将旧逻辑视作默认 namespace：

```text
namespace = application
```

这样可以降低改造成本，并为旧数据迁移保留空间。
